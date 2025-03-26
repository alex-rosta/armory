package templates

import (
	"fmt"
	"html/template"
	"path/filepath"
)

// Manager handles the loading and access of HTML templates
type Manager struct {
	templates *template.Template
}

// NewManager creates a new template manager
func NewManager(templatesDir string) (*Manager, error) {
	templatesPath := filepath.Join(templatesDir, "*.html")
	tmpl, err := template.ParseGlob(templatesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Manager{templates: tmpl}, nil
}

// Get returns the loaded templates
func (m *Manager) Get() *template.Template {
	return m.templates
}
