#!/bin/bash
echo "Generating Swagger documentation..."
swag init -g cmd/api/main.go

echo "Starting application..."
go run ./cmd/api
