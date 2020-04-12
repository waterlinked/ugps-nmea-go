package main

import (
	"time"
)

// outSerializer is the interface for outputting varying NMEA sentences
type outSerializer interface {
	serialize(sattelitePosition) string
	noPosition() string
}

// QualityNoFix represents no fix in an GGA sentence
const QualityNoFix = 0

// TargetStatusTracking represents sucessfully tracking a target
const TargetStatusTracking = 'T'

// TargetStatusLost represents loosing the tracked target
const TargetStatusLost = 'L'

type ggaSerializer struct{}

func (s ggaSerializer) serialize(pos sattelitePosition) string {
	sentence := GAGGA{
		TimeUTC:            time.Now().UTC(),
		Latitude:           LatLng(pos.Lat),
		Longitude:          LatLng(pos.Lon),
		QualityIndicator:   pos.FixQuality,
		Hdop:               pos.Hdop,
		NbOfSatellitesUsed: int(pos.NumSats),
	}
	out := sentence.Serialize()
	return out
}

func (s ggaSerializer) noPosition() string {
	gga := GAGGA{
		TimeUTC:            time.Now().UTC(),
		Latitude:           LatLng(0),
		Longitude:          LatLng(0),
		QualityIndicator:   QualityNoFix,
		NbOfSatellitesUsed: int(0),
		Hdop:               0,
	}
	return gga.Serialize()
}

type tllSerializer struct{}

func (s tllSerializer) serialize(pos sattelitePosition) string {
	sentence := RATLL{
		TimeUTC:      time.Now().UTC(),
		Latitude:     LatLng(pos.Lat),
		Longitude:    LatLng(pos.Lon),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: TargetStatusTracking,
	}
	return sentence.Serialize()
}

func (s tllSerializer) noPosition() string {
	sentence := RATLL{
		TimeUTC:      time.Now().UTC(),
		Latitude:     LatLng(0),
		Longitude:    LatLng(0),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: TargetStatusLost,
	}
	return sentence.Serialize()
}

/*
// pilebonesSerialiser is a serializer using the library
// available on "github.com/pilebones/go-nmea"
type pilebonesSerialiser struct{}

func (s pilebonesSerialiser) serialize(pos sattelitePosition) string {
	gga := nmea.GPGGA{
		TimeUTC:            time.Now().UTC(),
		Latitude:           nmea.LatLong(pos.Lat),
		Longitude:          nmea.LatLong(pos.Lon),
		QualityIndicator:   nmea.QualityIndicator(pos.FixQuality),
		NbOfSatellitesUsed: uint64(pos.NumSats),
		HDOP:               pos.Hdop,
		Altitude:           0,
	}
	return gga.Serialize()
}

func (s pilebonesSerialiser) noPosition() string {
	gga := nmea.GPGGA{
		TimeUTC:            time.Now().UTC(),
		Latitude:           nmea.LatLong(0),
		Longitude:          nmea.LatLong(0),
		QualityIndicator:   nmea.QualityIndicator(QualityNoFix),
		NbOfSatellitesUsed: uint64(0),
		HDOP:               0,
		Altitude:           0,
	}
	return gga.Serialize()
}
*/
