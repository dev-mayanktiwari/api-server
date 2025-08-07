# Database Guide

Complete guide to database management, migrations, and data operations for the API Server project.

## 🗄️ Database Overview

### Technology Stack
- **Database**: PostgreSQL 15+
- **ORM**: GORM (Go Object-Relational Mapping)
- **Connection Pooling**: pgxpool
- **Migrations**: SQL scripts with versioning

### Database Schema

```sql
-- Users table (primary entity)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User sessions (for JWT tracking)
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Additional tables for features...
```

## 🚀 Database Setup

### Local Development Setup

#### 1. Using Docker (Recommended)
```bash
# Start PostgreSQL container
docker-compose up -d postgres

# Check if database is ready
docker-compose exec postgres pg_isready -U postgres

# Connect to database
docker-compose exec postgres psql -U postgres -d api_server
```

#### 2. Manual Installation
```bash
# Install PostgreSQL (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install postgresql postgresql-contrib

# Install PostgreSQL (macOS)
brew install postgresql
brew services start postgresql

# Create database and user
sudo -u postgres createdb api_server
sudo -u postgres createuser --superuser api_user
```

### Database Initialization

#### 1. Run Initialization Script
```bash
# Via Docker
docker-compose exec postgres psql -U postgres -d api_server -f /docker-entrypoint-initdb.d/init-db.sql

# Via local psql
psql -U postgres -d api_server -f scripts/init-db.sql
```

#### 2. Verify Database Setup
```sql
-- Connect to database
\c api_server;

-- List tables
\dt

-- Check users table structure
\d users;

-- Verify extensions
\dx;
```

## 📊 Database Configuration

### Connection Configuration

```yaml
# configs/development/config.yaml
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
  conn_max_idle_time: 5m
  log_level: "info"
```

### Environment Variables
```bash
# Database connection
export USER_SERVICE_DATABASE_HOST=localhost
export USER_SERVICE_DATABASE_PORT=5432
export USER_SERVICE_DATABASE_USERNAME=postgres
export USER_SERVICE_DATABASE_PASSWORD=postgres
export USER_SERVICE_DATABASE_DATABASE=api_server
export USER_SERVICE_DATABASE_SSLMODE=disable

# Connection pool settings
export USER_SERVICE_DATABASE_MAXOPENCONNS=25
export USER_SERVICE_DATABASE_MAXIDLECONNS=5
export USER_SERVICE_DATABASE_CONNMAXLIFETIME=30m
```

### Connection Pool Tuning

```go
// Database connection configuration
config := &database.Config{
    Host:               "localhost",
    Port:               5432,
    Username:           "postgres",
    Password:           "postgres",
    Database:           "api_server",
    SSLMode:            "disable",
    MaxOpenConns:       25,    // Maximum open connections
    MaxIdleConns:       5,     // Maximum idle connections
    ConnMaxLifetime:    30 * time.Minute, // Connection lifetime
    ConnMaxIdleTime:    5 * time.Minute,  // Idle connection timeout
}
```

## 🏗️ Database Models

### Domain Entity
```go
// internal/domain/entities/user.go
type User struct {
    ID                string
    Email             string
    Password          string
    FirstName         string
    LastName          string
    Role              UserRole
    Status            UserStatus
    EmailVerified     bool
    EmailVerifiedAt   *time.Time
    PasswordChangedAt *time.Time
    LastLoginAt       *time.Time
    LoginCount        int
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

type UserRole string
const (
    RoleUser  UserRole = "user"
    RoleAdmin UserRole = "admin"
)

type UserStatus string
const (
    StatusActive   UserStatus = "active"
    StatusInactive UserStatus = "inactive"
)
```

### GORM Model
```go
// internal/infrastructure/database/models.go
type UserModel struct {
    ID                string         `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
    Email             string         `gorm:"type:varchar(255);uniqueIndex;not null"`
    Password          string         `gorm:"type:varchar(255);not null"`
    FirstName         string         `gorm:"type:varchar(100);not null"`
    LastName          string         `gorm:"type:varchar(100);not null"`
    Role              string         `gorm:"type:varchar(50);not null;default:'user'"`
    Status            string         `gorm:"type:varchar(50);not null;default:'active'"`
    EmailVerified     bool           `gorm:"default:false"`
    EmailVerifiedAt   *time.Time     `gorm:"type:timestamp"`
    PasswordChangedAt *time.Time     `gorm:"type:timestamp"`
    LastLoginAt       *time.Time     `gorm:"type:timestamp"`
    LoginCount        int            `gorm:"default:0"`
    CreatedAt         time.Time      `gorm:"autoCreateTime"`
    UpdatedAt         time.Time      `gorm:"autoUpdateTime"`
}

