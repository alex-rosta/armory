package main

import (
	"fmt"
	"log"
	"net/http"

	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/handlers"
	"wowarmory/internal/middleware"
	"wowarmory/internal/redis"
	"wowarmory/internal/router"
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

	// Create Redis client
	redisClient, err := redis.NewClient(&cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Create handlers
	characterHandler, err := handlers.NewCharacterHandler(cfg, blizzardClient, redisClient)
	if err != nil {
		log.Fatalf("Failed to create character handler: %v", err)
	}

	// Create guild handler
	guildHandler, err := handlers.NewGuildHandler(cfg, warcraftlogsClient, redisClient)
	if err != nil {
		log.Fatalf("Failed to create guild handler: %v", err)
	}

	// Create recent searches handler
	recentSearchesHandler := handlers.NewRecentSearchesHandler(cfg, redisClient, characterHandler.GetTemplates())

	// Create router
	r := router.New(cfg)
	r.Setup(characterHandler, guildHandler, recentSearchesHandler)

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
