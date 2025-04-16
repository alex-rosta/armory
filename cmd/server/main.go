package main

import (
	"fmt"
	"log"
	"net/http"

	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/handlers"
	"wowarmory/internal/interfaces"
	"wowarmory/internal/middleware"
	"wowarmory/internal/redis"
	"wowarmory/internal/router"
	"wowarmory/internal/templates"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create API clients
	blizzardClient := api.NewBlizzardClient(cfg.ClientID, cfg.ClientSecret)
	warcraftlogsClient := api.NewWarcraftlogsClient(cfg.WarcraftlogsAPIToken)
	tokenClient := api.NewTokenClient(cfg.ClientID, cfg.ClientSecret)

	// Create Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Create template manager
	templateMgr, err := templates.NewManager(cfg.TemplatesDir)
	if err != nil {
		log.Fatalf("Failed to create template manager: %v", err)
	}

	// Create base handler
	baseHandler := handlers.NewBaseHandler(cfg, redisClient, templateMgr)

	// Create handlers
	characterHandler := handlers.NewCharacterHandler(baseHandler, blizzardClient)
	guildHandler := handlers.NewGuildHandler(baseHandler, warcraftlogsClient)
	recentSearchesHandler := handlers.NewRecentSearchesHandler(baseHandler)
	tokenHandler := handlers.NewTokenHandler(baseHandler, tokenClient)

	// Collect all handlers
	appHandlers := []interfaces.Handler{
		characterHandler,
		guildHandler,
		recentSearchesHandler,
		tokenHandler,
	}

	// Create router
	r := router.New(cfg)
	r.SetupHandlers(appHandlers)

	// Wrap router with middleware
	handler := middleware.RecoveryMiddleware(
		middleware.LoggingMiddleware(r),
	)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Listening on http://localhost%s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
