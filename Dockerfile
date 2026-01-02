# Build stage
FROM golang:1.24.4-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o storage-provider ./cmd/api

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create app directory and uploads directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/storage-provider .

# Copy example.env (optional, can be overridden with volume mount)
COPY --from=builder /app/example.env .

# Create uploads directory
RUN mkdir -p /app/uploads

# Expose port 5050
EXPOSE 5050

# Set environment variable for port
ENV SERVER_PORT=5050
ENV SERVER_HOST=0.0.0.0

# Run the application
CMD ["./storage-provider"]
