package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
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

func main() {
	fmt.Printf("Water Linked NMEA UGPS bridge (v%s %s.%s)\n", Version, BuildNum, SHA)
	flag.StringVar(&listen, "i", "0.0.0.0:7777", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1@4800) to listen for NMEA input. ")
	flag.StringVar(&output, "o", "127.0.0.1:2947", "UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1) to send NMEA output. ")
	flag.StringVar(&baseURL, "url", "http://192.168.2.94", "URL of Underwater GPS")
	//flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()

	inStatusCh := make(chan inputStats, 1)
	go inputLoop(listen, inStatusCh)
	outStatusCh := make(chan outStats, 1)
	if output != "" {
		go outputLoop(output, outStatusCh)
	}

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
	height = 10
	outStatus.SetRect(0, y, 80, y+height)
	y += height
	outStatus.TextStyle.Fg = ui.ColorGreen
	outStatus.BorderStyle.Fg = ui.ColorCyan

	draw := func() {
		ui.Render(p, inpStatus, outStatus)
	}

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
