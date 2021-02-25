package main

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

const (
	CLOSING = iota
	OPENING
	STOPPED
)

type SmartShutter struct {
	*accessory.Accessory
	WindowCovering *service.WindowCovering
	HoldPosition   *characteristic.HoldPosition
	Name           *characteristic.Name
}

/*
	TargetPosition :
	This characteristic describes the target position of accessories.
	This characteristic can be used with doors, windows, awnings or window coverings.
	For windows and doors, a value of 0 indicates that a window (or door) is fully closed
		while a value of 100 indicates a fully open position.
	For blinds/shades/awnings, a value of 0 indicates a position that permits the least light
		and a value of 100 indicates a position that allows most light.
*/

/*
	CurrentPosition :
	This characteristic describes the current position of accessories.
	This characteristic can be used with doors, windows, awnings or window coverings.
	For windows and doors, a value of 0 indicates that a window (or door) is fully closed
		while a value of 100 indicates a fully open position.
	For blinds/shades/awnings, a value of 0 indicates a position that permits the least light
		and a value of 100 indicates a position that allows most light.
*/

/*
	PositionState :
	This characteristic describes the state of the position of accessories.
	This characteristic can be used with doors, windows, awnings or window coverings for presentation purposes.
		Format : uint8
		Minimum value : 0
		Maximum value : 2
		Valid values :
			0 : Going to the minimum value specified in metadata (closing)
			1 : Going to the maximum value specified in metadata (opening)
			2 : Stopped
			3-255 : Reserved
*/

func NewSmartShutter(cio CalaosIO, id uint64) *SmartShutter {
	acc := SmartShutter{}

	info := accessory.Info{
		Name:         cio.Name,
		SerialNumber: cio.ID,
		Manufacturer: "Calaos",
		Model:        cio.IoType,
		ID:           id,
	}

	acc.Accessory = accessory.New(info, accessory.TypeWindowCovering)
	acc.WindowCovering = service.NewWindowCovering()

	acc.AddService(acc.WindowCovering.Service)

	acc.HoldPosition = characteristic.NewHoldPosition()
	acc.Name = characteristic.NewName()

	acc.WindowCovering.Service.AddCharacteristic(acc.HoldPosition.Characteristic)
	acc.WindowCovering.Service.AddCharacteristic(acc.Name.Characteristic)

	acc.Update(&cio)

	acc.WindowCovering.TargetPosition.OnValueRemoteUpdate(func(targetPosition int) {
		//TODO: we should retrieve current position from cio object to compare, not from homekit
		currentPosition := acc.WindowCovering.CurrentPosition.GetValue()
		log.Debug("current position : ", currentPosition, " target position : ", targetPosition)
		if targetPosition != currentPosition {
			// calaos and homekit shutter position values are inverted
			// calaos open = 0, closed = 100
			// homekit open = 100, closed = 0
			// we need to convert from homekit value to calaos with : 100 - x
			state := "set " + strconv.Itoa(100-targetPosition)
			log.Debug(state)
			cio.State = state
			CalaosUpdate(cio)
		}
	})

	return &acc
}

func (acc *SmartShutter) Update(cio *CalaosIO) error {
	if cio.GuiType == "shutter_smart" {
		// split state to get current shutter position
		words := strings.Fields(cio.State)
		command := words[0]
		value := words[len(words)-1]
		ival, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		// calaos and homekit shutter position are inverted
		// calaos open = 0, closed = 100
		// homekit open = 100, closed = 0
		// we need to convert from calaos value to homekit with : 100 - x

		//TODO: target position should be retrieved from "set x" calaos command, but this target position is not sent from calaos server
		// only the current position of the shutter is given by "up", "down" and "stop" and is sent every few milliseconds
		val := 100 - ival
		switch command {
		case "up":
			acc.WindowCovering.PositionState.SetValue(OPENING)
			acc.WindowCovering.TargetPosition.SetValue(val)
			acc.WindowCovering.CurrentPosition.SetValue(val)

		case "down":
			acc.WindowCovering.PositionState.SetValue(CLOSING)
			acc.WindowCovering.TargetPosition.SetValue(val)
			acc.WindowCovering.CurrentPosition.SetValue(val)

		case "stop":
			acc.WindowCovering.TargetPosition.SetValue(val)
			acc.WindowCovering.CurrentPosition.SetValue(val)
			acc.WindowCovering.PositionState.SetValue(STOPPED)
			acc.HoldPosition.SetValue(true)

		case "calibration":
			//TODO

		default:
		}
	}

	return nil
}

func (acc *SmartShutter) AccessoryGet() *accessory.Accessory {
	return acc.Accessory
}
