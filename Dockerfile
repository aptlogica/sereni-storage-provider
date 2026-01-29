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
RUN go mod tidy && go mod download

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

# Copy example.env (optional, can be overridden with volume mount)
COPY --from=builder /app/example.env .

# Copy swagger docs
COPY --from=builder /app/docs ./docs

# Create uploads directory
RUN mkdir -p /app/uploads

# Expose port 5050
EXPOSE 5050

# Set environment variable for port
ENV SERVER_PORT=5050
ENV SERVER_HOST=0.0.0.0

# Run the application
CMD ["./main"]
