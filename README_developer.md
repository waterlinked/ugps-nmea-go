# Water Linked Underwater GPS NMEA bridge (wl-95051)

## Test for release

- Run unit tests `go test`
- Start main application `test/test-run.sh`
- Start sending data to input stream `test/test-udp-send.sh`
- Verify data is outputted `test/test-udp-receive.sh`

### Verify with OpenCPN:

- Start main application `go run . -d -url https://demo.waterlinked.com -o localhost:2947`
- Install and start OpenCPN (apt install opencpn, opencpn)
- Go to "options" -> "connections"
- Click "Add connection".
- Select "Network"  Protocol: "UDP"  DataPort 2947
- Select "Apply"

Verify that GPS signal icon in top right corner is green
Verify boat moves in a circle.

- Start main application with RATLL output: `go run . -d -url https://demo.waterlinked.com -o localhost:2947 -sentence RATLL`
- Verify the ROV target is moving in a circle

## Release

- Push changes
- Verify "Actions" build and test successfully
- Download artifacts from "Build" actions.
- Click "Draft new release" on [github](https://github.com/waterlinked/ugps-nmea-go/releases)
- Create tag with release number (eg v1.2.3)
- Describe changes
- Attache artifacts which you downloaded above
- Publish release
