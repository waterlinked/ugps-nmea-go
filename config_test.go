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
  sentence: hdm
output:
  device: /dev/ttyUSB1@9600
  sentence: gpgga
ugps: http://127.0.0.1:8080
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
	assert.Equal(t, "gpgga", cfg.Output.Sentence)

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
