# Use the official Golang image as the base image for building
FROM golang:1.23.4-alpine AS build

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

# Use a minimal base image to reduce the size of the final image
FROM golang:1.23.4-alpine

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file and source files from the builder stage
COPY --from=build /app/main .
COPY --from=build /app/assets ./assets
COPY --from=build /app/views ./views
COPY --from=build /app/*.go ./

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./main"]