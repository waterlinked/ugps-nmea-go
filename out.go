package main

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/pilebones/go-nmea"
)

type outStats struct {
	getOk  int
	getErr int
	sendOk int
	isErr  bool
	errMsg string
}

const updateSeconds = 10

func outputLoop(writer io.Writer, outStatusCh chan outStats) {
	var prevLat float64
	var prevLon float64

	var stats outStats

	for {
		time.Sleep(100 * time.Millisecond)
		pos, err := getGlobalPosition()
		if err != nil {
			stats.isErr = true
			stats.errMsg = fmt.Sprintf("ERR get position from UGPS: %v", err)
			stats.getErr++
			outStatusCh <- stats
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

		gga := nmea.GPGGA{
			TimeUTC:            time.Now().UTC(),
			Latitude:           nmea.LatLong(pos.Lat),
			Longitude:          nmea.LatLong(pos.Lon),
			QualityIndicator:   nmea.QualityIndicator(pos.FixQuality),
			NbOfSatellitesUsed: uint64(pos.NumSats),
			Altitude:           0,
		}

		out := gga.Serialize()
		_, err = fmt.Fprintf(writer, "%s\n", out)
		if err != nil {
			stats.isErr = true
			stats.errMsg = fmt.Sprintf("NMEA out: %v", err)
		} else {
			stats.isErr = false
			stats.sendOk++
		}
		outStatusCh <- stats
	}
}
