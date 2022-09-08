package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	cfg := Config{}

	assert.False(t, cfg.InputEnabled())
	assert.False(t, cfg.OutputEnabled())

	err := readFile(&cfg, "does-not-exist")

	assert.Error(t, err)

	data := `input:
  device: /dev/ttyUSB0
  heading_sentence: hdm
output:
  device: /dev/ttyUSB1@9600
  position_sentence: gpgga
ugps_url: http://127.0.0.1:8080
`
	fn := "/tmp/config.yml.1"
	err = os.WriteFile(fn, []byte(data), 0644)
	assert.NoError(t, err)

	defer os.Remove(fn)

	err = readFile(&cfg, fn)
	assert.NoError(t, err)
	assert.Equal(t, "/dev/ttyUSB0", cfg.Input.Device)
	assert.Equal(t, "hdm", cfg.Input.HeadingSentence)

	assert.Equal(t, "/dev/ttyUSB1@9600", cfg.Output.Device)
	assert.Equal(t, "gpgga", cfg.Output.PositionSentence)

	assert.Equal(t, "http://127.0.0.1:8080", cfg.BaseURL)

	assert.True(t, cfg.InputEnabled())
	assert.True(t, cfg.OutputEnabled())
}

func TestConfigInvalid(t *testing.T) {
	cfg := Config{}

	data := `:`
	fn := "/tmp/config.yml.2"
	err := os.WriteFile(fn, []byte(data), 0644)
	assert.NoError(t, err)

	defer os.Remove(fn)

	err = readFile(&cfg, fn)
	assert.Error(t, err)
}

func TestExampleConfig(t *testing.T) {
	cfg := Config{}

	err := readFile(&cfg, "config_example.yml")
	assert.NoError(t, err)
	assert.NotEmpty(t, cfg.Input.Device)
	assert.NotEmpty(t, cfg.Input.HeadingSentence)
	assert.NotEmpty(t, cfg.Output.Device)
	assert.NotEmpty(t, cfg.Output.PositionSentence)
	assert.NotEmpty(t, cfg.BaseURL)
}
