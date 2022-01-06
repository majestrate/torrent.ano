package util

import (
	"fmt"
)

func IsSpace(r rune) bool {
	return r == '\n' || r == ' ' || r == '\t' || r == '\r'
}

var stringSizes = []string{
	"B",
	"KB",
	"MB",
	"GB",
}

func SizeString(sz uint64) string {
	unit := stringSizes[0]
	for _, str := range stringSizes {
		if sz > 1024 {
			sz /= 1024
		} else {
			unit = str
			break
		}
	}
	return fmt.Sprintf("%d %s", sz, unit)
}
