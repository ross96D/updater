//go:build !windows

package utils

import (
	"strings"
	"unicode/utf8"
)

const invalidChars = string('\x00')

func ValidPath(path string) bool {
	if !utf8.ValidString(path) {
		return false
	}

	for _, char := range invalidChars {
		if contains := strings.ContainsRune(path, char); contains {
			return false
		}
	}

	return true
}
