package handlers

import (
	"log"
	"net/http"
	"wowarmory/internal/interfaces"
)

// TokenHandler handles WoW Token price requests
type TokenHandler struct {
	*BaseHandler
	tokenClient interfaces.TokenAPI
}

var _ interfaces.Handler = (*TokenHandler)(nil)

// NewTokenHandler creates a new TokenHandler
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

// GetTokenPrice renders the current WoW Token price for EU and US
func (h *TokenHandler) GetTokenPrice(w http.ResponseWriter, r *http.Request) {
	prices, err := h.fetchPrices()
	if err != nil {
		log.Printf("error getting token prices: %v", err)
		h.RenderErrorPage(w, http.StatusBadGateway, "Token prices unavailable", "We couldn't reach Blizzard right now. Please try again in a moment.")
		return
	}

	h.Render(w, http.StatusOK, "token_page", map[string]interface{}{
		"PageTitle": "Token price",
		"ActiveTab": "token",
		"EUPrice":   prices["eu"],
		"USPrice":   prices["us"],
	})
}

// fetchPrices returns the token price in gold per region
func (h *TokenHandler) fetchPrices() (map[string]float64, error) {
	accessToken, err := h.tokenClient.GetAccessToken()
	if err != nil {
		return nil, err
	}

	prices := make(map[string]float64, 2)
	for _, region := range []string{"eu", "us"} {
		copper, err := h.tokenClient.GetTokenPrice(accessToken, region)
		if err != nil {
			return nil, err
		}
		prices[region] = copper / 10000
	}
	return prices, nil
}
