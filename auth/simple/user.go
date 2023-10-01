package simple

type User interface {
	Username() string
	IsAdmin() bool
	Serialize() ([]byte, error)
}

type UserDeserializer interface {
	Deserialize([]byte) (User, error)
}
