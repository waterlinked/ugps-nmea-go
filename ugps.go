package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type sattelitePosition struct {
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	Cog         float64 `json:"cog"`
	FixQuality  float64 `json:"fix_quality"`
	Hdop        float64 `json:"hdop"`
	NumSats     float64 `json:"numsats"`
	Orientation float64 `json:"orientation"`
	Sog         float64 `json:"sog"`
}

type acousticPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type externalDepth struct {
	Depth       float64 `json:"depth"`
	Temperature float64 `json:"temp"`
}

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

var baseURL = "http://127.0.0.1:8080"

var client = &http.Client{
	Timeout: time.Second * 1,
}

func getJSON(url string, target interface{}) error {
	r, err := client.Get(url)
	if err != nil {
		return err
	}
	if r.StatusCode != 200 {
		return fmt.Errorf("Expect status 200, got %d", r.StatusCode)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getGlobalPosition() (sattelitePosition, error) {
	url := baseURL + "/api/v1/position/global"

	var pos sattelitePosition
	if err := getJSON(url, &pos); err != nil {
		return pos, err
	}
	return pos, nil
}

func getAcousticPosition() (acousticPosition, error) {
	url := baseURL + "/api/v1/position/acoustic/filtered"

	var pos acousticPosition
	if err := getJSON(url, &pos); err != nil {
		return pos, err
	}
	return pos, nil
}

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
		fmt.Printf("%v\n", resp)
		return fmt.Errorf("Expected status 200 but got %d", resp.StatusCode)
	}
	return nil
}

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
		fmt.Printf("%v\n", resp)
		return fmt.Errorf("Expected status 200 but got %d", resp.StatusCode)
	}
	return nil

}
