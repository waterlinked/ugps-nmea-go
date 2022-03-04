package main

import (
	"testing"
	"time"
)

func TestTTL_Empty(t *testing.T) {
	r := RATLL{}
	res := r.Serialise()
	expected := "$RATLL,0,,,,,,000000.000,L*25"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
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
	res := r.Serialise()
	expected := "$RATLL,1,113.8,S,220.4,W,ROV,203458.651,T*76"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
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
	res := r.Serialise()
	expected := "$RATLL,1,113.8,N,220.4,E,ROV,203458.651,T*79"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}

func TestGGA_Empty(t *testing.T) {
	r := GAGGA{}
	res := r.Serialise()
	expected := "$GPGGA,000000.000,,,,,0,0,,0.00,M,,M,,*56"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
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
	res := r.Serialise()
	expected := "$GPGGA,203458.651,113.8,N,220.4,E,1,5,2.3,-5.37,M,,M,,*5F"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
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
	res := r.Serialise()
	expected := "$GPGGA,203458.651,113.8,S,220.4,W,5,5,2.3,-2.11,M,,M,,*57"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
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
	res := r.Serialise()
	expected := "$GPGGA,203458.651,113.8,N,220.4,E,1,0,,-0.30,M,,M,,*77"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}
