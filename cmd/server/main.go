package main

import (
	"fmt"
	"log"
	"net/http"

	"wowarmory/internal/api"
	"wowarmory/internal/config"
	"wowarmory/internal/handlers"
	"wowarmory/internal/middleware"
	"wowarmory/internal/router"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Blizzard API client
	blizzardClient := api.NewBlizzardClient(cfg.ClientID, cfg.ClientSecret)

	// Create character handler
	characterHandler, err := handlers.NewCharacterHandler(cfg, blizzardClient)
	if err != nil {
		log.Fatalf("Failed to create character handler: %v", err)
	}

	// Create router
	r := router.New(cfg)
	r.Setup(characterHandler)

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
