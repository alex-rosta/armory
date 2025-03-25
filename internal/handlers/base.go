package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"time"
	"wowarmory/internal/config"
	"wowarmory/internal/redis"
	"wowarmory/internal/templates"
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

// GetTemplates returns the template manager's templates
func (h *BaseHandler) GetTemplates() *template.Template {
	return h.templates.Get()
}

// RenderTemplate renders a template with the given data
func (h *BaseHandler) RenderTemplate(w http.ResponseWriter, templateName string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html")
	return h.templates.Get().ExecuteTemplate(w, templateName, data)
}

// RenderWithLayout renders a content template within the master layout
func (h *BaseHandler) RenderWithLayout(w http.ResponseWriter, contentTemplate string, data map[string]interface{}) error {
	// Add the content template to the data
	data["ContentTemplate"] = contentTemplate
	return h.RenderTemplate(w, "master_layout.html", data)
}

// RenderError renders the error template
func (h *BaseHandler) RenderError(w http.ResponseWriter, activeTab, url string) error {
	layoutData := map[string]interface{}{
		"PageTitle": "Error",
		"ActiveTab": activeTab,
		"url":       url,
	}
	w.WriteHeader(http.StatusNotFound)
	return h.RenderWithLayout(w, "error", layoutData)
}

// RecordSearch records a search in Redis
func (h *BaseHandler) RecordSearch(r *http.Request, region, realm, name string) error {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.redisClient.RecordSearch(ctx, region, realm, name); err != nil {
		// Log the error but don't fail the request
		fmt.Printf("Error recording search: %v\n", err)
	}
	return nil
}
