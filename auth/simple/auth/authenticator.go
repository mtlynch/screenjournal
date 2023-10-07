package auth

func (a authenticator) Authenticate(username, password string) error {
	h, err := a.store.ReadPasswordHash(username)
	if err != nil {
		return err
	}

	if ok := h.MatchesPlaintext(password); !ok {
		return ErrIncorrectPassword
	}

	return nil
}
