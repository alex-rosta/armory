package templates

import "testing"

func TestDisplay(t *testing.T) {
	for in, want := range map[string]string{
		"defias-brotherhood":  "Defias Brotherhood",
		"tempests":            "Tempests",
		"divine intervention": "Divine Intervention",
		"":                    "",
		"émile":               "Émile",
	} {
		if got := Display(in); got != want {
			t.Errorf("Display(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestClassSlug(t *testing.T) {
	for in, want := range map[string]string{"Death Knight": "death-knight", "Mage": "mage", "Demon Hunter": "demon-hunter", "": ""} {
		if got := ClassSlug(in); got != want {
			t.Errorf("ClassSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGold(t *testing.T) {
	for in, want := range map[float64]string{0: "0", 999: "999", 1000: "1,000", 281234.6: "281,235", 12345678: "12,345,678", -1234: "-1,234"} {
		if got := Gold(in); got != want {
			t.Errorf("Gold(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestNewManagerParsesAllTemplates(t *testing.T) {
	m, err := NewManager("../templates")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"header", "footer", "search_form", "character_page", "guild_page", "token_page", "recent_page", "error_page"} {
		if m.Get().Lookup(name) == nil {
			t.Errorf("template %q not defined", name)
		}
	}
}
