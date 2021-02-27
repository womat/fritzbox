package fritzbox

import (
	"github.com/womat/tools"
	"testing"
	"time"
)

const (
	host     = "fritz.box"
	user     = ""
	password = ""
)

func TestConnect(t *testing.T) {
	fb := New()
	if err := fb.Connect(host, user, password); err != nil {
		t.Errorf("cann't connect to fritzbox: %v", err)
		return
	}

	if fb.String() == "0000000000000000" {
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

	_, err := fb.Devices()
	if err != nil {
		t.Errorf("cann't get devicelist: %v", err)
		return
	}

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

	_, err := fb.NewDevice("wallbox")
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

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

	_, err = d.Info()
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

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

func TestOnline(t *testing.T) {
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

	i, err := d.Present()
	if err != nil {
		t.Errorf("cann't get State: %v", err)
		return
	}

	detail, err := d.Info()
	if err != nil {
		t.Errorf("cann't get device: %v", err)
		return
	}

	if i != detail.Present {
		t.Errorf("state1 and state2 aren't equal, t1: %v, t2: %v", i, detail.Present)
	}

	if err := fb.Close(); err != nil {
		t.Errorf("cann't disconnect to fritzbox: %v", err)
		return
	}
}

func TestUTF16(t *testing.T) {
	type pattern struct {
		utf8  string
		utf16 []byte
	}

	tests := []pattern{
		{utf8: "abcde€", utf16: []byte{0x61, 0x00, 0x62, 0x00, 0x63, 0x00, 0x64, 0x00, 0x65, 0x00, 0xac, 0x20}},
		{utf8: "abcde", utf16: []byte{0x61, 0x00, 0x62, 0x00, 0x63, 0x00, 0x64, 0x00, 0x65, 0x00}},
		{utf8: "Äß}gg", utf16: []byte{0xc4, 0x00, 0xdf, 0x00, 0x7d, 0x00, 0x67, 0x00, 0x67, 0x00}},
	}

	for _, test := range tests {
		if !tools.IsEqual(stringToUTF16(test.utf8), test.utf16) {
			t.Errorf("expected %x, got: %x", test.utf16, stringToUTF16(test.utf8))
		}
	}
}
