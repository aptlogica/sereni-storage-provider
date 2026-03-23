APP_NAME=storage-provider
CMD_PATH=./cmd/api
GO=go
COVER_PROFILE=coverage.out
COVER_HTML=coverage.html

.PHONY: help all build run clean swag tidy test test-coverage coverage coverage-func

help: ## Display this help message
	@echo "Available targets:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the application"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make coverage       - Alias for test-coverage"
	@echo "  make coverage-func  - Show coverage by function"
	@echo "  make swag           - Generate swagger documentation"
	@echo "  make tidy           - Tidy go.mod"
	@echo "  make clean          - Clean build artifacts"

all: build

build:
	$(GO) build -o $(APP_NAME) $(CMD_PATH)

run:
	$(GO) run $(CMD_PATH)

swag:
	swag init -g cmd/api/main.go

tidy:
	$(GO) mod tidy

test:
	$(GO) test -v -race ./tests/...

test-coverage:
	$(GO) test -v -race -coverprofile=$(COVER_PROFILE) -covermode=atomic -coverpkg=./... ./tests/...
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

coverage: test-coverage

coverage-func:
	$(GO) tool cover -func=$(COVER_PROFILE)

clean:
	$(GO) clean
	rm -f $(APP_NAME) $(APP_NAME).exe
	rm -f $(COVER_PROFILE) $(COVER_HTML)
