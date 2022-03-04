package main

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Latitude or Longitude
type LatLng float64

// Degrees and minutes
func (l LatLng) DM() (int, float64) {
	if l < 0 {
		l = -l
	}

	d := math.Floor(float64(l))
	m := (float64(l) - d) * 60

	return int(d), m
}

func (l LatLng) Serialise() string {
	if l == 0 {
		return ""
	}
	d, m := l.DM()
	return strings.Trim(fmt.Sprintf("%d%f", d, m), "0")
}

// CardinalPoint returns the cardinal point: N/S if isLat else E/W
func (l LatLng) CardinalPoint(isLat bool) string {
	if l == 0 {
		return ""
	}

	// Lattitude
	if isLat {
		if l < 0 {
			return "S"
		}
		return "N"
	}

	// Longitude
	if l < 0 {
		return "W"
	}
	return "E"
}

func assembleSentence(fields []string) string {
	sentence := strings.Join(fields, ",")
	out := "$" + sentence + "*"
	var csum uint8
	for i := 0; i < len(sentence); i++ {
		csum ^= sentence[i]
	}
	checksum := fmt.Sprintf("%02X", csum)
	return out + checksum

}

/*
RATLL struct represents the "--TTL" NMEA sentence

https://gpsd.gitlab.io/gpsd/NMEA.html#_tll_target_latitude_and_longitude

Field Number:

1. Target Number (0-99)
2. Target Latitude
3. N=north, S=south
4. Target Longitude
5. E=east, W=west
6. Target name
7. UTC of data
8. Status (L=lost, Q=acquisition, T=tracking)
9. R= reference target; null (,,)= otherwise
*/
type RATLL struct {
	TargetNum    int
	Latitude     LatLng // In decimal format
	Longitude    LatLng // In decimal format
	TargetName   string
	TimeUTC      time.Time // Aggregation of TimeUTC data field
	TargetStatus byte      // L=lost, Q=acuisition, T=tracking
}

func (sentence RATLL) Serialise() string {

	fields := make([]string, 0)
	fields = append(fields, "RATLL")

	fields = append(fields, fmt.Sprintf("%d", sentence.TargetNum))

	fields = append(fields,
		strings.Trim(sentence.Latitude.Serialise(), "0"), sentence.Latitude.CardinalPoint(true),
		strings.Trim(sentence.Longitude.Serialise(), "0"), sentence.Longitude.CardinalPoint(false),
	)

	fields = append(fields, fmt.Sprintf("%s", sentence.TargetName))

	fields = append(fields, sentence.TimeUTC.Format("150405.000"))
	if sentence.TargetStatus == 'T' {
		fields = append(fields, "T")
	} else {
		fields = append(fields, "L")
	}

	return assembleSentence(fields)
}

/*
GAGGA structure represents --GGA message
https://gpsd.gitlab.io/gpsd/NMEA.html#_gga_global_positioning_system_fix_data

Fields:
1. UTC of this position report
2. Latitude
3. N or S (North or South)
4. Longitude
5. E or W (East or West)
6. GPS Quality Indicator (non null)
	0 - fix not available,
	1 - GPS fix,
	2 - Differential GPS fix (values above 2 are 2.3 features)
	3 = PPS fix
	4 = Real Time Kinematic
	5 = Float RTK
	6 = estimated (dead reckoning)
	7 = Manual input mode
	8 = Simulation mode
7. Number of satellites in use, 00 - 12
8. Horizontal Dilution of precision (meters)
9. Antenna Altitude above/below mean-sea-level (geoid) (in meters)
10. Units of antenna altitude, meters
11. Geoidal separation, the difference between the WGS-84 earth ellipsoid and mean-sea-level (geoid), "-" means mean-sea-level below ellipsoid
12. Units of geoidal separation, meters
13. Age of differential GPS data, time in seconds since last SC104 type 1 or 9 update, null field when DGPS is not used
14. Differential reference station ID, 0000-1023
15. Checksum

Example: $GPGGA,015540.000,3150.68378,N,11711.93139,E,1,17,0.6,0051.6,M,0.0,M,,*58

*/

type GAGGA struct {
	TimeUTC                time.Time
	Latitude               LatLng
	Longitude              LatLng
	QualityIndicator       float64
	NumberOfSatellitesUsed int
	Altitude               float64
	Hdop                   float64
}

func (sentence GAGGA) Serialise() string {

	fields := make([]string, 0)
	fields = append(fields, "GPGGA")

	fields = append(fields, sentence.TimeUTC.Format("150405.000"))

	fields = append(fields,
		strings.Trim(sentence.Latitude.Serialise(), "0"), sentence.Latitude.CardinalPoint(true),
		strings.Trim(sentence.Longitude.Serialise(), "0"), sentence.Longitude.CardinalPoint(false),
	)

	fields = append(fields, fmt.Sprintf("%.0f", sentence.QualityIndicator))
	fields = append(fields, fmt.Sprintf("%d", sentence.NumberOfSatellitesUsed))
	if sentence.Hdop > 0 {
		fields = append(fields, fmt.Sprintf("%.1f", sentence.Hdop))
	} else {
		fields = append(fields, "")
	}
	fields = append(fields, fmt.Sprintf("%.2f", sentence.Altitude), "M")
	fields = append(fields, "", "M")
	fields = append(fields, "", "")

	return assembleSentence(fields)
}
