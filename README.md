# Water Linked Underwater GPS NMEA bridge (wl-95051)

## About

This application can be used to let the [Water Linked Underwater GPS](https://waterlinked.com/underwater-gps/) use an external GPS/compass as input GPS and send the Locator position to a chart plotter.

The application reads NMEA 0183 input from a serial/UDP connection and sends it to Water Linked Underwater GPS to allow it to use compass (HDT sentence) and GPS (GGA sentence) as an external source. Once this application is running the Underwater GPS must be configured to use this external source in the [settings](https://waterlinked.github.io/underwater-gps/gui/settings/)

The application also reads the latitude/longitude of the Locator from the Underwater GPS and sends it via serial or UDP as a NMEA sentence (type of sentence is configurable).

## Installation

[Download the application](https://github.com/waterlinked/ugps-nmea-go/releases) for your platform:

| Name | Platform |
|------|----------|
| nmea_ugps_linux_armv6 | Linux ARMv6 - Raspberry PI etc |
| nmea_ugps_linux_amd64 | Linux 64 bit |
| nmea_ugps_windows_amd64.exe | Windows 64 bit |

## Running

The application is run on the command line and can be configured via the config file (config.yml). Copy `config_example.yml` to `config.yml` and modify to suit your setup.

```yaml
#
# Example config file
#
input:
# Input from COM port - device: COM1@9600
# Input from UDP - device: 127.0.0.1:2948
  device: COM1
# Sentences: hdm, hdt, ths
  heading_sentence: hdm
output:
# Output to UDP - device: 127.0.0.1:2947
# Output to COM port - device: COM1@9600
  device: 127.0.0.1:2947
# Sentences: gpgga, ratll
  position_sentence: gpgga
ugps_url: http://192.168.2.94
```

If the configuration file is not found, parameters from command line are used.

Versions before 1.6.0 used only command line arguments for configuration.
Command line arguments in the 1.6.0 release are compatible with earlier versions.


## Screenshot

When running the application it typically looks like this:

![Screenshot](/screenshot/screenshot.png)
