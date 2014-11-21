package objects

import (
	"bytes"
	"encoding/base64"
)

type AuthCredentials struct {
	RawInput []byte
}

func (a *AuthCredentials) Username() []byte {
	return []byte("xabber")
}

func (a *AuthCredentials) PasswordString() string {
	foo,_ := base64.StdEncoding.DecodeString(string(a.RawInput))
	value := bytes.Split(foo, []byte("\x00"))
	return string(value[2])
}

func (a *AuthCredentials) Password() []byte {
	foo,_ := base64.StdEncoding.DecodeString(string(a.RawInput))
	value := bytes.Split(foo, []byte("\x00"))
	return value[2]
}

