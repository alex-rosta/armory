# Use the official Golang image as the base image for building
FROM golang:1.23.4-alpine AS build

# Install necessary packages
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /app/wowarmory cmd/server/main.go

# Use a smaller base image for the final container
FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the built Go app and assets from the build stage
COPY --from=build /app/wowarmory .
COPY --from=build /app/assets ./assets
COPY --from=build /app/internal/templates ./internal/templates

# Expose the port the app runs on
EXPOSE 3000

# Environment variables with defaults
ENV REDIS_ADDR=redis:6379
ENV REDIS_PASSWORD=
ENV REDIS_DB=0

# Command to run the executable
CMD ["./wowarmory"]
