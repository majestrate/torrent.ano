package util

func IsSpace(r rune) bool {
	return r == '\n' || r == ' ' || r == '\t' || r == '\r'
}
