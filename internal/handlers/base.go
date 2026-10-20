package handlers

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"strings"
	"time"
	"wowarmory/internal/config"
	"wowarmory/internal/redis"
	"wowarmory/internal/templates"
	"wowarmory/internal/validation"
)

// BaseHandler contains common handler functionality
type BaseHandler struct {
	config      *config.Config
	redisClient *redis.Client
	templates   *templates.Manager
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(cfg *config.Config, redisClient *redis.Client, templateMgr *templates.Manager) *BaseHandler {
	return &BaseHandler{
		config:      cfg,
		redisClient: redisClient,
		templates:   templateMgr,
	}
}

// RenderTemplateStatus renders a template with the given status code.
// Output is buffered so a template failure never produces a partial page.
func (h *BaseHandler) RenderTemplateStatus(w http.ResponseWriter, status int, templateName string, data interface{}) error {
	var buf bytes.Buffer
	if err := h.templates.Get().ExecuteTemplate(&buf, templateName, data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}

// RenderTemplate renders a template with the given data
func (h *BaseHandler) RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	return h.RenderTemplateStatus(w, http.StatusOK, templateName, data)
}

// RenderWithLayout renders a content template within the master layout
func (h *BaseHandler) RenderWithLayout(w http.ResponseWriter, contentTemplate string, data map[string]interface{}) error {
	data["ContentTemplate"] = contentTemplate
	return h.RenderTemplate(w, "master_layout.html", data)
}

// RenderError renders the not-found error page within the master layout
func (h *BaseHandler) RenderError(w http.ResponseWriter, activeTab, url string) error {
	layoutData := map[string]interface{}{
		"PageTitle":       "Error",
		"ActiveTab":       activeTab,
		"url":             url,
		"ContentTemplate": "error",
	}
	return h.RenderTemplateStatus(w, http.StatusNotFound, "master_layout.html", layoutData)
}

// RecordSearch records a search in Redis; failures are logged and never fail the request
func (h *BaseHandler) RecordSearch(r *http.Request, searchType string, region, realm, name string) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.redisClient.RecordSearch(ctx, searchType, region, realm, name); err != nil {
		log.Printf("error recording search: %v", err)
	}
}

// lookupParams reads the lowercased region, realm and entity name query parameters
func lookupParams(r *http.Request, nameParam string) (region, realm, name string) {
	q := r.URL.Query()
	return strings.ToLower(q.Get("region")), strings.ToLower(q.Get("realm")), strings.ToLower(q.Get(nameParam))
}

// validLookup reports whether the lookup parameters are safe to send upstream and store
func validLookup(region, realm, name string) bool {
	return validation.Region(region) && validation.Name(realm) && validation.Name(name)
}
