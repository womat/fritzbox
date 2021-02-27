package fritzbox

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
	"unicode/utf16"
)

const (
	defaultSid       = "0000000000000000"
	loginURL         = "http://%s/login_sid.lua"
	logoutURL        = "http://%s/login_sid.lua?logout=1&sid=%s"
	loginResponseURL = "http://%s/login_sid.lua?user=%s&response=%s-%s"

	timeout = 2 * time.Second
)

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

	hash := fmt.Sprintf("%x", md5.Sum(stringToUTF16(id+"-"+password)))
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
	client := http.Client{Timeout: timeout}

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() { _ = res.Body.Close() }()

	return io.ReadAll(res.Body)
}

func stringToUTF16(s string) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, utf16.Encode([]rune(s)))
	return buf.Bytes()
}
