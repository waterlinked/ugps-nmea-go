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
	listen  string
	output  string
	verbose bool

	Version  string = "0.0.0"
	BuildNum string = "local"
	SHA      string = "local"
)

func debugPrintf(arguments string, a ...interface{}) {
	if verbose {
		log.Printf(arguments, a...)
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
	fmt.Printf("Water Linked NMEA UGPS bridge (v%s %s.%s)\n", Version, BuildNum, SHA)
	flag.StringVar(&listen, "i", "0.0.0.0:7777", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1@4800) to listen for NMEA input. ")
	flag.StringVar(&output, "o", "", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1) to send NMEA output. ")
	flag.StringVar(&baseURL, "url", "http://192.168.2.94", "URL of Underwater GPS")
	//flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	// Same serial port for input and output?
	sameInOut := (listen == output) && !deviceIsUDP(listen)
	if sameInOut {
		fmt.Println("Same port for input and output", listen)
	}

	// Channels
	inStatusCh := make(chan inputStats, 1)
	masterCh := make(chan externalMaster, 1)
	outStatusCh := make(chan outStats, 1)

	// Output
	var writer io.Writer = nil

	// Setup input
	if deviceIsUDP(listen) {
		// Input from UDP
		go inputUDPLoop(listen, masterCh, inStatusCh)
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

		go inputSerialLoop(s, masterCh, inStatusCh)
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

	if writer != nil {
		go outputLoop(writer, outStatusCh)
	}

	RunUI(inStatusCh, outStatusCh)
}

func RunUI(inStatusCh chan inputStats, outStatusCh chan outStats) {
	// Let the goroutines initialize before starting GUI
	time.Sleep(50 * time.Millisecond)
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	y := 0
	height := 5

	p := widgets.NewParagraph()
	p.Title = "Water Linked Underwater GPS NMEA bridge"
	p.Text = fmt.Sprintf("PRESS q TO QUIT\nIn : %s\nOut: %s", listen, output)
	p.SetRect(0, y, 80, height)
	y += height
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	inpStatus := widgets.NewParagraph()
	inpStatus.Title = "Input status"
	inpStatus.Text = "Waiting for data"
	height = 12
	inpStatus.SetRect(0, y, 80, y+height)
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
	outStatus.SetRect(0, y, 80, y+height)
	y += height
	outStatus.TextStyle.Fg = ui.ColorGreen
	outStatus.BorderStyle.Fg = ui.ColorCyan

	draw := func() {
		ui.Render(p, inpStatus, outStatus)
	}

	// Intial draw before any events have occured
	draw()

	uiEvents := ui.PollEvents()

	for {
		select {
		case instats := <-inStatusCh:
			inpStatus.TextStyle.Fg = ui.ColorGreen
			inpStatus.Text = fmt.Sprintf("Supported NMEA sentences received:\n * GGA: %d\n * HDT: %d\n * THS: %d\nSent sucessfully to UGPS: %d",
				instats.typeGga, instats.typeHdt, instats.typeThs, instats.sendOk)
			if instats.typeHdt > 0 && instats.typeThs > 0 {
				inpStatus.Text += "\nWarning: BOTH HDT and THS received, this can give jumpy orientation"
			}
			if instats.isErr {
				inpStatus.TextStyle.Fg = ui.ColorRed
				inpStatus.Text += fmt.Sprintf("\n\n%s", instats.errorMsg)
			}
			draw()
		case outstats := <-outStatusCh:
			outStatus.Text = fmt.Sprintf("%d positions sent to NMEA out", outstats.sendOk)
			outStatus.TextStyle.Fg = ui.ColorGreen

			if outstats.isErr {
				outStatus.TextStyle.Fg = ui.ColorRed
				outStatus.Text += fmt.Sprintf("\n\n%v (%d)", outstats.errMsg, outstats.getErr)
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
