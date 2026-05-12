package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/rs/zerolog/log"
	"go.bug.st/serial"
)

type inputStats struct {
	src struct {
		posDesc         string
		posCount        int
		headDesc        string
		unparsableCount int
		errorMsg        string
	}
	dst struct {
		errorMsg string
		sendOk   int
	}
	retransmit struct {
		count    int
		errorMsg string
	}
}

const missingDataTimeout = 10

var (
	stats  inputStats
	latest externalMaster
)

// parseNMEA takes a string and return true if new data, else false
func parseNMEA(data []byte, headingParse nmeaHeadingParser) (bool, error) {
	line := strings.TrimSpace(string(data))

	s, err := nmea.Parse(line)
	if err != nil {
		log.Debug().Msgf("Parse err: %s (%s)", err, line)
		stats.src.unparsableCount++
		return false, nil
	}

	switch m := s.(type) {
	case nmea.GGA:
		log.Debug().Msgf("GGA: Lat/lon : %s %s\n", nmea.FormatGPS(m.Latitude), nmea.FormatGPS(m.Longitude))

		fix, err := strconv.ParseFloat(m.FixQuality, 64)
		if err != nil {
			log.Debug().Msgf("GGA invalid fix quality: %s -> %v\n", m.FixQuality, err)
			fix = 0
		}
		latest.Lat = m.Latitude
		latest.Lon = m.Longitude
		latest.NumSats = float64(m.NumSatellites)
		latest.FixQuality = fix
		latest.Hdop = m.HDOP
		//stats.typeGga++
		stats.src.posCount++
		stats.src.posDesc = fmt.Sprintf("GGA: %d", stats.src.posCount)
		return true, nil
	}
	success, err := headingParse.parseNMEA(s)
	stats.src.headDesc = headingParse.String()
	return success, err
}

func inputUDPLoop(listen string, headingParser nmeaHeadingParser, msg chan externalMaster, inStatsCh chan inputStats, retransmitConn net.Conn) {
	udpAddr, err := net.ResolveUDPAddr("udp4", listen)
	if err != nil {
		log.Debug().Msgf("Unable to resolve UDP address '%s': %s", listen, err)
		fmt.Printf("Unable to resolve UDP address '%s': %s\n", listen, err)
		os.Exit(1)
	}

	ln, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Debug().Msgf("Unable to listen on UDP address '%s': %s", listen, err)
		fmt.Printf("Unable to listen on UDP address '%s': %s\n", listen, err)
		os.Exit(1)
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
			stats.src.errorMsg = fmt.Sprintf("UDP err: %v\n", err)
			inStatsCh <- stats
			continue
		}

		data := buffer[:n]
		if retransmitConn != nil {
			retransmitConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
			_, err := retransmitConn.Write(data)
			if err != nil {
				log.Debug().Msgf("Retransmit error: %s", err)
				stats.retransmit.errorMsg = fmt.Sprintf("Retransmit error: %s", err)
			} else {
				stats.retransmit.count += 1
				stats.retransmit.errorMsg = ""
			}
		}
		stats.src.errorMsg = ""

		gotUpdate, err := parseNMEA(data, headingParser)
		if err != nil {
			stats.src.errorMsg = fmt.Sprintf("%v", err)
		} else if gotUpdate {
			select {
			case msg <- latest: // put message in channel
			default: // channel is full
			}

		}
		inStatsCh <- stats
	}
}

func inputSerialLoop(s serial.Port, headingParser nmeaHeadingParser, msg chan externalMaster, inStatsCh chan inputStats, retransmit io.Writer) {

	scanner := bufio.NewReader(s)
	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			stats.src.errorMsg = fmt.Sprintf("Serial err: %v\n", err)
			inStatsCh <- stats
			continue
		}
		if retransmit != nil {
			_, err := retransmit.Write(line)
			if err != nil {
				log.Debug().Msgf("Retransmit error: %s", err)
				stats.retransmit.errorMsg = fmt.Sprintf("Retransmit error: %s", err)
			} else {
				stats.retransmit.count += 1
				stats.retransmit.errorMsg = ""
			}
		}

		gotUpdate, err := parseNMEA(line, headingParser)
		stats.src.errorMsg = ""

		if err != nil {
			stats.src.errorMsg = fmt.Sprintf("%v", err)
		} else if gotUpdate {
			select {
			case msg <- latest: // put message in channel
			default: // channel is full
			}
		}
		inStatsCh <- stats
	}
}

func inputLoop(masterCh chan externalMaster, inputStatusCh chan inputStats) {

	for {
		select {
		case <-time.After(missingDataTimeout * time.Second):
			stats.src.errorMsg = fmt.Sprintf("Got no input after %d seconds, is data being sent?", missingDataTimeout)
			inputStatusCh <- stats
		case curr := <-masterCh:
			err := setExternalMaster(curr)
			if err == nil {
				stats.dst.sendOk++
				stats.dst.errorMsg = ""
			} else {
				log.Debug().Msgf("%v", err)
				stats.dst.errorMsg = fmt.Sprintf("%v", err)
				inputStatusCh <- stats
			}
			time.Sleep(50 * time.Millisecond) // Sleep to make maximum "setExternalMaster" frequency of 20 Hz
		}
	}
}
