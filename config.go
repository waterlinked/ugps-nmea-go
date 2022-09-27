package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Input struct {
		Device          string `yaml:"device"`
		HeadingSentence string `yaml:"heading_sentence"`
		Retransmit      string `yaml:"retransmit"`
	} `yaml:"input"`
	Output struct {
		Device           string `yaml:"device"`
		PositionSentence string `yaml:"position_sentence"`
	} `yaml:"output"`
	BaseURL string `yaml:"ugps_url"`
}

func readFile(cfg *Config, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed opening %s: %w", filename, err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(cfg)
	if err != nil {
		return fmt.Errorf("failed parsing %s: %w", filename, err)
	}
	return nil
}

func (c Config) InputEnabled() bool {
	return c.Input.Device != ""
}

func (c Config) RetransmitEnabled() bool {
	return c.Input.Retransmit != ""
}

func (c Config) OutputEnabled() bool {
	return c.Output.Device != ""
}
