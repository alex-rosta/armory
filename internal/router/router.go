package router

import (
	"net/http"

	"wowarmory/internal/config"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/middleware"
)

// Router handles HTTP routing for the application
type Router struct {
	config *config.Config
	mux    *http.ServeMux
}

// Ensure Router implements RouteRegistrar interface
var _ interfaces.RouteRegistrar = (*Router)(nil)

// New creates a new router
func New(cfg *config.Config) *Router {
	return &Router{
		config: cfg,
		mux:    http.NewServeMux(),
	}
}

// HandleFunc implements the RouteRegistrar interface
func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	r.mux.HandleFunc(pattern, handler)
}

// SetupHandlers registers routes for all handlers
func (r *Router) SetupHandlers(handlers []interfaces.Handler) {
	// Set up static file server
	fileServer := http.FileServer(http.Dir(r.config.AssetsDir))
	r.mux.Handle("/assets/", http.StripPrefix("/assets/", middleware.ContentTypeMiddleware(fileServer)))

	// Register routes for each handler
	for _, handler := range handlers {
		handler.RegisterRoutes(r)
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
