package models

import (
	"errors"
	"fmt"
	"sort"
	"wowarmory/internal/api"
)

// ErrGuildNotFound is returned when Warcraftlogs has no guild matching the query
var ErrGuildNotFound = errors.New("guild not found")

// GuildData represents the data for a guild
type GuildData struct {
	Name            string
	Region          string
	Realm           string
	MemberCount     int
	ServerRank      int
	ServerRankColor string
	RegionRank      int
	RegionRankColor string
	WorldRank       int
	WorldRankColor  string
	Members         []GuildMember
}

// GuildMember represents a member of a guild
type GuildMember struct {
	Name string
	Type string
}

// NewGuildData creates a new GuildData from the API response
func NewGuildData(guildResponse interface{}, region, realm string) (*GuildData, error) {
	response, ok := guildResponse.(*api.GuildResponse)
	if !ok || response == nil {
		return nil, fmt.Errorf("unexpected guild response type %T", guildResponse)
	}

	guild := response.GuildData.Guild
	if guild.Name == "" {
		return nil, ErrGuildNotFound
	}

	progress := guild.ZoneRanking.Progress
	guildData := &GuildData{
		Name:            guild.Name,
		Region:          region,
		Realm:           realm,
		MemberCount:     guild.Members.Total,
		ServerRank:      progress.ServerRank.Number,
		ServerRankColor: progress.ServerRank.Color,
		RegionRank:      progress.RegionRank.Number,
		RegionRankColor: progress.RegionRank.Color,
		WorldRank:       progress.WorldRank.Number,
		WorldRankColor:  progress.WorldRank.Color,
	}

	// Most recent raid is the first attendance entry
	if len(guild.Attendance.Data) > 0 {
		seen := make(map[string]bool)
		for _, player := range guild.Attendance.Data[0].Players {
			if seen[player.Name] {
				continue
			}
			seen[player.Name] = true
			guildData.Members = append(guildData.Members, GuildMember{Name: player.Name, Type: player.Type})
		}
		sort.Slice(guildData.Members, func(i, j int) bool {
			return guildData.Members[i].Name < guildData.Members[j].Name
		})
	}

	return guildData, nil
}
