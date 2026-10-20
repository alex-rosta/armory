package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"wowarmory/internal/templates"
)

type fakeBlizzard struct{ profileErr error }

func (f fakeBlizzard) GetAccessToken() (string, error) { return "tok", nil }
func (f fakeBlizzard) GetCharacterProfile(_, _, _, _ string) (map[string]interface{}, error) {
	return nil, f.profileErr
}

type fakeWarcraftLogs struct{}

func (fakeWarcraftLogs) GetGuild(context.Context, string, string, string) (interface{}, error) {
	return nil, errors.New("upstream down")
}

func newTestBase(t *testing.T) *BaseHandler {
	t.Helper()
	tm, err := templates.NewManager("../templates")
	if err != nil {
		t.Fatal(err)
	}
	// Redis is never reached: these tests only cover failure and validation paths
	return NewBaseHandler(nil, nil, tm)
}

func get(h http.HandlerFunc, target string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestCharacterHandler(t *testing.T) {
	h := NewCharacterHandler(newTestBase(t), fakeBlizzard{profileErr: errors.New("not found")})

	t.Run("form", func(t *testing.T) {
		rec := get(h.LookupCharacter, "/")
		if rec.Code != 200 || !strings.Contains(rec.Body.String(), "Search for your character") {
			t.Fatalf("code=%d", rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q", ct)
		}
	})

	t.Run("unknown path is 404", func(t *testing.T) {
		if rec := get(h.LookupCharacter, "/does-not-exist"); rec.Code != 404 {
			t.Fatalf("code=%d", rec.Code)
		}
	})

	t.Run("invalid parameters are 400", func(t *testing.T) {
		for _, target := range []string{
			"/?region=evil.com&realm=x&character=y",
			"/?region=eu&realm=x%2F..&character=y",
			"/?region=eu&realm=x&character=" + strings.Repeat("a", 100),
		} {
			if rec := get(h.LookupCharacter, target); rec.Code != 400 {
				t.Errorf("%s: code=%d", target, rec.Code)
			}
		}
		if rec := get(h.GetCharacterTemplate, "/character?region=eu&realm=x"); rec.Code != 400 {
			t.Errorf("missing character: code=%d", rec.Code)
		}
	})

	t.Run("upstream failure renders themed 404 page", func(t *testing.T) {
		rec := get(h.LookupCharacter, "/?region=eu&realm=darkspear&character=ghost")
		body := rec.Body.String()
		if rec.Code != 404 || !strings.Contains(body, "Not Found") || !strings.Contains(body, "<html") {
			t.Fatalf("code=%d body=%q", rec.Code, body)
		}
		if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
			t.Fatalf("Content-Type = %q", ct)
		}
	})

	t.Run("htmx fragment failure", func(t *testing.T) {
		rec := get(h.GetCharacterTemplate, "/character?region=eu&realm=darkspear&character=ghost")
		if rec.Code != 404 || strings.Contains(rec.Body.String(), "<html") {
			t.Fatalf("code=%d", rec.Code)
		}
	})
}

func TestGuildHandler(t *testing.T) {
	h := NewGuildHandler(newTestBase(t), fakeWarcraftLogs{})

	if rec := get(h.LookupGuild, "/guild-lookup"); rec.Code != 200 {
		t.Fatalf("form code=%d", rec.Code)
	}
	if rec := get(h.LookupGuild, "/guild-lookup?region=eu&realm=darkspear&guild=%3Cscript%3E"); rec.Code != 400 {
		t.Fatalf("invalid guild code=%d", rec.Code)
	}
	if rec := get(h.GetGuildTemplate, "/guild?region=eu&realm=darkspear&guild=ghosts"); rec.Code != 404 {
		t.Fatalf("upstream failure code=%d", rec.Code)
	}
}

func TestRecentTabHighlight(t *testing.T) {
	h := NewRecentSearchesHandler(newTestBase(t))
	rec := get(h.GetRecentSearchesPage, "/recent-searches")
	if !strings.Contains(rec.Body.String(), `nav-link-wow-active" href="/recent-searches"`) {
		t.Fatal("Recent tab is not highlighted as active")
	}
}
