package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/tarm/serial"
)

var (
	listen          string
	headingSentence string
	output          string
	sentence        string
	debug           bool

	Version  string = "0.0.0"
	BuildNum string = "local"
	SHA      string = "local"
)

const dbgLen = 5

var dbgMsg []string = make([]string, 0)

func debugPrintf(arguments string, a ...interface{}) {
	if debug {
		if len(dbgMsg) > dbgLen {
			dbgMsg = dbgMsg[1:dbgLen]
		}
		s := time.Now().Format("15:04:05") + " " + fmt.Sprintf(arguments, a...)
		dbgMsg = append(dbgMsg, strings.TrimSpace(s))
		//log.Printf(arguments, a...)
	}
}

// deviceIsUDP uses ":" to decide if this is UDP address or serial device
func deviceIsUDP(device string) bool {
	return len(strings.Split(device, ":")) > 1
}

func baudAndPortFromDevice(device string) (string, int) {
	baudrate := 115200
	port := device
	// Is the baudrate specified?
	parts := strings.Split(device, "@")
	if len(parts) > 1 {
		b, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Printf("Unable to parse baudrate: %s as numeric value\n", parts[1])
			os.Exit(1)
		}
		baudrate = b
		port = parts[0]
	}
	return port, baudrate
}

func main() {
	availableSerialisers := make(map[string]nmeaSerialiser)
	availableSerialisers["RATLL"] = tllSerialiser{}
	availableSerialisers["GPGGA"] = ggaSerialiser{}
	supportedSentences := nmeaSerialisers(availableSerialisers)

	availableHeadingSentences := make(map[string]headingParser)
	availableHeadingSentences["HDM"] = &hdmParser{}
	availableHeadingSentences["HDT"] = &hdtParser{}
	availableHeadingSentences["THS"] = &thsParser{}
	supportedHeadings := keysForMapP(availableHeadingSentences)

	fmt.Printf("Water Linked NMEA UGPS bridge (v%s %s.%s)\n", Version, BuildNum, SHA)
	flag.StringVar(&listen, "i", "0.0.0.0:7777", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1@4800) to listen for NMEA input. ")
	flag.StringVar(&output, "o", "", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1) to send NMEA output. ")
	flag.StringVar(&sentence, "sentence", "GPGGA", "NMEA output sentence to use. Supported: "+supportedSentences)
	flag.StringVar(&headingSentence, "heading", "HDT", "Input sentence type to use for heading. Supported: "+supportedHeadings)
	flag.StringVar(&baseURL, "url", "http://192.168.2.94", "URL of Underwater GPS")
	flag.BoolVar(&debug, "d", false, "debug")
	flag.Parse()

	// Same serial port for input and output?
	sameInOut := (listen == output) && !deviceIsUDP(listen)
	if sameInOut {
		fmt.Println("Same port for input and output", listen)
	}

	serialiser, exists := availableSerialisers[strings.ToUpper(sentence)]
	if !exists {
		fmt.Printf("Unsupported sentence '%s'. Supported are: %s\n", sentence, supportedSentences)
		os.Exit(1)
	}

	hParser, exists := availableHeadingSentences[strings.ToUpper(headingSentence)]
	if !exists {
		fmt.Printf("Unsupported heading sentence '%s'. Supported are: %s\n", headingSentence, supportedHeadings)
		os.Exit(1)
	}

	// Channels
	inStatusCh := make(chan inputStats, 1)
	masterCh := make(chan externalMaster, 1)

	// Output
	var writer io.Writer = nil

	// Setup input
	if deviceIsUDP(listen) {
		// Input from UDP
		go inputUDPLoop(listen, hParser, masterCh, inStatusCh)
	} else {
		// Input from serial port
		port, baudrate := baudAndPortFromDevice(listen)

		c := &serial.Config{Name: port, Baud: baudrate}
		s, err := serial.OpenPort(c)
		if err != nil {
			fmt.Printf("Error opening serial port: %v\n", err)
			os.Exit(1)
		}
		defer s.Close()

		go inputSerialLoop(s, hParser, masterCh, inStatusCh)
		if sameInOut {
			// Output is to same serial port as input
			writer = s
		}
	}
	go inputLoop(masterCh, inStatusCh)

	// Setup output
	if output == "" {
		// Output disabled
	} else if deviceIsUDP(output) {
		// Output to UDP
		conn, err := net.Dial("udp", output)
		if err != nil {
			fmt.Printf("Error connecting to UDP: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()
		writer = conn

	} else if !sameInOut {
		// Output to different serial port
		port, baudrate := baudAndPortFromDevice(output)

		c := &serial.Config{Name: port, Baud: baudrate}
		s, err := serial.OpenPort(c)
		if err != nil {
			fmt.Printf("Error opening serial port: %v\n", err)
			os.Exit(1)
		}
		defer s.Close()
		writer = s
	}

	outputter := NewOutputter(writer, serialiser)
	if writer != nil {
		go NewOutputter(writer, serialiser).OutputLoop()
	}

	RunUI(inStatusCh, outputter.outputStatusChannel)
}

// Bool2Int returns 1 if true, else 0
func Bool2Int(val bool) int {
	if val {
		return 1
	}
	return 0
}

// RunUI updates the GUI
func RunUI(inStatusCh chan inputStats, outputStatusChannel chan outputStats) {
	// Let the goroutines initialize before starting GUI
	time.Sleep(50 * time.Millisecond)
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	y := 0
	height := 5
	width := 80

	p := widgets.NewParagraph()
	p.Title = "Water Linked Underwater GPS NMEA bridge"
	p.Text = fmt.Sprintf("PRESS q TO QUIT\nIn : %s\nOut: %s", listen, output)
	p.SetRect(0, y, width, height)
	y += height
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	inpStatus := widgets.NewParagraph()
	inpStatus.Title = "Input status"
	inpStatus.Text = "Waiting for data"
	height = 12
	inpStatus.SetRect(0, y, width, y+height)
	y += height
	inpStatus.TextStyle.Fg = ui.ColorGreen
	inpStatus.BorderStyle.Fg = ui.ColorCyan

	outStatus := widgets.NewParagraph()
	outStatus.Title = "Output status"
	outStatus.Text = "Waiting for data"
	if output == "" {
		outStatus.Text = "Output not enabled"
	}
	height = 10
	outStatus.SetRect(0, y, width, y+height)
	y += height
	outStatus.TextStyle.Fg = ui.ColorGreen
	outStatus.BorderStyle.Fg = ui.ColorCyan

	height = 15
	dbgText := widgets.NewList()
	dbgText.Title = "Debug"
	dbgText.Rows = dbgMsg
	dbgText.WrapText = true
	dbgText.SetRect(0, y, width, y+height)
	y += height
	//dbgText.TextStyle.Fg = ui.ColorGreen
	dbgText.BorderStyle.Fg = ui.ColorCyan

	draw := func() {
		ui.Render(p, inpStatus, outStatus)
		if debug {
			ui.Render(dbgText)
		}
	}

	// Intial draw before any events have occured
	draw()

	uiEvents := ui.PollEvents()

	for {
		select {
		case instats := <-inStatusCh:
			inpStatus.TextStyle.Fg = ui.ColorGreen
			inpStatus.Text = fmt.Sprintf(
				"Supported NMEA sentences received:\n"+
					" * Position   : %s\n"+
					" * Heading    : %s\n"+
					" * Parse error: %d\n"+
					"Sent sucessfully to UGPS: %d\n\n"+
					"%s",
				instats.posDesc, instats.headDesc, instats.unparsableCount, instats.sendOk, instats.errorMsg)
			if instats.isErr {
				inpStatus.TextStyle.Fg = ui.ColorRed
			}
			if debug {
				dbgText.Rows = dbgMsg
			}
			draw()
		case outstats := <-outputStatusChannel:
			outStatus.Text = fmt.Sprintf("%d positions sent to NMEA out", outstats.sendOk)
			outStatus.TextStyle.Fg = ui.ColorGreen

			if outstats.isErr {
				outStatus.TextStyle.Fg = ui.ColorRed
				outStatus.Text += fmt.Sprintf("\n\n%v (%d)", outstats.errMsg, outstats.getErr)
			}
			if debug {
				dbgText.Rows = dbgMsg
			}
			draw()
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}

func nmeaSerialisers(m map[string]nmeaSerialiser) string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	joined := strings.Join(keys, ", ")
	return joined
}

func keysForMapP(m map[string]headingParser) string {
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	joined := strings.Join(keys, ", ")
	return joined
}
