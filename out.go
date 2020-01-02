package main

import (
	"fmt"
	"io"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pilebones/go-nmea"
	"github.com/tarm/serial"
)

type outStats struct {
	getOk  int
	getErr int
	sendOk int
	isErr  bool
	errMsg string
}

const updateSeconds = 10

func outputLoop(output string, outStatusCh chan outStats) {
	var prevLat float64
	var prevLon float64

	var stats outStats

	var writer io.Writer

	// Use ":" to decide if this is UDP address or serial device
	if len(strings.Split(output, ":")) > 1 {
		conn, err := net.Dial("udp", output)
		if err != nil {
			fmt.Printf("Error connecting to UDP: %v\n", err)
			os.Exit(1)
		}
		defer conn.Close()
		writer = conn

	} else {
		baudrate := 115200
		port := output
		// Is the baudrate specified?
		parts := strings.Split(output, "@")
		if len(parts) > 1 {
			b, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Unable to parse baudrate: %s as numeric value\n", parts[1])
				os.Exit(1)
			}
			baudrate = b
			port = parts[0]
		}
		c := &serial.Config{Name: port, Baud: baudrate}
		s, err := serial.OpenPort(c)
		if err != nil {
			fmt.Printf("Error opening serial port: %v\n", err)
			os.Exit(1)
		}
		defer s.Close()
		// Start listening on serial port
		writer = s
	}

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
