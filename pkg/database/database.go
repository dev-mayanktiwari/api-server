package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"github.com/dev-mayanktiwari/api-server/internal/config"
	appLogger "github.com/dev-mayanktiwari/api-server/pkg/logger"
)

// Database wraps gorm.DB with additional functionality
type Database struct {
	*gorm.DB
	config *config.Config
	logger *appLogger.Logger
}

// New creates a new database connection
func New(cfg *config.Config, appLogger *appLogger.Logger) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()
	
	// Configure GORM logger
	var gormLogLevel logger.LogLevel
	switch cfg.Logger.Level {
	case "error":
		gormLogLevel = logger.Error
	case "warn":
		gormLogLevel = logger.Warn
	case "info":
		gormLogLevel = logger.Info
	default:
		gormLogLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(gormLogLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB for connection pooling
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("Successfully connected to database")

	return &Database{
		DB:     db,
		config: cfg,
		logger: appLogger,
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	
	d.logger.Info("Database connection closed")
	return nil
}

// Migrate runs database migrations
func (d *Database) Migrate(models ...interface{}) error {
	d.logger.Info("Running database migrations...")
	
	if err := d.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	d.logger.Info("Database migrations completed successfully")
	return nil
}

// HealthCheck checks if the database is healthy
func (d *Database) HealthCheck() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	
	return nil
}

// GetStats returns database connection statistics
func (d *Database) GetStats() map[string]interface{} {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return map[string]interface{}{
			"error": "failed to get underlying sql.DB",
		}
	}
	
	stats := sqlDB.Stats()
	return map[string]interface{}{
		"max_open_connections": stats.MaxOpenConnections,
		"open_connections":     stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_idle_time_closed": stats.MaxIdleTimeClosed,
		"max_lifetime_closed":  stats.MaxLifetimeClosed,
	}
}

// Transaction executes a function within a database transaction
func (d *Database) Transaction(fn func(*gorm.DB) error) error {
	return d.DB.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithContext returns a new database instance with context
func (d *Database) WithContext(ctx interface{}) *Database {
	return &Database{
		DB:     d.DB.WithContext(ctx.(interface{ Deadline() (time.Time, bool); Done() <-chan struct{}; Err() error; Value(interface{}) interface{} })),
		config: d.config,
		logger: d.logger,
	}
}