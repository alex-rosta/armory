package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestTokenSourceCachesToken(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if user, pass, ok := r.BasicAuth(); !ok || user != "id" || pass != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		fmt.Fprint(w, `{"access_token":"tok","expires_in":3600}`)
	}))
	defer srv.Close()

	s := newTokenSource("id", "secret")
	s.tokenURL = srv.URL

	for i := 0; i < 3; i++ {
		tok, err := s.Token()
		if err != nil || tok != "tok" {
			t.Fatalf("Token() = %q, %v", tok, err)
		}
	}
	if calls != 1 {
		t.Fatalf("token endpoint called %d times, want 1", calls)
	}
}

func TestTokenSourceDoesNotCacheShortLivedToken(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		fmt.Fprint(w, `{"access_token":"tok","expires_in":30}`)
	}))
	defer srv.Close()

	s := newTokenSource("id", "secret")
	s.tokenURL = srv.URL
	_, _ = s.Token()
	_, _ = s.Token()
	if calls != 2 {
		t.Fatalf("token endpoint called %d times, want 2", calls)
	}
}

func TestTokenSourceErrors(t *testing.T) {
	if _, err := newTokenSource("", "").Token(); err == nil {
		t.Fatal("expected error for missing credentials")
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad credentials", http.StatusUnauthorized)
	}))
	defer srv.Close()
	s := newTokenSource("id", "secret")
	s.tokenURL = srv.URL
	if _, err := s.Token(); err == nil {
		t.Fatal("expected error for 401")
	}
}

func newTestBlizzardClient(srv *httptest.Server) *BlizzardClient {
	c := NewBlizzardClient("id", "secret")
	// Region is substituted into the format; the test server ignores it
	c.apiURLFormat = strings.Replace(srv.URL, "%", "%%", -1) + "/%s"
	return c
}

func TestGetCharacterProfile(t *testing.T) {
	t.Run("rejects invalid region before any request", func(t *testing.T) {
		c := NewBlizzardClient("id", "secret")
		for _, region := range []string{"evil.example.com/x?", "EU", "xx"} {
			if _, err := c.GetCharacterProfile("tok", region, "darkspear", "tempests"); err == nil {
				t.Errorf("region %q: expected error", region)
			}
		}
	})

	t.Run("merges all endpoints", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer tok" {
				http.Error(w, "no auth", http.StatusUnauthorized)
				return
			}
			switch {
			case strings.HasSuffix(r.URL.Path, "/character-media"):
				fmt.Fprint(w, `{"assets":[{"key":"main-raw","value":"https://example/img.png"}]}`)
			case strings.HasSuffix(r.URL.Path, "/statistics"):
				fmt.Fprint(w, `{"health":100,"name":"from-stats"}`)
			default:
				fmt.Fprint(w, `{"name":"Tempests","level":80}`)
			}
		}))
		defer srv.Close()

		got, err := newTestBlizzardClient(srv).GetCharacterProfile("tok", "eu", "darkspear", "tempests")
		if err != nil {
			t.Fatal(err)
		}
		if got["name"] != "Tempests" || got["health"] != float64(100) || got["assets"] == nil {
			t.Fatalf("unexpected merge result: %v", got)
		}
	})

	t.Run("profile failure is an error even if other endpoints succeed", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/character-media") || strings.HasSuffix(r.URL.Path, "/statistics") {
				fmt.Fprint(w, `{"assets":[]}`)
				return
			}
			http.NotFound(w, r)
		}))
		defer srv.Close()

		if _, err := newTestBlizzardClient(srv).GetCharacterProfile("tok", "eu", "darkspear", "ghost"); err == nil {
			t.Fatal("expected error when profile endpoint fails")
		}
	})

	t.Run("optional endpoint failure is tolerated", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/character-media") || strings.HasSuffix(r.URL.Path, "/statistics") {
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			}
			fmt.Fprint(w, `{"name":"Tempests"}`)
		}))
		defer srv.Close()

		got, err := newTestBlizzardClient(srv).GetCharacterProfile("tok", "eu", "darkspear", "tempests")
		if err != nil || got["name"] != "Tempests" {
			t.Fatalf("got %v, %v", got, err)
		}
	})

	t.Run("realm and character are path-escaped", func(t *testing.T) {
		var gotPath string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.Contains(r.URL.Path, "/character-media") && !strings.Contains(r.URL.Path, "/statistics") {
				gotPath = r.URL.EscapedPath()
			}
			fmt.Fprint(w, `{}`)
		}))
		defer srv.Close()

		_, _ = newTestBlizzardClient(srv).GetCharacterProfile("tok", "eu", "a/../b", "x?y")
		if !strings.HasSuffix(gotPath, "/profile/wow/character/a%2F..%2Fb/x%3Fy") {
			t.Fatalf("path = %q", gotPath)
		}
	})
}

func TestGetTokenPrice(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/data/wow/token/index") {
			fmt.Fprint(w, `{"price":2500000000}`)
			return
		}
		fmt.Fprint(w, `{}`)
	}))
	defer srv.Close()

	c := NewTokenClient("id", "secret")
	c.apiURLFormat = srv.URL + "/%s"

	price, err := c.GetTokenPrice("tok", "eu")
	if err != nil || price != 2500000000 {
		t.Fatalf("price = %v, err = %v", price, err)
	}
	if _, err := c.GetTokenPrice("tok", "bad/region"); err == nil {
		t.Fatal("expected error for invalid region")
	}
}
