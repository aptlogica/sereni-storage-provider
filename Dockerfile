# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache git ca-certificates

# Install swag CLI for swagger docs generation
RUN go install github.com/swaggo/swag/cmd/swag@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs and build the application
RUN swag init -g cmd/server/main.go -o docs && \
    go mod tidy && \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Create uploads directory
RUN mkdir -p /app/uploads


# Environment variables are provided by .env via Docker Compose

# Expose port
EXPOSE 8083

# Run the application
CMD ["./main"]
