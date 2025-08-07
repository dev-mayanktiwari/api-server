# Development Guide

Complete guide for developers working on the API Server microservices project, including development workflows, coding standards, testing practices, and contribution guidelines.

## 🚀 Development Environment Setup

### Quick Start for Developers

```bash
# 1. Clone and setup dependencies
git clone <repository-url>
cd api-server
go mod tidy

# Setup shared libraries
cd shared && go mod tidy && cd ..

# Setup each service
cd services/api-gateway && go mod tidy && cd ../..
cd services/auth-service && go mod tidy && cd ../..
cd services/user-service && go mod tidy && cd ../..

# 2. Start infrastructure services
docker-compose -f docker-compose.microservices.yml up -d postgres redis

# 3. Start development with hot reload
cd services/user-service && air &
cd services/auth-service && air &
cd services/api-gateway && air &

# 4. Access development environment
# - User Service: http://localhost:8082
# - Auth Service: http://localhost:8081
# - API Gateway: http://localhost:8080
```

### Development Tools Setup

#### Install Essential Tools
```bash
# Hot reload for Go
go install github.com/cosmtrek/air@latest

# Linting and formatting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Testing framework
go install github.com/stretchr/testify@latest

# API testing tools
go install github.com/rakyll/hey@latest
```

#### Recommended VS Code Extensions
- **Go** (Google) - Go language support
- **Docker** (Microsoft) - Docker integration
- **YAML** (Red Hat) - YAML file support
- **REST Client** (Huachao Mao) - API testing
- **GitLens** (GitKraken) - Git integration
- **Thunder Client** - API testing alternative

## 🏗️ Project Architecture Deep Dive

### Microservices Structure
```
api-server/
├── services/                           # Individual Microservices
│   ├── api-gateway/                   # API Gateway Service
│   │   ├── cmd/server/main.go         # Entry point
│   │   ├── internal/
│   │   │   ├── application/           # Application layer
│   │   │   │   └── services/         # Proxy services
│   │   │   └── infrastructure/        # Infrastructure layer
│   │   │       └── http/handlers/    # HTTP handlers
│   │   ├── Dockerfile                 # Container definition
│   │   └── go.mod                     # Service dependencies
│   ├── auth-service/                  # Authentication Service
│   │   ├── cmd/server/main.go         # Entry point
│   │   ├── internal/
│   │   │   ├── application/           # Application layer
│   │   │   │   ├── dto/              # Data transfer objects
│   │   │   │   └── services/         # Use cases
│   │   │   ├── domain/               # Domain layer
│   │   │   │   ├── entities/         # Business entities
│   │   │   │   ├── repositories/     # Repository interfaces
│   │   │   │   └── services/         # Domain services
│   │   │   └── infrastructure/        # Infrastructure layer
│   │   │       ├── database/         # Database implementations
│   │   │       └── http/handlers/    # HTTP handlers
│   │   └── go.mod
│   └── user-service/                  # User Management Service
│       ├── cmd/server/main.go         # Entry point
│       ├── internal/                  # Clean architecture layers
│       │   ├── application/           # Application layer
│       │   ├── domain/               # Domain layer
│       │   └── infrastructure/        # Infrastructure layer
│       └── go.mod
├── shared/                            # Shared Libraries
│   └── pkg/                          # Reusable packages
│       ├── auth/                     # JWT utilities
│       ├── config/                   # Configuration management
│       ├── database/                 # Database utilities
│       ├── logger/                   # Structured logging
│       ├── middleware/               # HTTP middleware
│       └── response/                 # API response utilities
├── configs/                          # Environment configurations
├── k8s/                             # Kubernetes manifests
├── scripts/                         # Build and deployment scripts
└── tests/                           # Test suites
    ├── unit/                        # Unit tests
    ├── integration/                 # Integration tests
    ├── mocks/                       # Mock implementations
    └── utils/                       # Test utilities
```

### Clean Architecture Implementation

Each service follows Clean Architecture principles:

#### Domain Layer (Inner Layer)
```go
// entities/user.go
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    FirstName string    `json:"first_name"`
    LastName  string    `json:"last_name"`
    Role      UserRole  `json:"role"`
    Status    UserStatus `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// repositories/user_repository.go
type UserRepository interface {
    Create(ctx context.Context, user *User) (*User, error)
    GetByID(ctx context.Context, id string) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}
```

#### Application Layer (Use Cases)
```go
// services/user_service.go
type UserService struct {
    userRepo   domain.UserRepository
    logger     logger.Logger
    validator  validator.Validator
}

func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    // Business logic implementation
    // Validation, domain rules, etc.
}
```

#### Infrastructure Layer (External)
```go
// database/user_repository.go
type userRepository struct {
    db     *gorm.DB
    logger logger.Logger
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
    // Database implementation
}

