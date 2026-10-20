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
		// StripPrefix turns "/assets/" into "", which FileServer would serve as a directory listing
		if req.URL.Path == "" || strings.HasSuffix(req.URL.Path, "/") {
			http.NotFound(w, req)
			return
		}
		// Our own CSS/JS change with deploys and have no content hash in the URL, so always revalidate
		cacheControl := "no-cache"
		if strings.HasPrefix(req.URL.Path, "media/") || strings.HasPrefix(req.URL.Path, "vendor/") {
			cacheControl = "public, max-age=86400"
		}
		fileServer.ServeHTTP(&cacheOnSuccess{ResponseWriter: w, value: cacheControl}, req)
	})
}

// cacheOnSuccess sets Cache-Control only on non-error responses, so a 404 is never cached
type cacheOnSuccess struct {
	http.ResponseWriter
	value string
}

func (c *cacheOnSuccess) WriteHeader(status int) {
	if status < 400 {
		c.Header().Set("Cache-Control", c.value)
	}
	c.ResponseWriter.WriteHeader(status)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
