package handlers

import (
	"log"
	"net/http"
	"wowarmory/internal/interfaces"
)

type TokenHandler struct {
	*BaseHandler
	tokenClient interfaces.TokenAPI
}

var _ interfaces.Handler = (*TokenHandler)(nil)

// GetName returns the name of the handler
func (h *TokenHandler) GetName() string {
	return "TokenHandler"
}

func NewTokenHandler(base *BaseHandler, tokenClient interfaces.TokenAPI) *TokenHandler {
	return &TokenHandler{
		BaseHandler: base,
		tokenClient: tokenClient,
	}
}

// RegisterRoutes registers the handler's routes with the router
func (h *TokenHandler) RegisterRoutes(router interfaces.RouteRegistrar) {
	router.HandleFunc("/token", h.GetTokenPrice)
}

func (h *TokenHandler) GetTokenPrice(w http.ResponseWriter, r *http.Request) {
	accessToken, err := h.tokenClient.GetAccessToken()
	if err != nil {
		log.Printf("error getting access token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	euPrice, err := h.tokenClient.GetTokenPrice(accessToken, "eu")
	if err != nil {
		log.Printf("error getting EU token price: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	usPrice, err := h.tokenClient.GetTokenPrice(accessToken, "us")
	if err != nil {
		log.Printf("error getting US token price: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prices are in copper; convert to gold
	euGold := euPrice / 10000
	usGold := usPrice / 10000

	layoutData := map[string]interface{}{
		"ActiveTab":      "token",
		"PageTitle":      "Token Price",
		"ContainerClass": "token-container",
		"EUPrice":        euGold,
		"USPrice":        usGold,
	}

	if err := h.RenderWithLayout(w, "token", layoutData); err != nil {
		log.Printf("error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
