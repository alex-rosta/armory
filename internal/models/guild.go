package models

import (
	"fmt"
	"wowarmory/internal/api"
)

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
	// Type assertion for our known response structure
	response, ok := guildResponse.(*api.GuildResponse)
	if response.GuildData.Guild.Name != "" && ok {
		// We have a strongly typed response, use it directly
		guildData := &GuildData{
			Name:            response.GuildData.Guild.Name,
			Region:          region,
			Realm:           realm,
			MemberCount:     response.GuildData.Guild.Members.Total,
			ServerRank:      response.GuildData.Guild.ZoneRanking.Progress.ServerRank.Number,
			ServerRankColor: response.GuildData.Guild.ZoneRanking.Progress.ServerRank.Color,
			RegionRank:      response.GuildData.Guild.ZoneRanking.Progress.RegionRank.Number,
			RegionRankColor: response.GuildData.Guild.ZoneRanking.Progress.RegionRank.Color,
			WorldRank:       response.GuildData.Guild.ZoneRanking.Progress.WorldRank.Number,
			WorldRankColor:  response.GuildData.Guild.ZoneRanking.Progress.WorldRank.Color,
		}

		// Process members from attendance data
		if len(response.GuildData.Guild.Attendance.Data) > 0 {
			// Use a map to deduplicate members
			memberMap := make(map[string]string)

			// Take the most recent raid (first entry)
			for _, player := range response.GuildData.Guild.Attendance.Data[0].Players {
				memberMap[player.Name] = player.Type
			}

			// Convert map to slice
			for name, playerType := range memberMap {
				guildData.Members = append(guildData.Members, GuildMember{
					Name: name,
					Type: playerType,
				})
			}
		}

		return guildData, nil
	}

	// Fallback to the generic map approach if the typed assertion failed
	mapResponse, ok := guildResponse.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format")
	}

	guildData, ok := mapResponse["guildData"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("guildData not found in response")
	}

	guild, ok := guildData["guild"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("guild not found in guildData")
	}

	// Extract guild name
	name, _ := guild["name"].(string)

	// Extract member count
	members, ok := guild["members"].(map[string]interface{})
	memberCount := 0
	if ok {
		total, ok := members["total"].(float64)
		if ok {
			memberCount = int(total)
		}
	}

	// Extract ranking information
	serverRank, serverRankColor := 0, ""
	regionRank, regionRankColor := 0, ""
	worldRank, worldRankColor := 0, ""

	zoneRanking, ok := guild["zoneRanking"].(map[string]interface{})
	if ok {
		progress, ok := zoneRanking["progress"].(map[string]interface{})
		if ok {
			// Server rank
			serverRankData, ok := progress["serverRank"].(map[string]interface{})
			if ok {
				if number, ok := serverRankData["number"].(float64); ok {
					serverRank = int(number)
				}
				serverRankColor, _ = serverRankData["color"].(string)
			}

			// Region rank
			regionRankData, ok := progress["regionRank"].(map[string]interface{})
			if ok {
				if number, ok := regionRankData["number"].(float64); ok {
					regionRank = int(number)
				}
				regionRankColor, _ = regionRankData["color"].(string)
			}

			// World rank
			worldRankData, ok := progress["worldRank"].(map[string]interface{})
			if ok {
				if number, ok := worldRankData["number"].(float64); ok {
					worldRank = int(number)
				}
				worldRankColor, _ = worldRankData["color"].(string)
			}
		}
	}

	// Extract guild members
	var guildMembers []GuildMember
	attendance, ok := guild["attendance"].(map[string]interface{})
	if ok {
		dataArray, ok := attendance["data"].([]interface{})
		if ok && len(dataArray) > 0 {
			// Take the most recent attendance data (first element in the array)
			mostRecentData := dataArray[0]
			dataMap, ok := mostRecentData.(map[string]interface{})
			if ok {
				players, ok := dataMap["players"].([]interface{})
				if ok {
					// Create a map to deduplicate players
					memberMap := make(map[string]bool)

					for _, player := range players {
						playerMap, ok := player.(map[string]interface{})
						if ok {
							name, ok := playerMap["name"].(string)
							if ok && !memberMap[name] {
								// Add player to the list if not already added
								memberMap[name] = true
								guildMembers = append(guildMembers, GuildMember{
									Name: name,
									Type: "Raider", // Default type since it's not in the current API response
								})
							}
						}
					}
				}
			}
		}
	}

	return &GuildData{
		Name:            name,
		Region:          region,
		Realm:           realm,
		MemberCount:     memberCount,
		ServerRank:      serverRank,
		ServerRankColor: serverRankColor,
		RegionRank:      regionRank,
		RegionRankColor: regionRankColor,
		WorldRank:       worldRank,
		WorldRankColor:  worldRankColor,
		Members:         guildMembers,
	}, nil
}
