# API Server Makefile
# Production-ready build automation for microservices architecture

# Variables
SHELL := /bin/bash
.DEFAULT_GOAL := help

# Project information
PROJECT_NAME := api-server
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go variables
GO_VERSION := 1.23.2
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0

# Directories
BIN_DIR := bin
BUILD_DIR := build
COVERAGE_DIR := coverage
DOCS_DIR := docs
TOOLS_DIR := tools

# Services
SERVICES := api-gateway auth-service user-service
SERVICE_PORTS := 8080 8081 8082

# Docker variables
DOCKER_REGISTRY := ghcr.io/dev-mayanktiwari
DOCKER_TAG := $(VERSION)
DOCKERFILE_DIR := deployments/docker

# Build flags
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT_HASH) -X main.date=$(BUILD_DATE)"
BUILD_FLAGS := $(LDFLAGS) -trimpath

# Test flags
TEST_FLAGS := -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic
TEST_PACKAGES := ./...
TEST_TIMEOUT := 300s

# Colors for output
RED := \033[31m
GREEN := \033[32m
YELLOW := \033[33m
BLUE := \033[34m
MAGENTA := \033[35m
CYAN := \033[36m
WHITE := \033[37m
RESET := \033[0m

##@ General

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\n$(BLUE)Usage:$(RESET)\n  make $(CYAN)<target>$(RESET)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(CYAN)%-15s$(RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: version
version: ## Show version information
	@echo "$(GREEN)Project:$(RESET) $(PROJECT_NAME)"
	@echo "$(GREEN)Version:$(RESET) $(VERSION)"
	@echo "$(GREEN)Commit:$(RESET) $(COMMIT_HASH)"
	@echo "$(GREEN)Build Date:$(RESET) $(BUILD_DATE)"
	@echo "$(GREEN)Go Version:$(RESET) $(GO_VERSION)"
	@echo "$(GREEN)OS/Arch:$(RESET) $(GOOS)/$(GOARCH)"

##@ Development

.PHONY: setup-dev
setup-dev: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(RESET)"
	@mkdir -p $(BIN_DIR) $(BUILD_DIR) $(COVERAGE_DIR)
	@$(MAKE) install-tools
	@$(MAKE) deps
	@$(MAKE) setup-hooks
	@echo "$(GREEN)Development environment ready!$(RESET)"

.PHONY: deps
deps: ## Install dependencies
	@echo "$(GREEN)Installing dependencies...$(RESET)"
	@go mod download
	@go mod tidy
	@cd shared && go mod download && go mod tidy
	@for service in $(SERVICES); do \
		echo "Installing dependencies for $$service..."; \
		cd services/$$service && go mod download && go mod tidy && cd ../..; \
	done

.PHONY: update-deps
update-deps: ## Update dependencies
	@echo "$(GREEN)Updating dependencies...$(RESET)"
	@go get -u ./...
	@go mod tidy
	@cd shared && go get -u ./... && go mod tidy && cd ..
	@for service in $(SERVICES); do \
		echo "Updating dependencies for $$service..."; \
		cd services/$$service && go get -u ./... && go mod tidy && cd ../..; \
	done

.PHONY: install-tools
install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(RESET)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/sast-scan/cmd/gosec@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/cosmtrek/air@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest

.PHONY: setup-hooks
setup-hooks: ## Setup git pre-commit hooks
	@echo "$(GREEN)Setting up git hooks...$(RESET)"
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'make pre-commit' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "$(GREEN)Pre-commit hooks installed!$(RESET)"

##@ Building

.PHONY: build
build: ## Build all services
	@echo "$(GREEN)Building all services...$(RESET)"
	@$(MAKE) build-shared
	@for service in $(SERVICES); do \
		$(MAKE) build-service SERVICE=$$service; \
	done

.PHONY: build-shared
build-shared: ## Build shared library
	@echo "$(GREEN)Building shared library...$(RESET)"
	@cd shared && go build $(BUILD_FLAGS) ./...

.PHONY: build-service
build-service: ## Build specific service (requires SERVICE variable)
	@if [ -z "$(SERVICE)" ]; then echo "$(RED)SERVICE variable is required$(RESET)"; exit 1; fi
	@echo "$(GREEN)Building $(SERVICE)...$(RESET)"
	@cd services/$(SERVICE) && \
		CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(BUILD_FLAGS) -o ../../$(BIN_DIR)/$(SERVICE) ./cmd/server

.PHONY: build-monolith
build-monolith: ## Build monolith version for development
	@echo "$(GREEN)Building monolith...$(RESET)"
	@CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GOARCH) \
		go build $(BUILD_FLAGS) -o $(BIN_DIR)/monolith ./cmd/monolith

.PHONY: build-prod
build-prod: ## Build for production (all platforms)
	@echo "$(GREEN)Building for production...$(RESET)"
	@mkdir -p $(BUILD_DIR)
	@for service in $(SERVICES); do \
		for os in linux darwin windows; do \
			for arch in amd64 arm64; do \
				if [ "$$os" = "windows" ]; then ext=".exe"; else ext=""; fi; \
				echo "Building $$service for $$os/$$arch..."; \
				cd services/$$service && \
				CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build $(BUILD_FLAGS) -o ../../$(BUILD_DIR)/$$service-$$os-$$arch$$ext ./cmd/server && \
				cd ../..; \
			done; \
		done; \
	done

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning build artifacts...$(RESET)"
	@rm -rf $(BIN_DIR) $(BUILD_DIR) $(COVERAGE_DIR)
	@go clean -cache
	@docker system prune -f

##@ Running

.PHONY: run
run: build-monolith ## Build and run monolith
	@echo "$(GREEN)Starting monolith...$(RESET)"
	@ENVIRONMENT=development ./$(BIN_DIR)/monolith

.PHONY: run-service
run-service: ## Run specific service (requires SERVICE variable)
	@if [ -z "$(SERVICE)" ]; then echo "$(RED)SERVICE variable is required$(RESET)"; exit 1; fi
	@$(MAKE) build-service SERVICE=$(SERVICE)
	@echo "$(GREEN)Starting $(SERVICE)...$(RESET)"
	@ENVIRONMENT=development ./$(BIN_DIR)/$(SERVICE)

.PHONY: run-all
run-all: ## Run all services in background
	@echo "$(GREEN)Starting all services...$(RESET)"
	@for service in $(SERVICES); do \
		$(MAKE) build-service SERVICE=$$service; \
		ENVIRONMENT=development ./$(BIN_DIR)/$$service & \
		echo "Started $$service"; \
	done
	@echo "$(GREEN)All services started!$(RESET)"
	@echo "$(YELLOW)Use 'make stop-all' to stop all services$(RESET)"

.PHONY: stop-all
stop-all: ## Stop all running services
	@echo "$(GREEN)Stopping all services...$(RESET)"
	@for service in $(SERVICES); do \
		pkill -f "./$(BIN_DIR)/$$service" || true; \
	done

.PHONY: dev
dev: ## Run in development mode with hot reload
	@echo "$(GREEN)Starting development mode with hot reload...$(RESET)"
	@air -c .air.toml

##@ Testing

.PHONY: test
test: ## Run all tests
	@echo "$(GREEN)Running tests...$(RESET)"
	@mkdir -p $(COVERAGE_DIR)
	@go test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) $(TEST_PACKAGES)
	@$(MAKE) test-shared
	@for service in $(SERVICES); do \
		$(MAKE) test-service SERVICE=$$service; \
	done

.PHONY: test-shared
test-shared: ## Test shared library
	@echo "$(GREEN)Testing shared library...$(RESET)"
	@cd shared && go test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-service
test-service: ## Test specific service (requires SERVICE variable)
	@if [ -z "$(SERVICE)" ]; then echo "$(RED)SERVICE variable is required$(RESET)"; exit 1; fi
	@echo "$(GREEN)Testing $(SERVICE)...$(RESET)"
	@cd services/$(SERVICE) && go test $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./...

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(RESET)"
	@go test -tags=integration $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./tests/...

.PHONY: test-e2e
test-e2e: ## Run end-to-end tests
	@echo "$(GREEN)Running end-to-end tests...$(RESET)"
	@docker-compose -f docker-compose.test.yml up -d
	@sleep 10
	@go test -tags=e2e $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) ./tests/e2e/...
	@docker-compose -f docker-compose.test.yml down

.PHONY: test-coverage
test-coverage: test ## Generate test coverage report
	@echo "$(GREEN)Generating coverage report...$(RESET)"
	@go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@go tool cover -func=$(COVERAGE_DIR)/coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@echo "$(GREEN)Coverage report generated: $(COVERAGE_DIR)/coverage.html$(RESET)"

.PHONY: test-unit
test-unit: ## Run only unit tests
	@echo "$(GREEN)Running unit tests...$(RESET)"
	@go test -short $(TEST_FLAGS) -timeout $(TEST_TIMEOUT) $(TEST_PACKAGES)

.PHONY: benchmark
benchmark: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(RESET)"
	@go test -bench=. -benchmem -run=^Benchmark $(TEST_PACKAGES)

##@ Quality Assurance

.PHONY: fmt
fmt: ## Format code
	@echo "$(GREEN)Formatting code...$(RESET)"
	@go fmt ./...
	@goimports -w .
	@cd shared && go fmt ./... && goimports -w . && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && go fmt ./... && goimports -w . && cd ../..; \
	done

.PHONY: lint
lint: ## Lint code
	@echo "$(GREEN)Linting code...$(RESET)"
	@golangci-lint run ./...
	@cd shared && golangci-lint run ./... && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && golangci-lint run ./... && cd ../..; \
	done

.PHONY: vet
vet: ## Vet code
	@echo "$(GREEN)Vetting code...$(RESET)"
	@go vet ./...
	@cd shared && go vet ./... && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && go vet ./... && cd ../..; \
	done

.PHONY: security
security: ## Run security checks
	@echo "$(GREEN)Running security checks...$(RESET)"
	@gosec -quiet ./...
	@cd shared && gosec -quiet ./... && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && gosec -quiet ./... && cd ../..; \
	done

.PHONY: pre-commit
pre-commit: fmt vet lint security test-unit ## Run pre-commit checks
	@echo "$(GREEN)All pre-commit checks passed!$(RESET)"

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker images for all services
	@echo "$(GREEN)Building Docker images...$(RESET)"
	@for service in $(SERVICES); do \
		echo "Building $$service image..."; \
		docker build -f services/$$service/Dockerfile -t $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG) .; \
		docker tag $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG) $(DOCKER_REGISTRY)/$$service:latest; \
	done

.PHONY: docker-build-service
docker-build-service: ## Build Docker image for specific service (requires SERVICE variable)
	@if [ -z "$(SERVICE)" ]; then echo "$(RED)SERVICE variable is required$(RESET)"; exit 1; fi
	@echo "$(GREEN)Building $(SERVICE) Docker image...$(RESET)"
	@docker build -f services/$(SERVICE)/Dockerfile -t $(DOCKER_REGISTRY)/$(SERVICE):$(DOCKER_TAG) .
	@docker tag $(DOCKER_REGISTRY)/$(SERVICE):$(DOCKER_TAG) $(DOCKER_REGISTRY)/$(SERVICE):latest

.PHONY: docker-push
docker-push: ## Push Docker images to registry
	@echo "$(GREEN)Pushing Docker images...$(RESET)"
	@for service in $(SERVICES); do \
		echo "Pushing $$service image..."; \
		docker push $(DOCKER_REGISTRY)/$$service:$(DOCKER_TAG); \
		docker push $(DOCKER_REGISTRY)/$$service:latest; \
	done

.PHONY: docker-up
docker-up: ## Start services with docker-compose
	@echo "$(GREEN)Starting services with docker-compose...$(RESET)"
	@docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop services with docker-compose
	@echo "$(GREEN)Stopping services with docker-compose...$(RESET)"
	@docker-compose down

.PHONY: docker-logs
docker-logs: ## Show docker-compose logs
	@docker-compose logs -f

.PHONY: docker-clean
docker-clean: ## Clean Docker images and volumes
	@echo "$(GREEN)Cleaning Docker images and volumes...$(RESET)"
	@docker-compose down -v
	@docker system prune -f
	@docker volume prune -f

##@ Database

.PHONY: db-setup
db-setup: ## Setup database with docker-compose
	@echo "$(GREEN)Setting up database...$(RESET)"
	@docker-compose up -d postgres
	@echo "$(GREEN)Waiting for database to be ready...$(RESET)"
	@sleep 10
	@$(MAKE) migrate

.PHONY: migrate
migrate: ## Run database migrations
	@echo "$(GREEN)Running database migrations...$(RESET)"
	@go run ./cmd/migrate

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	@echo "$(GREEN)Rolling back database migrations...$(RESET)"
	@go run ./cmd/migrate -down

.PHONY: db-reset
db-reset: ## Reset database completely
	@echo "$(GREEN)Resetting database...$(RESET)"
	@docker-compose down -v postgres
	@docker-compose up -d postgres
	@sleep 10
	@$(MAKE) migrate

.PHONY: db-shell
db-shell: ## Open database shell
	@docker-compose exec postgres psql -U postgres -d api_server

##@ Documentation

.PHONY: docs
docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(RESET)"
	@$(MAKE) docs-api
	@$(MAKE) docs-swagger

.PHONY: docs-api
docs-api: ## Generate API documentation
	@echo "$(GREEN)Generating API documentation...$(RESET)"
	@swag init -g cmd/server/main.go -o docs/swagger

.PHONY: docs-swagger
docs-swagger: ## Start Swagger UI server
	@echo "$(GREEN)Starting Swagger UI...$(RESET)"
	@docker run -d -p 8090:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(PWD)/docs/swagger:/docs swaggerapi/swagger-ui

.PHONY: docs-serve
docs-serve: ## Serve documentation locally
	@echo "$(GREEN)Serving documentation...$(RESET)"
	@cd docs && python3 -m http.server 8080

##@ Deployment

.PHONY: deploy-dev
deploy-dev: ## Deploy to development environment
	@echo "$(GREEN)Deploying to development...$(RESET)"
	@docker-compose -f deployments/docker-compose/development.yml up -d

.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	@echo "$(GREEN)Deploying to staging...$(RESET)"
	@docker-compose -f deployments/docker-compose/staging.yml up -d

.PHONY: deploy-prod
deploy-prod: build-prod docker-build docker-push ## Deploy to production
	@echo "$(GREEN)Deploying to production...$(RESET)"
	@kubectl apply -f deployments/k8s/

.PHONY: k8s-deploy
k8s-deploy: ## Deploy to Kubernetes
	@echo "$(GREEN)Deploying to Kubernetes...$(RESET)"
	@kubectl apply -f deployments/k8s/

.PHONY: k8s-delete
k8s-delete: ## Delete from Kubernetes
	@echo "$(GREEN)Deleting from Kubernetes...$(RESET)"
	@kubectl delete -f deployments/k8s/

##@ Monitoring

.PHONY: health
health: ## Check health of all services
	@echo "$(GREEN)Checking health of all services...$(RESET)"
	@for port in $(SERVICE_PORTS); do \
		echo "Checking service on port $$port..."; \
		curl -s http://localhost:$$port/health | jq . || echo "Service on port $$port is not responding"; \
	done

.PHONY: logs
logs: ## Show logs from all services
	@echo "$(GREEN)Showing logs...$(RESET)"
	@docker-compose logs -f

.PHONY: metrics
metrics: ## Show metrics
	@echo "$(GREEN)Showing metrics...$(RESET)"
	@curl -s http://localhost:9090/metrics

##@ Utilities

.PHONY: generate
generate: ## Run go generate
	@echo "$(GREEN)Running go generate...$(RESET)"
	@go generate ./...
	@cd shared && go generate ./... && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && go generate ./... && cd ../..; \
	done

.PHONY: tidy
tidy: ## Tidy modules
	@echo "$(GREEN)Tidying modules...$(RESET)"
	@go mod tidy
	@cd shared && go mod tidy && cd ..
	@for service in $(SERVICES); do \
		cd services/$$service && go mod tidy && cd ../..; \
	done

.PHONY: install-local
install-local: build ## Install binaries locally
	@echo "$(GREEN)Installing binaries locally...$(RESET)"
	@for service in $(SERVICES); do \
		cp $(BIN_DIR)/$$service $(GOPATH)/bin/; \
	done

.PHONY: quick-start
quick-start: setup-dev db-setup ## Quick start for new developers
	@echo "$(GREEN)Quick start completed!$(RESET)"
	@echo "$(YELLOW)Run 'make run' to start the monolith or 'make run-all' to start all services$(RESET)"

# Service-specific targets
.PHONY: run-gateway run-auth run-user
run-gateway: ## Run API Gateway
	@$(MAKE) run-service SERVICE=api-gateway

run-auth: ## Run Auth Service
	@$(MAKE) run-service SERVICE=auth-service

run-user: ## Run User Service
	@$(MAKE) run-service SERVICE=user-service