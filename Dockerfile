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
RUN go build -o /app/main .

# Use a smaller base image for the final container
FROM alpine:latest

# Install necessary packages
RUN apk add --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the built Go app from the build stage
COPY --from=build /app/main .

# Command to run the executable
CMD ["./main"]