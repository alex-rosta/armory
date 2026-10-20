package router

import (
	"net/http"
	"strings"

	"wowarmory/internal/config"
	"wowarmory/internal/interfaces"
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
	r.mux.Handle("/assets/", http.StripPrefix("/assets/", staticHandler(r.config.AssetsDir)))

	// Register routes for each handler
	for _, handler := range handlers {
		handler.RegisterRoutes(r)
	}
}

// staticHandler serves files from dir without directory listings
func staticHandler(dir string) http.Handler {
	fileServer := http.FileServer(http.Dir(dir))
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasSuffix(req.URL.Path, "/") {
			http.NotFound(w, req)
			return
		}
		// Our own CSS/JS change with deploys and have no content hash in the URL, so always revalidate
		if strings.HasPrefix(req.URL.Path, "media/") || strings.HasPrefix(req.URL.Path, "vendor/") {
			w.Header().Set("Cache-Control", "public, max-age=86400")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fileServer.ServeHTTP(w, req)
	})
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
