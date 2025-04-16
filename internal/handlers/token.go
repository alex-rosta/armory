package handlers

import (
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
	// Get access token
	accessToken, err := h.tokenClient.GetAccessToken()
	if err != nil {
		http.Error(w, "Error getting access token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get EU token price
	euPrice, err := h.tokenClient.GetTokenPrice(accessToken, "eu")
	if err != nil {
		http.Error(w, "Error getting EU token price: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Get US token price
	usPrice, err := h.tokenClient.GetTokenPrice(accessToken, "us")
	if err != nil {
		http.Error(w, "Error getting US token price: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Format prices (convert to gold)
	euGold := euPrice / 10000
	usGold := usPrice / 10000

	// Render token template with token prices
	layoutData := map[string]interface{}{
		"ActiveTab":      "token",
		"PageTitle":      "Token Price",
		"ContainerClass": "token-container",
		"EUPrice":        euGold,
		"USPrice":        usGold,
	}

	if err := h.RenderWithLayout(w, "token", layoutData); err != nil {
		http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
	}
}
