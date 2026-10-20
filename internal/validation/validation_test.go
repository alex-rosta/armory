package validation

import (
	"strings"
	"testing"
)

func TestName(t *testing.T) {
	valid := []string{"tempests", "defias-brotherhood", "divine intervention", "azjol'nerub", "Ångström", "예시"}
	for _, s := range valid {
		if !Name(s) {
			t.Errorf("Name(%q) = false, want true", s)
		}
	}

	invalid := []string{
		"", "a/b", "../etc", "a?b=c", "a&b", "<script>", "name\nnext", "%2e%2e", "a;b",
		strings.Repeat("a", MaxNameLength+1),
	}
	for _, s := range invalid {
		if Name(s) {
			t.Errorf("Name(%q) = true, want false", s)
		}
	}
}

func TestRegion(t *testing.T) {
	for _, r := range []string{"us", "eu"} {
		if !Region(r) {
			t.Errorf("Region(%q) = false, want true", r)
		}
	}
	for _, r := range []string{"", "EU", "evil.example.com/", "eu.", "xx", "kr", "tw"} {
		if Region(r) {
			t.Errorf("Region(%q) = true, want false", r)
		}
	}
}