// http/handlers/user_handler.go
type UserHandler struct {
    userService application.UserService
    logger      logger.Logger
}

func (h *UserHandler) CreateUser(c *gin.Context) {
    // HTTP handler implementation
}
```

## 📝 Coding Standards

### Go Style Guide

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html) guidelines.

#### Key Principles

1. **Package Names**: Short, concise, lowercase, no underscores
```go
// Good
package auth
package database

// Bad
package auth_service
package databaseUtils
```

2. **Interface Names**: Single method interfaces end with `-er`
```go
// Good
type Reader interface {
    Read([]byte) (int, error)
}

type UserRepository interface {
    Create(context.Context, *User) error
    GetByID(context.Context, string) (*User, error)
}
```

3. **Variable Names**: Use camelCase, be descriptive but concise
```go
// Good
userID := "123"
httpClient := &http.Client{}

// Bad
userId := "123"
HTTP_CLIENT := &http.Client{}
```

4. **Error Handling**: Always handle errors explicitly
```go
// Good
user, err := userService.GetUser(ctx, userID)
if err != nil {
    return nil, fmt.Errorf("failed to get user: %w", err)
}

// Bad
user, _ := userService.GetUser(ctx, userID)
```

### Code Organization

#### File Naming
- Use snake_case for filenames: `user_handler.go`, `auth_service.go`
- Test files: `user_handler_test.go`
- Interface files: `user_repository.go` (for repository interfaces)

#### Project Structure
```go
// services/user-service/internal/domain/entities/user.go
package entities

// services/user-service/internal/domain/repositories/user_repository.go
package repositories

// services/user-service/internal/application/services/user_service.go
package services

// services/user-service/internal/infrastructure/database/user_repository.go
package database

// services/user-service/internal/infrastructure/http/handlers/user_handler.go
package handlers
```

### Code Formatting

#### Use gofmt and goimports
```bash
# Format all Go files
find . -name "*.go" -not -path "./vendor/*" | xargs gofmt -s -w

# Organize imports
find . -name "*.go" -not -path "./vendor/*" | xargs goimports -w
```

#### Linting Configuration
```yaml
# .golangci.yml
linters-settings:
  golint:
    min-confidence: 0
  gocyclo:
    min-complexity: 10
  dupl:
    threshold: 100

linters:
  enable:
    - errcheck
    - gosimple
    - govet
    - ineffassign
    - staticcheck
    - unused
    - gocritic
    - gofmt
    - goimports
```

## 🧪 Testing Practices

### Testing Strategy

#### 1. Unit Tests
Test individual functions and methods in isolation:

```go
// services/user-service/internal/application/services/user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := &mocks.UserRepository{}
    mockValidator := &mocks.Validator{}
    service := NewUserService(mockRepo, mockValidator)
    
    req := &dto.CreateUserRequest{
        Email:     "test@example.com",
        Password:  "password123",
        FirstName: "Test",
        LastName:  "User",
    }
    
    expectedUser := &domain.User{
        ID:        "user-id",
        Email:     req.Email,
        FirstName: req.FirstName,
        LastName:  req.LastName,
    }
    
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
        Return(expectedUser, nil)
    mockValidator.On("Validate", req).Return(nil)
    
    // Act
    result, err := service.CreateUser(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expectedUser.Email, result.Email)
    mockRepo.AssertExpectations(t)
    mockValidator.AssertExpectations(t)
}
```

#### 2. Integration Tests
Test service interactions with real database:

```go
// tests/integration/user_service_integration_test.go
//go:build integration
// +build integration

