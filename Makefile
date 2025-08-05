# Variables
APP_NAME := api-server
BINARY_NAME := $(APP_NAME)
CMD_PATH := ./cmd/server

# Go related variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod

# Build flags
BUILD_FLAGS := -v
LDFLAGS := -ldflags="-s -w"

# Colors for output
GREEN := \033[32m
YELLOW := \033[33m
RED := \033[31m
RESET := \033[0m

.PHONY: all build clean test deps run dev docker-build docker-run docker-up docker-down db-setup db-migrate db-seed db-reset watch install-tools fmt lint vet security help

# Default target
all: clean deps test build

# Install dependencies
deps:
	@echo "$(GREEN)Installing dependencies...$(RESET)"
	$(GOMOD) download
	$(GOMOD) tidy

# Build the application
build:
	@echo "$(GREEN)Building $(APP_NAME)...$(RESET)"
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_PATH)

# Build for production (optimized)
build-prod:
	@echo "$(GREEN)Building $(APP_NAME) for production...$(RESET)"
	CGO_ENABLED=0 GOOS=linux $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_NAME) $(CMD_PATH)

# Run the application
run: build
	@echo "$(GREEN)Running $(APP_NAME)...$(RESET)"
	./$(BINARY_NAME)

# Run in development mode (just go run)
dev:
	@echo "$(GREEN)Running $(APP_NAME) in development mode...$(RESET)"
	$(GOCMD) run $(CMD_PATH)/main.go

# Run tests
test:
	@echo "$(GREEN)Running tests...$(RESET)"
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(YELLOW)Coverage report generated: coverage.html$(RESET)"

# Run tests and generate coverage for CI
test-ci:
	@echo "$(GREEN)Running tests for CI...$(RESET)"
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

# Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(RESET)"
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "$(GREEN)Formatting code...$(RESET)"
	$(GOCMD) fmt ./...

# Lint code (requires golangci-lint)
lint:
	@echo "$(GREEN)Linting code...$(RESET)"
	golangci-lint run

# Vet code
vet:
	@echo "$(GREEN)Vetting code...$(RESET)"
	$(GOCMD) vet ./...

# Security check (requires gosec)
security:
	@echo "$(GREEN)Running security check...$(RESET)"
	gosec ./...

# Check for outdated dependencies
check-deps:
	@echo "$(GREEN)Checking for outdated dependencies...$(RESET)"
	$(GOCMD) list -u -m all

# Update dependencies
update-deps:
	@echo "$(GREEN)Updating dependencies...$(RESET)"
	$(GOGET) -u ./...
	$(GOMOD) tidy

# Docker related commands
docker-build:
	@echo "$(GREEN)Building Docker image...$(RESET)"
	docker build -t $(APP_NAME):latest .

docker-run:
	@echo "$(GREEN)Running Docker container...$(RESET)"
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

# Docker compose commands
docker-up:
	@echo "$(GREEN)Starting services with docker-compose...$(RESET)"
	docker-compose up -d

docker-down:
	@echo "$(YELLOW)Stopping services...$(RESET)"
	docker-compose down

docker-logs:
	@echo "$(GREEN)Showing logs...$(RESET)"
	docker-compose logs -f

# Database related commands
db-setup:
	@echo "$(GREEN)Setting up database with docker-compose...$(RESET)"
	docker-compose up -d postgres redis
	@echo "$(YELLOW)Waiting for database to be ready...$(RESET)"
	sleep 10

db-migrate:
	@echo "$(GREEN)Running database migrations...$(RESET)"
	./$(BINARY_NAME) migrate

db-seed:
	@echo "$(GREEN)Seeding database with sample data...$(RESET)"
	./$(BINARY_NAME) seed

db-reset:
	@echo "$(YELLOW)Resetting database...$(RESET)"
	docker-compose down postgres
	docker volume rm api-server_postgres_data || true
	docker-compose up -d postgres
	sleep 15
	make db-migrate
	make db-seed

# Hot reload (requires air)
watch:
	@echo "$(GREEN)Starting hot reload with Air...$(RESET)"
	air

# Install development tools
install-tools:
	@echo "$(GREEN)Installing development tools...$(RESET)"
	$(GOGET) -u github.com/cosmtrek/air@latest
	$(GOGET) -u github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) -u github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Quick start for development
quick-start: deps db-setup
	@echo "$(GREEN)Quick start completed! Run 'make dev' to start the server$(RESET)"

# Full deployment
deploy:
	@echo "$(GREEN)Deploying application...$(RESET)"
	make build-prod
	make docker-build
	make docker-up

# Show help
help:
	@echo "$(GREEN)Available commands:$(RESET)"
	@echo ""
	@echo "$(GREEN)Development:$(RESET)"
	@echo "  $(YELLOW)dev$(RESET)            - Run in development mode"
	@echo "  $(YELLOW)watch$(RESET)          - Start hot reload with Air"
	@echo "  $(YELLOW)quick-start$(RESET)    - Setup database and dependencies"
	@echo ""
	@echo "$(GREEN)Build & Run:$(RESET)"
	@echo "  $(YELLOW)build$(RESET)          - Build the application"
	@echo "  $(YELLOW)build-prod$(RESET)     - Build for production"
	@echo "  $(YELLOW)run$(RESET)            - Build and run the application"
	@echo ""
	@echo "$(GREEN)Testing & Quality:$(RESET)"
	@echo "  $(YELLOW)test$(RESET)           - Run tests"
	@echo "  $(YELLOW)test-coverage$(RESET)  - Run tests with coverage"
	@echo "  $(YELLOW)fmt$(RESET)            - Format code"
	@echo "  $(YELLOW)lint$(RESET)           - Lint code"
	@echo "  $(YELLOW)vet$(RESET)            - Vet code"
	@echo "  $(YELLOW)security$(RESET)       - Run security checks"
	@echo ""
	@echo "$(GREEN)Dependencies:$(RESET)"
	@echo "  $(YELLOW)deps$(RESET)           - Install dependencies"
	@echo "  $(YELLOW)update-deps$(RESET)    - Update dependencies"
	@echo "  $(YELLOW)check-deps$(RESET)     - Check for outdated dependencies"
	@echo "  $(YELLOW)install-tools$(RESET)  - Install development tools"
	@echo ""
	@echo "$(GREEN)Docker:$(RESET)"
	@echo "  $(YELLOW)docker-build$(RESET)   - Build Docker image"
	@echo "  $(YELLOW)docker-run$(RESET)     - Run Docker container"
	@echo "  $(YELLOW)docker-up$(RESET)      - Start with docker-compose"
	@echo "  $(YELLOW)docker-down$(RESET)    - Stop docker-compose"
	@echo "  $(YELLOW)docker-logs$(RESET)    - Show docker-compose logs"
	@echo ""
	@echo "$(GREEN)Database:$(RESET)"
	@echo "  $(YELLOW)db-setup$(RESET)       - Setup database with docker-compose"
	@echo "  $(YELLOW)db-migrate$(RESET)     - Run database migrations"
	@echo "  $(YELLOW)db-seed$(RESET)        - Seed database with sample data"
	@echo "  $(YELLOW)db-reset$(RESET)       - Reset database completely"
	@echo ""
	@echo "$(GREEN)Utilities:$(RESET)"
	@echo "  $(YELLOW)clean$(RESET)          - Clean build artifacts"
	@echo "  $(YELLOW)deploy$(RESET)         - Full deployment"
	@echo "  $(YELLOW)help$(RESET)           - Show this help message"