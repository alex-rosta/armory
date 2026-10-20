package interfaces

import (
	"net/http"
)

// Handler is the interface for HTTP handlers
type Handler interface {
	// RegisterRoutes registers the handler's routes with the router
	RegisterRoutes(router RouteRegistrar)
}

// RouteRegistrar is the interface for registering routes
type RouteRegistrar interface {
	// HandleFunc registers a handler function for a given pattern
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}
