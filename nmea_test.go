package main

import (
	"testing"
	"time"
)

func TestTTL_Empty(t *testing.T) {
	r := RATLL{}
	res := r.Serialize()
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
	res := r.Serialize()
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
	res := r.Serialize()
	expected := "$RATLL,1,113.8,N,220.4,E,ROV,203458.651,T*79"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}

func TestGGA_Empty(t *testing.T) {
	r := GAGGA{}
	res := r.Serialize()
	expected := "$GPGGA,000000.000,,,,,0,0,,,M,,M,,*48"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}

func TestGGA_Regular(t *testing.T) {
	r := GAGGA{
		TimeUTC:            time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		Latitude:           1.23,
		Longitude:          2.34,
		QualityIndicator:   1,
		Hdop:               2.3,
		NbOfSatellitesUsed: 5,
	}
	res := r.Serialize()
	expected := "$GPGGA,203458.651,113.8,N,220.4,E,1,5,2.3,,M,,M,,*6D"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}

func TestGGA_Negative(t *testing.T) {
	r := GAGGA{
		TimeUTC:            time.Date(2009, 11, 17, 20, 34, 58, 651387237, time.UTC),
		Latitude:           -1.23,
		Longitude:          -2.34,
		QualityIndicator:   5,
		Hdop:               2.3,
		NbOfSatellitesUsed: 5,
	}
	res := r.Serialize()
	expected := "$GPGGA,203458.651,113.8,S,220.4,W,5,5,2.3,,M,,M,,*66"
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
	}
	res := r.Serialize()
	expected := "$GPGGA,203458.651,113.8,N,220.4,E,1,0,,,M,,M,,*47"
	if res != expected {
		t.Errorf("Expected '%s' got '%s'", expected, res)
	}
}