func TestUserService_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDatabase(t)
    defer cleanupTestDatabase(t, db)
    
    // Create service with real dependencies
    userRepo := database.NewUserRepository(db)
    userService := services.NewUserService(userRepo, validator.New())
    
    // Test user creation
    req := &dto.CreateUserRequest{
        Email:     "integration@example.com",
        Password:  "password123",
        FirstName: "Integration",
        LastName:  "Test",
    }
    
    user, err := userService.CreateUser(context.Background(), req)
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
    
    // Test user retrieval
    retrievedUser, err := userService.GetUser(context.Background(), user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Email, retrievedUser.Email)
}
```

#### 3. API Tests
Test HTTP endpoints:

```go
// tests/api/user_api_test.go
func TestUserAPI_Register(t *testing.T) {
    // Setup test server
    router := setupTestRouter()
    
    // Test user registration
    reqBody := `{
        "email": "api@example.com",
        "password": "password123",
        "first_name": "API",
        "last_name": "Test"
    }`
    
    req := httptest.NewRequest("POST", "/api/v1/users/register", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.True(t, response["success"].(bool))
}
```

### Running Tests

```bash
# Run all unit tests
go test ./... -short

# Run integration tests
go test -tags=integration ./tests/integration/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific service tests
go test ./services/user-service/...

# Run tests with race detection
go test -race ./...

# Verbose output
go test -v ./...
```

### Test Utilities

#### Database Test Setup
```go
// tests/utils/database.go
func SetupTestDatabase(t *testing.T) *gorm.DB {
    db, err := gorm.Open(postgres.Open(getTestDSN()), &gorm.Config{})
    require.NoError(t, err)
    
    // Run migrations
    err = db.AutoMigrate(&domain.User{})
    require.NoError(t, err)
    
    return db
}

func CleanupTestDatabase(t *testing.T, db *gorm.DB) {
    db.Exec("DELETE FROM users")
}
```

#### Mock Generation
```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks for interfaces
mockery --dir=./services/user-service/internal/domain/repositories --all --output=./tests/mocks

# Generate mock for specific interface
mockery --name=UserRepository --dir=./services/user-service/internal/domain/repositories --output=./tests/mocks
```

## 🔄 Development Workflow

### Git Workflow

#### Branch Naming
- `feature/service-name-feature-description`
- `bugfix/service-name-issue-description`
- `hotfix/service-name-critical-fix`

Examples:
```bash
git checkout -b feature/user-service-password-reset
git checkout -b bugfix/auth-service-token-validation
git checkout -b hotfix/api-gateway-memory-leak
```

#### Commit Messages
```bash
# Format: type(service): description
git commit -m "feat(user-service): add password reset functionality"
git commit -m "fix(auth-service): resolve JWT token validation issue"
git commit -m "docs(api-gateway): update API documentation"
git commit -m "test(user-service): add unit tests for user registration"
```

### Development Process

#### 1. Feature Development
```bash
# 1. Create feature branch
git checkout -b feature/user-service-email-verification

# 2. Start development environment
docker-compose -f docker-compose.dev.yml up -d postgres redis

# 3. Start service with hot reload
cd services/user-service && air

# 4. Write tests first (TDD approach)
# Create test file: internal/application/services/email_service_test.go

# 5. Implement feature
# Create implementation: internal/application/services/email_service.go

# 6. Run tests
go test ./internal/application/services/...

# 7. Test integration
curl -X POST http://localhost:8082/api/v1/users/verify-email \
  -H "Content-Type: application/json" \
  -d '{"token": "verification-token"}'

# 8. Commit changes
git add .
git commit -m "feat(user-service): add email verification functionality"

# 9. Push and create PR
git push origin feature/user-service-email-verification
```

#### 2. Code Review Checklist
- [ ] **Tests**: Unit tests cover new functionality
- [ ] **Documentation**: Code is well-documented
- [ ] **Error Handling**: All errors are properly handled
- [ ] **Logging**: Appropriate logging levels and messages
- [ ] **Security**: No secrets or sensitive data in code
- [ ] **Performance**: No obvious performance issues
- [ ] **Dependencies**: No unnecessary dependencies added

### Service Development

#### Adding New Endpoint
1. **Define Domain Entity** (if needed)
```go
// internal/domain/entities/user_profile.go
type UserProfile struct {
    UserID      string `json:"user_id"`
    Bio         string `json:"bio"`
    AvatarURL   string `json:"avatar_url"`
    // ... other fields
}
```

2. **Define Repository Interface**
```go
// internal/domain/repositories/user_profile_repository.go
type UserProfileRepository interface {
    GetByUserID(ctx context.Context, userID string) (*UserProfile, error)
    Update(ctx context.Context, profile *UserProfile) error
}
```

3. **Create DTO**
```go
// internal/application/dto/user_profile_dto.go
type UpdateUserProfileRequest struct {
    Bio       string `json:"bio" validate:"max=500"`
    AvatarURL string `json:"avatar_url" validate:"url"`
}

type UserProfileResponse struct {
    UserID    string `json:"user_id"`
    Bio       string `json:"bio"`
    AvatarURL string `json:"avatar_url"`
    UpdatedAt string `json:"updated_at"`
}
```

4. **Implement Application Service**
```go
// internal/application/services/user_profile_service.go
type UserProfileService struct {
    profileRepo domain.UserProfileRepository
    logger      logger.Logger
}

func (s *UserProfileService) UpdateProfile(ctx context.Context, userID string, req *dto.UpdateUserProfileRequest) (*dto.UserProfileResponse, error) {
    // Implementation
}
```

5. **Implement Infrastructure**
```go
// internal/infrastructure/database/user_profile_repository.go
type userProfileRepository struct {
    db *gorm.DB
}

func (r *userProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.UserProfile, error) {
    // Database implementation
}
```

6. **Create HTTP Handler**
```go
// internal/infrastructure/http/handlers/user_profile_handler.go
type UserProfileHandler struct {
    profileService application.UserProfileService
}

func (h *UserProfileHandler) UpdateProfile(c *gin.Context) {
    // HTTP handler implementation
}
```

7. **Register Route**
```go
// internal/infrastructure/http/routes.go
func SetupRoutes(router *gin.Engine, handlers *Handlers) {
    api := router.Group("/api/v1/users")
    {
        api.PUT("/profile", middleware.AuthRequired(), handlers.UserProfile.UpdateProfile)
    }
}
```

## 🔧 Debugging and Monitoring

### Local Debugging

#### VS Code Debug Configuration
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug User Service",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "./services/user-service/cmd/server",
            "env": {
                "USER_SERVICE_DATABASE_HOST": "localhost",
                "USER_SERVICE_LOGGING_LEVEL": "debug"
            }
        }
    ]
}
```

#### Debug with Delve
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugger
cd services/user-service
dlv debug ./cmd/server

# Set breakpoints and continue
(dlv) break internal/application/services/user_service.go:45
(dlv) continue
```

### Logging Best Practices

#### Structured Logging
```go
// Use consistent logging with context
func (s *UserService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    requestID := middleware.GetRequestID(ctx)
    
    s.logger.Info("Creating user",
        zap.String("request_id", requestID),
        zap.String("email", req.Email),
        zap.String("operation", "create_user"),
    )
    
    user, err := s.userRepo.Create(ctx, domainUser)
    if err != nil {
        s.logger.Error("Failed to create user",
            zap.String("request_id", requestID),
            zap.String("email", req.Email),
            zap.Error(err),
        )
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    s.logger.Info("User created successfully",
        zap.String("request_id", requestID),
        zap.String("user_id", user.ID),
        zap.String("email", user.Email),
    )
    
    return response, nil
}
```

#### Log Levels
- **DEBUG**: Detailed information for debugging
- **INFO**: General operational messages
- **WARN**: Warning conditions
- **ERROR**: Error conditions that might still allow operation
- **FATAL**: Very serious errors that will abort the program

### Performance Monitoring

#### Adding Metrics
```go
// Use prometheus metrics
var (
    httpRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests.",
        },
        []string{"service", "method", "endpoint", "status_code"},
    )
    
    httpRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request latency distributions.",
        },
        []string{"service", "method", "endpoint"},
    )
)
```

#### Database Query Monitoring
```go
// Add query logging in repository
func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
    start := time.Now()
    defer func() {
        duration := time.Since(start)
        r.logger.Debug("Database query executed",
            zap.String("operation", "create_user"),
            zap.Duration("duration", duration),
        )
    }()
    
    result := r.db.WithContext(ctx).Create(user)
    return user, result.Error
}
```

## 🚀 Deployment and CI/CD

### Docker Best Practices

#### Multi-stage Dockerfile
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Runtime stage
FROM alpine:3.18

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN adduser -D -s /bin/sh api-user
USER api-user

EXPOSE 8082

CMD ["./main"]
```

