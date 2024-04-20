//go:build windows

package share

import (
	"strings"
	"unicode/utf8"
)

const invalidChars = "<>:\"|?*\x00"

func ValidPath(path string) bool {
	if !utf8.ValidString(path) {
		return false
	}

	if len(path) > 2 && path[1] == ':' {
		firstLetter := path[0]
		if !((firstLetter >= 'a' && firstLetter <= 'z') || (firstLetter >= 'A' && firstLetter <= 'Z')) {
			return false
		}
		path = path[2:]
	}

	for _, char := range invalidChars {
		if contains := strings.ContainsRune(path, char); contains {
			return false
		}
	}

	return true
}
