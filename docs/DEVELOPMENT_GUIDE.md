# Development Guide

Complete guide for developers working on the API Server project, including development workflows, coding standards, testing practices, and contribution guidelines.

## 🚀 Development Environment

### Quick Start for Developers

```bash
# 1. Clone and setup
git clone <repository-url>
cd api-server
go mod tidy && cd shared && go mod tidy && cd ..

# 2. Start development environment
docker-compose -f docker-compose.dev.yml up -d

# 3. Start development with hot reload
cd services/user-service
air

# 4. Access development tools
# - API: http://localhost:8082
# - pgAdmin: http://localhost:5050
# - RedisInsight: http://localhost:8001
```

### Development Tools Setup

#### Install Air for Hot Reload
```bash
go install github.com/cosmtrek/air@latest
```

#### Install Testing Tools
```bash
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/stretchr/testify@latest
```

#### VS Code Extensions (Recommended)
- Go (Google)
- Docker (Microsoft)
- YAML (Red Hat)
- REST Client (Huachao Mao)
- GitLens (GitKraken)

## 🏗️ Project Structure Deep Dive

```
api-server/
├── services/user-service/           # User Service (Clean Architecture)
│   ├── cmd/server/main.go          # Application entry point
│   ├── internal/
│   │   ├── application/            # Application Layer
│   │   │   ├── dto/               # Data Transfer Objects
│   │   │   └── services/          # Application Services (Use Cases)
│   │   ├── domain/                # Domain Layer
│   │   │   ├── entities/          # Business Entities
│   │   │   ├── repositories/      # Repository Interfaces
│   │   │   └── services/          # Domain Services
│   │   └── infrastructure/        # Infrastructure Layer
│   │       ├── database/          # Database Implementations
│   │       └── http/handlers/     # HTTP Handlers
│   └── go.mod                     # Service Dependencies
├── shared/pkg/                     # Shared Libraries
│   ├── auth/                      # JWT Authentication
│   ├── config/                    # Configuration Management
│   ├── database/                  # Database Utilities
│   ├── logger/                    # Structured Logging
│   ├── middleware/                # HTTP Middleware
│   └── response/                  # API Response Utilities
├── tests/                         # Testing Framework
│   ├── unit/                      # Unit Tests
│   ├── integration/               # Integration Tests
│   ├── mocks/                     # Mock Implementations
│   └── utils/                     # Test Utilities
└── configs/                       # Environment Configurations
    ├── development/
    ├── staging/
    └── production/
```

## 📝 Coding Standards

### Go Style Guide

We follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html) guidelines.

#### Key Principles

1. **Package Names**: Short, concise, lowercase, no underscores
2. **Interface Names**: Single method interfaces end with `-er` (e.g., `Reader`, `Writer`)
3. **Variable Names**: Use camelCase, be descriptive but concise
4. **Constants**: Use camelCase or SCREAMING_SNAKE_CASE for exported constants
5. **Error Handling**: Always handle errors explicitly

