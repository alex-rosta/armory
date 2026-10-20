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

// NewRecentSearchesHandler creates a new RecentSearchesHandler
func NewRecentSearchesHandler(base *BaseHandler) *RecentSearchesHandler {
	return &RecentSearchesHandler{
		BaseHandler: base,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *RecentSearchesHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/recent-searches", h.GetRecentSearchesPage)
}

// GetRecentSearchesPage renders the recent searches page
func (h *RecentSearchesHandler) GetRecentSearchesPage(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	searches, err := h.searchStore.GetRecentSearches(ctx)
	if err != nil {
		log.Printf("error getting recent searches: %v", err)
		h.RenderErrorPage(w, http.StatusServiceUnavailable, "Recent searches unavailable", "We couldn't load recent searches right now. Please try again in a moment.")
		return
	}

	h.Render(w, http.StatusOK, "recent_page", map[string]interface{}{
		"PageTitle": "Recent searches",
		"ActiveTab": "recent",
		"Searches":  searches,
	})
}
