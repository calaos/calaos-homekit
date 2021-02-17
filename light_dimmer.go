package main

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type LightDimmer struct {
	*accessory.Accessory
	LightDimmer *service.Lightbulb
	Brightness  *characteristic.Brightness
	Name        *characteristic.Name
}

func NewLightDimmer(cio CalaosIO, id uint64) *LightDimmer {
	acc := LightDimmer{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
		ID:           id,
	}

	acc.Accessory = accessory.New(info, accessory.TypeLightbulb)
	acc.LightDimmer = service.NewLightbulb()

	acc.AddService(acc.LightDimmer.Service)

	acc.Brightness = characteristic.NewBrightness()
	acc.Name = characteristic.NewName()

	acc.LightDimmer.Service.AddCharacteristic(acc.Brightness.Characteristic)
	acc.LightDimmer.Service.AddCharacteristic(acc.Name.Characteristic)

	if v, err := strconv.ParseFloat(cio.State, 32); err == nil {
		ival := int(v)
		acc.Brightness.SetValue(ival)
		if ival == 0 {
			acc.LightDimmer.On.SetValue(false)
		} else {
			acc.LightDimmer.On.SetValue(true)
		}
	}

	acc.LightDimmer.On.OnValueRemoteUpdate(func(on bool) {
		if on == true {
			log.Debug("Switch is on")
			cio.State = "true"
			CalaosUpdate(cio)
		} else {
			log.Debug("Switch is off")
			cio.State = "false"
			CalaosUpdate(cio)
		}

	})

	acc.Brightness.OnValueRemoteUpdate(func(val int) {
		cio.State = "set " + strconv.Itoa(val)
		CalaosUpdate(cio)
	})

	return &acc
}

func (acc *LightDimmer) Update(cio *CalaosIO) error {
	log.Debug("try to update val ", cio.State)
	v, err := strconv.Atoi(cio.State)
	if err == nil {
		acc.Brightness.SetValue(v)
		if v == 0 {
			acc.LightDimmer.On.SetValue(false)
		}
	}
	return err
}

func (acc *LightDimmer) AccessoryGet() *accessory.Accessory {
	return acc.Accessory
}
