package main

import (
	"log"

	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type Light struct {
	*service.Lightbulb
	Name     *characteristic.Name
	CalaosId string
}

func NewLight(cio CalaosIO) *Light {
	svc := Light{}
	svc.Lightbulb = service.NewLightbulb()

	svc.Name = characteristic.NewName()
	svc.Name.SetValue(cio.Name)
	svc.AddCharacteristic(svc.Name.Characteristic)
	svc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		if on {
			log.Println("Client changed switch to on")
			cio.State = "true"
		} else {
			log.Println("Client changed switch to off")
			cio.State = "false"
		}
		CalaosUpdate(cio)
	})

	return &svc
}

func (svc *Light) Service() *service.Service {
	return svc.Lightbulb.Service
}

func (svc *Light) Update(cio *CalaosIO) error {
	v := false
	if cio.State == "true" {
		v = true
	} else {
		v = false
	}
	svc.Lightbulb.On.SetValue(v)
	return nil
}

func (svc *Light) ID() string {
	return svc.CalaosId
}
