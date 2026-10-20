# Use the official Golang image as the base image for building
FROM golang:1.27-alpine AS build

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

# Build the Go app as a static, stripped binary
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app/wowarmory cmd/server/main.go

# Use a smaller base image for the final container
FROM alpine:3.21

# Install necessary packages and create an unprivileged user
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the built Go app and assets from the build stage
COPY --from=build /app/wowarmory .
COPY --from=build /app/assets ./assets
COPY --from=build /app/internal/templates ./internal/templates

USER app

# Expose the port the app runs on
EXPOSE 3000

# Command to run the executable
CMD ["./wowarmory"]