#### Code Formatting

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Organize imports
goimports -w .
```

### Clean Architecture Layers

#### 1. Domain Layer (`internal/domain/`)
**Purpose**: Core business logic, independent of external concerns

```go
// entities/user.go
type User struct {
    ID        string
    Email     string
    FirstName string
    LastName  string
    Role      UserRole
    Status    UserStatus
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business methods
func (u *User) IsAdmin() bool {
    return u.Role == RoleAdmin
}

func (u *User) IsActive() bool {
    return u.Status == StatusActive
}
```

**Rules**:
- ✅ Pure business logic
- ✅ No external dependencies
- ❌ No database imports
- ❌ No HTTP imports
- ❌ No third-party libraries

#### 2. Application Layer (`internal/application/`)
**Purpose**: Use cases and application-specific business rules

```go
// services/user_app_service.go
type UserApplicationService struct {
    userRepo     repositories.UserRepository
    domainSvc    services.UserDomainService
    jwtManager   auth.JWTManager
    logger       logger.Logger
}

func (s *UserApplicationService) CreateUser(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, error) {
    // Validation
    if err := req.Validate(); err != nil {
        return nil, err
    }
    
    // Business logic
    user, err := entities.NewUser(req.Email, req.Password, req.FirstName, req.LastName)
    if err != nil {
        return nil, err
    }
    
    // Repository interaction
    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }
    
    return dto.NewUserResponse(user), nil
}
```

**Rules**:
- ✅ Orchestrates domain objects
- ✅ Uses repository interfaces
- ✅ Handles use cases
- ❌ No HTTP concerns
- ❌ No database implementation details

#### 3. Infrastructure Layer (`internal/infrastructure/`)
**Purpose**: External concerns, implementations, frameworks

```go
// database/user_repository_impl.go
type userRepository struct {
    db     database.DB
    logger logger.Logger
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
    model := &UserModel{
        ID:        user.ID,
        Email:     user.Email,
        FirstName: user.FirstName,
        // ... other fields
    }
    
    return r.db.Create(model).Error
}
```

**Rules**:
- ✅ Implements domain interfaces
- ✅ Framework-specific code
- ✅ External service integrations
- ✅ Database implementations

### Naming Conventions

#### Files and Directories
```
user_service.go          # Snake case for files
UserService             # Pascal case for types
createUser              # Camel case for methods
USER_ROLE_ADMIN         # Screaming snake for constants
```

#### Database
```sql
-- Tables: snake_case, plural
users
user_sessions

-- Columns: snake_case
first_name
created_at
```

#### API Endpoints
```
GET  /api/v1/users/profile      # Kebab case for URLs
POST /api/v1/users/change-password
```

## 🧪 Testing Strategy

### Test Structure

```
tests/
├── unit/                    # Unit tests (fast, isolated)
│   └── user_service_test.go
├── integration/             # Integration tests (with database)
│   └── user_api_test.go
├── mocks/                   # Generated mocks
│   └── user_repository_mock.go
└── utils/                   # Test utilities
    └── test_helper.go
```

### Unit Testing

```go
// tests/unit/user_service_test.go
func TestUserApplicationService_CreateUser(t *testing.T) {
    // Setup
    mockRepo := &mocks.MockUserRepository{}
    userSvc := services.NewUserApplicationService(mockRepo, ...)

    // Configure mock
    mockRepo.On("ExistsByEmail", mock.Anything, "test@example.com").Return(false, nil)
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*entities.User")).Return(nil)

    // Execute
    req := &dto.CreateUserRequest{
        Email:     "test@example.com",
        Password:  "password123",
        FirstName: "Test",
        LastName:  "User",
    }
    result, err := userSvc.CreateUser(context.Background(), req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "test@example.com", result.Email)
    mockRepo.AssertExpectations(t)
}
```

### Integration Testing

```go
// tests/integration/user_api_test.go
func TestUserRegistrationIntegration(t *testing.T) {
    // Setup test database
    db := utils.SetupTestDB(t)
    defer utils.CleanupDatabase(t, db)

    // Setup test server
    router := utils.SetupTestRouter()
    // ... setup dependencies and routes

    // Test user registration
    userRequest := map[string]interface{}{
        "email":      "test@example.com",
        "password":   "password123",
        "first_name": "Test",
        "last_name":  "User",
    }

    w := utils.MakeRequest(t, router, "POST", "/api/v1/users/register", userRequest, nil)
    response := utils.AssertSuccessResponse(t, w, 201)
    
    // Verify response
    assert.Equal(t, "User created successfully", response["message"])
}
```

### Running Tests

```bash
# Unit tests only
go test ./tests/unit/...

# Integration tests (requires database)
go test -tags=integration ./tests/integration/...

# All tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific test
go test -v ./tests/unit -run TestUserApplicationService_CreateUser

# Generate test report
go test -json ./... | tee test-report.json
```

### Mock Generation

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks
mockery --name=UserRepository --dir=services/user-service/internal/domain/repositories --output=tests/mocks
```

