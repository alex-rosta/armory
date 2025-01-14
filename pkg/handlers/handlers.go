package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"wowchecker/pkg/api"
	"wowchecker/pkg/models"
)

func GetStringValue(data map[string]interface{}, key string, subkey string) string {
	if val, ok := data[key]; ok {
		if subval, ok := val.(map[string]interface{})[subkey]; ok {
			return subval.(string)
		}
	}
	return ""
}

func LookupCharacter(w http.ResponseWriter, r *http.Request) {
	region := r.URL.Query().Get("region")
	realm := r.URL.Query().Get("realm")
	character := r.URL.Query().Get("character")

	data := models.CharacterData{}

	if region != "" && realm != "" && character != "" {
		accessToken, err := api.GetAccessToken(region)
		if err != nil {
			http.Error(w, "Error getting access token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		profileData, err := api.GetCharacterProfile(accessToken, region, realm, character)
		if err != nil {
			http.Error(w, "Error getting character profile: "+err.Error(), http.StatusInternalServerError)
			return
		}

		titles := models.TitlesOwned{
			Titles: []struct {
				Name string `json:"name"`
			}{},
		}

		mounts := models.MountsOwned{
			Mounts: []struct {
				Mount struct {
					Name string `json:"name"`
				} `json:"mount"`
			}{},
		}

		if titlesData, ok := profileData["titles"].([]interface{}); ok {
			for _, title := range titlesData {
				titleMap := title.(map[string]interface{})
				titles.Titles = append(titles.Titles, struct {
					Name string `json:"name"`
				}{
					Name: titleMap["name"].(string),
				})
			}
		}

		if mountsData, ok := profileData["mounts"].([]interface{}); ok {
			for _, mount := range mountsData {
				mountMap := mount.(map[string]interface{})
				mounts.Mounts = append(mounts.Mounts, struct {
					Mount struct {
						Name string `json:"name"`
					} `json:"mount"`
				}{
					Mount: struct {
						Name string `json:"name"`
					}{
						Name: mountMap["mount"].(map[string]interface{})["name"].(string),
					},
				})
			}
		}

		characterimages := models.CharacterMedia{
			Assets: []struct {
				Key   string `json:"key"`
				Value string `json:"value"`
			}{},
		}

		if characterImagesData, ok := profileData["assets"].([]interface{}); ok {
			for _, image := range characterImagesData {
				imageMap := image.(map[string]interface{})
				characterimages.Assets = append(characterimages.Assets, struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				}{
					Key:   imageMap["key"].(string),
					Value: imageMap["value"].(string),
				})
			}
		} else {
			fmt.Println("Character media data not found") // Debugging statement
		}

		data = models.CharacterData{
			Name:              profileData["name"].(string),
			Level:             int(profileData["level"].(float64)),
			ItemLevel:         int(profileData["average_item_level"].(float64)),
			AchievementPoints: int(profileData["achievement_points"].(float64)),
			ActiveSpec:        struct{ Name string }{Name: profileData["active_spec"].(map[string]interface{})["name"].(string)},
			Class:             struct{ Name string }{Name: profileData["character_class"].(map[string]interface{})["name"].(string)},
			Titles:            titles,
			Mounts:            mounts,
			CharacterImages:   characterimages,
			MainRawImage:      characterimages.GetMainRawImage(),
			Guild:             struct{ Name string }{Name: GetStringValue(profileData, "guild", "name")},
			Faction:           struct{ Name string }{Name: GetStringValue(profileData, "faction", "name")},
			Health:            int(profileData["health"].(float64)),
			PowerType:         struct{ Name string }{Name: GetStringValue(profileData, "power_type", "name")},
			Power:             int(profileData["power"].(float64)),
			Stamina:           struct{ Effective int }{Effective: int(profileData["stamina"].(map[string]interface{})["effective"].(float64))},
		}
	}

	tmpl, err := template.ParseFiles("views/layout.html", "views/form.html", "views/character.html")
	if err != nil {
		http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
