# Microservices Architecture

This project has been transformed from a monolith into a microservices architecture with the following services:

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  Auth Service   │    │  User Service   │
│     :8080       │────│     :8081       │────│     :8082       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   PostgreSQL    │
                    │     :5432       │
                    └─────────────────┘
```

## Services

### 1. API Gateway (:8080)
- **Purpose**: Entry point for all client requests
- **Features**: 
  - Request routing to appropriate services
  - Authentication middleware
  - Rate limiting
  - CORS handling
  - Request logging
- **Technology**: Go + Gin

### 2. Auth Service (:8081)
- **Purpose**: JWT token management and authentication
- **Features**:
  - Token generation
  - Token validation
  - Token refresh
- **Technology**: Go + Gin
- **Database**: None (stateless)

### 3. User Service (:8082)
- **Purpose**: User management and profile operations
- **Features**:
  - User registration
  - User login (with auth service integration)
  - Profile management
  - Password management
  - User listing (admin)
- **Technology**: Go + Gin + GORM
- **Database**: PostgreSQL (user_db)

## Inter-Service Communication

Services communicate via HTTP REST APIs:

1. **API Gateway → Auth Service**: Token validation
2. **API Gateway → User Service**: User operations
3. **User Service → Auth Service**: Token generation during login

## Deployment Options

### Local Development (Docker Compose)

```bash
# Build all services
make build-all

# Start services
make run-local

# View logs
make logs-local

# Stop services
make stop-local
```

### Kubernetes Deployment

```bash
# Deploy to Kubernetes
make deploy-k8s

# Check status
kubectl get all -n microservices

# Remove deployment
make undeploy-k8s
```

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/refresh` - Refresh tokens

### Users (Requires Authentication)
- `GET /api/v1/users` - List users (admin only)
- `GET /api/v1/users/:id` - Get user profile
- `PUT /api/v1/users/:id` - Update user profile
- `DELETE /api/v1/users/:id` - Delete user (admin only)
- `POST /api/v1/users/:id/change-password` - Change password

### Health Checks
- `GET /health` - Service health status

## Environment Configuration

### Auth Service
```bash
AUTH_SERVER_PORT=8081
AUTH_JWT_SECRET_KEY=your-secret-key
AUTH_JWT_ACCESS_TOKEN_DURATION=15  # minutes
AUTH_JWT_REFRESH_TOKEN_DURATION=168  # hours
```

### User Service
```bash
USER_SERVER_PORT=8082
USER_DATABASE_HOST=postgres
USER_DATABASE_NAME=user_db
USER_SERVICES_AUTH_SERVICE=http://auth-service:8081
```

### API Gateway
```bash
GATEWAY_SERVER_PORT=8080
GATEWAY_SERVICES_AUTH_SERVICE=http://auth-service:8081
GATEWAY_SERVICES_USER_SERVICE=http://user-service:8082
```

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- Request ID tracking
- Rate limiting
- CORS protection
- Input validation

## Monitoring and Logging

- Structured JSON logging
- Request/response logging with correlation IDs
- Health check endpoints
- Graceful shutdown handling

## Development Workflow

1. **Local Development**: Use Docker Compose for full stack
2. **Testing**: Individual service testing with Make targets
3. **Deployment**: Kubernetes manifests for production

## Database Schema

### Auth Service
- No persistent storage (stateless JWT service)

### User Service
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR UNIQUE NOT NULL,
    password VARCHAR NOT NULL,
    first_name VARCHAR NOT NULL,
    last_name VARCHAR NOT NULL,
    role VARCHAR DEFAULT 'user' NOT NULL,
    is_active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## Scaling Considerations

- **Auth Service**: Horizontally scalable (stateless)
- **User Service**: Horizontally scalable with database connection pooling
- **API Gateway**: Load balancer with multiple instances
- **Database**: Single PostgreSQL instance with potential for read replicas

## Migration from Monolith

The original monolith has been decomposed by:
1. Extracting auth logic into dedicated Auth Service
2. Moving user operations to User Service
3. Creating API Gateway for request routing
4. Maintaining database separation (auth_db, user_db)
5. Implementing service-to-service communication

This maintains all original functionality while providing better:
- Scalability
- Maintainability
- Deployment flexibility
- Technology diversity