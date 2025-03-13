package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Port         int
	ClientID     string
	ClientSecret string
	TemplatesDir string
	AssetsDir    string
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

	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("missing required environment variables: CLIENT_ID and CLIENT_SECRET")
	}

	// Set default directories
	templatesDir := "internal/templates"
	assetsDir := "assets"

	return &Config{
		Port:         port,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TemplatesDir: templatesDir,
		AssetsDir:    assetsDir,
	}, nil
}
