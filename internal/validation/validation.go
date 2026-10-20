// Package validation checks user-supplied lookup parameters before they reach upstream APIs or storage.
package validation

import (
	"unicode"
	"unicode/utf8"
	"wowarmory/internal/api"
)

// MaxNameLength is the maximum length in characters of a realm, character or guild name
const MaxNameLength = 64

// Region reports whether region is a supported region code
func Region(region string) bool {
	return api.ValidRegion(region)
}

// Name reports whether s is a plausible realm, character or guild name:
// letters, digits, spaces, hyphens and apostrophes only.
func Name(s string) bool {
	if s == "" || utf8.RuneCountInString(s) > MaxNameLength {
		return false
	}
	for _, r := range s {
		switch {
		case unicode.IsLetter(r), unicode.IsMark(r), unicode.IsDigit(r):
		case r == ' ', r == '-', r == '\'':
		default:
			return false
		}
	}
	return true
}
