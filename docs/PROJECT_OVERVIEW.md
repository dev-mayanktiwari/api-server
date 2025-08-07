# Project Overview

## What is API Server?

API Server is a **production-ready microservices application** built with **Go**, implementing **Clean Architecture** principles across multiple services. It provides a complete user management system with JWT authentication, role-based access control, and enterprise-grade deployment capabilities.

## 🎯 Purpose & Goals

This project demonstrates:
- **True Microservices Architecture** with proper service separation
- **Clean Architecture** implementation in Go with domain-driven design
- **Enterprise-grade security** with JWT authentication and RBAC
- **Production-ready deployment** with Docker and Kubernetes
- **Zero code duplication** through shared libraries
- **Service-to-service communication** patterns
- **Comprehensive monitoring** and observability

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                Load Balancer (Nginx)                       │
│                     Port: 80/443                           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  API Gateway                                │
│                  Port: 8080                                 │
│ • Request Routing    • JWT Validation    • Rate Limiting    │
│ • CORS Handling      • Correlation IDs   • User Context     │
└──────────┬─────────────────────┬────────────────────────────┘
           │                     │
┌──────────▼──────────┐ ┌───────▼──────────────────────────────┐
│   Auth Service      │ │         User Service                 │
│   Port: 8081        │ │         Port: 8082                   │
│ • User Login        │ │ • User Registration                  │
│ • JWT Generation    │ │ • Profile Management                 │
│ • Token Validation  │ │ • Password Operations                │
│ • Token Refresh     │ │ • Admin User Management              │
│ • Session Management│ │ • User Statistics                    │
└─────────────────────┘ └──────────────────────────────────────┘
           │                     │
           └─────────┬───────────┘
                     │
┌────────────────────▼────────────────────────────────────────┐
│              Shared Infrastructure                          │
│  ┌─────────────────────┐  ┌─────────────────────┐          │
│  │   PostgreSQL 15     │  │      Redis 7        │          │
│  │ • User Data         │  │ • Sessions          │          │
│  │ • JWT Sessions      │  │ • Rate Limiting     │          │
│  │ • Audit Logs        │  │ • Caching           │          │
│  └─────────────────────┘  └─────────────────────┘          │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ Technology Stack

### Core Technologies
- **Language**: Go 1.21+ with modern practices
- **Web Framework**: Gin (high-performance HTTP router)
- **Database**: PostgreSQL 15+ with GORM ORM
- **Cache**: Redis 7+ for sessions and rate limiting
- **Authentication**: JWT with refresh tokens

### Microservices Components
- **API Gateway**: Request routing and authentication
- **Auth Service**: JWT management and user authentication
- **User Service**: User CRUD operations and management
- **Load Balancer**: Nginx for traffic distribution

### Development & Deployment
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Kubernetes with auto-scaling
- **Development**: Hot reload with Air
- **Monitoring**: Health checks and structured logging
- **CI/CD**: Docker builds and Kubernetes deployment

## 📦 Microservices Overview

### 🌐 API Gateway (Port 8080)
**Purpose**: Central entry point for all client requests

**Responsibilities**:
- Route requests to appropriate downstream services
- Validate JWT tokens for protected endpoints
- Implement rate limiting per client IP
- Add correlation IDs for distributed tracing
- Handle CORS policies and security headers
- Propagate user context to downstream services

**Key Features**:
- Proxy service with proper header handling
- Service health check aggregation
- Request/response logging with performance metrics
- Graceful shutdown and connection draining

### 🔐 Auth Service (Port 8081)
**Purpose**: Dedicated authentication and JWT token management

**Responsibilities**:
- User login and logout operations
- JWT token generation with configurable expiration
- Token validation and refresh functionality
- Session management with database storage
- Integration with User Service for credential validation

**Architecture**:
- **Domain Layer**: Token entities and authentication business rules
- **Application Layer**: Auth use cases and token services
- **Infrastructure Layer**: Database repositories and HTTP handlers

### 👥 User Service (Port 8082)
**Purpose**: Complete user lifecycle and data management

**Responsibilities**:
- User registration with validation
- User profile management (CRUD operations)
- Password change and security operations
- Admin user management capabilities
- User statistics and reporting

**Architecture**:
- **Domain Layer**: User entities, roles, and business logic
- **Application Layer**: User use cases and application services
- **Infrastructure Layer**: Database repositories and REST handlers

### ⚖️ Load Balancer (Nginx)
**Purpose**: Traffic distribution and SSL termination

**Responsibilities**:
- Route traffic to API Gateway instances
- SSL/TLS termination for HTTPS
- Static content serving capabilities
- Health check routing to individual services
- Security headers and rate limiting

## 🔗 Shared Libraries Architecture

