.PHONY: build run test docker-up docker-down migrate-up migrate-down clean

# Build the application
build:
	go build -o bin/server ./cmd/server

# Run the application locally
run:
	go run ./cmd/server

# Run tests
test:
	go test -v ./...

# Run tests with race detector
test-race:
	go test -race -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Start docker-compose
docker-up:
	docker-compose up --build

# Stop docker-compose
docker-down:
	docker-compose down

# Run migrations up (if using golang-migrate CLI - migrations are auto-run on app start)
migrate-up:
	migrate -path internal/migrations -database "postgres://avito:avito@localhost:5432/avito_db?sslmode=disable" up

# Run migrations down (if using golang-migrate CLI)
migrate-down:
	migrate -path internal/migrations -database "postgres://avito:avito@localhost:5432/avito_db?sslmode=disable" down

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Install golangci-lint
install-linter:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

