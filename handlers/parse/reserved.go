package parse

import "slices"

var reservedWords = []string{"undefined", "null"}

func isReservedWord(s string) bool {
	return slices.Contains(reservedWords, s)
}
