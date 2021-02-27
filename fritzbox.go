package fritzbox

// Session is the SessionInfo (SID and a list of all Devices) of Fritzbox AHA
type Session struct {
	host     string
	user     string
	password string
	sid      string
}

func New() *Session {
	return &Session{}
}

func (s *Session) String() string {
	return s.sid
}

func (s *Session) Connect(host, user, password string) (err error) {
	s.host = host
	s.user = user
	s.password = password
	return s.ReConnect()
}

func (s *Session) ReConnect() (err error) {
	s.sid, err = login(s.host, s.user, s.password)
	return
}

// Close send a logout request to the Fritzbox and delete the SID Token
func (s *Session) Close() error {
	err := logout(s.host, s.sid)
	s.sid = defaultSid
	return err
}
