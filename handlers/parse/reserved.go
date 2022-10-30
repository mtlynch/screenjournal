package parse

var reservedWords = []string{"undefined", "null", "root", "admin", "add", "edit", "delete", "copy"}

func isReservedWord(s string) bool {
	for _, w := range reservedWords {
		if s == w {
			return true
		}
	}
	return false
}
