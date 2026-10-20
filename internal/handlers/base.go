package handlers

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/models"
	"wowarmory/internal/templates"
	"wowarmory/internal/validation"
)

// BaseHandler contains common handler functionality
type BaseHandler struct {
	config      *config.Config
	searchStore interfaces.SearchStore
	templates   *templates.Manager
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(cfg *config.Config, searchStore interfaces.SearchStore, templateMgr *templates.Manager) *BaseHandler {
	return &BaseHandler{
		config:      cfg,
		searchStore: searchStore,
		templates:   templateMgr,
	}
}

// Render renders a named template with the given status code.
// Output is buffered so a template failure never produces a partial page.
func (h *BaseHandler) Render(w http.ResponseWriter, status int, templateName string, data interface{}) {
	var buf bytes.Buffer
	if err := h.templates.Get().ExecuteTemplate(&buf, templateName, data); err != nil {
		log.Printf("error executing template %s: %v", templateName, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

// RenderErrorPage renders a friendly full-page error
func (h *BaseHandler) RenderErrorPage(w http.ResponseWriter, status int, title, message string) {
	h.Render(w, status, "error_page", map[string]interface{}{
		"PageTitle": title,
		"ActiveTab": "",
		"Title":     title,
		"Message":   message,
		"Status":    status,
	})
}

// RecordSearch records a search; failures are logged and never fail the request
func (h *BaseHandler) RecordSearch(r *http.Request, searchType, region, realm, name string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.searchStore.RecordSearch(ctx, searchType, region, realm, name); err != nil {
		log.Printf("error recording search: %v", err)
	}
}

// SearchForm holds the values rendered by the shared search form template
type SearchForm struct {
	Action          string
	NameParam       string
	NameLabel       string
	NamePlaceholder string
	SubmitLabel     string
	Region          string
	Realm           string
	Name            string
	Error           string
}

// Lookup holds normalized region, realm and entity name query parameters
type Lookup struct {
	Region string
	Realm  string
	Name   string
}

const invalidLookupMessage = "Check your search: choose EU or US, and use only letters, numbers, spaces, hyphens and apostrophes (up to 64 characters)."

// parseLookup reads and normalizes the lookup query parameters
func parseLookup(r *http.Request, nameParam string) Lookup {
	q := r.URL.Query()
	return Lookup{
		Region: strings.ToLower(strings.TrimSpace(q.Get("region"))),
		Realm:  normalizeRealm(q.Get("realm")),
		Name:   strings.Join(strings.Fields(strings.ToLower(q.Get(nameParam))), " "),
	}
}

// normalizeRealm converts a typed realm name to its API slug, e.g. "Quel'Thalas" -> "quelthalas", "Defias Brotherhood" -> "defias-brotherhood"
func normalizeRealm(s string) string {
	s = strings.NewReplacer("'", "", "\u2019", "").Replace(strings.ToLower(s))
	return strings.Join(strings.Fields(s), "-")
}

// Empty reports whether no realm and no name were given, i.e. the user has not searched yet
func (l Lookup) Empty() bool {
	return l.Realm == "" && l.Name == ""
}

// Valid reports whether the lookup is safe to send upstream and store
func (l Lookup) Valid() bool {
	return validation.Region(l.Region) && validation.Name(l.Realm) && validation.Name(l.Name)
}

// lookupConfig describes one kind of lookup page (character or guild)
type lookupConfig struct {
	Tab        string
	Template   string
	ResultKey  string
	Title      string
	SearchType interfaces.SearchType
	NameParam  string
	Form       SearchForm
	// ExternalURL points to a third-party page the user can check when nothing was found
	ExternalURL func(Lookup) string
	// Fetch loads the result and returns it together with a page title
	Fetch func(ctx context.Context, q Lookup) (result interface{}, title string, err error)
}

// serveLookup renders a search page, with results or a message below the form when a search was made
func (h *BaseHandler) serveLookup(w http.ResponseWriter, r *http.Request, cfg lookupConfig) {
	q := parseLookup(r, cfg.NameParam)

	form := cfg.Form
	form.Region, form.Realm, form.Name = q.Region, q.Realm, q.Name
	if !validation.Region(form.Region) {
		form.Region = "eu"
	}

	page := map[string]interface{}{
		"PageTitle": cfg.Title,
		"ActiveTab": cfg.Tab,
		"Form":      &form,
	}
	status := http.StatusOK

	switch {
	case q.Empty():
	case !q.Valid():
		form.Error = invalidLookupMessage
		status = http.StatusBadRequest
	default:
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()

		result, title, err := cfg.Fetch(ctx, q)
		switch {
		case err == nil:
			page[cfg.ResultKey] = result
			page["PageTitle"] = title
			h.RecordSearch(r, string(cfg.SearchType), q.Region, q.Realm, q.Name)
		case errors.Is(err, api.ErrNotFound), errors.Is(err, models.ErrGuildNotFound):
			page["NotFound"] = true
			page["ExternalURL"] = cfg.ExternalURL(q)
			status = http.StatusNotFound
		default:
			log.Printf("error loading %s: %v", cfg.SearchType, err)
			page["Unavailable"] = true
			status = http.StatusBadGateway
		}
	}

	h.Render(w, status, cfg.Template, page)
}
