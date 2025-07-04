package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/adrianmo/go-nmea"
	"github.com/stretchr/testify/assert"
)

func TestPrependZero(t *testing.T) {
	assert.Equal(t, "01", PrependZero(1, 2, "%.0f"))
	assert.Equal(t, "001", PrependZero(1, 3, "%.0f"))
	assert.Equal(t, "0001.2", PrependZero(1.2, 4, "%.1f"))
}

func TestLatLon(t *testing.T) {
	// http://aprs.gids.nl/nmea/
	// http://www.hiddenvision.co.uk/ez/
	// Latitude	4124.8963, N	41d 24.8963' N or 41d 24' 54" N
	// =>41.414938
	// Longitude	08151.6838, W	81d 51.6838' W or 81d 51' 41" W
	// => 81.861396
	lat := Lat(41.414938)
	assert.Equal(t, "4124.8963", lat.Serialise(4))
	assert.Equal(t, "N", lat.CardinalPoint())

	lng := Lng(-81.861396)
	assert.Equal(t, "08151.6838", lng.Serialise(4))
	assert.Equal(t, "W", lng.CardinalPoint())

	//-40.663289, 175.914934
	//41°24'53.8"N 81°51'41.0"E

	lat = Lat(-40.663289)
	assert.Equal(t, "4039.79734", lat.Serialise(5))
	assert.Equal(t, "S", lat.CardinalPoint())

	lng = Lng(175.914934)
	assert.Equal(t, "17554.89604", lng.Serialise(5))
	assert.Equal(t, "E", lng.CardinalPoint())
}

func TestTTL_Empty(t *testing.T) {
	r := RATLL{}
	res := r.Serialise()
	expected := "$RATLL,0,,,,,,000000.000,L*25"
	assert.Equal(t, expected, res)
}

func TestTTL_Negative(t *testing.T) {
	r := RATLL{
		Latitude:     -1.23,
		Longitude:    -2.34,
		TimeUTC:      time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: 'T',
	}
	res := r.SerialiseDecimals(1)
	expected := "$RATLL,1,0113.8,S,00220.4,W,ROV,203458.651,T*46"
	assert.Equal(t, expected, res)
}

func TestTTL_Positive(t *testing.T) {
	r := RATLL{
		Latitude:     1.23,
		Longitude:    2.34,
		TimeUTC:      time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		TargetName:   "ROV",
		TargetNum:    1,
		TargetStatus: 'T',
	}
	res := r.SerialiseDecimals(1)
	expected := "$RATLL,1,0113.8,N,00220.4,E,ROV,203458.651,T*49"
	assert.Equal(t, expected, res)
}

func TestGGA_Empty(t *testing.T) {
	r := GAGGA{}
	res := r.Serialise()
	expected := "$GPGGA,000000.000,,,,,0,0,,0.00,M,,M,,*56"
	assert.Equal(t, expected, res)
}

func TestGGA_Regular(t *testing.T) {
	r := GAGGA{
		TimeUTC:                time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		Latitude:               1.23,
		Longitude:              2.34,
		QualityIndicator:       1,
		Hdop:                   2.3,
		NumberOfSatellitesUsed: 5,
		Altitude:               -5.367,
	}
	res := r.SerialiseDecimals(1)
	expected := "$GPGGA,203458.651,0113.8,N,00220.4,E,1,5,2.3,-5.37,M,,M,,*6F"
	assert.Equal(t, expected, res)
}

func TestGGA_Negative(t *testing.T) {
	r := GAGGA{
		TimeUTC:                time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		Latitude:               -1.23,
		Longitude:              -2.34,
		QualityIndicator:       5,
		Hdop:                   2.3,
		NumberOfSatellitesUsed: 5,
		Altitude:               -2.111,
	}
	res := r.SerialiseDecimals(1)
	expected := "$GPGGA,203458.651,0113.8,S,00220.4,W,5,5,2.3,-2.11,M,,M,,*67"
	assert.Equal(t, expected, res)
}

func TestGGA_NoHdop(t *testing.T) {
	r := GAGGA{
		TimeUTC:          time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		Latitude:         1.23,
		Longitude:        2.34,
		QualityIndicator: 1,
		Hdop:             0,
		Altitude:         -0.3,
	}
	res := r.SerialiseDecimals(1)
	expected := "$GPGGA,203458.651,0113.8,N,00220.4,E,1,0,,-0.30,M,,M,,*47"
	assert.Equal(t, expected, res)
}

func TestGGA_Docs(t *testing.T) {
	// https://orolia.com/manuals/VSP/Content/NC_and_SS/Com/Topics/APPENDIX/NMEA_GGAmess.htm
	// 4807.038,N	Latitude 48 deg 07.038' N
	// 01131.000, E	Longitude 11 deg 31.000' E
	// expected = "$GPGGA,123519.00,4807.038,N,01131.000,E,1,08,0.9,545.4,M,-164.0,M,,,,*47"
	r := GAGGA{
		TimeUTC:                time.Date(2022, 04, 26, 12, 35, 19, 0, time.UTC),
		Latitude:               48.1173,
		Longitude:              11.516666,
		QualityIndicator:       1,
		NumberOfSatellitesUsed: 8,
		Hdop:                   0.9,
		Altitude:               545.4,
	}
	res := r.SerialiseDecimals(3)

	// originally "$GPGGA,123519.00,4807.038,N,01131.000,E,1,08,0.9,545.4,M,-164.0,M,,,,*47"
	// modified with supported fields:
	expected := "$GPGGA,123519.000,4807.038,N,01131.000,E,1,8,0.9,545.40,M,,M,,*4C"
	assert.Equal(t, expected, res)
}

func TestGGA_Real(t *testing.T) {
	r := GAGGA{
		TimeUTC:          time.Date(2022, 04, 26, 20, 34, 58, 651387237, time.UTC),
		Latitude:         64.07552231,
		Longitude:        11.25716543,
		QualityIndicator: 1,
		Hdop:             0,
		Altitude:         -0.5,
	}
	res := r.SerialiseDecimals(6)

	expected := "$GPGGA,203458.651,6404.531339,N,01115.429926,E,1,0,,-0.50,M,,M,,*40"
	assert.Equal(t, expected, res)
}

func TestGGA_LeadingZeros(t *testing.T) {
	r := GAGGA{
		TimeUTC:          time.Date(2022, 04, 26, 20, 34, 58, 651387237, time.UTC),
		Latitude:         -4.07552231,
		Longitude:        -1.025716543,
		QualityIndicator: 1,
		Hdop:             2,
		Altitude:         -0.5,
	}
	res := r.SerialiseDecimals(6)
	expected := "$GPGGA,203458.651,0404.531339,S,00101.542993,W,1,0,2.0,-0.50,M,,M,,*63"
	assert.Equal(t, expected, res)

	res = r.SerialiseDecimals(3)
	expected = "$GPGGA,203458.651,0404.531,S,00101.543,W,1,0,2.0,-0.50,M,,M,,*68"
	assert.Equal(t, expected, res)

}

func TestFuzz(t *testing.T) {

	for i := 1; i < 5000; i++ {
		r := GAGGA{
			Latitude:  Lat(180*rand.Float64() - 90),
			Longitude: Lng(360*rand.Float64() - 180),
		}
		res := r.Serialise()
		back, err := nmea.Parse(res)
		assert.NoError(t, err)

		assert.IsType(t, nmea.GGA{}, back)
		nm := back.(nmea.GGA)
		assert.InDelta(t, float64(r.Latitude), nm.Latitude, 0.00001)
		assert.InDelta(t, float64(r.Longitude), nm.Longitude, 0.00001)
	}
}
