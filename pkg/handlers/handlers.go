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
		accessToken, err := api.GetAccessToken()
		if err != nil {
			http.Error(w, "Error getting access token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		profileData, err := api.GetCharacterProfile(accessToken, region, realm, character)
		if err != nil {
			http.Error(w, "Error getting character profile: "+err.Error(), http.StatusInternalServerError)
			return
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
