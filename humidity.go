package main

import (
	"strconv"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/service"
)

type Humidity struct {
	*accessory.A
	HumiditySensor *service.HumiditySensor
}

func NewHumiditySensor(cio CalaosIO, id uint64) *Humidity {
	acc := Humidity{}
	// info := accessory.Info{
	// 	Name:         cio.Name,
	// 	SerialNumber: cio.ID,
	// 	Manufacturer: "Calaos",
	// 	Model:        cio.IoType,
	// }

	// acc.A = accessory.NewHumidifier(info).A
	acc.HumiditySensor = service.NewHumiditySensor()
	acc.HumiditySensor.Id = id

	acc.Update(&cio)

	return &acc
}

func (acc *Humidity) Update(cio *CalaosIO) error {
	if h, err := strconv.ParseFloat(cio.State, 32); err == nil {
		acc.HumiditySensor.CurrentRelativeHumidity.SetValue(h)
	} else {
		acc.HumiditySensor.CurrentRelativeHumidity.SetValue(0.0)
	}
	return nil
}

func (acc *Humidity) AccessoryGet() *accessory.A {
	return acc.A
}
