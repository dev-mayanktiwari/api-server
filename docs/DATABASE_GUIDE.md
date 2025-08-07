# Database Guide

Complete guide to database management, schema design, migrations, and operations for the API Server microservices project.

## 🗄️ Database Overview

### Technology Stack
- **Database**: PostgreSQL 15+
- **ORM**: GORM (Go Object-Relational Mapping)
- **Connection Pooling**: pgxpool for efficient connection management
- **Extensions**: UUID generation, pgcrypto for security
- **Migration Strategy**: SQL scripts with versioning

### Database Architecture
The database is shared across microservices but each service has clear data ownership:
- **User Service**: Owns user data, profiles, and user-related operations
- **Auth Service**: Owns session data, tokens, and authentication-related data
- **Shared Tables**: Audit logs, rate limiting data

## 📊 Database Schema

### Core Tables

#### Users Table
Primary table for user management across all services.

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,              -- bcrypt hashed
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',    -- 'user' | 'admin'
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active' | 'inactive' | 'suspended'
    email_verified BOOLEAN DEFAULT FALSE,
    email_verified_at TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    login_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `idx_users_email` - Unique email lookups
- `idx_users_status` - Status filtering for admin queries
- `idx_users_role` - Role-based access control
- `idx_users_created_at` - Date-based sorting and filtering

#### User Sessions Table
Manages JWT tokens and user sessions for the Auth Service.

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,            -- Hashed JWT token
    device_info JSONB,                           -- Device information
    ip_address INET,                             -- Client IP address
    user_agent TEXT,                             -- Client user agent
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

**Indexes**:
- `idx_user_sessions_user_id` - User session lookups
- `idx_user_sessions_token_hash` - Token validation
- `idx_user_sessions_expires_at` - Cleanup expired sessions

#### Password Reset Tokens
Secure password recovery functionality.

```sql
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,                 -- Secure random token
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,                  -- One-time use
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Email Verification Tokens
Email verification workflow support.

```sql
CREATE TABLE email_verification_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,                 -- Secure random token
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    used BOOLEAN DEFAULT FALSE,                  -- One-time use
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Audit Logs
Complete audit trail for security and compliance.

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,               -- 'create', 'update', 'delete', 'login', etc.
    resource VARCHAR(100),                      -- 'user', 'session', etc.
    resource_id UUID,                          -- ID of affected resource
    old_values JSONB,                          -- Previous state (for updates)
    new_values JSONB,                          -- New state
    ip_address INET,                           -- Client IP
    user_agent TEXT,                           -- Client user agent
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Rate Limits
Database-backed rate limiting for API endpoints.

```sql
CREATE TABLE rate_limits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key VARCHAR(255) NOT NULL,                  -- Rate limit key (IP, user_id, etc.)
    count INTEGER NOT NULL DEFAULT 1,           -- Current request count
    reset_time TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## 🚀 Database Setup

### Local Development Setup

#### 1. Using Docker (Recommended)
```bash
# Start PostgreSQL container with initialization
docker-compose -f docker-compose.microservices.yml up -d postgres

# Check if database is ready
docker-compose -f docker-compose.microservices.yml exec postgres pg_isready -U postgres

# Connect to database
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server
```

#### 2. Manual PostgreSQL Installation

##### Ubuntu/Debian
```bash
# Update package list
sudo apt-get update

# Install PostgreSQL and extensions
sudo apt-get install postgresql postgresql-contrib postgresql-15

# Start PostgreSQL service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database and user
sudo -u postgres createdb api_server
sudo -u postgres psql -c "ALTER USER postgres PASSWORD 'postgres';"
```

##### macOS
```bash
# Install via Homebrew
brew install postgresql@15

# Start PostgreSQL service
brew services start postgresql@15

# Create database
createdb api_server

# Connect as superuser
psql postgres
```

##### Windows
1. Download PostgreSQL installer from https://www.postgresql.org/download/windows/
2. Run installer and follow setup wizard
3. Remember the superuser password
4. Use pgAdmin or psql to create database

### Database Initialization

#### 1. Initialize Database Schema
```bash
# Via Docker
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server -f /docker-entrypoint-initdb.d/init-db.sql

# Via local psql
psql -U postgres -d api_server -f scripts/init-db.sql

# Via psql with specific host
psql -h localhost -U postgres -d api_server -f scripts/init-db.sql
```

#### 2. Verify Database Setup
```sql
-- Connect to database
\c api_server;

-- List all tables
\dt

-- Check table structure
\d users
\d user_sessions

-- Verify indexes
\di

-- Check extensions
\dx

