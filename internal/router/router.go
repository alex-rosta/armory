package router

import (
	"net/http"

	"wowarmory/internal/config"
	"wowarmory/internal/handlers"
	"wowarmory/internal/middleware"
)

// Router handles HTTP routing for the application
type Router struct {
	config *config.Config
	mux    *http.ServeMux
}

// New creates a new router
func New(cfg *config.Config) *Router {
	return &Router{
		config: cfg,
		mux:    http.NewServeMux(),
	}
}

// Setup sets up the routes for the application
func (r *Router) Setup(characterHandler *handlers.CharacterHandler, guildHandler *handlers.GuildHandler, recentSearchesHandler *handlers.RecentSearchesHandler) {
	// Set up static file server
	fileServer := http.FileServer(http.Dir(r.config.AssetsDir))
	r.mux.Handle("/assets/", http.StripPrefix("/assets/", middleware.ContentTypeMiddleware(fileServer)))

	// Set up routes
	r.mux.HandleFunc("/", characterHandler.LookupCharacter)
	r.mux.HandleFunc("/character", characterHandler.GetCharacterTemplate)
	r.mux.HandleFunc("/guild-lookup", guildHandler.LookupGuild)
	r.mux.HandleFunc("/guild", guildHandler.GetGuildTemplate)
	r.mux.HandleFunc("/recent-searches", recentSearchesHandler.GetRecentSearchesPage)
	r.mux.HandleFunc("/recent-searches-data", recentSearchesHandler.GetRecentSearches)
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
