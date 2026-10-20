package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wowarmory/internal/api"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/templates"
)

type fakeBlizzard struct {
	profile  map[string]interface{}
	err      error
	gotRealm string
	gotName  string
}

func (f *fakeBlizzard) GetAccessToken() (string, error) { return "tok", nil }
func (f *fakeBlizzard) GetCharacterProfile(_, _, realm, name string) (map[string]interface{}, error) {
	f.gotRealm, f.gotName = realm, name
	return f.profile, f.err
}

type fakeWarcraftLogs struct {
	resp interface{}
	err  error
}

func (f fakeWarcraftLogs) GetGuild(context.Context, string, string, string) (interface{}, error) {
	return f.resp, f.err
}

type fakeTokens struct{}

func (fakeTokens) GetAccessToken() (string, error) { return "tok", nil }
func (fakeTokens) GetTokenPrice(_, region string) (float64, error) {
	if region == "eu" {
		return 2812345 * 10000, nil
	}
	return 250000 * 10000, nil
}

type fakeStore struct {
	recorded []string
	searches []interfaces.SearchEntry
	err      error
}

func (s *fakeStore) RecordSearch(_ context.Context, searchType, region, realm, name string) error {
	s.recorded = append(s.recorded, searchType+":"+region+":"+realm+":"+name)
	return nil
}
func (s *fakeStore) GetRecentSearches(context.Context) ([]interfaces.SearchEntry, error) {
	return s.searches, s.err
}
func (s *fakeStore) Close() error { return nil }

func newTestBase(t *testing.T, store *fakeStore) *BaseHandler {
	t.Helper()
	tm, err := templates.NewManager("../templates", "../../assets")
	if err != nil {
		t.Fatal(err)
	}
	return NewBaseHandler(nil, store, tm)
}

