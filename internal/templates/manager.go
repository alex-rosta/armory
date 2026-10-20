package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
)

// Manager handles the loading and access of HTML templates
type Manager struct {
	templates *template.Template
	assetsDir string
	assetURLs sync.Map
}

// NewManager creates a new template manager; assetsDir is used to version asset URLs
func NewManager(templatesDir, assetsDir string) (*Manager, error) {
	m := &Manager{assetsDir: assetsDir}

	templatesPath := filepath.Join(templatesDir, "*.html")
	tmpl, err := template.New("").Funcs(funcs).Funcs(template.FuncMap{"asset": m.assetURL}).ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	m.templates = tmpl
	return m, nil
}

// assetURL returns the public URL of an asset with a content hash appended.
// Edge caches and browsers key on the full URL, so a changed file is never served stale.
func (m *Manager) assetURL(path string) string {
	if cached, ok := m.assetURLs.Load(path); ok {
		return cached.(string)
	}

	url := "/assets/" + path
	if data, err := os.ReadFile(filepath.Join(m.assetsDir, filepath.FromSlash(path))); err == nil {
		sum := sha256.Sum256(data)
		url += "?v=" + hex.EncodeToString(sum[:])[:10]
	}
	m.assetURLs.Store(path, url)
	return url
}

// Get returns the loaded templates
func (m *Manager) Get() *template.Template {
	return m.templates
}

var funcs = template.FuncMap{
	"display":   Display,
	"classSlug": ClassSlug,
	"gold":      Gold,
	"num":       func(n int) string { return Gold(float64(n)) },
	"upper":     strings.ToUpper,
}

// Display turns a stored lowercase slug such as "defias-brotherhood" into "Defias Brotherhood"
func Display(s string) string {
	words := strings.Fields(strings.ReplaceAll(s, "-", " "))
	for i, w := range words {
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		words[i] = string(r)
	}
	return strings.Join(words, " ")
}

// ClassSlug turns a class name such as "Death Knight" into a CSS-friendly "death-knight"
func ClassSlug(s string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(s)), " ", "-")
}

// Gold formats a gold amount with thousands separators
func Gold(amount float64) string {
	digits := fmt.Sprintf("%.0f", amount)
	var b strings.Builder
	for i, r := range digits {
		if i > 0 && (len(digits)-i)%3 == 0 && digits[i-1] != '-' {
			b.WriteByte(',')
		}
		b.WriteRune(r)
	}
	return b.String()
}
