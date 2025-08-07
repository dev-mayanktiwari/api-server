# Setup Guide

This comprehensive guide will help you set up the API Server project for development, testing, and production environments.

## 📋 Prerequisites

### Required Software

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| Go | 1.21+ | Programming language | [Download Go](https://golang.org/dl/) |
| Docker | 20.10+ | Containerization | [Install Docker](https://docs.docker.com/get-docker/) |
| Docker Compose | 2.0+ | Multi-container apps | Included with Docker Desktop |
| PostgreSQL | 15+ | Database (if running locally) | [PostgreSQL Downloads](https://www.postgresql.org/download/) |
| Redis | 7+ | Cache (if running locally) | [Redis Downloads](https://redis.io/download) |
| Git | Latest | Version control | [Git Downloads](https://git-scm.com/downloads) |

### Optional Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| kubectl | Kubernetes deployment | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| make | Build automation | Usually pre-installed on Unix systems |
| Air | Go hot reload | `go install github.com/cosmtrek/air@latest` |

## 🚀 Quick Start (Docker Compose)

The fastest way to get started:

### 1. Clone Repository
```bash
git clone <repository-url>
cd api-server
```

### 2. Start Development Environment
```bash
# Start all services with hot reload
docker-compose -f docker-compose.dev.yml up -d

# Check status
docker-compose -f docker-compose.dev.yml ps

# View logs
docker-compose -f docker-compose.dev.yml logs -f user-service
```

### 3. Test the API
```bash
# Health check
curl http://localhost:8082/health

# Register a user
curl -X POST http://localhost:8082/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### 4. Access Development Tools
- **API**: http://localhost:8082
- **pgAdmin**: http://localhost:5050 (admin@dev.com / admin)
- **RedisInsight**: http://localhost:8001

## 🔧 Manual Setup

### 1. Environment Setup

```bash
# Install Go dependencies
go mod tidy

# Setup shared module
cd shared
go mod tidy
cd ..

# Install development tools (optional)
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 2. Database Setup

#### Option A: Docker (Recommended)
```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Wait for database to be ready
docker-compose exec postgres pg_isready -U postgres
```

#### Option B: Local Installation
```bash
# Create database
createdb api_server

# Create user (optional)
psql -c "CREATE USER api_user WITH PASSWORD 'api_password';"
psql -c "GRANT ALL PRIVILEGES ON DATABASE api_server TO api_user;"
```

### 3. Initialize Database
```bash
# Run database initialization script
psql -U postgres -d api_server -f scripts/init-db.sql

# Or via Docker
docker-compose exec postgres psql -U postgres -d api_server -f /docker-entrypoint-initdb.d/init-db.sql
```

### 4. Configuration

#### Development Configuration
Create or modify `configs/development/config.yaml`:

```yaml
# Server configuration
server:
  host: "0.0.0.0"
  port: 8082
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 60s
  shutdown_timeout: 10s

# Database configuration
database:
  host: "localhost"
  port: 5432
  username: "postgres"
  password: "postgres"
  database: "api_server"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 30m

# JWT configuration
jwt:
  secret: "development-secret-key"
  issuer: "api-server"
  expiration_time: 24h
  refresh_time: 168h
  algorithm: "HS256"

# Logging configuration
logging:
  level: "debug"
  format: "console"
  output: "stdout"

# Environment
environment: "development"
service_name: "user-service"
version: "v1.0.0"
```

#### Environment Variables Override
```bash
# Database
export USER_SERVICE_DATABASE_HOST=localhost
export USER_SERVICE_DATABASE_PASSWORD=yourpassword

# JWT
export USER_SERVICE_JWT_SECRET=your-super-secret-key

# Logging
export USER_SERVICE_LOGGING_LEVEL=debug
```

### 5. Run Services

#### User Service
```bash
cd services/user-service
go run cmd/server/main.go

# Or with hot reload
air
```

#### With Make (if Makefile exists)
```bash
make run-user-service

# Or run all
make run
```

## 🧪 Testing Setup

### 1. Test Database
```bash
# Create test database
createdb api_server_test

# Or via Docker
docker-compose exec postgres createdb -U postgres api_server_test
```

### 2. Run Tests
```bash
# Unit tests
go test ./tests/unit/...

# Integration tests (requires database)
go test -tags=integration ./tests/integration/...

# All tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Test Configuration
Tests use environment variables:
```bash
export USER_SERVICE_DATABASE_HOST=localhost
export USER_SERVICE_DATABASE_DATABASE=api_server_test
export USER_SERVICE_JWT_SECRET=test-secret
```

## 🐳 Docker Development

### Development with Hot Reload
```bash
# Start development stack
docker-compose -f docker-compose.dev.yml up -d

# Rebuild after changes
docker-compose -f docker-compose.dev.yml up -d --build user-service

# View real-time logs
docker-compose -f docker-compose.dev.yml logs -f
```

### Production-like Environment
```bash
# Build production images
./scripts/docker-build.sh --version v1.0.0

# Start production stack
docker-compose -f docker-compose.prod.yml up -d
```

## 🎯 Environment-Specific Setup

### Development Environment
- **Purpose**: Local development with hot reload
- **Database**: Containerized PostgreSQL
- **Configuration**: `configs/development/config.yaml`
- **Features**: Debug logging, development tools

```bash
# Quick start
docker-compose -f docker-compose.dev.yml up -d

# Manual start
cd services/user-service && air
```

### Staging Environment
- **Purpose**: Pre-production testing
- **Database**: Managed PostgreSQL or container
- **Configuration**: `configs/staging/config.yaml`
- **Features**: Production-like setup, detailed logging

```bash
# Deploy to staging
kubectl apply -f k8s/ --namespace api-server-staging

# Or Docker Compose
ENVIRONMENT=staging docker-compose -f docker-compose.prod.yml up -d
```

### Production Environment
- **Purpose**: Live production deployment
- **Database**: Managed PostgreSQL service
- **Configuration**: `configs/production/config.yaml`
- **Features**: Optimized performance, error-only logging

```bash
# Kubernetes deployment
./scripts/deploy-kubernetes.sh deploy --namespace api-server

# Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

## 🔍 Verification Steps

### 1. Service Health
```bash
# Check service health
curl http://localhost:8082/health

# Expected response
{
  "status": "healthy",
  "service": "user-service",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### 2. Database Connection
```bash
# Check database readiness
curl http://localhost:8082/ready

# Direct database check
docker-compose exec postgres pg_isready -U postgres
```

### 3. API Functionality
```bash
# Register user
curl -X POST http://localhost:8082/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123", "first_name": "Test", "last_name": "User"}'

# Login
curl -X POST http://localhost:8082/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'
```

## 🚨 Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check what's using port 8082
lsof -i :8082

# Kill process using port
kill -9 $(lsof -ti:8082)
```

#### Database Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U postgres
```

#### Go Module Issues
```bash
# Clean module cache
go clean -modcache

# Recreate go.mod
rm go.mod go.sum
go mod init github.com/your-repo/api-server
go mod tidy
```

#### Docker Issues
```bash
# Clean Docker system
docker system prune -af

# Rebuild containers
docker-compose build --no-cache

# Check container logs
docker-compose logs service-name
```

### Log Analysis

#### Service Logs
```bash
# View user service logs
docker-compose logs -f user-service

# Filter for errors
docker-compose logs user-service | grep ERROR
```

#### Database Logs
```bash
# PostgreSQL logs
docker-compose logs postgres

# Check slow queries
docker-compose exec postgres psql -U postgres -c "SELECT * FROM pg_stat_activity WHERE state != 'idle';"
```

## 📚 Next Steps

After successful setup:

1. **Read [API Documentation](API_REFERENCE.md)** - Learn about available endpoints
2. **Review [Architecture Guide](ARCHITECTURE.md)** - Understand the system design
3. **Check [Development Guide](DEVELOPMENT_GUIDE.md)** - Learn development workflows
4. **Explore [Deployment Guide](DEPLOYMENT.md)** - Production deployment strategies

## 🤝 Getting Help

- **Check logs**: Always start with service and database logs
- **Verify configuration**: Ensure all config files are correctly set
- **Test connectivity**: Use curl or ping to test service availability
- **Review environment**: Check environment variables and their values

For additional help:
- Review error messages carefully
- Check the troubleshooting section
- Consult the project documentation
- Create an issue with detailed error logs