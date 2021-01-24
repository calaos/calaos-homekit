package main

import (
	"strconv"

	"github.com/brutella/hc/accessory"
)

type Temp struct {
	*accessory.Accessory
}

func NewTemperatureSensor(cio CalaosIO, id uint64) *Temp {
	acc := Temp{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
		ID:           id,
	}

	if t, err := strconv.ParseFloat(cio.State, 32); err == nil {
		acc.Accessory = accessory.NewTemperatureSensor(info, t, -50, 80, 0.1).Accessory
	} else {
		acc.Accessory = accessory.NewTemperatureSensor(info, 0, -50, 80, 0.1).Accessory
	}

	return &acc
}
