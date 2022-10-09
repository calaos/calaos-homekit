package main

import (
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/brutella/hap/accessory"
	"github.com/brutella/hap/characteristic"
)

type LightDimmer struct {
	*accessory.Lightbulb
	Brightness *characteristic.Brightness
	Name       *characteristic.Name
}

func NewLightDimmer(cio CalaosIO, id uint64) *LightDimmer {
	acc := LightDimmer{}
	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
	}

	acc.Lightbulb = accessory.NewLightbulb(info)
	acc.Lightbulb.Id = id

	acc.Brightness = characteristic.NewBrightness()
	acc.Name = characteristic.NewName()

	acc.Lightbulb.Lightbulb.AddC(acc.Brightness.C)
	acc.Lightbulb.Lightbulb.AddC(acc.Name.C)

	acc.Update(&cio)

	acc.Lightbulb.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
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
	if cio.GuiType == "light_dimmer" {
		v, err := strconv.Atoi(cio.State)
		if err == nil {
			acc.Brightness.SetValue(v)
			if v == 0 {
				acc.Lightbulb.Lightbulb.On.SetValue(false)
			} else {
				acc.Lightbulb.Lightbulb.On.SetValue(true)
			}
		}
		return err
	} else {
		v, err := strconv.ParseBool(cio.State)
		if err == nil {
			acc.Lightbulb.Lightbulb.On.SetValue(v)
			if v {
				acc.Brightness.SetValue(100)
			} else {
				acc.Brightness.SetValue(0)
			}
		}
		return err
	}
}

func (acc *LightDimmer) AccessoryGet() *accessory.A {
	return acc.A
}
