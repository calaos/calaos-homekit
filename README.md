# CalaosHomekit

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
./CalaosHomeKit -config config.json
```
