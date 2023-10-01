package simple

type User interface {
	Username() string
	IsAdmin() bool
}
