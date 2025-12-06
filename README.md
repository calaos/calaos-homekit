# CalaosHomekit

## Development

```
go get github.com/calaos/calaos-homekit
go build
```

Modify config.json to reflect your installation setup.
Host should be the ip address of your Calaos server.
Password and User, credentials of your Calaos server.
Port is the port on which calaos server is running (5454 by default)

PinCode is the pin code for pairing iOS device and your Calaos Homekit Gateway. It's asked when pairing.

Launch CalaosHomeKit

```
./calaos-homekit -config config.dev.json
```

Now you can Launch Home application on iOS

- Click on "Add Accessory".
- Click on "I Don't Have a Code or Cannot Scan"
- Calaos-Server appears as Bridge in the list of detected devices.
- Calaos-server is not an guenuine HomeKit device so you need to accept advertisement to be able to communicate with it.
- Enter the pin code in config.json

Now Input/Ouput marked as "visible" in calaos installer are proposed in Homekit. Note that Room is not imported inside HomeKit so you need to change it manually.
For now only IO with following GuiType and IOStyle are supported :

- temp
- input_analog / humidity
- light_dimmer
- light / without ioStyle
- shutter_smart

If you want more types, please ask.

## Depoloy to calaos server

build for linux

```
GOOS=linux GOARCH=amd64 go build
```

copy files to server

```
scp calaos-homekit root@<ip_calaos>:/usr/bin/CalaosHomeKit
scp calaos-homekit.service root@<ip_calaos>:/lib/systemd/system/calaos-homekit.service
scp config.json root@<ip_calaos>:/mnt/calaos/homekit/config.json
```

ask server to start homekit service at each boot

```
systemctl enable calaos-homekit --now
```

## Testing

The project includes unit tests (files ending in `_test.go`). To run tests:

### Run all tests

```bash
go test -v
```

Or with coverage:

```bash
go test -v -cover
```

### Run a specific test file

```bash
# Run main tests
go test -v -run TestCalaosJsonMsgLoginRequest

# Run light dimmer tests
go test -v -run TestLightDimmer

# Run temperature sensor tests
go test -v -run TestTemp

# Run humidity sensor tests
go test -v -run TestHumidity

# Run smart shutter tests
go test -v -run TestSmartShutter
```

### Run a specific test function

```bash
# Run a specific test by name
go test -v -run TestNewLightDimmer

# Run tests matching a pattern
go test -v -run "TestLightDimmer_Update"
```

### Run tests and generate coverage report

```bash
go test -coverprofile=coverage.out
go tool cover -html=coverage.out
```
