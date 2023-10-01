package screenjournal

import (
	"bytes"
	"encoding/json"
	"log"
)

type (
	Email    string
	Username string
	Password string

	// EmailSubscriber represents a user or entity that subscribes to events via
	// email notifications.
	EmailSubscriber struct {
		Username Username
		Email    Email
	}

	PasswordHash []byte

	User struct {
		IsAdmin_     bool
		Username_    Username
		Email        Email
		PasswordHash PasswordHash
	}

	simpleUser struct {
		Username string `json:"username"`
		IsAdmin  bool   `json:"isAdmin"`
	}
)

func (e Email) String() string {
	return string(e)
}

func (u Username) String() string {
	return string(u)
}

func (u Username) Equal(o Username) bool {
	return u.String() == o.String()
}

func (pw Password) String() string {
	return string(pw)
}

// Equal returns true if two passwords match. Only use this in testing, never as
// a way of authenticating actual user passwords.
func (pw Password) Equal(o Password) bool {
	return pw.String() == o.String()
}

func (ph PasswordHash) Bytes() []byte {
	return []byte(ph)
}

func (u User) IsAdmin() bool {
	return u.IsAdmin_
}

func (u User) IsEmpty() bool {
	return u.Username == ""
}

func (u User) Serialize() ([]byte, error) {
	su := simpleUser{
		Username: u.Username.String(),
		IsAdmin:  u.IsAdmin_,
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(su); err != nil {
		log.Fatalf("failed to serialize user to JSON: %v", err)
	}
	return b.Bytes(), nil
}
