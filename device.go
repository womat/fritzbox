package fritzbox

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Device is the SessionInfo and DeviceInfos) of a Fritzbox AHA Devices
type Device struct {
	client          *Client
	name            string
	ain             string
	functionBitmask int
	id              int
	fwVersion       string
	manufacturer    string
	productName     string
}

// DeviceInfo is the DeviceInfo of a Fritzbox AHA Devices
type DeviceInfo struct {
	Name            string
	Identifier      string
	Id              int
	FunctionBitMask int
	FWVersion       string
	Manufacturer    string
	ProductName     string
	Present         int
	OnOffDevice     struct {
		State      int
		Mode       int
		Lock       int
		DeviceLock int
	}
	Powermeter struct {
		Power   float64
		Energy  float64
		Voltage float64
	}
	Temperature float64
}

const (
	hanFunDevice      = 1
	alarmSensor       = 1 << 4
	radiatorRegulator = 1 << 6
	energyGage        = 1 << 7
	temperatureSensor = 1 << 8
	switchingSocket   = 1 << 9
	aVMDECTRepeater   = 1 << 10
	microphone        = 1 << 11
	hanFunUnit        = 1 << 13

	switchcmdURL        = "http://%s/webservices/homeautoswitch.lua?ain=%s&sid=%s&switchcmd=%s"
	cmdSwitchOn         = "setswitchon"
	cmdSwitchOff        = "setswitchoff"
	cmdSwitchToggle     = "setswitchtoggle"
	cmdTemperature      = "gettemperature"
	cmdSwitchState      = "getswitchstate"
	cmdSwitchName       = "getswitchname"
	cmdSwitchEnergy     = "getswitchenergy"
	cmdSwitchPresent    = "getswitchpresent"
	cmdBasicDeviceStats = "getbasicdevicestats"
	cmdSwitchPower      = "getswitchpower"
)

// New create a new Fritzbox DeviceInfo Client of an AHA DeviceInfo
// Name kann der Gerätename (Name) oder die AIN sein, es wird nach beiden gesucht
func (c *Client) NewDevice(name string) (*Device, error) {
	device := &Device{client: c}

	devices, err := c.Devices()
	if err != nil {
		return device, err
	}

	for _, d := range devices {
		if d.Name == name || d.Identifier == name {
			device.name = d.Name
			device.ain = d.Identifier
			device.id = d.Id
			device.functionBitmask = d.FunctionBitMask
			device.fwVersion = d.FWVersion
			device.manufacturer = d.Manufacturer
			device.productName = d.ProductName
			return device, nil
		}
	}

	return device, ErrDeviceNotFound
}

func (d *Device) Info() (DeviceInfo, error) {
	devices, err := d.client.Devices()
	if err != nil {
		return DeviceInfo{}, err
	}

	for _, x := range devices {
		if x.Identifier == d.ain {
			return x, nil
		}
	}

	return DeviceInfo{}, ErrDeviceNotFound
}

// SwitchOn switches on a switch device
func (d *Device) SwitchOn() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchOn)
}

// SwitchOff switches off a switch device
func (d *Device) SwitchOff() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchOff)
}

// SwitchToggle switches on/off a switch device; on if it it was off; off it is was on
func (d *Device) SwitchToggle() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchToggle)
}

// SwitchState returns the state of a switch device
// "0" oder "1" (Steckdose aus oder an), "inval" if unkno unbekannt
func (d *Device) SwitchState() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchState)
}

// Temperature measures the temperature of a temperature device
func (d *Device) Temperature() (f float64, err error) {
	if (d.functionBitmask & temperatureSensor) == 0 {
		return f, ErrTemperatureNotSupported
	}

	file, err := getFile(d.ahaURL(cmdTemperature))
	if err != nil {
		return f, err
	}

	f, err = strconv.ParseFloat(strings.TrimSpace(string(file)), 64)
	return f / 10, err
}

// Name returns the name of the actor
func (d *Device) String() string {
	return d.name
}

func (d *Device) Ain() string {
	return d.ain
}

func (d *Device) FWVersion() string {
	return d.fwVersion
}

func (d *Device) Manufacturer() string {
	return d.manufacturer
}

func (d *Device) ProductName() string {
	return d.productName
}

// Online returns the online state (true, false) of a Fritzbox AHA DeviceInfo
func (d *Device) Present() (int, error) {
	file, err := getFile(d.ahaURL(cmdSwitchPresent))

	switch {
	case err != nil:
		return Offline, err
	case len(file) == 0:
		return Offline, ErrUnknownAnswer
	case string(file[0]) == "0":
		return Offline, nil
	case string(file[0]) == "1":
		return Online, nil
	}

	return Offline, ErrUnknownAnswer
}

// Power returns the current power of a Fritzbox AHA DeviceInfo
func (d *Device) Power() (p float64, err error) {
	file, err := getFile(d.ahaURL(cmdSwitchPower))
	if err != nil {
		return p, err
	}

	p, err = strconv.ParseFloat(string(file), 64)
	return p / 1000, err
}

// Power returns the total energy of a Fritzbox AHA DeviceInfo
func (d *Device) Energy() (e float64, err error) {
	file, err := getFile(d.ahaURL(cmdSwitchEnergy))
	if err != nil {
		return e, err
	}
	return strconv.ParseFloat(string(file), 64)
}

func (d *Device) ahaSwitchCmd(cmd string) (i int, err error) {
	if (d.functionBitmask & switchingSocket) == 0 {
		return Off, ErrSwitchCommandNotSupported
	}

	return getSwitchState(d.ahaURL(cmd))
}

func getSwitchState(url string) (int, error) {
	file, err := getFile(url)
	asRunes := []rune(string(file))

	switch {
	case err != nil:
		return Off, err
	case string(asRunes[:5]) == "inval":
		return Off, ErrInvalidSwitchState
	case string(asRunes[:1]) == "1":
		return On, nil
	case string(asRunes[:1]) == "0":
		return Off, nil
	}
	return Off, ErrUnknownAnswer
}

func (d *Device) ahaURL(cmd string) string {
	return fmt.Sprintf(switchcmdURL, d.client.host, url.QueryEscape(d.ain), d.client.sid, cmd)
}
