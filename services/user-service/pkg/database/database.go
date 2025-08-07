package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"user-service/internal/config"
	appLogger "user-service/pkg/logger"
)

type Database struct {
	*gorm.DB
	config *config.Config
	logger *appLogger.Logger
}

func New(cfg *config.Config, appLogger *appLogger.Logger) (*Database, error) {
	dsn := cfg.GetDatabaseDSN()
	
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

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

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

func (d *Database) Migrate(models ...interface{}) error {
	d.logger.Info("Running database migrations...")
	
	if err := d.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	
	d.logger.Info("Database migrations completed successfully")
	return nil
}