func (UserModel) TableName() string {
    return "users"
}
```

### Model Conversion
```go
// Convert GORM model to domain entity
func (m *UserModel) ToEntity() *entities.User {
    return &entities.User{
        ID:                m.ID,
        Email:             m.Email,
        FirstName:         m.FirstName,
        LastName:          m.LastName,
        Role:              entities.UserRole(m.Role),
        Status:            entities.UserStatus(m.Status),
        EmailVerified:     m.EmailVerified,
        EmailVerifiedAt:   m.EmailVerifiedAt,
        PasswordChangedAt: m.PasswordChangedAt,
        LastLoginAt:       m.LastLoginAt,
        LoginCount:        m.LoginCount,
        CreatedAt:         m.CreatedAt,
        UpdatedAt:         m.UpdatedAt,
    }
}

// Convert domain entity to GORM model
func NewUserModelFromEntity(user *entities.User) *UserModel {
    return &UserModel{
        ID:                user.ID,
        Email:             user.Email,
        Password:          user.Password,
        FirstName:         user.FirstName,
        LastName:          user.LastName,
        Role:              string(user.Role),
        Status:            string(user.Status),
        EmailVerified:     user.EmailVerified,
        EmailVerifiedAt:   user.EmailVerifiedAt,
        PasswordChangedAt: user.PasswordChangedAt,
        LastLoginAt:       user.LastLoginAt,
        LoginCount:        user.LoginCount,
        CreatedAt:         user.CreatedAt,
        UpdatedAt:         user.UpdatedAt,
    }
}
```

## 🔄 Database Operations

### Repository Pattern Implementation

```go
// internal/infrastructure/database/user_repository_impl.go
type userRepository struct {
    db     database.DB
    logger logger.Logger
}

func NewUserRepository(db database.DB, logger logger.Logger) repositories.UserRepository {
    return &userRepository{
        db:     db,
        logger: logger,
    }
}

// Create user
func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
    model := NewUserModelFromEntity(user)
    
    if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
        r.logger.WithContext(ctx).WithError(err).Error("Failed to create user")
        return err
    }
    
    // Update entity with generated ID
    user.ID = model.ID
    user.CreatedAt = model.CreatedAt
    user.UpdatedAt = model.UpdatedAt
    
    return nil
}

// Get user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
    var model UserModel
    
    if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, domain.ErrUserNotFound
        }
        r.logger.WithContext(ctx).WithError(err).Error("Failed to get user by ID")
        return nil, err
    }
    
    return model.ToEntity(), nil
}

// Update user
func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
    model := NewUserModelFromEntity(user)
    
    if err := r.db.WithContext(ctx).Save(model).Error; err != nil {
        r.logger.WithContext(ctx).WithError(err).Error("Failed to update user")
        return err
    }
    
    user.UpdatedAt = model.UpdatedAt
    return nil
}

