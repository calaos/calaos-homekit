package main

import (
	"github.com/brutella/hap/accessory"
)

type CalaosAccessory interface {
	Update(*CalaosIO) error
	AccessoryGet() *accessory.A
}

type CalaosGateway struct {
	*accessory.Bridge
}

func NewCalaosGateway(name string) *CalaosGateway {
	acc := CalaosGateway{}

	info := accessory.Info{
		Name:         name,
		Manufacturer: "Calaos",
		Model:        "calaos-homekit",
		Firmware:     "3.0.0",
	}
	acc.Bridge = accessory.NewBridge(info)

	return &acc
}
