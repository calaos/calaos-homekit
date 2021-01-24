package main

import (
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type Humidity struct {
	*accessory.Accessory
	HumiditySensor *service.HumiditySensor
}

func NewHumiditySensor(cio CalaosIO, id uint64) *Humidity {
	acc := Humidity{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
		ID:           id,
	}

	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.HumiditySensor = service.NewHumiditySensor()

	acc.AddService(acc.HumiditySensor.Service)

	if h, err := strconv.ParseFloat(cio.State, 32); err == nil {
		acc.HumiditySensor.CurrentRelativeHumidity.SetValue(h)
	} else {
		acc.HumiditySensor.CurrentRelativeHumidity.SetValue(0.0)
	}

	return &acc
}