-- Test default admin user
SELECT id, email, role, status FROM users WHERE role = 'admin';
```

#### 3. Database Permissions
```sql
-- Grant permissions to application user (if created)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO api_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO api_user;

-- Grant permissions for functions
GRANT EXECUTE ON FUNCTION cleanup_expired_tokens() TO api_user;
GRANT EXECUTE ON FUNCTION update_updated_at_column() TO api_user;
```

## 🔧 Database Configuration

### Connection Configuration

#### Environment Variables
```bash
# Primary database connection
export DB_HOST=localhost
export DB_PORT=5432
export DB_USERNAME=postgres
export DB_PASSWORD=postgres
export DB_DATABASE=api_server
export DB_SSL_MODE=disable

# Connection pooling
export DB_MAX_OPEN_CONNS=25
export DB_MAX_IDLE_CONNS=5
export DB_CONN_MAX_LIFETIME=30m

# For testing
export DB_TEST_DATABASE=api_server_test
```

#### Service-Specific Configuration

##### User Service (configs/development/config.yaml)
```yaml
database:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "postgres"
  database: "api_server"
  sslmode: "disable"
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "30m"
  auto_migrate: true
  log_level: "info"
```

##### Auth Service (configs/auth-service/config.yaml)
```yaml
database:
  host: "postgres"
  port: 5432
  username: "postgres"
  password: "postgres"
  database: "api_server"
  sslmode: "disable"
  max_open_conns: 15
  max_idle_conns: 3
  conn_max_lifetime: "30m"
  auto_migrate: true
  log_level: "warn"
```

### Connection Pooling Best Practices

#### Pool Size Guidelines
- **Development**: 5-10 connections per service
- **Staging**: 15-25 connections per service  
- **Production**: 25-50 connections per service (based on load)

#### Connection Management
```go
// Example GORM configuration
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})

sqlDB, err := db.DB()

// Configure connection pool
sqlDB.SetMaxOpenConns(25)        // Maximum connections
sqlDB.SetMaxIdleConns(5)         // Idle connections
sqlDB.SetConnMaxLifetime(time.Hour) // Connection lifetime
```

## 🔄 Database Operations

### GORM Auto-Migration

#### User Service Models
```go
// Auto-migrate user-related tables
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &entities.User{},
        &entities.PasswordResetToken{},
        &entities.EmailVerificationToken{},
        &entities.AuditLog{},
    )
}
```

#### Auth Service Models
```go
// Auto-migrate auth-related tables
func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &entities.UserSession{},
        &entities.RateLimit{},
    )
}
```

### Manual Migrations

#### Creating Migration Scripts
```bash
# Create migration directory structure
mkdir -p migrations/{up,down}

