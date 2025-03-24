package models

// CharacterProfile represents the raw character profile data from the Blizzard API
type CharacterProfile struct {
	Level             int    `json:"level"`
	Name              string `json:"name"`
	ItemLevel         int    `json:"average_item_level"`
	AchievementPoints int    `json:"achievement_points"`
	Realm             struct {
		Name string `json:"name"`
	} `json:"realm"`
	Faction struct {
		Name string `json:"name"`
	} `json:"faction"`
	Guild struct {
		Name string `json:"name"`
	} `json:"guild"`
	Class struct {
		Name string `json:"name"`
	} `json:"character_class"`
	ActiveSpec struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"active_spec"`
	Health    int `json:"health"`
	PowerType struct {
		Name string `json:"name"`
	} `json:"power_type"`
	Power   int `json:"power"`
	Stamina struct {
		Effective int `json:"effective"`
	} `json:"stamina"`
}

// CharacterMedia represents the character media data from the Blizzard API
type CharacterMedia struct {
	Assets []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"assets"`
}

// GetMainRawImage returns the main-raw image URL from the character media
func (media *CharacterMedia) GetMainRawImage() string {
	for _, asset := range media.Assets {
		if asset.Key == "main-raw" {
			return asset.Value
		}
	}
	return ""
}

// CharacterData represents the processed character data for display
type CharacterData struct {
	Name              string
	Level             int
	ItemLevel         int
	AchievementPoints int
	Realm             struct {
		Name string
	}
	Region  string
	Faction struct {
		Name string
	}
	ActiveSpec struct {
		Name string
	}
	Class struct {
		Name string
	}
	CharacterImages CharacterMedia
	MainRawImage    string
	Guild           struct {
		Name string
	}
	Health    int
	PowerType struct {
		Name string
	}
	Power   int
	Stamina struct {
		Effective int
	}
}

// NewCharacterData creates a new CharacterData from a map of profile data
func NewCharacterData(profileData map[string]interface{}, region string) (CharacterData, error) {
	data := CharacterData{
		Region: region,
	}

	// Extract basic character information
	if name, ok := profileData["name"].(string); ok {
		data.Name = name
	}
	if level, ok := profileData["level"].(float64); ok {
		data.Level = int(level)
	}
	if itemLevel, ok := profileData["average_item_level"].(float64); ok {
		data.ItemLevel = int(itemLevel)
	}
	if achievementPoints, ok := profileData["achievement_points"].(float64); ok {
		data.AchievementPoints = int(achievementPoints)
	}

	// Extract active spec
	if activeSpec, ok := profileData["active_spec"].(map[string]interface{}); ok {
		if name, ok := activeSpec["name"].(string); ok {
			data.ActiveSpec.Name = name
		}
	}

	// Extract realm
	if realm, ok := profileData["realm"].(map[string]interface{}); ok {
		if name, ok := realm["name"].(string); ok {
			data.Realm.Name = name
		}
	}

	// Extract class
	if class, ok := profileData["character_class"].(map[string]interface{}); ok {
		if name, ok := class["name"].(string); ok {
			data.Class.Name = name
		}
	}

	// Extract guild
	if guild, ok := profileData["guild"].(map[string]interface{}); ok {
		if name, ok := guild["name"].(string); ok {
			data.Guild.Name = name
		}
	}

	// Extract faction
	if faction, ok := profileData["faction"].(map[string]interface{}); ok {
		if name, ok := faction["name"].(string); ok {
			data.Faction.Name = name
		}
	}

	// Extract health
	if health, ok := profileData["health"].(float64); ok {
		data.Health = int(health)
	}

	// Extract power type
	if powerType, ok := profileData["power_type"].(map[string]interface{}); ok {
		if name, ok := powerType["name"].(string); ok {
			data.PowerType.Name = name
		}
	}

	// Extract power
	if power, ok := profileData["power"].(float64); ok {
		data.Power = int(power)
	}

	// Extract stamina
	if stamina, ok := profileData["stamina"].(map[string]interface{}); ok {
		if effective, ok := stamina["effective"].(float64); ok {
			data.Stamina.Effective = int(effective)
		}
	}

	// Extract character images
	characterimages := CharacterMedia{
		Assets: []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{},
	}

	if characterImagesData, ok := profileData["assets"].([]interface{}); ok {
		for _, image := range characterImagesData {
			if imageMap, ok := image.(map[string]interface{}); ok {
				if key, ok := imageMap["key"].(string); ok {
					if value, ok := imageMap["value"].(string); ok {
						characterimages.Assets = append(characterimages.Assets, struct {
							Key   string `json:"key"`
							Value string `json:"value"`
						}{
							Key:   key,
							Value: value,
						})
					}
				}
			}
		}
	}

	data.CharacterImages = characterimages
	data.MainRawImage = characterimages.GetMainRawImage()

	return data, nil
}