func get(h http.HandlerFunc, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func assertContains(t *testing.T, rec *httptest.ResponseRecorder, want ...string) {
	t.Helper()
	body := rec.Body.String()
	for _, w := range want {
		if !strings.Contains(body, w) {
			t.Errorf("body missing %q", w)
		}
	}
}

func assertNotContains(t *testing.T, rec *httptest.ResponseRecorder, unwanted ...string) {
	t.Helper()
	body := rec.Body.String()
	for _, u := range unwanted {
		if strings.Contains(body, u) {
			t.Errorf("body unexpectedly contains %q", u)
		}
	}
}

var sampleProfile = map[string]interface{}{
	"name":               "Tempests",
	"level":              float64(80),
	"average_item_level": float64(640),
	"achievement_points": float64(12345),
	"health":             float64(1200000),
	"realm":              map[string]interface{}{"name": "Darkspear"},
	"faction":            map[string]interface{}{"name": "Horde"},
	"character_class":    map[string]interface{}{"name": "Death Knight"},
	"active_spec":        map[string]interface{}{"name": "Blood"},
}

func TestCharacterPage(t *testing.T) {
	t.Run("empty form", func(t *testing.T) {
		store := &fakeStore{}
		h := NewCharacterHandler(newTestBase(t, store), &fakeBlizzard{})
		rec := get(h.LookupCharacter, "/")
		if rec.Code != 200 {
			t.Fatalf("code=%d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q", ct)
		}
		assertContains(t, rec, "Look up character", `name="character"`, `nav-link nav-link-active`, "/assets/css/styles.css?v=", "/assets/js/app.js?v=")
		assertNotContains(t, rec, "htmx", "unpkg", "cdn.jsdelivr")
		if len(store.recorded) != 0 {
			t.Fatal("empty form must not record a search")
		}
	})

	t.Run("unknown path", func(t *testing.T) {
		h := NewCharacterHandler(newTestBase(t, &fakeStore{}), &fakeBlizzard{})
		rec := get(h.LookupCharacter, "/nope")
		if rec.Code != 404 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "Page not found")
	})

	t.Run("invalid input keeps form values and shows message", func(t *testing.T) {
		fb := &fakeBlizzard{}
		h := NewCharacterHandler(newTestBase(t, &fakeStore{}), fb)
		for _, target := range []string{
			"/?region=evil.com&realm=x&character=y",
			"/?region=eu&realm=x%2F..&character=y",
			"/?region=eu&realm=x&character=" + strings.Repeat("a", 100),
			"/?region=eu&realm=x&character=",
			"/?region=kr&realm=x&character=y",
		} {
			rec := get(h.LookupCharacter, target)
			if rec.Code != 400 {
				t.Errorf("%s: code=%d", target, rec.Code)
			}
			assertContains(t, rec, "Check your search")
		}
		if fb.gotName != "" {
			t.Fatal("invalid input must not reach the API")
		}
	})

	t.Run("realm is normalized to a slug", func(t *testing.T) {
		for typed, want := range map[string]string{
			"Defias Brotherhood": "defias-brotherhood",
			"Quel'Thalas":        "quelthalas",
			"  Darkspear ":       "darkspear",
		} {
			fb := &fakeBlizzard{profile: sampleProfile}
			h := NewCharacterHandler(newTestBase(t, &fakeStore{}), fb)
			rec := get(h.LookupCharacter, "/?region=eu&character=Tempests&realm="+strings.ReplaceAll(typed, " ", "+"))
			if rec.Code != 200 || fb.gotRealm != want {
				t.Errorf("%q: code=%d realm=%q, want %q", typed, rec.Code, fb.gotRealm, want)
			}
		}
	})

	t.Run("success", func(t *testing.T) {
		store := &fakeStore{}
		h := NewCharacterHandler(newTestBase(t, store), &fakeBlizzard{profile: sampleProfile})
		rec := get(h.LookupCharacter, "/?region=eu&realm=darkspear&character=tempests")
		if rec.Code != 200 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "Tempests", "class-death-knight", "12,345", "1,200,000", "<title>Tempests · WoW Armory</title>",
			"warcraftlogs.com/character/eu/darkspear/tempests")
		if len(store.recorded) != 1 || store.recorded[0] != "character:eu:darkspear:tempests" {
			t.Fatalf("recorded = %v", store.recorded)
		}
	})

	t.Run("upstream values are escaped", func(t *testing.T) {
		profile := map[string]interface{}{"name": "<script>alert(1)</script>", "character_class": map[string]interface{}{"name": "Mage"}}
		h := NewCharacterHandler(newTestBase(t, &fakeStore{}), &fakeBlizzard{profile: profile})
		rec := get(h.LookupCharacter, "/?region=eu&realm=darkspear&character=tempests")
		assertNotContains(t, rec, "<script>alert(1)")
	})

	t.Run("not found", func(t *testing.T) {
		store := &fakeStore{}
		h := NewCharacterHandler(newTestBase(t, store), &fakeBlizzard{err: api.ErrNotFound})
		rec := get(h.LookupCharacter, "/?region=eu&realm=darkspear&character=ghost")
		if rec.Code != 404 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "couldn't find a match", "worldofwarcraft.blizzard.com/en-gb/character/eu/darkspear/ghost", `value="ghost"`)
		if len(store.recorded) != 0 {
			t.Fatal("failed lookups must not be recorded")
		}
	})

	t.Run("upstream outage is not reported as not found", func(t *testing.T) {
		h := NewCharacterHandler(newTestBase(t, &fakeStore{}), &fakeBlizzard{err: errors.New("status code: 503")})
		rec := get(h.LookupCharacter, "/?region=eu&realm=darkspear&character=tempests")
		if rec.Code != 502 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "isn't responding")
		assertNotContains(t, rec, "503")
	})
}

