package api

import (
	"context"
	"fmt"
	"wowarmory/internal/interfaces"

	"github.com/machinebox/graphql"
)

// WarcraftlogsClient is a client for the Warcraftlogs API
type WarcraftlogsClient struct {
	client      *graphql.Client
	accessToken string
}

// Ensure WarcraftlogsClient implements WarcraftLogsAPI interface
var _ interfaces.WarcraftLogsAPI = (*WarcraftlogsClient)(nil)

// GetClientName returns the name of the client
func (c *WarcraftlogsClient) GetClientName() string {
	return "WarcraftLogsAPI"
}

// NewWarcraftlogsClient creates a new Warcraftlogs API client
func NewWarcraftlogsClient(accessToken string) *WarcraftlogsClient {
	client := graphql.NewClient("https://www.warcraftlogs.com/api/v2/client")

	return &WarcraftlogsClient{
		client:      client,
		accessToken: accessToken,
	}
}

// GuildMember represents a member of a guild
type GuildMember struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

// AttendanceEntry represents a single attendance entry
type AttendanceEntry struct {
	Players []GuildMember `json:"players"`
}

// GuildAttendance represents attendance data for a guild
type GuildAttendance struct {
	Data []AttendanceEntry `json:"data"`
}

// GuildMembers represents the total number of members in a guild
type GuildMembers struct {
	Total int `json:"total"`
}

// RankInfo represents ranking information
type RankInfo struct {
	Number int    `json:"number"`
	Color  string `json:"color"`
}

// GuildProgress represents progress information for a guild
type GuildProgress struct {
	ServerRank RankInfo `json:"serverRank"`
	RegionRank RankInfo `json:"regionRank"`
	WorldRank  RankInfo `json:"worldRank"`
}

// GuildZoneRanking represents zone ranking information for a guild
type GuildZoneRanking struct {
	Progress GuildProgress `json:"progress"`
}

// GuildData represents data for a guild
type GuildData struct {
	Name        string           `json:"name"`
	Attendance  GuildAttendance  `json:"attendance"`
	Members     GuildMembers     `json:"members"`
	ZoneRanking GuildZoneRanking `json:"zoneRanking"`
}

// GuildResponse represents the response from the Warcraftlogs API for a guild query
type GuildResponse struct {
	GuildData struct {
		Guild GuildData `json:"guild"`
	} `json:"guildData"`
}

// GetGuild gets information about a guild from the Warcraftlogs API
func (c *WarcraftlogsClient) GetGuild(ctx context.Context, name, serverSlug, serverRegion string) (interface{}, error) {
	// Create a new request
	req := graphql.NewRequest(`
		query GetGuild(
			$name: String!
			$serverSlug: String!
			$serverRegion: String!
		) 
		{
			guildData {
				guild(name: $name, serverSlug: $serverSlug, serverRegion: $serverRegion) {
					name
					attendance{
						data{
							players{
								name
								type
							}
						}
					}
					members{
						total
					}
					zoneRanking{
						progress{
							serverRank{
								number
								color
							}
							regionRank{
								number
								color
							}
							worldRank{
								number
								color
							}
						}
					}
				}
			}
		}
	`)

	// Set variables
	req.Var("name", name)
	req.Var("serverSlug", serverSlug)
	req.Var("serverRegion", serverRegion)

	// Set auth header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	// Run the query
	var response GuildResponse
	if err := c.client.Run(ctx, req, &response); err != nil {
		return nil, fmt.Errorf("error querying Warcraftlogs API: %w", err)
	}

	return &response, nil
}
