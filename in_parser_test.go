package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParserInputHDG(t *testing.T) {
	input := "$HCHDG,101.1,,,7.1,W*3C"

	headingParser := &hdgParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.True(t, gotUpdate)

	require.Equal(t, "HDG: 1", headingParser.String())
	require.Equal(t, 1, headingParser.count)
	require.Equal(t, 101.1, latest.Orientation)
}

func TestParserInputHDT(t *testing.T) {
	input := "$GPHDT,274.07,T*03"

	headingParser := &hdtParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.True(t, gotUpdate)

	require.Equal(t, "HDT: 1", headingParser.String())
	require.Equal(t, 1, headingParser.count)
	require.Equal(t, 274.07, latest.Orientation)
}

func TestParserInputHDM(t *testing.T) {
	input := "$HCHDM,277.19,M*13"

	headingParser := &hdmParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.True(t, gotUpdate)

	require.Equal(t, "HDM: 1", headingParser.String())
	require.Equal(t, 1, headingParser.count)
	require.Equal(t, 277.19, latest.Orientation)
}

func TestParserInputTHS(t *testing.T) {
	input := "$GPTHS,338.01,A*0E"

	headingParser := &thsParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.True(t, gotUpdate)

	require.Equal(t, "THS: 1", headingParser.String())
	require.Equal(t, 1, headingParser.count)
	require.Equal(t, 338.01, latest.Orientation)
}

func TestParserInputGGA(t *testing.T) {
	input := "$GPGGA,015540.000,3150.68378,N,11711.93139,E,1,17,0.6,0051.6,M,0.0,M,,*58"

	headingParser := &thsParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.True(t, gotUpdate)

	require.Equal(t, "THS: 0", headingParser.String())
	require.Equal(t, 0, headingParser.count)

	require.InDelta(t, 31.84473, latest.Lat, 0.0001)
	require.InDelta(t, 117.198856, latest.Lon, 0.0001)
	require.Equal(t, 1.0, latest.FixQuality)
	require.Equal(t, 0.6, latest.Hdop)
	require.Equal(t, 17.0, latest.NumSats)
}

func TestParserInvalid(t *testing.T) {
	input := "$GPGGA,*58"

	headingParser := &hdmParser{}
	gotUpdate, err := parseNMEA([]byte(input), headingParser)
	require.NoError(t, err)
	require.False(t, gotUpdate)
}
