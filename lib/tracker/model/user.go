package model

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"strings"
)

// hashed login credential
type LoginCred string

func (cred LoginCred) String() string {
	return string(cred)
}

func (cred LoginCred) Salt() string {
	str := cred.String()
	return str[1+strings.Index(str, ":"):]
}

func (cred LoginCred) Check(secret string) bool {
	return HashCred(secret, cred.Salt()).String() == cred.String()
}

func HashCred(secret, salt string) LoginCred {
	d := sha256.Sum256([]byte(secret + salt))
	return LoginCred(base64.StdEncoding.EncodeToString(d[:]) + ":" + salt)
}

func GenSalt() string {
	var buff [12]byte
	io.ReadFull(rand.Reader, buff[:])
	return base64.StdEncoding.EncodeToString(buff[:])
}

type User struct {
	Username string
	Login    LoginCred
}

// create a new user with username, hash credential with a fresh salt
func NewUser(username, password string) *User {
	return &User{
		Username: username,
		Login:    HashCred(password, GenSalt()),
	}
}
