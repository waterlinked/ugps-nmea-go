package main

import (
	"fmt"
	"io"
	"math"
	"time"
)

type outStats struct {
	getOk  int
	getErr int
	sendOk int
	isErr  bool
	errMsg string
}

const updateSeconds = 10

func outputLoop(writer io.Writer, outStatusCh chan outStats, ser outSerializer) {
	var prevLat float64
	var prevLon float64

	var stats outStats

	var out string

	for {
		time.Sleep(100 * time.Millisecond)
		pos, err := getGlobalPosition()
		if err != nil {
			stats.isErr = true
			stats.errMsg = fmt.Sprintf("ERR get position from UGPS: %v", err)
			debugPrintf(stats.errMsg)
			stats.getErr++
			outStatusCh <- stats

			out = ser.noPosition()
			fmt.Fprintf(writer, "%s\r\n", out)

			continue
		}
		stats.getOk++

		// Check if posision is update
		if math.Abs((pos.Lat-prevLat)) < 1e-12 &&
			math.Abs((pos.Lon-prevLon)) < 1e-12 {
			// No new position update yet
			continue
		}
		prevLat = pos.Lat
		prevLon = pos.Lon

		out := ser.serialize(pos)

		_, err = fmt.Fprintf(writer, "%s\r\n", out)
		if err != nil {
			stats.isErr = true
			stats.errMsg = fmt.Sprintf("NMEA out: %v", err)
			debugPrintf(stats.errMsg)
		} else {
			stats.isErr = false
			stats.sendOk++
		}
		outStatusCh <- stats
	}
}
