package fritzbox

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// DeviceSession is the SessionInfo and DeviceInfos) of a Fritzbox AHA Devices
type DeviceSession struct {
	session         *Session
	name            string
	ain             string
	functionBitmask int
	id              string
	fwVersion       string
	manufacturer    string
	productName     string
}

// Device is the DeviceInfo of a Fritzbox AHA Devices
type Device struct {
	Name            string
	Identifier      string
	Id              string
	FunctionBitMask int
	FWVersion       string
	Manufacturer    string
	ProductName     string
	Present         bool
	OnOffDevice     struct {
		State      int
		Mode       string
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

var (
	ErrDeviceNotFound            = errors.New("device not found")
	ErrSwitchCommandNotSupported = errors.New("device doesn't support switch commands")
	ErrInvalidSwitchState        = errors.New("invalid switch state")
	ErrUnknownAnswer             = errors.New("unknown answer of url request")
	ErrTemperatureNotSupported   = errors.New("device doesn't support temperature")
)

const (
	On      = 1
	Off     = 0
	Auto    = "auto"
	Manuel  = "manuell"
	Online  = 1
	Offline = 0
	Lock    = 1
	Unlock  = 0
)

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
)

const (
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

// New create a new Fritzbox Device Session of an AHA Device
// Name kann der Gerätename (Name) oder die AIN sein, es wird nach beiden gesucht
func (s *Session) NewDevice(name string) (*DeviceSession, error) {
	device := &DeviceSession{session: s}

	devices, err := s.Devices()
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

func (d *DeviceSession) Info() (Device, error) {
	devices, err := d.session.Devices()
	if err != nil {
		return Device{}, err
	}

	for _, x := range devices {
		if x.Identifier == d.ain {
			return x, nil
		}
	}

	return Device{}, ErrDeviceNotFound
}

// SwitchOn switches on a switch device
func (d *DeviceSession) SwitchOn() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchOn)
}

// SwitchOff switches off a switch device
func (d *DeviceSession) SwitchOff() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchOff)
}

// SwitchToggle switches on/off a switch device; on if it it was off; off it is was on
func (d *DeviceSession) SwitchToggle() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchToggle)
}

// SwitchState returns the state of a switch device
// "0" oder "1" (Steckdose aus oder an), "inval" if unkno unbekannt
func (d *DeviceSession) SwitchState() (int, error) {
	return d.ahaSwitchCmd(cmdSwitchState)
}

// Temperature measures the temperature of a temperature device
func (d *DeviceSession) Temperature() (f float64, err error) {
	if (d.functionBitmask & temperatureSensor) == 0 {
		return f, ErrTemperatureNotSupported
	}

	file, err := getFile(d.ahaURL(cmdTemperature))
	if err != nil {
		return f, err
	}

	f, err = strconv.ParseFloat(string(file), 64)
	return f / 10, err
}

// Name returns the name of the actor
func (d *DeviceSession) String() string {
	return d.name
}

func (d *DeviceSession) Ain() string {
	return d.ain
}

func (d *DeviceSession) FWVersion() string {
	return d.fwVersion
}

func (d *DeviceSession) Manufacturer() string {
	return d.manufacturer
}

func (d *DeviceSession) ProductName() string {
	return d.productName
}

// Online returns the online state (true, false) of a Fritzbox AHA Device
func (d *DeviceSession) State() (int, error) {
	file, err := getFile(d.ahaURL(cmdSwitchPresent))

	switch {
	case err != nil:
		return Offline, err
	case string(file) == "0":
		return Offline, nil
	case string(file) == "1":
		return Online, nil
	}

	return Offline, ErrUnknownAnswer
}

// Power returns the current power of a Fritzbox AHA Device
func (d *DeviceSession) Power() (p float64, err error) {
	file, err := getFile(d.ahaURL(cmdSwitchPower))
	if err != nil {
		return p, err
	}

	p, err = strconv.ParseFloat(string(file), 64)
	return p / 1000, err
}

// Power returns the total energy of a Fritzbox AHA Device
func (d *DeviceSession) Energy() (e float64, err error) {
	file, err := getFile(d.ahaURL(cmdSwitchEnergy))
	if err != nil {
		return e, err
	}
	return strconv.ParseFloat(string(file), 64)
}

func (d *DeviceSession) ahaSwitchCmd(cmd string) (i int, err error) {
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

func (d *DeviceSession) ahaURL(cmd string) string {
	return fmt.Sprintf(switchcmdURL, d.session.host, url.QueryEscape(d.ain), d.session.sid, cmd)
}
