package main

import (
	"fmt"
	"io"
	"math"
	"time"
)

type outputStats struct {
	getOk  int
	getErr int
	sendOk int
	isErr  bool
	errMsg string
}

type Outputter struct {
	writer              io.Writer
	stats               outputStats
	outputStatusChannel chan outputStats
	serialiser          nmeaSerialiser
}

func NewOutputter(writer io.Writer, serialiser nmeaSerialiser) *Outputter {
	return &Outputter{writer: writer, stats: outputStats{}, outputStatusChannel: make(chan outputStats, 1), serialiser: serialiser}
}

func (outputter *Outputter) handleError(err error, message string) {
	outputter.stats.isErr = true
	outputter.stats.errMsg = fmt.Sprintf("%s: %v", message, err)
	debugPrintf(outputter.stats.errMsg)
	outputter.stats.getErr++
	outputter.outputStatusChannel <- outputter.stats

	output := outputter.serialiser.noPosition()
	fmt.Fprintf(outputter.writer, "%s\r\n", output)
}

func (outputter *Outputter) OutputLoop() {
	var previousLatitude float64
	var previousLongitude float64

	for {
		time.Sleep(100 * time.Millisecond)
		globalPosition, err := getGlobalPosition()
		if err != nil {
			outputter.handleError(err, "Error fetching global position from UGPS")
			continue
		}
		acousticPosition, err := getAcousticPosition()
		if err != nil {
			outputter.handleError(err, "Error fetching acoustic position from UGPS")
			continue
		}
		outputter.stats.getOk++

		// Check if position has changed
		if math.Abs((globalPosition.Latitude-previousLatitude)) < 1e-12 &&
			math.Abs((globalPosition.Longitude-previousLongitude)) < 1e-12 {
			// Not changed
			continue
		}

		previousLatitude = globalPosition.Latitude
		previousLongitude = globalPosition.Longitude

		output := outputter.serialiser.serialise(globalPosition, acousticPosition)

		_, err = fmt.Fprintf(outputter.writer, "%s\r\n", output)
		if err != nil {
			outputter.handleError(err, "Error in writing NMEA string")
		} else {
			outputter.stats.isErr = false
			outputter.stats.sendOk++
			outputter.outputStatusChannel <- outputter.stats
		}

	}
}
