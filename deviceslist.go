package fritzbox

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

const getdevicelistinfosURL = "http://%s/webservices/homeautoswitch.lua?sid=%s&switchcmd=getdevicelistinfos"

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

// GetDeviceList send a "getdevicelistinfos" to the Fritzbox and store the device infos.
// The return code contains the Number of recognized devices
func (c *Client) Devices() ([]DeviceInfo, error) {
	var xmlFile xmlDeviceList

	url := fmt.Sprintf(getdevicelistinfosURL, c.host, c.sid)

	if err := getXMLStructure(url, &xmlFile); err != nil {
		return []DeviceInfo{}, err
	}

	devices := make([]DeviceInfo, 0, len(xmlFile.Device))

	for _, device := range xmlFile.Device {
		d := DeviceInfo{
			Name:            strings.ToLower(device.Name),
			Identifier:      device.Identifier,
			Id:              device.id(),
			FunctionBitMask: device.FunctionBitmask,
			FWVersion:       device.FwVersion,
			Manufacturer:    device.Manufacturer,
			ProductName:     device.ProductName,
			Present:         device.Present,
			Temperature:     device.temperature() + device.Temperature.Offset/10.0,

			OnOffDevice: struct {
				State      int
				Mode       int
				Lock       int
				DeviceLock int
			}{
				State:      device.switchState(),
				Mode:       device.switchMode(),
				Lock:       device.switchLock(),
				DeviceLock: device.switchDeviceLock(),
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
	switch {
	case (d.FunctionBitmask & switchingSocket) == 0:
		return Invalid
	case d.Switch.State == On, d.Switch.State == Off:
		return d.Switch.State
	}
	return Invalid
}

func (d *xmlDevice) switchMode() int {
	switch {
	case (d.FunctionBitmask & switchingSocket) == 0:
		return Invalid
	case d.Switch.Mode == "Auto":
		return Auto
	case d.Switch.Mode == "Manuell":
		return Manuel
	}
	return Invalid
}

func (d *xmlDevice) switchLock() int {
	return d.lock(d.Switch.Lock)
}

func (d *xmlDevice) switchDeviceLock() int {
	return d.lock(d.Switch.DeviceLock)
}

func (d *xmlDevice) lock(l int) int {
	switch {
	case (d.FunctionBitmask & switchingSocket) == 0:
		return Invalid
	case l == Lock, l == Unlock:
		return l
	}
	return Invalid
}

func (d *xmlDevice) id() int {
	i, err := strconv.Atoi(d.Id)
	if err != nil {
		return Invalid
	}

	return i
}

func (d *xmlDevice) present() int {
	if d.Present == Online {
		return Online
	}

	return Offline
}