// List with pagination
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]*entities.User, int64, error) {
    var models []UserModel
    var total int64
    
    // Get total count
    if err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    
    // Get paginated results
    if err := r.db.WithContext(ctx).
        Offset(offset).
        Limit(limit).
        Order("created_at DESC").
        Find(&models).Error; err != nil {
        return nil, 0, err
    }
    
    // Convert to entities
    users := make([]*entities.User, len(models))
    for i, model := range models {
        users[i] = model.ToEntity()
    }
    
    return users, total, nil
}
```

### Transaction Support

```go
// Application service with transaction
func (s *UserApplicationService) CreateUserWithProfile(ctx context.Context, req *dto.CreateUserWithProfileRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Create user
        user := entities.NewUser(req.Email, req.Password, req.FirstName, req.LastName)
        if err := tx.WithContext(ctx).Create(NewUserModelFromEntity(user)).Error; err != nil {
            return err
        }
        
        // Create profile
        profile := entities.NewUserProfile(user.ID, req.ProfileData)
        if err := tx.WithContext(ctx).Create(NewProfileModelFromEntity(profile)).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## 📈 Database Migrations

### Migration Strategy

We use SQL migration files with version numbers for database schema changes.

### Migration File Structure
```
scripts/migrations/
├── 001_initial_schema.sql          # Initial database schema
├── 002_add_user_sessions.sql       # Add user sessions table
├── 003_add_user_indexes.sql        # Add performance indexes
└── 004_add_audit_logs.sql          # Add audit logging
```

### Migration File Format

```sql
-- Migration: 001_initial_schema.sql
-- Description: Initial database schema with users table
-- Author: Developer Name
-- Date: 2024-01-01

BEGIN;

-- Create UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMIT;
```

### Running Migrations

#### Manual Migration
```bash
# Run specific migration
psql -U postgres -d api_server -f scripts/migrations/001_initial_schema.sql

# Run all migrations (in order)
for migration in scripts/migrations/*.sql; do
    echo "Running $migration"
    psql -U postgres -d api_server -f "$migration"
done
```

#### Docker Migration
```bash
# Copy migration to container and run
docker-compose exec postgres psql -U postgres -d api_server -f /migrations/001_initial_schema.sql

# Or mount migrations directory
docker run --rm -v $(pwd)/scripts/migrations:/migrations postgres:15-alpine \
    psql -h host -U postgres -d api_server -f /migrations/001_initial_schema.sql
```

### Migration Best Practices

1. **Always use transactions** for migration scripts
2. **Use IF NOT EXISTS** for CREATE statements
3. **Version migrations** with incremental numbers
4. **Test migrations** on copy of production data
5. **Include rollback scripts** for complex changes
6. **Document breaking changes** clearly

### Rollback Migrations

```sql
-- Rollback: 002_add_user_sessions.sql
-- Description: Rollback user sessions table creation
-- Date: 2024-01-02

BEGIN;

-- Drop user sessions table
DROP TABLE IF EXISTS user_sessions CASCADE;

-- Drop related indexes
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_token_hash;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;

COMMIT;
```

## 📊 Database Monitoring

### Health Checks

```go
// Database health check implementation
func (db *Database) HealthCheck(ctx context.Context) error {
    // Test basic connectivity
    sqlDB, err := db.DB()
    if err != nil {
        return fmt.Errorf("failed to get database instance: %w", err)
    }
    
    // Ping database
    if err := sqlDB.PingContext(ctx); err != nil {
        return fmt.Errorf("database ping failed: %w", err)
    }
    
    // Test query execution
    var result int
    if err := db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error; err != nil {
        return fmt.Errorf("test query failed: %w", err)
    }
    
    return nil
}
```

### Connection Pool Monitoring

```go
// Get database statistics
func (db *Database) GetStats() map[string]interface{} {
    sqlDB, _ := db.DB()
    stats := sqlDB.Stats()
    
    return map[string]interface{}{
        "open_connections":     stats.OpenConnections,
        "in_use_connections":   stats.InUse,
        "idle_connections":     stats.Idle,
        "wait_count":          stats.WaitCount,
        "wait_duration":       stats.WaitDuration.String(),
        "max_idle_closed":     stats.MaxIdleClosed,
        "max_lifetime_closed": stats.MaxLifetimeClosed,
    }
}
```

### Query Performance

```sql
-- Find slow queries
SELECT
    query,
    mean_exec_time,
    calls,
    total_exec_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements
WHERE mean_exec_time > 1000  -- Queries taking more than 1 second
ORDER BY mean_exec_time DESC
LIMIT 10;

-- Check connection activity
SELECT
    pid,
    usename,
    application_name,
    client_addr,
    state,
    query_start,
    query
FROM pg_stat_activity
WHERE state != 'idle'
ORDER BY query_start;
```

## 🎯 Database Best Practices

### Query Optimization

```go
// Good: Use specific columns
func (r *userRepository) GetUserEmail(ctx context.Context, id string) (string, error) {
    var email string
    err := r.db.WithContext(ctx).
        Model(&UserModel{}).
        Where("id = ?", id).
        Select("email").
        Scan(&email).Error
    return email, err
}

// Good: Use indexes efficiently
func (r *userRepository) GetActiveUsers(ctx context.Context) ([]*entities.User, error) {
    var models []UserModel
    err := r.db.WithContext(ctx).
        Where("status = ?", "active").  // Uses idx_users_status
        Order("created_at DESC").       // Uses idx_users_created_at
        Find(&models).Error
    // ...
}

// Good: Use prepared statements (GORM does this automatically)
func (r *userRepository) GetUsersByRole(ctx context.Context, role string) ([]*entities.User, error) {
    var models []UserModel
    err := r.db.WithContext(ctx).
        Where("role = ?", role).  // Automatically prepared
        Find(&models).Error
    // ...
}
```

### Avoid N+1 Queries

```go
// Bad: N+1 query problem
func (r *userRepository) GetUsersWithProfiles(ctx context.Context) ([]*entities.User, error) {
    users, err := r.List(ctx, 0, 100)
    if err != nil {
        return nil, err
    }
    
    // This will execute N additional queries
    for _, user := range users {
        profile, _ := r.GetUserProfile(ctx, user.ID)  // N+1 problem!
        user.Profile = profile
    }
    
    return users, nil
}

// Good: Use joins or preloading
func (r *userRepository) GetUsersWithProfiles(ctx context.Context) ([]*entities.User, error) {
    var models []UserModel
    err := r.db.WithContext(ctx).
        Preload("Profile").  // Load related data in single query
        Find(&models).Error
    // ...
}
```

### Connection Management

```go
// Good: Use context for timeouts
func (r *userRepository) CreateUser(ctx context.Context, user *entities.User) error {
    // Context will handle timeouts and cancellation
    return r.db.WithContext(ctx).Create(user).Error
}

// Good: Use transactions for related operations
func (s *UserService) CreateUserWithProfile(ctx context.Context, req *CreateUserRequest) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        // Both operations succeed or both fail
        user := &UserModel{...}
        if err := tx.Create(user).Error; err != nil {
            return err
        }
        
        profile := &ProfileModel{UserID: user.ID, ...}
        return tx.Create(profile).Error
    })
}
```

## 🔧 Database Tools and Utilities

### Database Scripts

```bash
# scripts/db-backup.sh
#!/bin/bash
BACKUP_DIR="./backups"
DB_NAME="api_server"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

mkdir -p $BACKUP_DIR
docker-compose exec postgres pg_dump -U postgres $DB_NAME > "$BACKUP_DIR/backup_${TIMESTAMP}.sql"
echo "Backup created: $BACKUP_DIR/backup_${TIMESTAMP}.sql"
```

```bash
# scripts/db-restore.sh
#!/bin/bash
if [ -z "$1" ]; then
    echo "Usage: $0 <backup_file>"
    exit 1
fi

docker-compose exec -T postgres psql -U postgres api_server < "$1"
echo "Database restored from: $1"
```

### Database Seeding

```sql
-- scripts/seed-data.sql
-- Insert default admin user
INSERT INTO users (
    email, 
    password, 
    first_name, 
    last_name, 
    role, 
    status,
    email_verified
) VALUES (
    'admin@api-server.com',
    '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', -- password: admin123
    'System',
    'Administrator',
    'admin',
    'active',
    true
) ON CONFLICT (email) DO NOTHING;

-- Insert test users for development
INSERT INTO users (email, password, first_name, last_name, role) VALUES
('user1@test.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'User1', 'user'),
('user2@test.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Test', 'User2', 'user')
ON CONFLICT (email) DO NOTHING;
```

## 🚨 Troubleshooting

### Common Database Issues

#### Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check PostgreSQL logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U postgres

# Connect manually
docker-compose exec postgres psql -U postgres -d api_server
```

#### Performance Issues
```sql
-- Check slow queries
SELECT * FROM pg_stat_statements WHERE mean_exec_time > 1000;

-- Check locks
SELECT * FROM pg_locks WHERE NOT granted;

-- Check connection count
SELECT count(*) FROM pg_stat_activity;
```

#### Migration Issues
```bash
# Check migration status (manual tracking needed)
psql -U postgres -d api_server -c "SELECT * FROM schema_migrations;"

# Rollback last migration
psql -U postgres -d api_server -f scripts/rollbacks/002_rollback.sql
```

For more database troubleshooting, check the [Setup Guide](SETUP_GUIDE.md) database section.