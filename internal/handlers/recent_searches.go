package handlers

import (
	"context"
	"log"
	"net/http"
	"time"
	"wowarmory/internal/interfaces"
)

// RecentSearchesHandler handles recent searches-related HTTP requests
type RecentSearchesHandler struct {
	*BaseHandler
}

// Ensure RecentSearchesHandler implements Handler interface
var _ interfaces.Handler = (*RecentSearchesHandler)(nil)

// GetName returns the name of the handler
func (h *RecentSearchesHandler) GetName() string {
	return "RecentSearchesHandler"
}

// NewRecentSearchesHandler creates a new RecentSearchesHandler
func NewRecentSearchesHandler(base *BaseHandler) *RecentSearchesHandler {
	return &RecentSearchesHandler{
		BaseHandler: base,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *RecentSearchesHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/recent-searches", h.GetRecentSearchesPage)
	router.HandleFunc("/recent-searches-data", h.GetRecentSearches)
}

// GetRecentSearchesPage handles the recent searches page request
func (h *RecentSearchesHandler) GetRecentSearchesPage(w http.ResponseWriter, r *http.Request) {
	layoutData := map[string]interface{}{
		"PageTitle": "Recent Searches",
		"ActiveTab": "recent-searches",
	}

	if err := h.RenderWithLayout(w, "recent_searches_container", layoutData); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// GetRecentSearches handles the htmx request for recent searches data
func (h *RecentSearchesHandler) GetRecentSearches(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	searches, err := h.redisClient.GetRecentSearches(ctx)
	if err != nil {
		log.Printf("error getting recent searches: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Searches": searches,
	}

	if err := h.RenderTemplate(w, "recent_searches", data); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
