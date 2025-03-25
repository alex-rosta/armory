package models

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
	// Type assertion to get the guild data from the response
	response, ok := guildResponse.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	guildData, ok := response["guildData"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	guild, ok := guildData["guild"].(map[string]interface{})
	if !ok {
		return nil, nil
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
		data, ok := attendance["data"].(map[string]interface{})
		if ok {
			players, ok := data["players"].([]interface{})
			if ok {
				for _, player := range players {
					playerMap, ok := player.(map[string]interface{})
					if ok {
						name, _ := playerMap["name"].(string)
						playerType, _ := playerMap["type"].(string)
						guildMembers = append(guildMembers, GuildMember{
							Name: name,
							Type: playerType,
						})
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
