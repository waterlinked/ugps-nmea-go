# Water Linked Underwater GPS NMEA bridge (wl-95051)

## About

This application can be used to let the [Water Linked Underwater GPS](https://waterlinked.com/underwater-gps/) use an external GPS/compass as input GPS and send the Locator position to a chart plotter.

The application reads NMEA 0183 input from a serial/UDP connection and sends it to Water Linked Underwater GPS to allow it to use compass (HDT sentence) and GPS (GGA sentence) as an external source. Once this application is running the Underwater GPS must be configured to use this external source in the [settings](https://waterlinked.github.io/docs/explorer-kit/gui/settings/)

The application also reads the latitude/longitude of the Locator from the Underwater GPS and sends it via serial or UDP as a NMEA sentence (type of sentence is configurable).

## Installation

[Download the application](https://github.com/waterlinked/ugps-nmea-go/releases) for your platform:

| Name | Platform |
|------|----------|
| nmea_ugps_linux_armv6 | Linux ARMv6 - Raspberry PI etc |
| nmea_ugps_linux_amd64 | Linux 64 bit |
| nmea_ugps_windows_amd64.exe | Windows 64 bit |

## Running

The application is run on the command line and can be configured via arguments. The arguments are:

```
  -i string
    	UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1@4800) to listen for NMEA input.  (default "0.0.0.0:7777")
  -o string
    	UDP device and port (host:port) OR serial device (COM7 /dev/ttyUSB1) to send NMEA output.
  -sentence string
    	NMEA output sentence to use. Supported: RATLL, GPGGA (default "GPGGA")
  -url string
    	URL of Underwater GPS (default "http://192.168.2.94")
```

Example using UART input from /dev/ttyUSB2 with baud rate 4800 and sending the output via UDP on port 9999 on localhost.

```
./nmea_ugps_linux_amd64 -i /dev/ttyUSB2@4800 -o 127.0.0.1:9999
```

On Windows, the easiest is to create a `start.bat` file, edit it with notepad to the desired settings and then double-click it in Explorer to start it. Example of what the file can look like:

```
nmea_ugps_windows_amd64.exe -i COM1 -o 127.0.0.1:2947
pause
```

## Screenshot

When running the application it typically looks like this:

```
┌─Water Linked Underwater GPS NMEA bridge──────────────────────────────────────┐
│PRESS q TO QUIT                                                               │
│In : 0.0.0.0:7777                                                             │
│Out: 127.0.0.1:2947                                                           │
└──────────────────────────────────────────────────────────────────────────────┘
┌─Input status─────────────────────────────────────────────────────────────────┐
│Supported NMEA sentences received:                                            │
│ * GGA: 5                                                                     │
│ * HDT: 6                                                                     │
│ * THS: 0                                                                     │
│Sent sucessfully to UGPS: 10                                                  │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
┌─Output status────────────────────────────────────────────────────────────────┐
│112 positions sent to NMEA out                                                │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```
