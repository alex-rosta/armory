# WoW Armory

A web application that allows users to look up World of Warcraft character and guild information using the Blizzard and Warcraftlogs API.
https://armory.rosta.dev

## Helmchart

https://github.com/alexrsit/armory-helm

## Features

- Character lookup by region, realm, and name
- Display of character information including level, item level, achievement points, etc.
- Display of character images
- Guild lookup by region, realm, and name
- Display Guild information such as realm, region and world ranking.
- Display recent raiders in the guild using the warcraftlogs GraphQL API.
- Global recent searches tracking with Redis (last 24 hours)
- Azure Cache for Redis (Tested, probably works on AWS or GCP aswell)

## Prerequisites

- Go 1.18 or higher
- Blizzard API credentials (Client ID and Client Secret)
- Redis server (for recent searches feature)

## Configuration

Create a `.env` file in the root directory with the following content:

```
# Blizzard API credentials
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret

# Warcraft Logs API token
WARCRAFTLOGS_API_TOKEN=your_warcraftlogs_api_token
# Get it here: https://www.warcraftlogs.com/api/docs the access token is valid for a year.

# App is exposed on this port
PORT=3000

# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
# Uncomment this if using Redis on Cloud Providers
#REDIS_CLOUD=true
```

## Running the Application

```bash
# Redis locally
docker pull redis:latest && docker run -p 6379:6379 redis

# Run directly
go run main.go

# Build and run
go build
./wowarmory

# Using Docker
docker build -t wowarmory .
docker run -p 3000:3000 --env-file .env wowarmory #pass local .env file

# Using Docker Compose (includes Redis)
docker-compose up -d --build
```

The application will be available at http://localhost:3000

## Testing

### Integration Tests

The project includes integration tests for the Blizzard API, Warcraftlogs API and Redis components. These tests verify that the application correctly integrates with external services.

To run the integration tests use:

```bash
make test
```

## Project Structure

```
wowarmory/
├── assets/                  # Static assets
│   ├── css/                 # CSS files
│   └── media/               # Images and other media
├── cmd/                     # Application entry points
│   └── server/              # Server entry point
│       └── main.go          # Main application file
├── internal/                # Private application code
│   ├── api/                 # API client for Blizzard API
│   ├── config/              # Configuration management
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # HTTP middleware
│   ├── models/              # Data models
│   ├── redis/               # Redis client and logic
│   ├── router/              # HTTP router
│   └── templates/           # HTML templates
├── tests/                   # Test files
│   ├── integration/         # Integration tests
├── .dockerignore            # Docker ignore file
├── .env                     # Environment variables (not in version control)
├── .env.example             # Example environment variables
├── .gitignore               # Git ignore file
├── Dockerfile               # Docker build file
├── go.mod                   # Go module file
├── go.sum                   # Go module checksum file
├── Makefile                 # Makefile for common tasks
└── README.md                # Project documentation
```
