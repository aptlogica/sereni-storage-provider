# Build stage
FROM golang:1.24.4-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

WORKDIR /app

# Install required packages
RUN apk add --no-cache git ca-certificates

# Install swag CLI for swagger docs generation (pinned to v1.16.4 - commit 0b9e347c196710ea155a147782bf51707a600c2c)
RUN git clone https://github.com/swaggo/swag.git /tmp/swag && \
    cd /tmp/swag && \
    git checkout 0b9e347c196710ea155a147782bf51707a600c2c && \
    go install ./cmd/swag && \
    rm -rf /tmp/swag

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate swagger docs and build the application
RUN swag init -g cmd/server/main.go -o docs && go mod tidy && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:3.20@sha256:a4f4213abb84c497377b8544c81b3564f313746700372ec4fe84653e4fb03805

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
