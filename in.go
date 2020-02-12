package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/tarm/serial"
)

type inputStats struct {
	typeGga  int
	typeHdt  int
	typeThs  int
	isErr    bool
	errorMsg string
	sendOk   int
}

const missingDataTimeout = 10

var (
	stats  inputStats
	latest externalMaster
)

// parseNMEA takes a string and return true if new data, else false
func parseNMEA(data []byte) (bool, error) {
	line := strings.TrimSpace(string(data))

	s, err := nmea.Parse(line)
	if err != nil {
		return false, err
	}

	switch m := s.(type) {
	case nmea.GPGGA:
		//debugPrintf("GGA: Lat/lon : %s %s\n", nmea.FormatGPS(m.Latitude), nmea.FormatGPS(m.Longitude))

		fix, err := strconv.ParseFloat(m.FixQuality, 64)
		if err != nil {
			log.Printf("GGA invalid fix quality: %s -> %v\n", m.FixQuality, err)
			fix = 0
		}
		latest.Lat = m.Latitude
		latest.Lon = m.Longitude
		latest.NumSats = float64(m.NumSatellites)
		latest.FixQuality = fix
		latest.Hdop = m.HDOP
		stats.typeGga++
		return true, nil

	case nmea.GPHDT:
		//debugPrintf("HDT: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		stats.typeHdt++
		return true, nil
	case nmea.THS:
		//debugPrintf("THS: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		stats.typeThs++
		return true, nil
	}
	return false, nil
}

func inputUDPLoop(listen string, msg chan externalMaster, inStatsCh chan inputStats) {
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

		gotUpdate, err := parseNMEA(buffer[:n])
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

func inputSerialLoop(port string, baudrate int, msg chan externalMaster, inStatsCh chan inputStats) {
	if baudrate == 0 {
		baudrate = 112500
	}

	c := &serial.Config{Name: port, Baud: baudrate}
	s, err := serial.OpenPort(c)
	if err != nil {
		fmt.Printf("Error opening serial port: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	scanner := bufio.NewReader(s)
	for {
		line, _, err := scanner.ReadLine()
		if err != nil {
			stats.errorMsg = fmt.Sprintf("Serial err: %v\n", err)
			stats.isErr = true
			inStatsCh <- stats
			continue
		}
		gotUpdate, err := parseNMEA(line)

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

func inputLoop(listen string, inputStatusCh chan inputStats) {
	masterCh := make(chan externalMaster, 1)

	// Use ":" to decide if this is UDP address or serial device
	if len(strings.Split(listen, ":")) > 1 {
		// Start listening on UDP
		go inputUDPLoop(listen, masterCh, inputStatusCh)
	} else {
		baudrate := 0
		// Is the baudrate specified?
		parts := strings.Split(listen, "@")
		if len(parts) > 1 {
			b, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Printf("Unable to parse baudrate: %s as numeric value\n", parts[1])
				os.Exit(1)
			}
			baudrate = b
			listen = parts[0]
		}
		// Start listening on serial port
		go inputSerialLoop(listen, baudrate, masterCh, inputStatusCh)
	}

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
				stats.isErr = true
				stats.errorMsg = fmt.Sprintf("%v", err)
				inputStatusCh <- stats
			}
		}
	}
}
