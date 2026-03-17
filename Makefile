APP_NAME=storage-provider
CMD_PATH=./cmd/api
GO=go
COVER_DIR=coverage
COVER_PROFILE=$(COVER_DIR)/coverage.out
COVER_HTML=$(COVER_DIR)/coverage.html

.PHONY: all build run clean swag tidy test test-coverage coverage coverage-func

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
	mkdir -p $(COVER_DIR)
	$(GO) test -v -race -coverprofile=$(COVER_PROFILE) -covermode=atomic ./...

test-coverage: test
	$(GO) tool cover -html=$(COVER_PROFILE) -o $(COVER_HTML)

coverage: test-coverage

coverage-func:
	$(GO) tool cover -func=$(COVER_PROFILE)

clean:
	$(GO) clean
	rm -f $(APP_NAME) $(APP_NAME).exe
	rm -rf $(COVER_DIR)
