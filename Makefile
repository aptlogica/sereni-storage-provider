APP_NAME=storage-provider
CMD_PATH=./cmd/api

.PHONY: all build run clean swag tidy test

all: build

build:
	go build -o $(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

swag:
	swag init -g cmd/api/main.go

tidy:
	go mod tidy

test:
	go test -v ./...

clean:
	go clean
	@if exist $(APP_NAME).exe del $(APP_NAME).exe
	@if exist $(APP_NAME) del $(APP_NAME)