# Create migration files
cat > migrations/up/001_add_user_preferences.sql << 'EOF'
CREATE TABLE user_preferences (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(50) DEFAULT 'light',
    language VARCHAR(10) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    notifications JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
EOF

cat > migrations/down/001_add_user_preferences.sql << 'EOF'
DROP INDEX IF EXISTS idx_user_preferences_user_id;
DROP TABLE IF EXISTS user_preferences;
EOF
```

#### Running Migrations
```bash
# Apply migration
psql -U postgres -d api_server -f migrations/up/001_add_user_preferences.sql

# Rollback migration
psql -U postgres -d api_server -f migrations/down/001_add_user_preferences.sql
```

### Database Maintenance

#### Cleanup Expired Data
```sql
-- Manual cleanup (or call function)
SELECT cleanup_expired_tokens();

-- Check what will be cleaned up
SELECT 'password_reset_tokens' as table_name, count(*) as expired_count
FROM password_reset_tokens 
WHERE expires_at < CURRENT_TIMESTAMP

UNION ALL

SELECT 'email_verification_tokens', count(*)
FROM email_verification_tokens 
WHERE expires_at < CURRENT_TIMESTAMP

UNION ALL

SELECT 'user_sessions', count(*)
FROM user_sessions 
WHERE expires_at < CURRENT_TIMESTAMP

UNION ALL

SELECT 'audit_logs (>90 days)', count(*)
FROM audit_logs 
WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '90 days';
```

#### Database Statistics
```sql
-- Table sizes
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public'
ORDER BY tablename, attname;

-- Index usage
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes 
ORDER BY idx_scan DESC;

-- Connection stats
SELECT 
    datname,
    numbackends as connections,
    xact_commit as committed_transactions,
    xact_rollback as rolled_back_transactions,
    blks_read as blocks_read,
    blks_hit as blocks_hit,
    tup_returned as tuples_returned,
    tup_fetched as tuples_fetched,
    tup_inserted as tuples_inserted,
    tup_updated as tuples_updated,
    tup_deleted as tuples_deleted
FROM pg_stat_database 
WHERE datname = 'api_server';
```

### Performance Optimization

#### Index Optimization
```sql
-- Find missing indexes
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public'
  AND n_distinct > 100
ORDER BY n_distinct DESC;

-- Find unused indexes
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_scan
FROM pg_stat_user_indexes 
WHERE idx_scan = 0
ORDER BY tablename;

-- Analyze query performance
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';
EXPLAIN ANALYZE SELECT * FROM users WHERE status = 'active' ORDER BY created_at DESC LIMIT 10;
```

#### Query Optimization Examples
```sql
-- Efficient user search with pagination
SELECT id, email, first_name, last_name, status, created_at
FROM users 
WHERE status = 'active'
  AND (first_name ILIKE '%search%' OR last_name ILIKE '%search%' OR email ILIKE '%search%')
ORDER BY created_at DESC
LIMIT 20 OFFSET 0;

-- Efficient session cleanup
DELETE FROM user_sessions 
WHERE expires_at < CURRENT_TIMESTAMP - INTERVAL '1 hour';

-- Count active users efficiently
SELECT 
    COUNT(*) FILTER (WHERE status = 'active') as active_users,
    COUNT(*) FILTER (WHERE status = 'inactive') as inactive_users,
    COUNT(*) FILTER (WHERE role = 'admin') as admin_users
FROM users;
```

## 🧪 Testing Database

### Test Database Setup
```bash
# Create test database
docker-compose -f docker-compose.microservices.yml exec postgres \
  createdb -U postgres api_server_test

# Initialize test database
docker-compose -f docker-compose.microservices.yml exec postgres \
  psql -U postgres -d api_server_test -f /docker-entrypoint-initdb.d/init-db.sql

# Alternative: Local test database
createdb api_server_test
psql -U postgres -d api_server_test -f scripts/init-db.sql
```

### Test Data Management

#### Test Fixtures
```go
// tests/utils/fixtures.go
package utils

import (
    "context"
    "gorm.io/gorm"
)

func CreateTestUser(db *gorm.DB) *domain.User {
    user := &domain.User{
        Email:     "test@example.com",
        Password:  "$2a$10$hashedpassword",
        FirstName: "Test",
        LastName:  "User",
        Role:      "user",
        Status:    "active",
    }
    db.Create(user)
    return user
}

func CreateTestAdmin(db *gorm.DB) *domain.User {
    admin := &domain.User{
        Email:     "admin@example.com", 
        Password:  "$2a$10$hashedpassword",
        FirstName: "Admin",
        LastName:  "User",
        Role:      "admin",
        Status:    "active",
    }
    db.Create(admin)
    return admin
}

func CleanupTestData(db *gorm.DB) {
    db.Exec("TRUNCATE users, user_sessions, audit_logs, rate_limits RESTART IDENTITY CASCADE")
}
```

#### Database Test Utilities
```go
// tests/utils/database.go
package utils

import (
    "fmt"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
    "testing"
)

func SetupTestDB(t *testing.T) *gorm.DB {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        "localhost", "5432", "postgres", "postgres", "api_server_test",
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // Auto-migrate test schema
    err = db.AutoMigrate(&domain.User{}, &domain.UserSession{})
    if err != nil {
        t.Fatalf("Failed to migrate test database: %v", err)
    }
    
    return db
}

func TeardownTestDB(t *testing.T, db *gorm.DB) {
    CleanupTestData(db)
    
    sqlDB, err := db.DB()
    if err == nil {
        sqlDB.Close()
    }
}
```

### Integration Test Examples
```go
func TestUserRepository_Create(t *testing.T) {
    db := SetupTestDB(t)
    defer TeardownTestDB(t, db)
    
    repo := database.NewUserRepository(db)
    
    user := &domain.User{
        Email:     "test@integration.com",
        Password:  "hashedpassword",
        FirstName: "Integration",
        LastName:  "Test",
    }
    
    createdUser, err := repo.Create(context.Background(), user)
    
    assert.NoError(t, err)
    assert.NotEmpty(t, createdUser.ID)
    assert.Equal(t, user.Email, createdUser.Email)
    
    // Verify in database
    var dbUser domain.User
    err = db.Where("email = ?", user.Email).First(&dbUser).Error
    assert.NoError(t, err)
    assert.Equal(t, user.Email, dbUser.Email)
}
```

## 🔐 Database Security

### Connection Security
```yaml
# Production database configuration
database:
  host: "your-postgres-host"
  port: 5432
  username: "api_server_user"
  password: "${DB_PASSWORD}"  # From environment
  database: "api_server"
  sslmode: "require"          # Force SSL
  sslcert: "/path/to/client.crt"
  sslkey: "/path/to/client.key"
  sslrootcert: "/path/to/ca.crt"
```

### Access Control
```sql
-- Create application-specific user
CREATE USER api_server_user WITH PASSWORD 'secure_password';

-- Grant minimal required permissions
GRANT CONNECT ON DATABASE api_server TO api_server_user;
GRANT USAGE ON SCHEMA public TO api_server_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO api_server_user;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO api_server_user;

-- Revoke unnecessary permissions
REVOKE CREATE ON SCHEMA public FROM api_server_user;
REVOKE ALL PRIVILEGES ON SCHEMA information_schema FROM api_server_user;
```

### Data Encryption
```sql
-- Encrypt sensitive data in application layer
-- Passwords are already hashed with bcrypt
-- Additional encryption for sensitive fields:

-- Example: Encrypt user data at rest
CREATE OR REPLACE FUNCTION encrypt_sensitive_data(data TEXT) 
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_encrypt(data, 'your-encryption-key');
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION decrypt_sensitive_data(encrypted_data TEXT) 
RETURNS TEXT AS $$
BEGIN
    RETURN pgp_sym_decrypt(encrypted_data, 'your-encryption-key');
END;
$$ LANGUAGE plpgsql;
```

## 📊 Monitoring and Maintenance

### Database Health Monitoring
```sql
-- Connection monitoring
SELECT 
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active_connections,
    count(*) FILTER (WHERE state = 'idle') as idle_connections
FROM pg_stat_activity 
WHERE datname = 'api_server';

-- Lock monitoring
SELECT 
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_statement,
    blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON (blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation)
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;

-- Slow query monitoring
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
ORDER BY total_time DESC 
LIMIT 10;
```

### Automated Maintenance
```bash
#!/bin/bash
# scripts/db-maintenance.sh

# Clean up expired tokens
psql -U postgres -d api_server -c "SELECT cleanup_expired_tokens();"

# Update table statistics
psql -U postgres -d api_server -c "ANALYZE;"

# Reindex if needed
psql -U postgres -d api_server -c "REINDEX DATABASE api_server;"

# Vacuum to reclaim space
psql -U postgres -d api_server -c "VACUUM ANALYZE;"

echo "Database maintenance completed at $(date)"
```

### Backup Strategy
```bash
#!/bin/bash
# scripts/backup-db.sh

BACKUP_DIR="/backups"
DATE=$(date +%Y%m%d_%H%M%S)
DB_NAME="api_server"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Create full database backup
pg_dump -U postgres -h localhost "$DB_NAME" > "$BACKUP_DIR/backup_${DB_NAME}_${DATE}.sql"

# Create compressed backup
pg_dump -U postgres -h localhost "$DB_NAME" | gzip > "$BACKUP_DIR/backup_${DB_NAME}_${DATE}.sql.gz"

# Schema-only backup
pg_dump -U postgres -h localhost --schema-only "$DB_NAME" > "$BACKUP_DIR/schema_${DB_NAME}_${DATE}.sql"

# Clean up old backups (keep 30 days)
find "$BACKUP_DIR" -name "backup_${DB_NAME}_*" -mtime +30 -delete

echo "Database backup completed: backup_${DB_NAME}_${DATE}.sql.gz"
```

### Restore Procedures
```bash
# Restore from backup
createdb api_server_restored
psql -U postgres -d api_server_restored -f backup_api_server_20240101_120000.sql

# Restore from compressed backup
createdb api_server_restored
gunzip -c backup_api_server_20240101_120000.sql.gz | psql -U postgres -d api_server_restored
```

## 🎯 Best Practices

### Development Best Practices
1. **Use Transactions**: Wrap related operations in transactions
2. **Connection Pooling**: Configure appropriate pool sizes
3. **Index Strategy**: Add indexes for frequently queried columns
4. **Query Optimization**: Use EXPLAIN ANALYZE for query optimization
5. **Data Validation**: Validate data at application layer and use DB constraints

### Production Best Practices
1. **Monitoring**: Implement comprehensive database monitoring
2. **Backups**: Regular automated backups with tested restore procedures
3. **Security**: Use least privilege access and SSL connections
4. **Maintenance**: Regular VACUUM, ANALYZE, and REINDEX operations
5. **Scaling**: Consider read replicas for high-traffic applications

### Schema Evolution Best Practices
1. **Migrations**: Use versioned migration scripts
2. **Backward Compatibility**: Ensure schema changes don't break existing services
3. **Testing**: Test migrations on staging before production
4. **Rollback Plans**: Always have rollback procedures ready

This comprehensive database guide provides all the information needed to manage the PostgreSQL database effectively across the microservices architecture.