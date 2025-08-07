# Project Overview

## What is API Server?

API Server is a production-ready microservices application built with **Go**, implementing **Clean Architecture** principles. It provides a complete user management system with JWT authentication, role-based access control, and enterprise-grade deployment capabilities.

## 🎯 Purpose

This project demonstrates:
- **Clean Architecture** implementation in Go
- **Microservices** design patterns
- **Production-ready** deployment configurations
- **Enterprise-grade** security and monitoring
- **Comprehensive testing** strategies

## 🏗️ High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    API Server System                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │   Client    │    │  API Gateway │    │ Load Balancer│    │
│  │Applications │◄──►│   (Nginx)   │◄──►│  (Optional) │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
├─────────────────────────────────────────────────────────────┤
│                   Microservices Layer                       │
│  ┌─────────────────────────────────────────────────────────┐│
│  │              User Service (Port 8082)                   ││
│  │  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐       ││
│  │  │   Domain    │ │Application  │ │Infrastructure│      ││
│  │  │   Layer     │ │   Layer     │ │    Layer     │      ││
│  │  └─────────────┘ └─────────────┘ └─────────────┘       ││
│  └─────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│                   Shared Libraries                          │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Auth │ Config │ Logger │ Database │ Middleware │ etc.  ││
│  └─────────────────────────────────────────────────────────┘│
├─────────────────────────────────────────────────────────────┤
│                   Data & Cache Layer                        │
│  ┌─────────────┐              ┌─────────────┐              │
│  │ PostgreSQL  │              │    Redis    │              │
│  │ (Database)  │              │   (Cache)   │              │
│  └─────────────┘              └─────────────┘              │
├─────────────────────────────────────────────────────────────┤
│                 Monitoring & Logging                        │
│  ┌─────────────┐              ┌─────────────┐              │
│  │ Prometheus  │              │   Grafana   │              │
│  │ (Metrics)   │              │(Dashboards) │              │
│  └─────────────┘              └─────────────┘              │
└─────────────────────────────────────────────────────────────┘
```

## 🛠️ Technology Stack

### Core Technologies
- **Language**: Go 1.21+
- **Web Framework**: Gin (HTTP routing)
- **Database**: PostgreSQL 15+ 
- **Cache**: Redis 7+
- **ORM**: GORM (Object-Relational Mapping)

### Development Tools
- **Hot Reload**: Air (development)
- **Testing**: testify (mocking and assertions)
- **Linting**: golangci-lint
- **Documentation**: Markdown

### Deployment & DevOps
- **Containerization**: Docker & Docker Compose
- **Orchestration**: Kubernetes
- **Reverse Proxy**: Nginx
- **Monitoring**: Prometheus + Grafana
- **CI/CD**: GitHub Actions (configurable)

## 📦 Services Overview

### User Service
**Purpose**: Complete user lifecycle management
**Port**: 8082
**Responsibilities**:
- User registration and authentication
- Profile management
- Password operations
- Admin user management
- JWT token handling

**Architecture Layers**:
- **Domain**: Business entities and repository contracts
- **Application**: Use cases and data transfer objects
- **Infrastructure**: Database implementations and HTTP handlers

## 🔑 Core Features

### Security
- **JWT Authentication**: Secure token-based auth
- **Role-Based Access Control (RBAC)**: User and admin roles
- **Password Hashing**: bcrypt with salt
- **Rate Limiting**: Per-client request throttling
- **Input Validation**: Comprehensive request validation
- **Security Headers**: CORS, CSP, and security headers

### Database
- **Connection Pooling**: Efficient database connections
- **Migrations**: Automated database schema management
- **Health Checks**: Database connectivity monitoring
- **Transaction Support**: ACID compliance

### Configuration
- **Environment-Specific**: Dev, staging, prod configs
- **Hot Reload**: Configuration changes without restart
- **Environment Variables**: Override any config value
- **Validation**: Configuration validation on startup

### Testing
- **Unit Tests**: Business logic testing with mocks
- **Integration Tests**: API endpoint testing with real database
- **Test Utilities**: Shared testing helpers and fixtures
- **Mocking**: Repository and service layer mocks

### Deployment
- **Multi-Stage Builds**: Optimized Docker images
- **Container Security**: Non-root users, read-only filesystems
- **Kubernetes Native**: Complete K8s deployment manifests
- **Auto-Scaling**: Horizontal Pod Autoscaler configuration
- **Health Probes**: Liveness, readiness, and startup probes

## 🌐 API Capabilities

### Public Endpoints
- User registration
- User authentication
- Health checks

### Protected Endpoints (User Role)
- Profile retrieval and updates
- Password changes
- Personal data management

### Admin Endpoints (Admin Role)
- User listing with pagination
- User management (CRUD operations)
- System administration

## 📊 Monitoring & Observability

### Logging
- **Structured Logging**: JSON format for production
- **Log Levels**: Debug, Info, Warn, Error
- **Correlation IDs**: Request tracing across services
- **Contextual Data**: User context and performance metrics

### Metrics
- **HTTP Metrics**: Request duration, count, and status codes
- **Database Metrics**: Connection pool statistics
- **Business Metrics**: User registration, login counts
- **System Metrics**: CPU, memory, and disk usage

### Health Monitoring
- **Service Health**: `/health` endpoint
- **Readiness Check**: `/ready` endpoint for K8s
- **Database Health**: Connection and query performance
- **Dependency Checks**: External service availability

## 🚀 Deployment Options

### Local Development
- Direct Go execution with hot reload
- Docker Compose with development tools
- Database and Redis via containers

### Staging/Testing
- Docker Compose with production-like configuration
- Kubernetes deployment with staging secrets
- Database migrations and seed data

### Production
- Kubernetes cluster deployment
- Load balancer and ingress configuration
- Monitoring and alerting setup
- Backup and disaster recovery

## 📝 Project Goals

1. **Educational**: Demonstrate clean architecture in Go
2. **Production-Ready**: Real-world deployment capabilities
3. **Scalable**: Microservices design for horizontal scaling
4. **Maintainable**: Clear separation of concerns and documentation
5. **Secure**: Enterprise-grade security implementations
6. **Observable**: Comprehensive monitoring and logging

## 🎯 Target Audience

- **Go Developers** learning clean architecture
- **DevOps Engineers** implementing microservices deployments
- **System Architects** designing scalable backend systems
- **Students** studying production-grade application development
- **Teams** building enterprise applications

This project serves as both a learning resource and a foundation for building production microservices applications with Go.