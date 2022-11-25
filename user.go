package screenjournal

type (
	Username string

	UserAuth struct {
		Username Username
	}
)

func (u Username) String() string {
	return string(u)
}

func (ua UserAuth) IsEmpty() bool {
	return ua.Username == ""
}
