package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GlobalPosition struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	Cog         float64 `json:"cog"`
	FixQuality  float64 `json:"fix_quality"`
	Hdop        float64 `json:"hdop"`
	NumSats     float64 `json:"numsats"`
	Orientation float64 `json:"orientation"`
	Sog         float64 `json:"sog"`
}

type AcousticPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

/*
type externalDepth struct {
	Depth       float64 `json:"depth"`
	Temperature float64 `json:"temp"`
}
*/

type externalMaster struct {
	Cog         float64 `json:"cog"`
	FixQuality  float64 `json:"fix_quality"`
	Hdop        float64 `json:"hdop"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	NumSats     float64 `json:"numsats"`
	Orientation float64 `json:"orientation"`
	Sog         float64 `json:"sog"`
}

var baseURL = "nothing yet"

var client = &http.Client{
	Timeout: time.Second * 1,
}

func getJSON(url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	if r.StatusCode == 500 {
		// 500 error happens if no Locator is detected
		return fmt.Errorf("Locator has no position? Expect status 200, got %d", r.StatusCode)
	} else if r.StatusCode != 200 {
		return fmt.Errorf("Expect status 200, got %d", r.StatusCode)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getGlobalPosition() (GlobalPosition, error) {
	url := baseURL + "/api/v1/position/global"

	var globalPosition GlobalPosition
	if err := getJSON(url, &globalPosition); err != nil {
		return globalPosition, err
	}
	return globalPosition, nil
}

func getAcousticPosition() (AcousticPosition, error) {
	url := baseURL + "/api/v1/position/acoustic/filtered"

	var acousticPosition AcousticPosition
	if err := getJSON(url, &acousticPosition); err != nil {
		return acousticPosition, err
	}
	return acousticPosition, nil
}

/*
func setDepth(depth float64) error {
	url := baseURL + "/api/v1/external/depth"

	extDepth := externalDepth{Depth: depth, Temperature: 10}
	encoded, _ := json.Marshal(extDepth)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected status 200 but got %d", resp.StatusCode)
	}
	return nil
}
*/

func setExternalMaster(ext externalMaster) error {
	url := baseURL + "/api/v1/external/master"

	encoded, _ := json.Marshal(ext)

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(encoded))
	if err != nil {
		return err
	}

	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("Expected status 200 but got %d", resp.StatusCode)
	}
	return nil
}
