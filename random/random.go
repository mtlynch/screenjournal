package random

import (
	cryptrand "crypto/rand"
	"log"
	"math/big"
)

func String(n int, characters []rune) string {
	b := make([]rune, n)
	max := big.NewInt(int64(len(characters)))
	for i := range b {
		idx, err := cryptrand.Int(cryptrand.Reader, max)
		if err != nil {
			log.Fatalf("failed to generate random string: %v", err)
		}
		b[i] = characters[idx.Int64()]
	}
	return string(b)
}

func Bytes(n int) []byte {
	b := make([]byte, n)
	if _, err := cryptrand.Read(b); err != nil {
		log.Fatalf("failed to generate random bytes: %v", err)
	}
	return b
}
