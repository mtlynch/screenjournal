package screenjournal

import (
	"time"

	"github.com/mtlynch/screenjournal/v2/random"
)

type (
	PasswordResetToken struct {
		value string
	}

	PasswordResetEntry struct {
		Username  Username
		Token     PasswordResetToken
		ExpiresAt time.Time
	}
)

const (
	PasswordResetTokenLength = 32
)

var (
	// PasswordResetTokenCharset contains the allowed characters for a password reset token.
	// Uses URL-safe characters to ensure the token works well in URLs.
	PasswordResetTokenCharset = []rune("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")
)

func (prt PasswordResetToken) String() string {
	return prt.value
}

func (prt PasswordResetToken) Empty() bool {
	return prt.String() == ""
}

func (prt PasswordResetToken) Equal(other PasswordResetToken) bool {
	return prt.String() == other.String()
}

func NewPasswordResetToken() PasswordResetToken {
	return PasswordResetToken{value: random.String(PasswordResetTokenLength, PasswordResetTokenCharset)}
}

func NewPasswordResetTokenFromString(token string) PasswordResetToken {
	return PasswordResetToken{value: token}
}

func (prr PasswordResetEntry) Empty() bool {
	return prr.Username == "" || prr.Token.Empty()
}

func (prr PasswordResetEntry) IsExpired() bool {
	return time.Now().After(prr.ExpiresAt)
}
