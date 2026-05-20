.PHONY: dev build run test lint migrate seed clean

APP_NAME=expense-tracker
BUILD_DIR=./bin

dev:
	@which air > /dev/null 2>&1 || go install github.com/air-verse/air@latest
	air

build:
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server

run: build
	$(BUILD_DIR)/$(APP_NAME)

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	@which golangci-lint > /dev/null 2>&1 || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run ./...

migrate:
	@echo "Running migrations..."
	go run ./cmd/migrate

seed:
	@echo "Seeding default data..."
	go run ./cmd/seed

tidy:
	go mod tidy

clean:
	rm -rf $(BUILD_DIR) coverage.out coverage.html

docker-up:
	docker compose up -d postgres redis

docker-down:
	docker compose down

.DEFAULT_GOAL := dev
