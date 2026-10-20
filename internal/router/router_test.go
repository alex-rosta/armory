package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"wowarmory/internal/config"
)

func TestStaticHandler(t *testing.T) {
	dir := t.TempDir()
	for path, body := range map[string]string{
		"css/styles.css":                   "body{}",
		"vendor/bootstrap-icons/icons.css": ".bi{}",
	} {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	r := New(&config.Config{AssetsDir: dir})
	r.SetupHandlers(nil)

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		return rec
	}

	if rec := get("/assets/css/styles.css"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-cache" ||
		rec.Header().Get("Content-Type") != "text/css; charset=utf-8" {
		t.Errorf("css: code=%d cache=%q type=%q", rec.Code, rec.Header().Get("Cache-Control"), rec.Header().Get("Content-Type"))
	}
	if rec := get("/assets/vendor/bootstrap-icons/icons.css"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "public, max-age=86400" {
		t.Errorf("vendor: code=%d cache=%q", rec.Code, rec.Header().Get("Cache-Control"))
	}

	if rec := get("/assets/css/styles.css?v=abc123"); rec.Code != 200 || rec.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" {
		t.Errorf("versioned: code=%d cache=%q", rec.Code, rec.Header().Get("Cache-Control"))
	}

	for _, path := range []string{"/assets/vendor/missing.css", "/assets/css/", "/assets/"} {
		rec := get(path)
		if rec.Code != 404 {
			t.Errorf("%s: code=%d, want 404", path, rec.Code)
		}
		if cc := rec.Header().Get("Cache-Control"); cc != "" {
			t.Errorf("%s: 404 must not carry Cache-Control, got %q", path, cc)
		}
	}
}
