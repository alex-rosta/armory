package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

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

	// Wrap router with middleware; the first one listed is the outermost
	limiter := middleware.NewRateLimiter(2, 20)
	handler := middleware.RecoveryMiddleware(
		middleware.SecurityHeadersMiddleware(
			middleware.LoggingMiddleware(cfg.ClientIPHeader)(
				middleware.RateLimitMiddleware(limiter, cfg.ClientIPHeader)(
					middleware.AllowReadOnlyMiddleware(r),
				),
			),
		),
	)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("Listening on http://localhost%s", addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Failed to start server: %v", err)
	}
	<-shutdownDone
}
