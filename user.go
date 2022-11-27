package screenjournal

import "crypto/subtle"

type (
	Username string
	Password string

	UserAuth struct {
		IsAdmin  bool
		Username Username
	}
)

func (u Username) String() string {
	return string(u)
}

func (u Username) IsEmpty() bool {
	return u.String() == ""
}

func (u Username) Equal(o Username) bool {
	return u.String() == o.String()
}

func (pw Password) String() string {
	return string(pw)
}

func (pw Password) Equal(o Password) bool {
	return subtle.ConstantTimeCompare([]byte(pw.String()), []byte(o.String())) == 1
}

func (ua UserAuth) IsEmpty() bool {
	return ua.Username == ""
}
