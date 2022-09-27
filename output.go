package main

import (
	"fmt"
	"io"
	"math"
	"time"
)

type outputStats struct {
	src struct {
		getOk    int
		getCount int
		getErr   int
		errMsg   string
	}
	dst struct {
		sendOk   int
		errCount int
		errMsg   string
	}
}

type Outputter struct {
	writer              io.Writer
	stats               outputStats
	outputStatusChannel chan outputStats
	serialiser          nmeaPositionSerialiser
}

func NewOutputter(writer io.Writer, serialiser nmeaPositionSerialiser) *Outputter {
	return &Outputter{writer: writer, stats: outputStats{}, outputStatusChannel: make(chan outputStats, 1), serialiser: serialiser}
}

func (outputter *Outputter) handleSrcError(err error, message string) {
	outputter.stats.src.errMsg = fmt.Sprintf("%s: %v", message, err)
	debugPrintf(outputter.stats.src.errMsg)
	outputter.stats.src.getErr++
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
			outputter.handleSrcError(err, "Error fetching global position from UGPS")
			continue
		}
		acousticPosition, err := getAcousticPosition()
		if err != nil {
			outputter.handleSrcError(err, "Error fetching acoustic position from UGPS")
			continue
		}
		outputter.stats.src.getOk++
		outputter.stats.src.errMsg = ""

		// Check if position has changed
		if math.Abs((globalPosition.Latitude-previousLatitude)) < 1e-12 &&
			math.Abs((globalPosition.Longitude-previousLongitude)) < 1e-12 {
			// Not changed
			continue
		}
		outputter.stats.src.getCount++

		previousLatitude = globalPosition.Latitude
		previousLongitude = globalPosition.Longitude

		output := outputter.serialiser.serialise(globalPosition, acousticPosition)

		_, err = fmt.Fprintf(outputter.writer, "%s\r\n", output)
		if err != nil {
			message := "Error in writing NMEA string"
			outputter.stats.dst.errMsg = fmt.Sprintf("%s: %v", message, err)
			outputter.stats.dst.errCount++
		} else {
			outputter.stats.dst.errMsg = ""
			outputter.stats.dst.sendOk++
		}
		outputter.outputStatusChannel <- outputter.stats
	}
}
