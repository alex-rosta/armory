package interfaces

import (
	"html/template"
	"net/http"
)

// Handler is the interface for HTTP handlers
type Handler interface {
	// GetName returns the name of the handler
	GetName() string

	// RegisterRoutes registers the handler's routes with the router
	RegisterRoutes(router RouteRegistrar)
}

// TemplateProvider is the interface for handlers that provide templates
type TemplateProvider interface {
	// GetTemplates returns the templates used by the handler
	GetTemplates() *template.Template
}

// RouteRegistrar is the interface for registering routes
type RouteRegistrar interface {
	// HandleFunc registers a handler function for a given pattern
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
