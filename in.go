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
	src        inputSourceStats
	dst        inputDestinationStats
	retransmit inputRetransmitStats
}

type inputSourceStats struct {
	posDesc         string
	posCount        int
	headDesc        string
	unparsableCount int
	errorMsg        string
}

type inputDestinationStats struct {
	errorMsg string
	sendOk   int
}

type inputRetransmitStats struct {
	count    int
	errorMsg string
}

type inputStatsUpdate struct {
	src            inputSourceStats
	hasSrc         bool
	srcErrorMsg    string
	hasSrcErrorMsg bool
	dst            inputDestinationStats
	hasDst         bool
	retransmit     inputRetransmitStats
	hasRetransmit  bool
}

const missingDataTimeout = 10

// parseNMEA takes a string and return true if new data, else false
func parseNMEA(data []byte, headingParse nmeaHeadingParser, latest *externalMaster, srcStats *inputSourceStats) (bool, error) {
	line := strings.TrimSpace(string(data))

	s, err := nmea.Parse(line)
	if err != nil {
		log.Debug().Msgf("Parse err: %s (%s)", err, line)
		srcStats.unparsableCount++
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
		srcStats.posCount++
		srcStats.posDesc = fmt.Sprintf("GGA: %d", srcStats.posCount)
		return true, nil
	}
	success, err := headingParse.parseNMEA(s, latest)
	srcStats.headDesc = headingParse.String()
	return success, err
}

func inputStatsLoop(updates <-chan inputStatsUpdate, inStatsCh chan inputStats) {
	var stats inputStats

	for update := range updates {
		if update.hasSrc {
			stats.src = update.src
		}
		if update.hasSrcErrorMsg {
			stats.src.errorMsg = update.srcErrorMsg
		}
		if update.hasDst {
			stats.dst = update.dst
		}
		if update.hasRetransmit {
			stats.retransmit = update.retransmit
		}

		publishLatestInputStats(inStatsCh, stats)
	}
}

func publishLatestInputStats(inStatsCh chan inputStats, stats inputStats) {
	select {
	case inStatsCh <- stats:
	default:
		select {
		case <-inStatsCh:
		default:
		}
		inStatsCh <- stats
	}
}

func inputUDPLoop(listen string, headingParser nmeaHeadingParser, msg chan externalMaster, statsUpdates chan<- inputStatsUpdate, retransmitConn net.Conn) {
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
	var srcStats inputSourceStats
	var retransmitStats inputRetransmitStats
	var latest externalMaster

	for {

		ln.SetReadDeadline(time.Now().Add(2 * time.Second))
		n, _, err := ln.ReadFromUDP(buffer)

		if err != nil {
			nerr := err.(net.Error)
			if nerr != nil && nerr.Timeout() {
				continue
			}
			srcStats.errorMsg = fmt.Sprintf("UDP err: %v\n", err)
			statsUpdates <- inputStatsUpdate{src: srcStats, hasSrc: true}
			continue
		}

		data := buffer[:n]
		if retransmitConn != nil {
			retransmitConn.SetWriteDeadline(time.Now().Add(1 * time.Second))
			_, err := retransmitConn.Write(data)
			if err != nil {
				log.Debug().Msgf("Retransmit error: %s", err)
				retransmitStats.errorMsg = fmt.Sprintf("Retransmit error: %s", err)
			} else {
				retransmitStats.count += 1
				retransmitStats.errorMsg = ""
			}
		}
		srcStats.errorMsg = ""

		gotUpdate, err := parseNMEA(data, headingParser, &latest, &srcStats)
		if err != nil {
			srcStats.errorMsg = fmt.Sprintf("%v", err)
		} else if gotUpdate {
			select {
			case msg <- latest: // put message in channel
			default: // channel is full
			}

		}
		statsUpdates <- inputStatsUpdate{
			src:           srcStats,
			hasSrc:        true,
			retransmit:    retransmitStats,
			hasRetransmit: true,
		}
	}
}

func inputSerialLoop(s serial.Port, headingParser nmeaHeadingParser, msg chan externalMaster, statsUpdates chan<- inputStatsUpdate, retransmit io.Writer) {

	scanner := bufio.NewReader(s)
	var srcStats inputSourceStats
	var retransmitStats inputRetransmitStats
	var latest externalMaster

	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			srcStats.errorMsg = fmt.Sprintf("Serial err: %v\n", err)
			statsUpdates <- inputStatsUpdate{src: srcStats, hasSrc: true}
			continue
		}
		if retransmit != nil {
			_, err := retransmit.Write(line)
			if err != nil {
				log.Debug().Msgf("Retransmit error: %s", err)
				retransmitStats.errorMsg = fmt.Sprintf("Retransmit error: %s", err)
			} else {
				retransmitStats.count += 1
				retransmitStats.errorMsg = ""
			}
		}

		gotUpdate, err := parseNMEA(line, headingParser, &latest, &srcStats)
		srcStats.errorMsg = ""

		if err != nil {
			srcStats.errorMsg = fmt.Sprintf("%v", err)
		} else if gotUpdate {
			select {
			case msg <- latest: // put message in channel
			default: // channel is full
			}
		}
		statsUpdates <- inputStatsUpdate{
			src:           srcStats,
			hasSrc:        true,
			retransmit:    retransmitStats,
			hasRetransmit: true,
		}
	}
}

func inputLoop(masterCh chan externalMaster, statsUpdates chan<- inputStatsUpdate) {
	var dstStats inputDestinationStats

	for {
		select {
		case <-time.After(missingDataTimeout * time.Second):
			statsUpdates <- inputStatsUpdate{
				srcErrorMsg:    fmt.Sprintf("Got no input after %d seconds, is data being sent?", missingDataTimeout),
				hasSrcErrorMsg: true,
			}
		case curr := <-masterCh:
			err := setExternalMaster(curr)
			if err == nil {
				dstStats.sendOk++
				dstStats.errorMsg = ""
			} else {
				log.Debug().Msgf("%v", err)
				dstStats.errorMsg = fmt.Sprintf("%v", err)
			}
			statsUpdates <- inputStatsUpdate{dst: dstStats, hasDst: true}
			time.Sleep(50 * time.Millisecond) // Sleep to make maximum "setExternalMaster" frequency of 20 Hz
		}
	}
}
