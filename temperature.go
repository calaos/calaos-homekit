package main

import (
	"strconv"

	"github.com/brutella/hap/accessory"
)

type Temp struct {
	*accessory.Thermometer
}

const TypeTemperatureSensor = "8A"

func NewTemperatureSensor(cio CalaosIO, id uint64) *Temp {
	acc := Temp{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
	}

	acc.Thermometer = accessory.NewTemperatureSensor(info)
	acc.Thermometer.Id = id
	acc.Thermometer.TempSensor.CurrentTemperature.SetMinValue(-50)
	acc.Thermometer.TempSensor.CurrentTemperature.SetMaxValue(50)
	acc.Thermometer.TempSensor.CurrentTemperature.SetStepValue(0.1)

	acc.Update(&cio)

	return &acc
}

func (acc *Temp) Update(cio *CalaosIO) error {
	t, err := strconv.ParseFloat(cio.State, 32)
	if err == nil {
		acc.TempSensor.CurrentTemperature.SetValue(t)
	}
	return err
}

func (acc *Temp) AccessoryGet() *accessory.A {
	return acc.Thermometer.A
}
