package main

import (
	"strconv"

	"github.com/brutella/hc/service"

	"fmt"
)

type HapIO interface {
	Update(*CalaosIO) error
	Service() *service.Service
	ID() string
}

func NewIO(cio CalaosIO) (HapIO, error) {

	if cio.GuiType == "temp" {
		svc := NewTemperatureSensor(cio.Name)
		v, _ := strconv.ParseFloat(cio.State, 64)
		svc.CurrentTemperature.SetValue(float64(v))
		svc.Name.SetValue(cio.Name)
		return svc, nil
	} else if cio.GuiType == "light" {
		svc := NewLight(cio)
		v := false
		if cio.State == "true" {
			v = true
		} else {
			v = false
		}
		svc.Lightbulb.On.SetValue(v)
		svc.Name.SetValue(cio.Name)
		svc.CalaosId = cio.ID
		return svc, nil
	}

	return nil, fmt.Errorf("Cannot create object: Value: %s, name: %s", cio.State, cio.Name)
}
