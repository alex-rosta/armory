package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/models"
)

// GuildHandler handles guild-related HTTP requests
type GuildHandler struct {
	*BaseHandler
	warcraftlogsClient interfaces.WarcraftLogsAPI
}

// Ensure GuildHandler implements Handler interface
var _ interfaces.Handler = (*GuildHandler)(nil)

// NewGuildHandler creates a new GuildHandler
func NewGuildHandler(base *BaseHandler, warcraftlogsClient interfaces.WarcraftLogsAPI) *GuildHandler {
	return &GuildHandler{
		BaseHandler:        base,
		warcraftlogsClient: warcraftlogsClient,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *GuildHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/guild-lookup", h.LookupGuild)
}

// LookupGuild shows the guild search form and, when parameters are given, the result
func (h *GuildHandler) LookupGuild(w http.ResponseWriter, r *http.Request) {
	h.serveLookup(w, r, lookupConfig{
		Tab:        "guild",
		Template:   "guild_page",
		ResultKey:  "Guild",
		Title:      "Guild lookup",
		SearchType: interfaces.GuildSearchType,
		NameParam:  "guild",
		Form: SearchForm{
			Action:          "/guild-lookup",
			NameParam:       "guild",
			NameLabel:       "Guild name",
			NamePlaceholder: "e.g. divine intervention",
			SubmitLabel:     "Look up guild",
		},
		ExternalURL: func(q Lookup) string { return warcraftlogsGuildURL(q.Region, q.Realm, q.Name) },
		Fetch: func(ctx context.Context, q Lookup) (interface{}, string, error) {
			guild, err := h.fetchGuild(ctx, q)
			if err != nil {
				return nil, "", err
			}
			return guild, guild.Name, nil
		},
	})
}

// fetchGuild loads and processes a guild from the Warcraftlogs API
func (h *GuildHandler) fetchGuild(ctx context.Context, q Lookup) (*models.GuildData, error) {
	guildResponse, err := h.warcraftlogsClient.GetGuild(ctx, q.Name, q.Realm, q.Region)
	if err != nil {
		return nil, fmt.Errorf("getting guild data: %w", err)
	}

	return models.NewGuildData(guildResponse, q.Region, q.Realm)
}

func warcraftlogsGuildURL(region, realm, guild string) string {
	return fmt.Sprintf("https://www.warcraftlogs.com/guild/%s/%s/%s",
		url.PathEscape(region), url.PathEscape(realm), url.PathEscape(guild))
}
