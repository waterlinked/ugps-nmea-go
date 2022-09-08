package main

import (
	"time"
)

type nmeaPositionSerialiser interface {
	serialise(GlobalPosition, AcousticPosition) string
	noPosition() string
}

// QualityNoFix represents no fix in an GGA sentence
const QualityNoFix = 0

// TargetStatusTracking represents successfully tracking a target
const TargetStatusTracking = 'T'

// TargetStatusLost represents loosing the tracked target
const TargetStatusLost = 'L'

type ggaSerialiser struct{}

func (serialiser ggaSerialiser) serialise(globalPosition GlobalPosition, acousticPosition AcousticPosition) string {
	sentence := GAGGA{
		TimeUTC:                time.Now().UTC(),
		Latitude:               Lat(globalPosition.Latitude),
		Longitude:              Lng(globalPosition.Longitude),
		QualityIndicator:       globalPosition.FixQuality,
		Hdop:                   globalPosition.Hdop,
		NumberOfSatellitesUsed: int(globalPosition.NumSats),
		Altitude:               -acousticPosition.Z,
	}
	out := sentence.Serialise()
	return out
}

func (serialiser ggaSerialiser) noPosition() string {
	gga := GAGGA{
		TimeUTC:                time.Now().UTC(),
		Latitude:               Lat(0),
		Longitude:              Lng(0),
		QualityIndicator:       QualityNoFix,
		NumberOfSatellitesUsed: int(0),
		Hdop:                   0,
		Altitude:               0,
	}
	return gga.Serialise()
}

type tllSerialiser struct{}

func (serialiser tllSerialiser) serialise(globalPosition GlobalPosition, acousticPosition AcousticPosition) string {
	sentence := RATLL{
		TimeUTC:      time.Now().UTC(),
		Latitude:     Lat(globalPosition.Latitude),
		Longitude:    Lng(globalPosition.Longitude),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: TargetStatusTracking,
	}
	return sentence.Serialise()
}

func (serialiser tllSerialiser) noPosition() string {
	sentence := RATLL{
		TimeUTC:      time.Now().UTC(),
		Latitude:     Lat(0),
		Longitude:    Lng(0),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: TargetStatusLost,
	}
	return sentence.Serialise()
}
