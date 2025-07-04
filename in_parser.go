package main

import (
	"fmt"

	"github.com/adrianmo/go-nmea"
)

// nmeaHeadingParser is the interface parsing heading input
type nmeaHeadingParser interface {
	// parseNMEA takes a nmea.Sentence and return true if new data, else false
	parseNMEA(sentence nmea.Sentence) (bool, error)
	// String returns a string representing the current status
	String() string
}

type hdmParser struct {
	count int
}
type hdtParser struct {
	count int
}
type thsParser struct {
	count int
}
type hdgParser struct {
	count int
}

func (p *hdmParser) parseNMEA(sentence nmea.Sentence) (bool, error) {
	switch m := sentence.(type) {
	case nmea.HDM:
		debugPrintf("HDM: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		p.count++
		return true, nil
	}
	return false, nil
}

func (p hdmParser) String() string {
	return fmt.Sprintf("HDM: %d", p.count)
}

func (p *hdtParser) parseNMEA(sentence nmea.Sentence) (bool, error) {
	switch m := sentence.(type) {
	case nmea.HDT:
		debugPrintf("HDT: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		p.count++
		return true, nil
	}
	return false, nil
}

func (p hdtParser) String() string {
	return fmt.Sprintf("HDT: %d", p.count)
}

func (p *thsParser) parseNMEA(sentence nmea.Sentence) (bool, error) {
	switch m := sentence.(type) {
	case nmea.THS:
		debugPrintf("THS: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		p.count++
		return true, nil
	}
	return false, nil
}

func (p thsParser) String() string {
	return fmt.Sprintf("THS: %d", p.count)
}

func (p *hdgParser) parseNMEA(sentence nmea.Sentence) (bool, error) {
	switch m := sentence.(type) {
	case nmea.HDG:
		debugPrintf("HDG: Heading : %f\n", m.Heading)
		latest.Orientation = m.Heading
		p.count++
		return true, nil
	}
	return false, nil
}

func (p hdgParser) String() string {
	return fmt.Sprintf("HDG: %d", p.count)
}
