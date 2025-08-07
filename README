# API Server - Production-Ready Microservices

A **production-ready microservices application** built with **Go**, implementing **Clean Architecture** principles across multiple services. Features complete service separation, JWT authentication, enterprise-grade security, and cloud-native deployment capabilities.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![Architecture](https://img.shields.io/badge/architecture-microservices-green.svg)](docs/PROJECT_OVERVIEW.md)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](https://docker.com)
[![Kubernetes](https://img.shields.io/badge/kubernetes-ready-326ce5.svg)](https://kubernetes.io)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## ✨ Enterprise Features

- **🏗️ True Microservices Architecture**: 4 independent services with clear boundaries
- **🔐 Dedicated Auth Service** (Port 8081): JWT management, login/logout, session handling
- **👥 Dedicated User Service** (Port 8082): User CRUD operations, profile management  
- **🌐 API Gateway** (Port 8080): Request routing, authentication, rate limiting
- **⚖️ Load Balancer** (Nginx): SSL termination, traffic distribution, health routing
- **📦 Zero Code Duplication**: Shared libraries for common functionality
- **🔄 Service Communication**: HTTP-based with user context propagation
- **📊 Enterprise Monitoring**: Health checks, structured logging, audit trails
- **🚀 Container Native**: Individual optimized Docker containers per service
- **☸️ Kubernetes Ready**: Complete production manifests with auto-scaling

## 🏗️ Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Balancer (Nginx)                    │
│                        Port: 80                             │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  API Gateway                                │
│                  Port: 8080                                 │
│  • Request Routing   • JWT Validation   • Rate Limiting     │
└──────────┬─────────────────────┬────────────────────────────┘
           │                     │
┌──────────▼──────────┐ ┌───────▼──────────────────────────────┐
│   Auth Service      │ │         User Service                 │
│   Port: 8081        │ │         Port: 8082                   │
│ • JWT Generation    │ │ • User CRUD Operations               │
│ • Token Validation  │ │ • Profile Management                 │  
│ • Token Refresh     │ │ • Password Changes                   │
│ • User Login        │ │ • User Registration                  │
└─────────────────────┘ └──────────────────────────────────────┘
           │                     │
           └─────────┬───────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│                Shared Database                              │
│              PostgreSQL + Redis                             │
└─────────────────────────────────────────────────────────────┘
```

## 📦 Service Architecture

```
api-server/
├── services/                     # Microservices
│   ├── api-gateway/             # API Gateway & Request Routing  
│   │   ├── cmd/server/          # Gateway entry point
│   │   ├── internal/
│   │   │   ├── application/     # Proxy services
│   │   │   └── infrastructure/  # HTTP handlers
│   │   └── Dockerfile
│   ├── auth-service/            # Authentication & JWT Management
│   │   ├── cmd/server/          # Auth service entry point
│   │   ├── internal/
│   │   │   ├── application/     # Auth business logic
│   │   │   ├── domain/          # Auth entities
│   │   │   └── infrastructure/  # Database & HTTP
│   │   └── Dockerfile
│   └── user-service/            # User Management (CRUD Only)
│       ├── cmd/server/          # User service entry point
│       ├── internal/
│       │   ├── application/     # User business logic
│       │   ├── domain/          # User entities
│       │   └── infrastructure/  # Database & HTTP
│       └── Dockerfile
├── shared/                      # Shared Libraries (No Duplication)
│   └── pkg/                     # Common utilities
│       ├── auth/               # JWT utilities
│       ├── config/             # Configuration management
│       ├── database/           # Database connections
│       ├── logger/             # Structured logging
│       ├── middleware/         # HTTP middleware
│       └── response/           # API response utilities
├── configs/                     # Service-specific configurations
│   ├── api-gateway/            # Gateway configuration
│   ├── auth-service/           # Auth service configuration
│   ├── nginx/                  # Nginx load balancer config
│   └── development/            # User service configuration
├── k8s/                        # Kubernetes manifests
├── scripts/                    # Database migrations & scripts
└── docs/                       # Complete documentation
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Git** - [Install Git](https://git-scm.com/downloads)

### 🐳 Complete Microservices Stack (Recommended)

```bash
# 1. Clone the repository
git clone <repository-url>
cd api-server

# 2. Start complete microservices stack (all services + infrastructure)
docker-compose -f docker-compose.microservices.yml up -d

# 3. Verify all services are running
docker-compose -f docker-compose.microservices.yml ps

# 4. Test the complete system via load balancer
curl http://localhost/health
```

**🎯 Access Points:**
- **🌐 Main Application**: http://localhost (Nginx Load Balancer)
- **🔧 API Gateway**: http://localhost:8080 (Direct access)
- **🔐 Auth Service**: http://localhost:8081 (Direct access)
- **👥 User Service**: http://localhost:8082 (Direct access)
- **🗄️ Database Admin**: http://localhost:5050 (admin@dev.com / admin)
- **⚡ Redis Admin**: http://localhost:8001

### 📡 API Endpoints

All requests go through the **API Gateway** via **Load Balancer**:

#### **Authentication (Auth Service)**
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/auth/login` | User login & get JWT | No |
| POST | `/api/v1/auth/refresh` | Refresh JWT token | No |
| POST | `/api/v1/auth/logout` | Logout & invalidate token | No |
| POST | `/api/v1/auth/validate` | Validate JWT token | No |
| GET | `/api/v1/auth/me` | Get current user info | Yes |

#### **User Management (User Service)**
| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| POST | `/api/v1/users/register` | Register new user | No |
| GET | `/api/v1/users/profile` | Get user profile | Yes (User) |
| PUT | `/api/v1/users/profile` | Update user profile | Yes (User) |
| POST | `/api/v1/users/change-password` | Change password | Yes (User) |
| GET | `/api/v1/users` | List users (admin) | Yes (Admin) |
| GET | `/api/v1/users/{id}` | Get user by ID | Yes (Admin) |
| PUT | `/api/v1/users/{id}` | Update user | Yes (Admin) |
| DELETE | `/api/v1/users/{id}` | Delete user | Yes (Admin) |

### 🔑 Complete Authentication Flow

```bash
# 1. Register a new user (Load Balancer → API Gateway → User Service)
curl -X POST http://localhost/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'

# 2. Login to get JWT tokens (Load Balancer → API Gateway → Auth Service)
TOKEN_RESPONSE=$(curl -s -X POST http://localhost/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com", 
    "password": "password123"
  }')

# 3. Extract token and test protected endpoint
TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.data.access_token')
curl -X GET http://localhost/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN"

# 4. Test admin endpoints (requires admin user)
curl -X GET http://localhost/api/v1/users/users?page=1&limit=10 \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

## 🏢 Service Responsibilities

### 🌐 API Gateway (`port 8080`)
- **Request Routing**: Routes requests to appropriate services
- **Authentication**: Validates JWT tokens for protected routes
- **Rate Limiting**: Implements per-client rate limiting
- **Load Balancing**: Distributes requests across service instances
- **CORS Handling**: Manages cross-origin requests

### 🔐 Auth Service (`port 8081`)
- **User Authentication**: Validates login credentials
- **JWT Management**: Generates, validates, and refreshes JWT tokens
- **Token Storage**: Manages refresh tokens in database
- **Session Management**: Handles user sessions and logout

### 👥 User Service (`port 8082`)
- **User Registration**: Creates new user accounts
- **Profile Management**: Handles user CRUD operations
- **Password Management**: Manages password changes
- **User Administration**: Admin user management operations
- **Credential Validation**: Validates user credentials for auth service

### ⚖️ Load Balancer (Nginx)
- **SSL Termination**: Handles HTTPS certificates
- **Load Distribution**: Routes traffic to API Gateway instances
- **Static Content**: Serves static files if needed
- **Health Checks**: Monitors service health

## 🧪 Testing

```bash
# Test individual services
make test-auth-service
make test-user-service
make test-api-gateway

# Run all tests
make test-all

# Integration tests with running services
make test-integration
```

## 🚀 Deployment Options

### 🐳 Docker Development
```bash
# Development with hot reload
docker-compose -f docker-compose.dev.yml up -d

# Production build
docker-compose -f docker-compose.microservices.yml up -d
```

### ☸️ Kubernetes Production
```bash
# Deploy all services to Kubernetes
kubectl apply -f k8s/

# Scale individual services
kubectl scale deployment auth-service --replicas=3
kubectl scale deployment user-service --replicas=5
kubectl scale deployment api-gateway --replicas=2
```

## 📊 Monitoring & Health Checks

Each service provides comprehensive health endpoints:

```bash
# Service health (via load balancer)
curl http://localhost/services/auth/health
curl http://localhost/services/user/health

# Direct service health
curl http://localhost:8081/health  # Auth Service
curl http://localhost:8082/health  # User Service
curl http://localhost:8080/health  # API Gateway
```

## 🔧 Configuration

Each service has dedicated configuration:

- **API Gateway**: `configs/api-gateway/config.yaml`
- **Auth Service**: `configs/auth-service/config.yaml`  
- **User Service**: `configs/development/config.yaml`
- **Nginx**: `configs/nginx/nginx.conf`

Override with environment variables:
```bash
export AUTH_SERVICE_JWT_SECRET=your-secret
export USER_SERVICE_DATABASE_HOST=localhost
export API_GATEWAY_RATE_LIMIT=200
```

## 🛡️ Security

- **JWT Authentication**: Stateless authentication with configurable expiration
- **Role-Based Access**: User and admin role separation
- **Password Hashing**: bcrypt with salt for password security
- **Rate Limiting**: Per-client request throttling
- **CORS Protection**: Configurable cross-origin resource sharing
- **Security Headers**: Comprehensive HTTP security headers

## 📚 Complete Documentation

Comprehensive documentation covering all aspects of the microservices architecture:

- **[📋 Project Overview](docs/PROJECT_OVERVIEW.md)** - Architecture, technology stack, and design decisions
- **[🚀 Setup Guide](docs/SETUP_GUIDE.md)** - Complete setup instructions for all environments  
- **[🛠️ Development Guide](docs/DEVELOPMENT_GUIDE.md)** - Development workflows, coding standards, and testing
- **[📡 API Reference](docs/API_REFERENCE.md)** - Complete API documentation with examples
- **[🗄️ Database Guide](docs/DATABASE_GUIDE.md)** - Database schema, operations, and maintenance

### 🎯 Quick Navigation

| Need to... | Go to |
|------------|-------|
| **Understand the architecture** | [Project Overview](docs/PROJECT_OVERVIEW.md) |
| **Set up locally** | [Setup Guide](docs/SETUP_GUIDE.md) |
| **Start developing** | [Development Guide](docs/DEVELOPMENT_GUIDE.md) |
| **Integrate with APIs** | [API Reference](docs/API_REFERENCE.md) |
| **Work with database** | [Database Guide](docs/DATABASE_GUIDE.md) |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/microservice-name`)
3. Follow the microservices architecture patterns
4. Ensure all services have comprehensive tests
5. Update service-specific documentation
6. Submit a pull request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 🎯 Project Status

✅ **Production Ready** - Complete microservices architecture with enterprise features  
✅ **Well Documented** - Comprehensive documentation for all components  
✅ **Container Native** - Docker and Kubernetes deployment ready  
✅ **Security Focused** - JWT authentication, RBAC, audit logging  
✅ **Developer Friendly** - Hot reload, testing utilities, clear architecture  

**Built with ❤️ using Go microservices architecture** 🚀

*True service separation • Zero code duplication • Enterprise-grade security • Cloud-native deployment*