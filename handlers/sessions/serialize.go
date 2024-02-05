package sessions

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/mtlynch/screenjournal/v2/screenjournal"
)

type serializableSession struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
}

func serializeSession(sess Session) []byte {
	ss := serializableSession{
		Username: sess.Username.String(),
		IsAdmin:  sess.IsAdmin,
	}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(ss); err != nil {
		log.Fatalf("failed to serialize session to JSON: %v", err)
	}
	return b.Bytes()
}

func deserializeSession(b []byte) (Session, error) {
	var ss serializableSession
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&ss); err != nil {
		return Session{}, err
	}

	return Session{
		Username: screenjournal.Username(ss.Username),
		IsAdmin:  ss.IsAdmin,
	}, nil
}
