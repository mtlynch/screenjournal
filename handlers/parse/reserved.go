package parse

var reservedWords = []string{"undefined", "null"}

func isReservedWord(s string) bool {
	return isWordInSlice(s, reservedWords)
}

func isWordInSlice(s string, ss []string) bool {
	for _, w := range ss {
		if s == w {
			return true
		}
	}
	return false
}
