package handlers

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/mtlynch/screenjournal/v2/auth/simple"
	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type (
	simpleUser struct {
		IsAdmin_  bool                   `json:"isAdmin"`
		Username_ screenjournal.Username `json:"username"`
	}

	simpleUserSerializer struct{}
)

func NewSimpleUser(u screenjournal.User) simpleUser {
	return simpleUser{
		IsAdmin_:  u.IsAdmin,
		Username_: u.Username,
	}
}

func (u simpleUser) IsAdmin() bool {
	return u.IsAdmin_
}

func (u simpleUser) Username() string {
	return u.Username_.String()
}

func NewSimpleUserSerializer() simpleUserSerializer {
	return simpleUserSerializer{}
}

func (s simpleUserSerializer) Serialize(user simple.User) ([]byte, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(user); err != nil {
		log.Fatalf("failed to serialize user to JSON: %v", err)
	}
	return b.Bytes(), nil
}

func (s simpleUserSerializer) Deserialize(b []byte) (simple.User, error) {
	var su simpleUser
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&su); err != nil {
		return simpleUser{}, err
	}

	return su, nil
}
