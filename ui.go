package main

import (
	"fmt"
	"log"
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
	height := 7
	width := 80

	p := widgets.NewParagraph()
	p.Title = applicationName()
	p.Text = fmt.Sprintf("PRESS q TO QUIT.\nConfig source: %s\nUnderwater GPS: %s\nIn : %s\nOut: %s\n", cfgSource, cfg.BaseURL, cfg.Input.Device, cfg.Output.Device)
	p.SetRect(0, y, width, height)
	y += height
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	inpStatus := widgets.NewParagraph()
	inpStatus.Title = "Input status"
	if cfg.InputEnabled() {
		inpStatus.Text = "Waiting for data"
	} else {
		inpStatus.Text = "Input not enabled"
	}
	height = 12
	inpStatus.SetRect(0, y, width, y+height)
	y += height
	inpStatus.TextStyle.Fg = ui.ColorGreen
	inpStatus.BorderStyle.Fg = ui.ColorCyan

	outStatus := widgets.NewParagraph()
	outStatus.Title = "Output status"
	outStatus.Text = "Waiting for data"
	if !cfg.OutputEnabled() {
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
	//y += height
	//dbgText.TextStyle.Fg = ui.ColorGreen
	dbgText.BorderStyle.Fg = ui.ColorCyan

	hideDebug := widgets.NewParagraph()
	hideDebug.Text = ""
	hideDebug.SetRect(0, y, width, y+height)
	hideDebug.Border = false

	draw := func() {
		ui.Render(p, inpStatus, outStatus)
		if debug {
			ui.Render(dbgText)
		} else {
			ui.Render(hideDebug)
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
					"Sent successfully to Underwater GPS: %d\n\n"+
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
	p.Title = "Water Linked Underwater GPS NMEA bridge"
	p.Text = "PRESS q TO QUIT."
	p.SetRect(0, y, width, height)
	y += height
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	errorPara := widgets.NewParagraph()
	errorPara.Title = "Error occured"
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
