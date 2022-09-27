package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"net"
	neturl "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"go.bug.st/serial"
)

var (
	debug bool

	Version  string = "0.0.0"
	BuildNum string = "local"
	SHA      string = "local"
)

const dbgLen = 10

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

func keys[V any](m map[string]V) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	joined := strings.Join(keys, ", ")
	return joined
}

func applicationName() string {
	return fmt.Sprintf("Water Linked NMEA UGPS bridge (v%s %s.%s)", Version, BuildNum, SHA)
}

func main() {
	var (
		listen          string
		headingSentence string
		output          string
		sentence        string
		url             string
		cfgFilename     string
	)

	availableSerialisers := make(map[string]nmeaPositionSerialiser)
	availableSerialisers["RATLL"] = tllSerialiser{}
	availableSerialisers["GPGGA"] = ggaSerialiser{}
	supportedSentences := keys(availableSerialisers)

	availableHeadingSentences := make(map[string]nmeaHeadingParser)
	availableHeadingSentences["HDM"] = &hdmParser{}
	availableHeadingSentences["HDT"] = &hdtParser{}
	availableHeadingSentences["THS"] = &thsParser{}
	supportedHeadings := keys(availableHeadingSentences)

	fmt.Println(applicationName())
	flag.StringVar(&listen, "i", "", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1@4800) to listen for NMEA input. ")
	flag.StringVar(&output, "o", "", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1) to send NMEA output. ")
	flag.StringVar(&sentence, "sentence", "GPGGA", "NMEA output sentence to use. Supported: "+supportedSentences)
	flag.StringVar(&headingSentence, "heading", "HDT", "Input sentence type to use for heading. Supported: "+supportedHeadings)
	flag.StringVar(&url, "url", "http://192.168.2.94", "URL of Underwater GPS")
	flag.StringVar(&cfgFilename, "c", "config.yml", "Configuration file to use")
	flag.BoolVar(&debug, "d", false, "debug")
	flag.Parse()

	cfgSource := fmt.Sprintf("config file '%s'", cfgFilename)
	cfg := Config{}
	err := readFile(&cfg, cfgFilename)
	if err != nil {
		var pathError *fs.PathError
		if errors.As(err, &pathError) {
			fmt.Printf("no config file can be loaded ('%s').\nusing command line arguments\n", err)
			cfgSource = "command line arguments"
			cfg.Input.Device = listen
			cfg.Input.HeadingSentence = headingSentence
			cfg.Output.Device = output
			cfg.Output.PositionSentence = sentence
			cfg.BaseURL = url
		} else {
			RunUIError(fmt.Sprintf("config file parse error:\n%s", err))
			os.Exit(1)
		}
	}

	baseURL = cfg.BaseURL
	u, err := neturl.Parse(baseURL)
	if err != nil {
		RunUIError(fmt.Sprintf("Url should be in form http://1.2.3.4. Got '%s': %s", baseURL, err))
		os.Exit(1)
	}
	if u.Scheme == "" {
		RunUIError(fmt.Sprintf("Url should be in form http://1.2.3.4. Got '%s'", baseURL))
		os.Exit(1)
	}

	// Same serial port for input and output?
	sameInOut := (cfg.Input.Device == cfg.Output.Device) && !deviceIsUDP(cfg.Input.Device)
	if sameInOut {
		fmt.Println("Same port for input and output", cfg.Input.Device)
	}

	serialiser, exists := availableSerialisers[strings.ToUpper(cfg.Output.PositionSentence)]
	if !exists {
		msg := fmt.Sprintf("Unsupported sentence '%s'. Supported are: %s\n", cfg.Output.PositionSentence, supportedSentences)
		RunUIError(msg)

		os.Exit(1)
	}

	hParser, exists := availableHeadingSentences[strings.ToUpper(cfg.Input.HeadingSentence)]
	if !exists {
		msg := fmt.Sprintf("Unsupported heading sentence '%s'. Supported are: %s\n", cfg.Input.HeadingSentence, supportedHeadings)
		RunUIError(msg)
		os.Exit(1)
	}

	// Channels
	inStatusCh := make(chan inputStats, 1)
	masterCh := make(chan externalMaster, 1)

	// Output
	var writer io.Writer = nil

	// Setup input
	if cfg.InputEnabled() {
		var retransmit net.Conn
		if cfg.RetransmitEnabled() {
			if !deviceIsUDP(cfg.Input.Retransmit) {
				msg := fmt.Sprintf("Retransmit only supports UDP. Got serial port as configuration: %v\n", cfg.Input.Retransmit)
				RunUIError(msg)
				os.Exit(1)
			}
			conn, err := net.Dial("udp", cfg.Input.Retransmit)
			if err != nil {
				msg := fmt.Sprintf("Error connecting to UDP: %s:%v\n", err, cfg.Input.Retransmit)
				RunUIError(msg)
				os.Exit(1)
			}
			defer conn.Close()
			retransmit = conn
		}
		if deviceIsUDP(cfg.Input.Device) {
			// Input from UDP
			go inputUDPLoop(cfg.Input.Device, hParser, masterCh, inStatusCh, retransmit)
		} else {
			// Input from serial port
			port, baudrate := baudAndPortFromDevice(cfg.Input.Device)

			c := &serial.Mode{BaudRate: baudrate}
			s, err := serial.Open(port, c)
			if err != nil {
				msg := fmt.Sprintf("Error opening serial port %s: %v\n", port, err)
				RunUIError(msg)
				os.Exit(1)
			}
			defer s.Close()

			go inputSerialLoop(s, hParser, masterCh, inStatusCh, retransmit)
			if sameInOut {
				// Output is to same serial port as input
				writer = s
			}
		}
		go inputLoop(masterCh, inStatusCh)
	}

	// Setup output
	if cfg.Output.Device == "" {
		// Output disabled
	} else if deviceIsUDP(cfg.Output.Device) {
		// Output to UDP
		conn, err := net.Dial("udp", cfg.Output.Device)
		if err != nil {
			msg := fmt.Sprintf("Error connecting to UDP: %s:%v\n", err, cfg.Output.Device)
			RunUIError(msg)
			os.Exit(1)
		}
		defer conn.Close()
		writer = conn

	} else if !sameInOut {
		// Output to different serial port
		port, baudrate := baudAndPortFromDevice(cfg.Output.Device)

		c := &serial.Mode{BaudRate: baudrate}
		s, err := serial.Open(port, c)
		if err != nil {
			msg := fmt.Sprintf("Error opening serial port %s: %v\n", port, err)
			RunUIError(msg)
			os.Exit(1)
		}
		defer s.Close()
		writer = s
	}

	outputter := NewOutputter(writer, serialiser)
	if writer != nil {
		go outputter.OutputLoop()
	}

	RunUI(cfg, inStatusCh, outputter.outputStatusChannel, cfgSource)
}
