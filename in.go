package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/waterlinked/go-nmea"
	"go.bug.st/serial"
)

type inputStats struct {
	posDesc         string
	posCount        int
	headDesc        string
	unparsableCount int
	isErr           bool
	errorMsg        string
	sendOk          int
}

const missingDataTimeout = 10

var (
	stats  inputStats
	latest externalMaster
)

// parseNMEA takes a string and return true if new data, else false
func parseNMEA(data []byte, h headingParser) (bool, error) {
	line := strings.TrimSpace(string(data))

	s, err := nmea.Parse(line)
	if err != nil {
		debugPrintf("Parse err: %s (%s)", err, line)
		stats.unparsableCount++
		return false, nil
	}

	switch m := s.(type) {
	case nmea.GGA:
		debugPrintf("GGA: Lat/lon : %s %s\n", nmea.FormatGPS(m.Latitude), nmea.FormatGPS(m.Longitude))

		fix, err := strconv.ParseFloat(m.FixQuality, 64)
		if err != nil {
			debugPrintf("GGA invalid fix quality: %s -> %v\n", m.FixQuality, err)
			fix = 0
		}
		latest.Lat = m.Latitude
		latest.Lon = m.Longitude
		latest.NumSats = float64(m.NumSatellites)
		latest.FixQuality = fix
		latest.Hdop = m.HDOP
		//stats.typeGga++
		stats.posCount++
		stats.posDesc = fmt.Sprintf("GGA: %d", stats.posCount)
		return true, nil
	}
	success, err := h.parseNMEA(s)
	stats.headDesc = h.String()
	return success, err
}

func inputUDPLoop(listen string, hParser headingParser, msg chan externalMaster, inStatsCh chan inputStats) {
	udpAddr, err := net.ResolveUDPAddr("udp4", listen)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	buffer := make([]byte, 1024)

	for {

		ln.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := ln.ReadFromUDP(buffer)

		if err != nil {
			nerr := err.(net.Error)
			if nerr != nil && nerr.Timeout() {
				continue
			}
			stats.errorMsg = fmt.Sprintf("UDP err: %v\n", err)
			stats.isErr = true
			inStatsCh <- stats
			continue
		}

		gotUpdate, err := parseNMEA(buffer[:n], hParser)
		if err != nil {
			stats.errorMsg = fmt.Sprintf("%v", err)
			stats.isErr = true
			inStatsCh <- stats
		} else if gotUpdate {
			msg <- latest

			stats.errorMsg = ""
			stats.isErr = false
			inStatsCh <- stats
		}
	}
}

func inputSerialLoop(s serial.Port, hParser headingParser, msg chan externalMaster, inStatsCh chan inputStats) {

	scanner := bufio.NewReader(s)
	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			stats.errorMsg = fmt.Sprintf("Serial err: %v\n", err)
			stats.isErr = true
			inStatsCh <- stats
			continue
		}
		gotUpdate, err := parseNMEA(line, hParser)

		if err != nil {
			stats.errorMsg = fmt.Sprintf("%v", err)
			stats.isErr = true
			inStatsCh <- stats
		} else if gotUpdate {
			msg <- latest

			stats.errorMsg = ""
			stats.isErr = false
			inStatsCh <- stats
		}
	}
}

func inputLoop(masterCh chan externalMaster, inputStatusCh chan inputStats) {

	for {
		select {
		case <-time.After(missingDataTimeout * time.Second):
			stats.isErr = true
			stats.errorMsg = fmt.Sprintf("Got no input after %d seconds, is data being sent?", missingDataTimeout)
			inputStatusCh <- stats
		case curr := <-masterCh:
			err := setExternalMaster(curr)
			if err == nil {
				stats.sendOk++
			} else {
				debugPrintf("%v", err)
				stats.isErr = true
				stats.errorMsg = fmt.Sprintf("%v", err)
				inputStatusCh <- stats
			}
		}
	}
}
