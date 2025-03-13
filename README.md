# WoW Armory

A web application that allows users to look up World of Warcraft character information using the Blizzard API.
https://wow.rosta.dev

## Features

- Character lookup by region, realm, and name
- Display of character information including level, item level, achievement points, etc.
- Display of character images
- Support for both Horde and Alliance factions

## Prerequisites

- Go 1.18 or higher
- Blizzard API credentials (Client ID and Client Secret)

## Configuration

Create a `.env` file in the root directory with the following content:

```
CLIENT_ID=your_client_id
CLIENT_SECRET=your_client_secret
PORT=3000
```

## Running the Application

```bash
# Run directly
go run main.go

# Build and run
go build
./wowarmory

# Using Docker
docker build -t wowarmory .
docker run -p 3000:3000 --env-file .env wowarmory
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
│   ├── router/              # HTTP router
│   └── templates/           # HTML templates
├── pkg/                     # Public libraries
├── views/                   # HTML templates (to be moved to internal/templates)
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

## License

This project is licensed under the MIT License - see the LICENSE file for details.
