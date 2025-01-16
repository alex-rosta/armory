package models

type CharacterProfile struct {
	Level             int    `json:"level"`
	Name              string `json:"name"`
	ItemLevel         int    `json:"average_item_level"`
	AchievementPoints int    `json:"achievement_points"`
	Faction           struct {
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

type CharacterMedia struct {
	Assets []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"assets"`
}

type CharacterData struct {
	Name              string
	Level             int
	ItemLevel         int
	AchievementPoints int
	Faction           struct {
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

func (cm *CharacterMedia) GetMainRawImage() string {
	for _, asset := range cm.Assets {
		if asset.Key == "main-raw" {
			return asset.Value
		}
	}
	return ""
}
