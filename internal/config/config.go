package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port                 int
	ClientID             string
	ClientSecret         string
	TemplatesDir         string
	AssetsDir            string
	Redis                RedisConfig
	WarcraftlogsAPIToken string
}

// RedisConfig holds Redis-specific configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Warning: Error loading .env file, continuing without it. Ignore this if running as container.")
	}

	// Get port from environment variable or use default
	portStr := os.Getenv("PORT")
	port := 3000 // Default port
	if portStr != "" {
		portInt, err := strconv.Atoi(portStr)
		if err == nil {
			port = portInt
		}
	}

	// Get client ID and secret from environment variables
	clientID := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	warcraftlogsAPIToken := os.Getenv("WARCRAFTLOGS_API_TOKEN")

	if clientID == "" || clientSecret == "" || warcraftlogsAPIToken == "" {
		return nil, fmt.Errorf("missing required environment variables: CLIENT_ID, CLIENT_SECRET or WARCRAFTLOGS_API_TOKEN")
	}

	// Set default directories
	templatesDir := "internal/templates"
	assetsDir := "assets"

	// Get Redis configuration - handle both REDIS_URL and individual env vars
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := strings.Trim(os.Getenv("REDIS_PASSWORD"), "'")
	redisDB := 0

	// Check if REDIS_URL is provided (fly.io format)
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		// Parse the Redis URL
		u, err := url.Parse(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse REDIS_URL: %w", err)
		}

		// Extract host and port
		redisAddr = u.Host

		// Extract password
		if u.User != nil {
			if pwd, ok := u.User.Password(); ok {
				redisPassword = pwd
			}
		}

		// Extract database number from path
		if u.Path != "" && u.Path != "/" {
			dbStr := strings.TrimPrefix(u.Path, "/")
			if dbInt, err := strconv.Atoi(dbStr); err == nil {
				redisDB = dbInt
			}
		}
	} else {
		// Use individual environment variables (backward compatibility)
		if redisAddr == "" {
			redisAddr = "localhost:6379" // Default Redis address
		}

		redisDBStr := os.Getenv("REDIS_DB")
		if redisDBStr != "" {
			redisDBInt, err := strconv.Atoi(redisDBStr)
			if err == nil {
				redisDB = redisDBInt
			}
		}
	}

	return &Config{
		Port:                 port,
		ClientID:             clientID,
		ClientSecret:         clientSecret,
		TemplatesDir:         templatesDir,
		AssetsDir:            assetsDir,
		WarcraftlogsAPIToken: warcraftlogsAPIToken,
		Redis: RedisConfig{
			Addr:     redisAddr,
			Password: redisPassword,
			DB:       redisDB,
		},
	}, nil
}
