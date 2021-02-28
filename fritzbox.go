package fritzbox

import "errors"

var (
	ErrDeviceNotFound            = errors.New("device not found")
	ErrSwitchCommandNotSupported = errors.New("device doesn't support switch commands")
	ErrTemperatureNotSupported   = errors.New("device doesn't support temperature")
	ErrInvalidSwitchState        = errors.New("invalid switch state")
	ErrUnknownAnswer             = errors.New("unknown answer of request")
	ErrLoginFailed               = errors.New("login failed")
)

const (
	On      = 1
	Off     = 0
	Auto    = 1
	Manuel  = 2
	Online  = 1
	Offline = 0
	Lock    = 1
	Unlock  = 0
	Invalid = -1
)

// Client is the SessionInfo (SID and a list of all Devices) of Fritzbox AHA
type Client struct {
	host     string
	user     string
	password string
	sid      string
}

func New() *Client {
	return &Client{}
}

func (c *Client) String() string {
	return c.sid
}

func (c *Client) Connect(host, user, password string) (err error) {
	c.host = host
	c.user = user
	c.password = password
	return c.ReConnect()
}

func (c *Client) ReConnect() (err error) {
	c.sid, err = login(c.host, c.user, c.password)
	return
}

// Close send a logout request to the Fritzbox and delete the SID Token
func (c *Client) Close() error {
	err := logout(c.host, c.sid)
	c.sid = defaultSid
	return err
}
