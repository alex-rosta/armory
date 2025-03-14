# Makefile for WoW Armory

# Variables
BINARY_NAME=wowarmory
MAIN_PATH=cmd/server/main.go

# Build the application
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Clean the binary
clean:
	go clean
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test ./...

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Build and run the application
dev: build
	./$(BINARY_NAME)

css: 
	npx tailwindcss -i ./assets/css/tailwind.css -o ./assets/css/styles.css --watch

# Build Docker image
docker-build:
	docker build -t $(BINARY_NAME) .

# Run Docker container
docker-run:
	docker run -p 3000:3000 --env-file .env $(BINARY_NAME)

# Default target
.PHONY: build run clean test fmt vet dev docker-build docker-run