## 🔧 Development Workflow

### 1. Feature Development

```bash
# 1. Create feature branch
git checkout -b feature/user-profile-update

# 2. Make changes following clean architecture
# - Add/modify domain entities
# - Update application services
# - Implement infrastructure changes

# 3. Write tests
# - Unit tests for business logic
# - Integration tests for API endpoints

# 4. Run tests and linting
go test ./...
golangci-lint run

# 5. Commit and push
git add .
git commit -m "feat: add user profile update functionality"
git push origin feature/user-profile-update
```

### 2. Database Changes

```bash
# 1. Update entity models
# 2. Create migration script in scripts/migrations/
# 3. Update repository implementations
# 4. Test with fresh database

# Apply migrations
psql -U postgres -d api_server -f scripts/migrations/001_add_user_fields.sql
```

### 3. Configuration Changes

```bash
# 1. Update config struct in shared/pkg/config/
# 2. Update YAML files in configs/
# 3. Update environment variable documentation
# 4. Test with different environments
```

### 4. Adding New Endpoints

```bash
# 1. Add DTO in application/dto/
# 2. Add use case in application/services/
# 3. Add handler in infrastructure/http/handlers/
# 4. Register route in main.go
# 5. Write integration tests
# 6. Update API documentation
```

## 🐛 Debugging

### Local Debugging

#### VS Code Debug Configuration (`.vscode/launch.json`)
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch User Service",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/services/user-service/cmd/server/main.go",
            "env": {
                "USER_SERVICE_ENVIRONMENT": "development",
                "USER_SERVICE_DATABASE_HOST": "localhost"
            },
            "args": []
        }
    ]
}
```

#### Delve Command Line
```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug user service
cd services/user-service
dlv debug cmd/server/main.go

# Set breakpoint and continue
(dlv) break main.main
(dlv) continue
```

### Docker Debugging

#### Enable Debug Mode
```bash
# Start with debug build
docker-compose -f docker-compose.dev.yml up -d

# Attach debugger
docker-compose exec user-service dlv attach --headless --listen=:2345 --api-version=2 1
```

#### Remote Debugging
```bash
# Connect to debug port
dlv connect localhost:2345
```

### Logging for Debugging

```go
// Add debug logs in your code
logger.WithFields(logger.Fields{
    "user_id": userID,
    "operation": "update_profile",
}).Debug("Starting profile update")

// Log with context
logger.WithContext(ctx).
    WithField("request_id", requestID).
    Info("Processing request")
```

## 🔍 Performance Guidelines

### Database Optimization

```go
// Use indexes for queries
func (r *userRepository) GetUsersByRole(ctx context.Context, role entities.UserRole) ([]*entities.User, error) {
    // This query should have index on 'role' column
    var users []UserModel
    err := r.db.Where("role = ?", role).Find(&users).Error
    return convertToEntities(users), err
}

// Use pagination for large datasets
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*entities.User, int64, error) {
    var users []UserModel
    var total int64
    
    r.db.Model(&UserModel{}).Count(&total)
    err := r.db.Offset(offset).Limit(limit).Find(&users).Error
    
    return convertToEntities(users), total, err
}
```

### Memory Management

```go
// Avoid memory leaks in goroutines
func (s *UserApplicationService) ProcessBulkUsers(ctx context.Context, users []User) error {
    // Use context for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Process users
    }
    
    // Process in batches to avoid memory issues
    const batchSize = 100
    for i := 0; i < len(users); i += batchSize {
        end := i + batchSize
        if end > len(users) {
            end = len(users)
        }
        
        if err := s.processBatch(ctx, users[i:end]); err != nil {
            return err
        }
    }
    
    return nil
}
```

### API Performance

```go
// Use appropriate HTTP status codes
func (h *UserHandler) GetUser(c *gin.Context) {
    userID := c.Param("id")
    
    user, err := h.userService.GetUser(c.Request.Context(), userID)
    if err != nil {
        if errors.Is(err, domain.ErrUserNotFound) {
            response.Error(c, http.StatusNotFound, "User not found")
            return
        }
        response.Error(c, http.StatusInternalServerError, "Internal server error")
        return
    }
    
    response.Success(c, http.StatusOK, "User retrieved successfully", user)
}

