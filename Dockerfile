# Build stage
FROM golang:1.27.3-alpine3.24@sha256:fa3d596898d9f126af7c01d14ee7597bab3f39cccf392ef36b69fe77cd681f18 AS builder
RUN go version

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
FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

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
