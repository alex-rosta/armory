# WoW Armory

A web application that allows users to look up World of Warcraft character information using the Blizzard API.
https://armory.rosta.dev

## Helmchart

https://github.com/alexrsit/armory-helm

## Features

- Character lookup by region, realm, and name
- Display of character information including level, item level, achievement points, etc.
- Display of character images
- Global recent searches tracking with Redis (last 24 hours)

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

# App is exposed on this port
PORT=3000

# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
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
    ├── redis/               # Redis client and logic
│   ├── router/              # HTTP router
│   └── templates/           # HTML templates
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

