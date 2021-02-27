package fritzbox

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type xmlDeviceList struct {
	XMLName xml.Name    `xml:"devicelist"`
	Device  []xmlDevice `xml:"device"`
}

type xmlDevice struct {
	XMLName         xml.Name       `xml:"device"`
	Identifier      string         `xml:"identifier,attr"`
	Id              string         `xml:"id,attr"`
	FunctionBitmask int            `xml:"functionbitmask,attr"`
	FwVersion       string         `xml:"fwversion,attr"`
	Manufacturer    string         `xml:"manufacturer,attr"`
	ProductName     string         `xml:"productname,attr"`
	Present         int            `xml:"present"`
	Name            string         `xml:"name"`
	Switch          xmlSwitch      `xml:"switch"`
	Powermeter      xmlPowermeter  `xml:"powermeter"`
	Temperature     xmlTemperature `xml:"temperature"`
}

type xmlSwitch struct {
	XMLName    xml.Name `xml:"switch"`
	State      int      `xml:"state"`
	Mode       string   `xml:"mode"`
	Lock       int      `xml:"loc"`
	DeviceLock int      `xml:"deviceloc"`
}

type xmlPowermeter struct {
	XMLName xml.Name `xml:"powermeter"`
	Power   float64  `xml:"power"`
	Energy  float64  `xml:"energy"`
	Voltage float64  `xml:"voltage"`
}

type xmlTemperature struct {
	XMLName xml.Name `xml:"temperature"`
	Celsius float64  `xml:"celsius"`
	Offset  float64  `xml:"offset"`
}

const (
	getdevicelistinfosURL = "http://%s/webservices/homeautoswitch.lua?sid=%s&switchcmd=getdevicelistinfos"
)

// GetDeviceList send a "getdevicelistinfos" to the Fritzbox and store the device infos.
// The return code contains the Number of recognized devices
func (s *Session) Devices() ([]Device, error) {
	var xmlFile xmlDeviceList

	url := fmt.Sprintf(getdevicelistinfosURL, s.host, s.sid)

	if err := getXMLStructure(url, &xmlFile); err != nil {
		return []Device{}, err
	}

	devices := make([]Device, 0, len(xmlFile.Device))

	for _, device := range xmlFile.Device {
		d := Device{
			Name:            strings.ToLower(device.Name),
			Identifier:      device.Identifier,
			Id:              device.Id,
			FunctionBitMask: device.FunctionBitmask,
			FWVersion:       device.FwVersion,
			Manufacturer:    device.Manufacturer,
			ProductName:     device.ProductName,
			Present:         device.Present == 1,
			Temperature:     device.temperature() + device.Temperature.Offset/10.0,

			OnOffDevice: struct {
				State      int
				Mode       string
				Lock       int
				DeviceLock int
			}{
				State:      device.switchState(),
				Mode:       device.switchMode(),
				Lock:       device.switchLock(),
				DeviceLock: device.Switch.DeviceLock,
			},
			Powermeter: struct {
				Power   float64
				Energy  float64
				Voltage float64
			}{Power: device.Powermeter.Power / 1000.0,
				Energy:  device.Powermeter.Energy,
				Voltage: device.Powermeter.Voltage / 1000.0,
			},
		}

		devices = append(devices, d)
	}
	return devices, nil
}

func (d *xmlDevice) temperature() float64 {
	if (d.FunctionBitmask & temperatureSensor) == 0 {
		return 0
	}

	return (d.Temperature.Celsius + d.Temperature.Offset) / 10
}

func (d *xmlDevice) switchState() int {
	if (d.FunctionBitmask&switchingSocket) == switchingSocket && d.Switch.State == On {
		return On
	}
	return Off
}

func (d *xmlDevice) switchMode() string {
	if (d.FunctionBitmask & switchingSocket) == 0 {
		return Auto
	}
	return d.Switch.Mode
}

func (d *xmlDevice) switchLock() int {
	if (d.FunctionBitmask & switchingSocket) == 0 {
		return Unlock
	}
	return d.Switch.Lock
}