### CI/CD Pipeline

#### GitHub Actions Example
```yaml
# .github/workflows/ci.yml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: api_server_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: |
        go test -v ./...
        go test -tags=integration -v ./tests/integration/...
      env:
        USER_SERVICE_DATABASE_HOST: localhost
        USER_SERVICE_DATABASE_DATABASE: api_server_test
        USER_SERVICE_JWT_SECRET: test-secret
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
  build:
    needs: test
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Build Docker images
      run: |
        docker build -t api-server/user-service:${{ github.sha }} ./services/user-service
        docker build -t api-server/auth-service:${{ github.sha }} ./services/auth-service
        docker build -t api-server/api-gateway:${{ github.sha }} ./services/api-gateway
```

## 📚 Contributing Guidelines

### Before Contributing

1. **Read Documentation**: Understand the architecture and coding standards
2. **Setup Environment**: Follow the development setup guide
3. **Run Tests**: Ensure all tests pass locally
4. **Check Linting**: Run golangci-lint on your changes

### Contribution Process

1. **Fork Repository**: Create your own fork
2. **Create Branch**: Use descriptive branch names
3. **Write Tests**: Add tests for new functionality
4. **Follow Standards**: Adhere to coding standards
5. **Update Documentation**: Update relevant documentation
6. **Submit PR**: Create detailed pull request

### Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project coding standards
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No sensitive data exposed
```

This development guide provides a comprehensive foundation for working on the API Server microservices project. Follow these practices to maintain code quality and ensure smooth collaboration.