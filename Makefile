BUILD_FOLDER=build

# This how we want to name the binary output
BINARY=${BUILD_FOLDER}/nmea_ugps

# Pass variables for version number, sha id and build number
VERSION=1.7.0
SHA=$(shell git rev-parse --short HEAD)
# Set fallback build num if not set by environment variable
BUILDNUM?=local

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildNum=${BUILDNUM} -X main.SHA=${SHA}"


all: build

build:
	mkdir -p ${BUILD_FOLDER}
	CGO=0 GOOS=linux GOARCH=arm GOARM=6 go build ${LDFLAGS} -o ${BINARY}_linux_armv6
	CGO=0 GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY}_linux_amd64
	CGO=0 GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o ${BINARY}_windows_amd64.exe

test:
	go test

# Cleans our project
clean:
	rm -r ${BUILD_FOLDER}

.PHONY: all clean build test
