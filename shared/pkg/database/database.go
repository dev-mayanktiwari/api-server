// Package database provides database connection and management utilities for the API server.
// It supports PostgreSQL with connection pooling, health checks, and migration support.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dev-mayanktiwari/api-server/shared/pkg/config"
	"github.com/dev-mayanktiwari/api-server/shared/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// DB wraps gorm.DB with additional functionality
type DB struct {
	*gorm.DB
	config *config.DatabaseConfig
	logger *logger.Logger
}

// Config represents database configuration
type Config = config.DatabaseConfig

// DefaultConfig returns default database configuration
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		Username:        "postgres",
		Password:        "postgres",
		Database:        "api_server",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}
}

// Connect establishes a connection to the PostgreSQL database
func Connect(cfg *Config, log *logger.Logger) (*DB, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	
	if log == nil {
		log = logger.Default()
	}
	
	// Build connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode,
	)
	
	// Configure GORM logger
	var gormLogLevel gormLogger.LogLevel
	switch log.GetConfig().Level {
	case "debug":
		gormLogLevel = gormLogger.Info
	case "info":
		gormLogLevel = gormLogger.Warn
	default:
		gormLogLevel = gormLogger.Error
	}
	
	gormConfig := &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:                              true,
		DisableForeignKeyConstraintWhenMigrating: false,
	}
	
	// Open connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	
	// Get underlying sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	
	log.WithFields(logger.Fields{
		"host":     cfg.Host,
		"port":     cfg.Port,
		"database": cfg.Database,
	}).Info("Connected to database")
	
	return &DB{
		DB:     db,
		config: cfg,
		logger: log,
	}, nil
}

// GetConfig returns the database configuration
func (db *DB) GetConfig() *Config {
	return db.config
}

// GetLogger returns the logger instance
func (db *DB) GetLogger() *logger.Logger {
	return db.logger
}

// HealthCheck performs a health check on the database connection
func (db *DB) HealthCheck(ctx context.Context) error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	// Check if we can ping the database
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	// Check connection stats
	stats := sqlDB.Stats()
	if stats.OpenConnections == 0 {
		return fmt.Errorf("no open database connections")
	}
	
	return nil
}

// GetStats returns database connection statistics
func (db *DB) GetStats() map[string]interface{} {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": err.Error(),
		}
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections":     stats.MaxOpenConnections,
		"open_connections":         stats.OpenConnections,
		"in_use":                  stats.InUse,
		"idle":                    stats.Idle,
		"wait_count":              stats.WaitCount,
		"wait_duration":           stats.WaitDuration.String(),
		"max_idle_closed":         stats.MaxIdleClosed,
		"max_idle_time_closed":    stats.MaxIdleTimeClosed,
		"max_lifetime_closed":     stats.MaxLifetimeClosed,
	}
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	
	db.logger.Info("Database connection closed")
	return nil
}

// WithContext returns a new DB instance with the given context
func (db *DB) WithContext(ctx context.Context) *DB {
	return &DB{
		DB:     db.DB.WithContext(ctx),
		config: db.config,
		logger: db.logger.WithContext(ctx),
	}
}

// Transaction executes the given function within a database transaction
func (db *DB) Transaction(ctx context.Context, fn func(*DB) error) error {
	return db.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txDB := &DB{
			DB:     tx,
			config: db.config,
			logger: db.logger.WithContext(ctx),
		}
		return fn(txDB)
	})
}

// Migrate runs database migrations for the given models
func (db *DB) Migrate(models ...interface{}) error {
	db.logger.Info("Running database migrations")
	
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %w", model, err)
		}
	}
	
	db.logger.WithFields(logger.Fields{
		"models_count": len(models),
	}).Info("Database migrations completed")
	
	return nil
}

// Repository provides base repository functionality
type Repository struct {
	DB     *DB
	Logger *logger.Logger
}

// NewRepository creates a new base repository
func NewRepository(db *DB) *Repository {
	return &Repository{
		DB:     db,
		Logger: db.logger,
	}
}

// WithTx returns a new repository instance with the given transaction
func (r *Repository) WithTx(tx *DB) *Repository {
	return &Repository{
		DB:     tx,
		Logger: r.Logger,
	}
}

// LogQuery logs a database query with execution time
func (r *Repository) LogQuery(query string, duration time.Duration, err error) {
	r.Logger.LogDatabaseQuery(query, duration, err)
}

// Helper functions for common database operations

// FindByID finds a record by ID
func FindByID[T any](db *gorm.DB, id interface{}, result *T) error {
	return db.Where("id = ?", id).First(result).Error
}

// FindByField finds a record by a specific field
func FindByField[T any](db *gorm.DB, field string, value interface{}, result *T) error {
	return db.Where(field+" = ?", value).First(result).Error
}

// ExistsByID checks if a record exists by ID
func ExistsByID(db *gorm.DB, model interface{}, id interface{}) (bool, error) {
	var count int64
	err := db.Model(model).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// ExistsByField checks if a record exists by a specific field
func ExistsByField(db *gorm.DB, model interface{}, field string, value interface{}) (bool, error) {
	var count int64
	err := db.Model(model).Where(field+" = ?", value).Count(&count).Error
	return count > 0, err
}

// CountRecords counts records with optional conditions
func CountRecords(db *gorm.DB, model interface{}, conditions ...interface{}) (int64, error) {
	var count int64
	query := db.Model(model)
	
	if len(conditions) > 0 {
		query = query.Where(conditions[0], conditions[1:]...)
	}
	
	err := query.Count(&count).Error
	return count, err
}

// PaginatedFind finds records with pagination
func PaginatedFind[T any](db *gorm.DB, offset, limit int, results *[]T, totalCount *int64) error {
	// Get total count
	if err := db.Model(new(T)).Count(totalCount).Error; err != nil {
		return err
	}
	
	// Get paginated results
	return db.Offset(offset).Limit(limit).Find(results).Error
}

// SoftDelete performs a soft delete on a record
func SoftDelete(db *gorm.DB, model interface{}, id interface{}) error {
	return db.Where("id = ?", id).Delete(model).Error
}

// HardDelete performs a hard delete on a record
func HardDelete(db *gorm.DB, model interface{}, id interface{}) error {
	return db.Unscoped().Where("id = ?", id).Delete(model).Error
}