Located in `shared/pkg/`, these libraries eliminate code duplication:

### 🔐 Auth Library
- JWT token utilities (generate, validate, parse)
- Password hashing with bcrypt
- Authentication middleware with role checking
- User context extraction and propagation

### ⚙️ Configuration Management
- Environment-based configuration using Viper
- Hot configuration reloading
- Environment variable overrides
- Configuration validation

### 🗄️ Database Utilities
- PostgreSQL connection with pooling
- GORM integration with auto-migration
- Repository pattern helpers
- Health check utilities

### 📊 Additional Libraries
- **Logger**: Structured JSON logging with correlation IDs
- **Middleware**: CORS, rate limiting, security headers
- **Response**: Standardized API response formats
- **Errors**: Application error handling

## 🔑 Core Features

### Security & Authentication
- **JWT Authentication**: Stateless tokens with refresh capability
- **Role-Based Access Control**: User and admin roles
- **Password Security**: bcrypt hashing with salt
- **Rate Limiting**: Per-client request throttling
- **Input Validation**: Comprehensive request validation
- **Security Headers**: CORS, CSP, and security headers

### Database & Persistence
- **Connection Pooling**: Efficient database connections
- **Auto-Migration**: Automated schema management
- **Transaction Support**: ACID compliance
- **Health Monitoring**: Database connectivity checks
- **Audit Logging**: Complete user action tracking

### Service Communication
- **HTTP-based**: RESTful service communication
- **Context Propagation**: User context across services
- **Correlation IDs**: Request tracing and debugging
- **Error Handling**: Proper error propagation
- **Timeout Management**: Service call timeouts

### Deployment & Operations
- **Container-Native**: Docker with multi-stage builds
- **Kubernetes-Ready**: Complete K8s manifests
- **Health Probes**: Liveness, readiness, startup checks
- **Graceful Shutdown**: Proper connection draining
- **Resource Limits**: Memory and CPU constraints

## 🌐 API Capabilities

### Public Endpoints
- User registration with validation
- User authentication and JWT generation
- Service health and readiness checks

### Protected Endpoints (User Role)
- Profile retrieval and updates
- Password change operations
- Personal data management

### Admin Endpoints (Admin Role)
- User listing with pagination and filtering
- User management operations (CRUD)
- System administration and monitoring

## 📊 Monitoring & Observability

### Logging
- **Structured Logging**: JSON format for production
- **Log Levels**: Debug, Info, Warn, Error, Fatal
- **Correlation IDs**: Request tracing across services
- **User Context**: User ID and role in logs
- **Performance Metrics**: Response times and errors

### Health Monitoring
- **Service Health**: `/health` endpoint on each service
- **Readiness Checks**: `/ready` with dependency validation
- **Database Health**: Connection and query performance
- **Service Discovery**: Health aggregation at gateway level

### Metrics Collection
- HTTP request metrics (count, duration, status codes)
- Database connection pool statistics
- Authentication success/failure rates
- User registration and activity metrics

## 🚀 Deployment Architecture

### Development Environment
- **Docker Compose**: Complete development stack
- **Hot Reload**: Air for Go hot reloading
- **Development Tools**: pgAdmin, Redis Insight
- **Debug Configuration**: Enhanced logging and debugging

### Production Environment
- **Kubernetes Cluster**: Auto-scaling and load balancing
- **Multiple Replicas**: High availability configuration
- **Resource Management**: CPU and memory limits
- **Security Context**: Non-root containers, read-only filesystems

### Container Strategy
- **Multi-stage Builds**: Optimized production images
- **Security**: Non-root users, minimal attack surface
- **Health Checks**: Built-in container health monitoring
- **Configuration**: Environment-based configuration

## 🎯 Project Goals & Benefits

### Educational Value
- **Clean Architecture**: Demonstrates proper layered architecture
- **Microservices Patterns**: Service separation and communication
- **Go Best Practices**: Modern Go development patterns
- **DevOps Integration**: Complete deployment pipeline

### Production Readiness
- **Scalable Architecture**: Horizontal scaling capabilities
- **Security**: Enterprise-grade security implementations
- **Monitoring**: Comprehensive observability
- **Deployment**: Cloud-native deployment strategies

### Development Experience
- **Developer-Friendly**: Easy local setup and development
- **Documentation**: Comprehensive guides and API docs
- **Testing**: Unit and integration testing frameworks
- **Tooling**: Modern development tools integration

## 🎯 Target Audience

- **Go Developers** learning clean architecture and microservices
- **DevOps Engineers** implementing container-native applications
- **System Architects** designing scalable backend systems
- **Development Teams** building enterprise applications
- **Students** studying production-grade software development

This project serves as both a learning resource and a foundation for building production microservices applications with modern Go practices and cloud-native deployment strategies.