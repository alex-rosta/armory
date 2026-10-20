package main

import (
	"context"
	"log"
	"net/http"

	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/handlers"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/middleware"
	"wowarmory/internal/router"
	"wowarmory/internal/templates"
)

type blizz struct{}

func (blizz) GetAccessToken() (string, error) { return "t", nil }
func (blizz) GetCharacterProfile(_, _, realm, name string) (map[string]interface{}, error) {
	if name == "ghost" {
		return nil, api.ErrNotFound
	}
	if name == "down" {
		return nil, context.DeadlineExceeded
	}
	return map[string]interface{}{
		"name": "Tempests", "level": float64(80), "average_item_level": float64(640),
		"achievement_points": float64(12345), "health": float64(1200000), "power": float64(100000),
		"realm": map[string]interface{}{"name": "Defias Brotherhood"}, "faction": map[string]interface{}{"name": "Horde"},
		"character_class": map[string]interface{}{"name": "Death Knight"}, "active_spec": map[string]interface{}{"name": "Blood"},
		"guild": map[string]interface{}{"name": "Divine Intervention"}, "power_type": map[string]interface{}{"name": "Runic Power"},
		"stamina": map[string]interface{}{"effective": float64(98765)},
	}, nil
}

type wcl struct{}

func (wcl) GetGuild(context.Context, string, string, string) (interface{}, error) {
	r := &api.GuildResponse{}
	r.GuildData.Guild.Name = "Divine Intervention"
	r.GuildData.Guild.Members.Total = 120
	p := &r.GuildData.Guild.ZoneRanking.Progress
	p.ServerRank = api.RankInfo{Number: 3, Color: "legendary"}
	p.RegionRank = api.RankInfo{Number: 41, Color: "epic"}
	p.WorldRank = api.RankInfo{Number: 1234, Color: "rare"}
	r.GuildData.Guild.Attendance.Data = []api.AttendanceEntry{{Players: []api.GuildMember{{Name: "Amy", Type: "Priest"}, {Name: "Zed", Type: "Mage"}, {Name: "Bob", Type: "Warrior"}}}}
	return r, nil
}

type tok struct{}

func (tok) GetAccessToken() (string, error)            { return "t", nil }
func (tok) GetTokenPrice(_, _ string) (float64, error) { return 2812345 * 10000, nil }

type store struct{}

func (store) RecordSearch(context.Context, string, string, string, string) error { return nil }
func (store) GetRecentSearches(context.Context) ([]interfaces.SearchEntry, error) {
	return []interfaces.SearchEntry{
		{Type: "character", Name: "tempests", Realm: "defias-brotherhood", Region: "eu", Timestamp: "20 Oct 26 14:00 CEST"},
		{Type: "guild", Name: "divine intervention", Realm: "darkspear", Region: "eu", Timestamp: "20 Oct 26 13:00 CEST"},
	}, nil
}
func (store) Close() error { return nil }

func main() {
	tm, err := templates.NewManager("internal/templates")
	if err != nil {
		log.Fatal(err)
	}
	base := handlers.NewBaseHandler(nil, store{}, tm)
	r := router.New(&config.Config{AssetsDir: "assets"})
	r.SetupHandlers([]interfaces.Handler{
		handlers.NewCharacterHandler(base, blizz{}),
		handlers.NewGuildHandler(base, wcl{}),
		handlers.NewRecentSearchesHandler(base),
		handlers.NewTokenHandler(base, tok{}),
	})
	log.Println("preview on http://localhost:3998")
	log.Fatal(http.ListenAndServe("127.0.0.1:3998", middleware.SecurityHeadersMiddleware(r)))
}
