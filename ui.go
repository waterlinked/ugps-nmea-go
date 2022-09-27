package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

// RunUI updates the GUI
func RunUI(cfg Config, inStatusCh chan inputStats, outputStatusChannel chan outputStats, cfgSource string) {
	// Let the goroutines initialize before starting GUI
	time.Sleep(50 * time.Millisecond)
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	y := 0
	height := 5
	width := 120
	halfWidth := width / 2

	p := widgets.NewParagraph()
	p.Title = applicationName()
	p.Text = fmt.Sprintf("PRESS q TO QUIT.\nConfig from: %s\n", cfgSource)

	p.SetRect(0, y, width, height)
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	y += height
	height = 10
	inSrcHeight := height
	if cfg.RetransmitEnabled() {
		inSrcHeight = height * 2

	}

	inpSrcStatus := widgets.NewParagraph()
	inpSrcStatus.Title = "GPS/GPS Compass in"
	if cfg.InputEnabled() {
		inpSrcStatus.Text = "Waiting for data"
	} else {
		inpSrcStatus.Text = "Input not enabled"
	}

	inpSrcStatus.SetRect(0, y, halfWidth, y+inSrcHeight)
	inpSrcStatus.TextStyle.Fg = ui.ColorGreen
	inpSrcStatus.BorderStyle.Fg = ui.ColorCyan

	inpArrow := widgets.NewParagraph()
	inpArrow.Border = false
	inpArrow.Text = "=>"
	inpArrow.SetRect(halfWidth, y, halfWidth+5, y+height)

	inpDestStatus := widgets.NewParagraph()
	inpDestStatus.Title = "GPS/GPS Compass out to UGPS"

	inpDestStatus.SetRect(halfWidth+5, y, width, y+height)
	inpDestStatus.TextStyle.Fg = ui.ColorGreen
	inpDestStatus.BorderStyle.Fg = ui.ColorCyan

	inpRetransmitStatus := widgets.NewParagraph()
	if cfg.RetransmitEnabled() {
		inpRetransmitStatus.Title = "Retransmit Input"

		inpRetransmitStatus.SetRect(halfWidth+5, y+height, width, y+inSrcHeight)
		inpRetransmitStatus.TextStyle.Fg = ui.ColorGreen
		inpRetransmitStatus.BorderStyle.Fg = ui.ColorCyan
	}

	//y += height
	y += inSrcHeight
	height = 10

	outSrcStatus := widgets.NewParagraph()
	outSrcStatus.Title = "Locator Position in from UGPS"
	outSrcStatus.Text = "Waiting for data"
	if !cfg.OutputEnabled() {
		outSrcStatus.Text = "Output not enabled"
	}
	outSrcStatus.SetRect(0, y, halfWidth, y+height)
	outSrcStatus.TextStyle.Fg = ui.ColorGreen
	outSrcStatus.BorderStyle.Fg = ui.ColorCyan

	outArrow := widgets.NewParagraph()
	outArrow.Border = false
	outArrow.Text = "=>"
	outArrow.SetRect(halfWidth, y, halfWidth+5, y+height)

	outDestStatus := widgets.NewParagraph()
	outDestStatus.Title = "Locator Position out to NMEA"
	outDestStatus.Text = "Waiting for data"
	if !cfg.OutputEnabled() {
		outDestStatus.Text = "Output not enabled"
	}
	outDestStatus.SetRect(halfWidth+5, y, width, y+height)
	outDestStatus.TextStyle.Fg = ui.ColorGreen
	outDestStatus.BorderStyle.Fg = ui.ColorCyan

	y += height
	height = 15

	dbgText := widgets.NewList()
	dbgText.Title = "Debug"
	dbgText.Rows = dbgMsg
	dbgText.WrapText = true
	dbgText.SetRect(0, y, width, y+height)
	dbgText.BorderStyle.Fg = ui.ColorCyan

	hideDebug := widgets.NewParagraph()
	hideDebug.Text = ""
	hideDebug.SetRect(0, y, width, y+height)
	hideDebug.Border = false

	draw := func() {
		ui.Render(p, inpSrcStatus, inpArrow, inpDestStatus, outSrcStatus, outArrow, outDestStatus, inpRetransmitStatus)
		if debug {
			dbgText.Rows = dbgMsg
			ui.Render(dbgText)
		} else {
			ui.Render(hideDebug)
		}
	}

	// Initial draw before any events have occurred
	draw()

	uiEvents := ui.PollEvents()

	for {
		select {
		case inStats := <-inStatusCh:
			inpSrcStatus.TextStyle.Fg = ui.ColorGreen
			inpSrcStatus.Text = fmt.Sprintf("Source: %s\n\n", cfg.Input.Device) +
				"Supported NMEA sentences received:\n" +
				fmt.Sprintf(" * Topside Position   : %s\n", inStats.src.posDesc) +
				fmt.Sprintf(" * Topside Heading    : %s\n", inStats.src.headDesc) +
				fmt.Sprintf(" * Parse error: %d\n\n", inStats.src.unparsableCount) +
				inStats.src.errorMsg
			if inStats.src.errorMsg != "" {
				inpSrcStatus.TextStyle.Fg = ui.ColorRed
			}
			inpDestStatus.TextStyle.Fg = ui.ColorGreen
			inpDestStatus.Text = fmt.Sprintf("Destination: %s\n\n", cfg.BaseURL) +
				fmt.Sprintf("Sent successfully to\n Underwater GPS: %d\n\n", inStats.dst.sendOk) +
				inStats.dst.errorMsg
			if inStats.dst.errorMsg != "" {
				inpDestStatus.TextStyle.Fg = ui.ColorRed
			}

			inpRetransmitStatus.Text = fmt.Sprintf("Destination: %s\n\n", cfg.Input.Retransmit) +
				fmt.Sprintf("Count: %d\n%s", inStats.retransmit.count, inStats.retransmit.errorMsg)
			inpRetransmitStatus.TextStyle.Fg = ui.ColorGreen
			if inStats.retransmit.errorMsg != "" {
				inpRetransmitStatus.TextStyle.Fg = ui.ColorRed
			}
			draw()
		case outStats := <-outputStatusChannel:
			outSrcStatus.Text = fmt.Sprintf("Source: %s\n\n", cfg.BaseURL) +
				fmt.Sprintf("Positions from Underwater GPS:\n  %d\n", outStats.src.getCount)
			outSrcStatus.TextStyle.Fg = ui.ColorGreen

			if outStats.src.errMsg != "" {
				outSrcStatus.TextStyle.Fg = ui.ColorRed
				outSrcStatus.Text += fmt.Sprintf("\n\n%v (%d)", outStats.src.errMsg, outStats.src.getErr)
			}

			outDestStatus.Text = fmt.Sprintf("Destination: %s\n\n", cfg.Output.Device) +
				"Sent:\n" +
				fmt.Sprintf(" * Locator/ROV Position : %s: %d\n", strings.ToUpper(cfg.Output.PositionSentence), outStats.dst.sendOk)
			outDestStatus.TextStyle.Fg = ui.ColorGreen

			if outStats.dst.errMsg != "" {
				outDestStatus.TextStyle.Fg = ui.ColorRed
				outDestStatus.Text += fmt.Sprintf("\n\n%s", outStats.dst.errMsg)
			}
			draw()
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "d":
				dbgMsg = nil
				dbgText.Rows = dbgMsg
				debug = !debug

				draw()
			}
		}
	}
}

// RunUIError shows a dialog box with an error message
func RunUIError(message string) {
	// Let the goroutines initialize before starting GUI
	time.Sleep(50 * time.Millisecond)
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	y := 0
	height := 7
	width := 80

	p := widgets.NewParagraph()
	p.Title = applicationName()
	p.Text = "PRESS q TO QUIT."
	p.SetRect(0, y, width, height)
	y += height
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	errorPara := widgets.NewParagraph()
	errorPara.Title = "Error occurred"
	errorPara.Text = message
	height = 10
	errorPara.SetRect(0, y, width, y+height)
	errorPara.TextStyle.Fg = ui.ColorRed
	errorPara.BorderStyle.Fg = ui.ColorCyan

	draw := func() {
		ui.Render(p, errorPara)
	}

	// Intial draw before any events have occured
	draw()

	uiEvents := ui.PollEvents()

	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			}
		}
	}
}
