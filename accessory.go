package main

import (
	"github.com/brutella/hc/accessory"
)

type CalaosAccessory interface {
	Update(*CalaosIO) error
	AccessoryGet() *accessory.Accessory
}

type CalaosGateway struct {
	*accessory.Accessory
}

func NewCalaosGateway() *CalaosGateway {
	acc := CalaosGateway{}

	info := accessory.Info{
		Name:         "Calaos Gateway",
		Model:        "v3",
		Manufacturer: "Calaos",
	}
	acc.Accessory = accessory.New(info, accessory.TypeBridge)

	return &acc
}
