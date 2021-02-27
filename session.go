package fritzbox

import (
	"crypto/md5"
	"encoding/xml"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/unicode"
	"io"
	"net/http"
)

const (
	defaultSid       = "0000000000000000"
	loginURL         = "http://%s/login_sid.lua"
	logoutURL        = "http://%s/login_sid.lua?logout=1&sid=%s"
	loginResponseURL = "http://%s/login_sid.lua?user=%s&response=%s-%s"
)

var ErrLoginFailed = errors.New("login failed")

type xmlSessionInfo struct {
	XMLName   xml.Name `info:"Session"`
	SID       string   `info:"SID"`
	Challenge string   `info:"Challenge"`
	Rights    xmlRight `info:"Rights"`
	BlockTime int      `info:"BlockTime"`
}

type xmlRight struct {
	XMLName xml.Name `info:"Rights"`
	Name    []string `info:"Name"`
	Access  []string `info:"Access"`
}

func logout(host, sid string) error {
	var info xmlSessionInfo
	return getXMLStructure(fmt.Sprintf(logoutURL, host, sid), &info)
}

func login(host, username, password string) (string, error) {
	var id string
	var err error

	id, err = getChallenge(host)
	if err != nil {
		return defaultSid, err
	}

	hash := fmt.Sprintf("%x", md5.Sum(convert2utf16(id+"-"+password)))
	url := fmt.Sprintf(loginResponseURL, host, username, id, hash)

	var info xmlSessionInfo
	if err = getXMLStructure(url, &info); err != nil {
		return "", err
	}
	if info.SID == defaultSid || info.SID == "" {
		return defaultSid, ErrLoginFailed
	}

	return info.SID, err
}

func getChallenge(hostname string) (string, error) {
	url := fmt.Sprintf(loginURL, hostname)

	var info xmlSessionInfo
	err := getXMLStructure(url, &info)
	return info.Challenge, err
}

func getXMLStructure(url string, xmlStructure interface{}) error {
	xmlFile, err := getFile(url)
	if err != nil {
		return err
	}

	return xml.Unmarshal(xmlFile, &xmlStructure)
}

func getFile(url string) (content []byte, err error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() { _ = res.Body.Close() }()

	return io.ReadAll(res.Body)
}

func convert2utf16(s string) []byte {
	var utf16data = make([]byte, len(s)*2)

	for i := range s {
		utf16data[i*2] = s[i]
		utf16data[i*2+1] = 0
	}

	return utf16data
}

func utf8ToUTF16(s string) []byte {
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	utf8, _ := decoder.String(s)
	return []byte(utf8)
}