// Implement caching for frequently accessed data
func (h *UserHandler) GetProfile(c *gin.Context) {
    userID := auth.GetUserIDFromContext(c)
    
    // Check cache first
    if cached := h.cache.Get(fmt.Sprintf("user:%s", userID)); cached != nil {
        response.Success(c, http.StatusOK, "Profile retrieved from cache", cached)
        return
    }
    
    // Fetch from database
    user, err := h.userService.GetUser(c.Request.Context(), userID)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to get profile")
        return
    }
    
    // Cache for future requests
    h.cache.Set(fmt.Sprintf("user:%s", userID), user, 5*time.Minute)
    
    response.Success(c, http.StatusOK, "Profile retrieved successfully", user)
}
```

## 📚 Useful Commands

### Development Commands

```bash
# Start development environment
make dev                                    # If Makefile exists
docker-compose -f docker-compose.dev.yml up -d

# Hot reload development
cd services/user-service && air

# Run tests
go test ./...                              # All tests
go test -v ./tests/unit/...               # Unit tests only
go test -tags=integration ./tests/integration/...  # Integration tests

# Code quality
go fmt ./...                              # Format code
golangci-lint run                         # Run linter
go mod tidy                              # Clean dependencies
```

### Database Commands

```bash
# Connect to database
docker-compose exec postgres psql -U postgres -d api_server

# Run migrations
psql -U postgres -d api_server -f scripts/init-db.sql

# Backup database
docker-compose exec postgres pg_dump -U postgres api_server > backup.sql

# Restore database
docker-compose exec -i postgres psql -U postgres api_server < backup.sql
```

### Docker Commands

```bash
# Build images
docker build -t api-server/user-service .

# View logs
docker-compose logs -f user-service

# Execute commands in container
docker-compose exec user-service sh

# Clean up
docker-compose down -v                    # Stop and remove volumes
docker system prune -af                   # Clean everything
```

## 🤝 Contributing

### Pull Request Process

1. **Fork and Clone**: Fork the repository and clone your fork
2. **Branch**: Create a feature branch from `main`
3. **Develop**: Make your changes following the coding standards
4. **Test**: Ensure all tests pass and add new tests for new features
5. **Document**: Update documentation for API changes
6. **Commit**: Use conventional commit messages
7. **Push**: Push to your fork and create a pull request

### Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples**:
```
feat(user): add user profile update endpoint

Add PUT /api/v1/users/profile endpoint to allow users to update
their profile information including first name and last name.

Closes #123
```

### Code Review Checklist

- [ ] Code follows Go style guide
- [ ] Clean architecture layers are respected
- [ ] All tests pass
- [ ] New features have tests
- [ ] API changes are documented
- [ ] Error handling is appropriate
- [ ] Logging is meaningful
- [ ] No security vulnerabilities
- [ ] Performance considerations addressed

## 🚨 Troubleshooting

### Common Issues

#### Hot Reload Not Working
```bash
# Check Air configuration
cat .air.toml

# Restart Air
pkill air
air
```

#### Database Connection Issues
```bash
# Check if database is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U postgres
```

#### Import Path Issues
```bash
# Clean module cache
go clean -modcache

# Reinitialize modules
rm go.mod go.sum
go mod init github.com/your-org/api-server
go mod tidy
```

#### Test Failures
```bash
# Run specific test with verbose output
go test -v ./tests/unit -run TestUserService

# Run tests with race detection
go test -race ./...

# Clean test cache
go clean -testcache
```

For more troubleshooting help, check the [Setup Guide](SETUP_GUIDE.md) troubleshooting section.