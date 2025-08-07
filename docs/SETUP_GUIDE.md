# Setup Guide

Complete guide to set up the API Server microservices application for development, testing, and production environments.

## 📋 Prerequisites

### Required Software

| Tool | Version | Purpose | Installation |
|------|---------|---------|--------------|
| Go | 1.21+ | Programming language | [Download Go](https://golang.org/dl/) |
| Docker | 20.10+ | Containerization | [Install Docker](https://docs.docker.com/get-docker/) |
| Docker Compose | 2.0+ | Multi-container orchestration | Included with Docker Desktop |
| Git | Latest | Version control | [Git Downloads](https://git-scm.com/downloads) |

### Optional Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| kubectl | Kubernetes deployment | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| make | Build automation | Usually pre-installed on Unix systems |
| Air | Go hot reload | `go install github.com/cosmtrek/air@latest` |

## 🚀 Quick Start (Recommended)

The fastest way to get the complete microservices stack running:

### 1. Clone Repository
```bash
git clone <repository-url>
cd api-server
```

### 2. Start Complete Microservices Stack
```bash
# Start all services with load balancer
docker-compose -f docker-compose.microservices.yml up -d

# Check all services status
docker-compose -f docker-compose.microservices.yml ps

# View logs for all services
docker-compose -f docker-compose.microservices.yml logs -f
```

### 3. Verify Services Are Running
```bash
# Check load balancer (main entry point)
curl http://localhost/health

# Check individual services
curl http://localhost:8080/health  # API Gateway
curl http://localhost:8081/health  # Auth Service
curl http://localhost:8082/health  # User Service
```

### 4. Test the API
```bash
# Register a new user (via load balancer → API Gateway → User Service)
curl -X POST http://localhost/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }'

# Login to get JWT token (via load balancer → API Gateway → Auth Service)
curl -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# Use JWT token to get profile (via load balancer → API Gateway → User Service)
curl -X GET http://localhost/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 5. Access Development Tools
- **Application**: http://localhost (Nginx Load Balancer)
- **API Gateway**: http://localhost:8080
- **Auth Service**: http://localhost:8081
- **User Service**: http://localhost:8082
- **pgAdmin**: http://localhost:5050 (admin@dev.com / admin)
- **Redis Insight**: http://localhost:8001

## 🏗️ Manual Development Setup

For active development with code changes:

### 1. Environment Setup
```bash
# Install Go dependencies for all services
go mod tidy

# Setup shared module dependencies
cd shared && go mod tidy && cd ..

# Setup each service
cd services/api-gateway && go mod tidy && cd ../..
cd services/auth-service && go mod tidy && cd ../..
cd services/user-service && go mod tidy && cd ../..

# Install development tools
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 2. Start Infrastructure Services
```bash
# Start PostgreSQL, Redis, and Nginx
docker-compose -f docker-compose.microservices.yml up -d postgres redis nginx

# Wait for database to be ready
docker-compose -f docker-compose.microservices.yml exec postgres pg_isready -U postgres
```

### 3. Initialize Database
```bash
# Run database initialization script
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server -f /docker-entrypoint-initdb.d/init-db.sql

# Verify database setup
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server -c "\dt"
```

### 4. Run Services Manually (with Hot Reload)

#### Terminal 1: User Service
```bash
cd services/user-service
air  # or go run cmd/server/main.go
```

#### Terminal 2: Auth Service
```bash
cd services/auth-service
air  # or go run cmd/server/main.go
```

#### Terminal 3: API Gateway
```bash
cd services/api-gateway
air  # or go run cmd/server/main.go
```

## 🐳 Docker Development Environments

### Development with Hot Reload
```bash
# Development stack with hot reload enabled
docker-compose -f docker-compose.dev.yml up -d

# Rebuild specific service after changes
docker-compose -f docker-compose.dev.yml up -d --build user-service

# View real-time logs
docker-compose -f docker-compose.dev.yml logs -f
```

### Production-like Environment
```bash
# Build production-optimized images
./scripts/docker-build.sh

# Start production stack
docker-compose -f docker-compose.microservices.yml up -d
```

## ⚙️ Configuration

### Environment-Specific Configurations

#### API Gateway Configuration
File: `configs/api-gateway/config.yaml`
```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 30s
  write_timeout: 30s

services:
  auth_service_url: "http://auth-service:8081"
  user_service_url: "http://user-service:8082"

jwt:
  secret: "your-jwt-secret-key"
  issuer: "api-server"

rate_limiting:
  requests_per_second: 100
  burst: 200
```

#### Auth Service Configuration
File: `configs/auth-service/config.yaml`
```yaml
server:
  host: "0.0.0.0"
  port: 8081

database:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "postgres"
  database: "api_server"
  sslmode: "disable"

jwt:
  secret: "your-jwt-secret-key"
  expiration_time: "24h"
  refresh_time: "168h"
```

#### User Service Configuration
File: `configs/development/config.yaml`
```yaml
server:
  host: "0.0.0.0"
  port: 8082

database:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "postgres"
  database: "api_server"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5

jwt:
  secret: "your-jwt-secret-key"
  issuer: "api-server"

logging:
  level: "debug"
  format: "console"
```

### Environment Variables Override
```bash
# Database configuration
export USER_SERVICE_DATABASE_HOST=localhost
export USER_SERVICE_DATABASE_PASSWORD=yourpassword
export AUTH_SERVICE_DATABASE_HOST=localhost

# JWT configuration
export USER_SERVICE_JWT_SECRET=your-super-secret-key
export AUTH_SERVICE_JWT_SECRET=your-super-secret-key
export API_GATEWAY_JWT_SECRET=your-super-secret-key

# Service URLs (for development)
export API_GATEWAY_AUTH_SERVICE_URL=http://localhost:8081
export API_GATEWAY_USER_SERVICE_URL=http://localhost:8082

# Logging
export USER_SERVICE_LOGGING_LEVEL=debug
export AUTH_SERVICE_LOGGING_LEVEL=debug
```

## 🧪 Testing Setup

### 1. Test Database Setup
```bash
# Create test database
docker-compose -f docker-compose.microservices.yml exec postgres \
  createdb -U postgres api_server_test

# Initialize test database
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server_test -f /docker-entrypoint-initdb.d/init-db.sql
```

### 2. Run Tests
```bash
# Unit tests for all services
go test ./services/user-service/tests/unit/...
go test ./services/auth-service/tests/unit/...
go test ./services/api-gateway/tests/unit/...

# Integration tests (requires running database)
go test -tags=integration ./services/user-service/tests/integration/...
go test -tags=integration ./services/auth-service/tests/integration/...

# All tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 3. Test Environment Variables
```bash
# Set test-specific environment
export USER_SERVICE_DATABASE_DATABASE=api_server_test
export AUTH_SERVICE_DATABASE_DATABASE=api_server_test
export USER_SERVICE_JWT_SECRET=test-secret
export AUTH_SERVICE_JWT_SECRET=test-secret
```

## ☸️ Kubernetes Deployment

### 1. Local Kubernetes Setup
```bash
# Using kind (Kubernetes in Docker)
kind create cluster --name api-server-cluster

# Or using minikube
minikube start --cpus=4 --memory=8192
```

### 2. Deploy to Kubernetes
```bash
# Deploy all services and infrastructure
kubectl apply -f k8s/

# Check deployment status
kubectl get pods -l app=api-server

# Access services via port forwarding
kubectl port-forward svc/api-gateway 8080:8080
kubectl port-forward svc/user-service 8082:8082
```

### 3. Production Kubernetes Setup
```bash
# Deploy with production configuration
./scripts/deploy-kubernetes.sh --environment production

# Scale services based on load
kubectl scale deployment api-gateway --replicas=3
kubectl scale deployment auth-service --replicas=2
kubectl scale deployment user-service --replicas=3
```

## 🔍 Verification & Testing

### 1. Health Checks
```bash
# Check all services via load balancer
curl http://localhost/health

# Check API Gateway health
curl http://localhost:8080/health

# Check Auth Service health
curl http://localhost:8081/health

# Check User Service health
curl http://localhost:8082/health

# Check service readiness
curl http://localhost:8080/ready
curl http://localhost:8081/ready
curl http://localhost:8082/ready
```

### 2. Database Connectivity
```bash
# Check database connection
docker-compose -f docker-compose.microservices.yml exec postgres pg_isready -U postgres

# Connect to database
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server

# Check Redis connectivity
docker-compose -f docker-compose.microservices.yml exec redis redis-cli ping
```

### 3. Service Communication
```bash
# Test full authentication flow
# 1. Register user (API Gateway → User Service)
USER_RESPONSE=$(curl -s -X POST http://localhost/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User"
  }')

# 2. Login (API Gateway → Auth Service → User Service)
TOKEN_RESPONSE=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

# 3. Extract token and test protected endpoint
TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.data.token')
curl -X GET http://localhost/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN"
```

## 🚨 Troubleshooting

### Common Issues

#### Port Conflicts
```bash
# Check what's using the ports
lsof -i :80    # Nginx
lsof -i :8080  # API Gateway
lsof -i :8081  # Auth Service
lsof -i :8082  # User Service

# Kill processes if needed
sudo kill -9 $(lsof -ti:80)
```

#### Service Connection Issues
```bash
# Check Docker networks
docker network ls
docker network inspect api-server_default

# Check service logs
docker-compose -f docker-compose.microservices.yml logs api-gateway
docker-compose -f docker-compose.microservices.yml logs auth-service
docker-compose -f docker-compose.microservices.yml logs user-service
```

#### Database Issues
```bash
# Check PostgreSQL status
docker-compose -f docker-compose.microservices.yml ps postgres

# View database logs
docker-compose -f docker-compose.microservices.yml logs postgres

# Test database connection
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -c "SELECT version();"
```

#### Go Module Issues
```bash
# Clean Go module cache
go clean -modcache

# Update dependencies
go mod tidy
go mod download

# Rebuild vendor directory
go mod vendor
```

### Log Analysis
```bash
# Follow logs for all services
docker-compose -f docker-compose.microservices.yml logs -f

# Filter logs for specific service
docker-compose -f docker-compose.microservices.yml logs user-service | grep ERROR

# Check service startup logs
docker-compose -f docker-compose.microservices.yml logs --tail=50 api-gateway
```

## 🎯 Environment-Specific Setup

### Development Environment
- **Purpose**: Local development with hot reload
- **Database**: Containerized PostgreSQL and Redis
- **Configuration**: Debug logging, development tools
- **Services**: All services with hot reload enabled

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# Run services manually with hot reload
cd services/user-service && air
```

### Staging Environment
- **Purpose**: Pre-production testing
- **Database**: Managed services or containers
- **Configuration**: Production-like settings with detailed logging
- **Services**: Containerized services with resource limits

```bash
# Deploy to staging
kubectl apply -f k8s/ --namespace api-server-staging

# Or with Docker Compose
ENVIRONMENT=staging docker-compose -f docker-compose.microservices.yml up -d
```

### Production Environment
- **Purpose**: Live production deployment
- **Database**: Managed PostgreSQL and Redis services
- **Configuration**: Optimized performance, error-only logging
- **Services**: Auto-scaling, load balancing, monitoring

```bash
# Deploy to production Kubernetes
./scripts/deploy-kubernetes.sh --environment production --namespace api-server

# Or with Docker Swarm
docker stack deploy -c docker-compose.microservices.yml api-server
```

## 📚 Next Steps

After successful setup:

1. **[Development Guide](DEVELOPMENT_GUIDE.md)** - Learn development workflows and coding standards
2. **[API Reference](API_REFERENCE.md)** - Explore all available endpoints and data models
3. **[Database Guide](DATABASE_GUIDE.md)** - Database operations and management
4. **[Project Overview](PROJECT_OVERVIEW.md)** - Understand the architecture and design decisions

## 🤝 Getting Help

If you encounter issues:

1. **Check Logs**: Always start with service logs using `docker-compose logs`
2. **Verify Configuration**: Ensure all config files and environment variables are correct
3. **Test Connectivity**: Use curl to test service endpoints and health checks
4. **Database Status**: Verify PostgreSQL and Redis are running and accessible
5. **Port Availability**: Ensure required ports are not in use by other services

For persistent issues, review the troubleshooting section or check the service-specific logs for detailed error information.