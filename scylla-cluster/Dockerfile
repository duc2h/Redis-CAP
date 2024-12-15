# Use the official Golang image as the base image
FROM golang:1.22 as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o main .

# Use Ubuntu 22.04 as the base image for the final stage
FROM ubuntu:22.04

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/main .

# Install necessary tools for interacting with Scylla
RUN apt-get update && apt-get install -y \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Expose the port the application runs on
EXPOSE 8080

# Command to run the Go application
CMD ["./main"]