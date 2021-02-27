package fritzbox

import (
	"fmt"
	"testing"
	"time"
)

const (
	host     = "fritz.box"
	user     = "smarthome"
	password = "7Wl6UW5TsOr5Ba6uMbOO"
)

func TestConnect(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	fmt.Printf("%s\n", fb)

	if fmt.Sprintf("%s", fb) == "0000000000000000" {
		t.Errorf("invalid SID")
		return
	}

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}

	if err := fb.ReConnect(); err != nil {
		t.Errorf("cann't reconnect to fritzbox: %v", err)
		return
	}

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}

func TestDevices(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	d, err := fb.Devices()
	if err != nil {
		t.Errorf("cann't get devicelist: %v", err)
		return
	}

	fmt.Printf("%v\n", d)

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}

func TestWallbox(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	d, err := fb.NewDevice("wallbox")
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

	fmt.Printf("%s\n", d)

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}

func TestWaermepumpe(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	d, err := fb.NewDevice("wärmepumpe")
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

	detail, err := d.Info()
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

	fmt.Printf("%v\n", detail)

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}

func TestWaeschetrockner(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	d, err := fb.NewDevice("wäschetrockner")
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

	if s, err := d.SwitchOff(); err != nil || s != Off {
		t.Errorf("cann't switch off: %v", err)
		return
	}
	time.Sleep(200 * time.Millisecond)

	if s, err := d.SwitchState(); err != nil || s != Off {
		t.Errorf("cann't switch off: %v", err)
		return
	}

	if s, err := d.SwitchOn(); err != nil || s != On {
		t.Errorf("cann't switch on: %v", err)
		return
	}
	time.Sleep(200 * time.Millisecond)
	if s, err := d.SwitchState(); err != nil || s != On {
		t.Errorf("cann't switch on: %v", err)
		return
	}

	if s, err := d.SwitchOff(); err != nil || s != Off {
		t.Errorf("cann't switch off: %v", err)
		return
	}
	time.Sleep(200 * time.Millisecond)

	if s, err := d.SwitchState(); err != nil || s != Off {
		t.Errorf("cann't switch off: %v", err)
		return
	}

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}