func TestGuildPage(t *testing.T) {
	found := &api.GuildResponse{}
	found.GuildData.Guild.Name = "Divine Intervention"
	found.GuildData.Guild.Members.Total = 120
	found.GuildData.Guild.ZoneRanking.Progress.WorldRank = api.RankInfo{Number: 1234, Color: "epic"}
	found.GuildData.Guild.Attendance.Data = []api.AttendanceEntry{{Players: []api.GuildMember{{Name: "Amy", Type: "Priest"}}}}

	t.Run("form", func(t *testing.T) {
		h := NewGuildHandler(newTestBase(t, &fakeStore{}), fakeWarcraftLogs{})
		rec := get(h.LookupGuild, "/guild-lookup")
		if rec.Code != 200 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "Look up guild", `name="guild"`)
	})

	t.Run("success", func(t *testing.T) {
		store := &fakeStore{}
		h := NewGuildHandler(newTestBase(t, store), fakeWarcraftLogs{resp: found})
		rec := get(h.LookupGuild, "/guild-lookup?region=eu&realm=Defias+Brotherhood&guild=Divine+Intervention")
		if rec.Code != 200 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "Divine Intervention", "#1,234", "rank-epic", "Amy", "Defias Brotherhood")
		if len(store.recorded) != 1 || store.recorded[0] != "guild:eu:defias-brotherhood:divine intervention" {
			t.Fatalf("recorded = %v", store.recorded)
		}
	})

	t.Run("unknown guild is a not-found page", func(t *testing.T) {
		h := NewGuildHandler(newTestBase(t, &fakeStore{}), fakeWarcraftLogs{resp: &api.GuildResponse{}})
		rec := get(h.LookupGuild, "/guild-lookup?region=eu&realm=darkspear&guild=ghosts")
		if rec.Code != 404 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "couldn't find a match", "warcraftlogs.com/guild/eu/darkspear/ghosts")
	})

	t.Run("upstream error", func(t *testing.T) {
		h := NewGuildHandler(newTestBase(t, &fakeStore{}), fakeWarcraftLogs{err: errors.New("boom")})
		if rec := get(h.LookupGuild, "/guild-lookup?region=eu&realm=darkspear&guild=ghosts"); rec.Code != 502 {
			t.Fatalf("code=%d", rec.Code)
		}
	})

	t.Run("invalid guild", func(t *testing.T) {
		h := NewGuildHandler(newTestBase(t, &fakeStore{}), fakeWarcraftLogs{})
		if rec := get(h.LookupGuild, "/guild-lookup?region=eu&realm=darkspear&guild=%3Cscript%3E"); rec.Code != 400 {
			t.Fatalf("code=%d", rec.Code)
		}
	})
}

func TestRecentSearchesPage(t *testing.T) {
	t.Run("lists searches with working links", func(t *testing.T) {
		store := &fakeStore{searches: []interfaces.SearchEntry{
			{Type: "character", Name: "tempests", Realm: "defias-brotherhood", Region: "eu", Timestamp: "20 Oct 26 14:00 CEST"},
			{Type: "guild", Name: "divine intervention", Realm: "darkspear", Region: "eu", Timestamp: "20 Oct 26 13:00 CEST"},
		}}
		h := NewRecentSearchesHandler(newTestBase(t, store))
		rec := get(h.GetRecentSearchesPage, "/recent-searches")
		if rec.Code != 200 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertContains(t, rec, "Tempests", "Defias Brotherhood", "Divine Intervention",
			`href="/?region=eu&amp;realm=defias-brotherhood&amp;character=tempests"`,
			`href="/guild-lookup?region=eu&amp;realm=darkspear&amp;guild=divine%20intervention"`,
			`nav-link nav-link-active" aria-current="page"`)
	})

	t.Run("empty", func(t *testing.T) {
		h := NewRecentSearchesHandler(newTestBase(t, &fakeStore{}))
		assertContains(t, get(h.GetRecentSearchesPage, "/recent-searches"), "Nothing here yet")
	})

	t.Run("store failure", func(t *testing.T) {
		h := NewRecentSearchesHandler(newTestBase(t, &fakeStore{err: errors.New("redis down")}))
		rec := get(h.GetRecentSearchesPage, "/recent-searches")
		if rec.Code != 503 {
			t.Fatalf("code=%d", rec.Code)
		}
		assertNotContains(t, rec, "redis down")
	})
}

func TestNormalizeRealm(t *testing.T) {
	for in, want := range map[string]string{
		"Defias Brotherhood":  "defias-brotherhood",
		"Quel'Thalas":         "quelthalas",
		"Cho’gall":            "chogall",
		"Pozzo dell'Eternità": "pozzo-delleternita",
		"Marécage de Zangar":  "marecage-de-zangar",
		"  Azjol-Nerub  ":     "azjol-nerub",
		"Los    Errantes":     "los-errantes",
		"":                    "",
	} {
		if got := normalizeRealm(in); got != want {
			t.Errorf("normalizeRealm(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTokenPage(t *testing.T) {
	h := NewTokenHandler(newTestBase(t, &fakeStore{}), fakeTokens{})
	rec := get(h.GetTokenPrice, "/token")
	if rec.Code != 200 {
		t.Fatalf("code=%d", rec.Code)
	}
	assertContains(t, rec, "2,812,345", "250,000")
}
