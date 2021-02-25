package main

import (
	"strconv"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
)

type Temp struct {
	*accessory.Accessory
	TempSensor *service.TemperatureSensor
}

const TypeTemperatureSensor = "8A"

func NewTemperatureSensor(cio CalaosIO, id uint64) *Temp {
	acc := Temp{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
		ID:           id,
	}

	acc.Accessory = accessory.New(info, accessory.TypeSensor)
	acc.TempSensor = service.NewTemperatureSensor()
	acc.TempSensor.CurrentTemperature.SetMinValue(-50)
	acc.TempSensor.CurrentTemperature.SetMaxValue(50)
	acc.TempSensor.CurrentTemperature.SetStepValue(0.1)

	acc.Update(&cio)

	acc.AddService(acc.TempSensor.Service)

	return &acc
}

func (acc *Temp) Update(cio *CalaosIO) error {

	t, err := strconv.ParseFloat(cio.State, 32)
	if err == nil {
		acc.TempSensor.CurrentTemperature.SetValue(t)
	}
	return err
}

func (acc *Temp) AccessoryGet() *accessory.Accessory {
	return acc.Accessory
}
