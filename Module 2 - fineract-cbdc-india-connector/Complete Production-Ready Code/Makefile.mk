.PHONY: all build test run clean docker-build docker-run lint

BINARY_NAME=server
BUILD_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")

all: test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/server/main.go
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test:
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Tests complete. Coverage report: coverage.html"

run:
	@go run cmd/server/main.go

lint:
	@echo "Running linter..."
	@golangci-lint run ./...

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

docker-build:
	@echo "Building Docker image..."
	@docker build -t fineract-cbdc-india-connector:latest -f deployments/Dockerfile .

docker-run:
	@echo "Running Docker container..."
	@docker run -p 8080:8080 \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e CBDC_API_KEY=$(CBDC_API_KEY) \
		-e CBDC_API_SECRET=$(CBDC_API_SECRET) \
		-e FINERACT_TOKEN=$(FINERACT_TOKEN) \
		fineract-cbdc-india-connector:latest

docker-compose-up:
	@docker-compose -f deployments/docker-compose.yaml up -d

docker-compose-down:
	@docker-compose -f deployments/docker-compose.yaml down

generate-swagger:
	@swag init -g cmd/server/main.go

deps:
	@go mod download
	@go mod tidy