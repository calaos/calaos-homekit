package main

import (
	"strconv"

	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type TemperatureSensor struct {
	*service.TemperatureSensor
	Name     *characteristic.Name
	CalaosId string
}

func NewTemperatureSensor(name string) *TemperatureSensor {
	svc := TemperatureSensor{}
	svc.TemperatureSensor = service.NewTemperatureSensor()
	svc.TemperatureSensor.CurrentTemperature.SetMinValue(-100)
	svc.TemperatureSensor.CurrentTemperature.SetMaxValue(300)

	svc.Name = characteristic.NewName()
	svc.Name.SetValue(name)
	svc.AddCharacteristic(svc.Name.Characteristic)

	return &svc
}

func (svc *TemperatureSensor) Service() *service.Service {
	return svc.TemperatureSensor.Service
}

func (svc *TemperatureSensor) Update(cio *CalaosIO) error {

	v, _ := strconv.ParseFloat(cio.State, 64)
	svc.CurrentTemperature.SetValue(float64(v))

	return nil
}

func (svc *TemperatureSensor) ID() string {
	return svc.CalaosId
}
