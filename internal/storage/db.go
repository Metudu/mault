package storage

import (
	"context"
	"fmt"
	"mault/internal/cerror"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager handles database connections and lifecycle
type DatabaseManager struct {
	db     *gorm.DB
	dbPath string
	mu     sync.RWMutex
	config *DatabaseConfig
}

// DatabaseConfig holds configuration options for the database
type DatabaseConfig struct {
	BaseDir        string
	DatabaseName   string
	MaxConnections int
	LogLevel       logger.LogLevel
	Timeout        time.Duration
}

// DefaultDatabaseConfig returns the default configuration
func DefaultDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		DatabaseName:   "mault.db",
		MaxConnections: 1, // SQLite works best with single connection
		LogLevel:       logger.Silent,
		Timeout:        30 * time.Second,
	}
}

var (
	globalManager *DatabaseManager
	managerOnce   sync.Once
)

// GetDatabaseManager returns a singleton database manager instance
func GetDatabaseManager(config *DatabaseConfig) (*DatabaseManager, error) {
	var initErr error
	
	managerOnce.Do(func() {
		if config == nil {
			config = DefaultDatabaseConfig()
		}
		
		globalManager, initErr = newDatabaseManager(config)
	})
	
	return globalManager, initErr
}

// newDatabaseManager creates a new database manager
func newDatabaseManager(config *DatabaseConfig) (*DatabaseManager, error) {
	manager := &DatabaseManager{
		config: config,
	}
	
	if err := manager.initialize(); err != nil {
		return nil, err
	}
	
	return manager, nil
}

// initialize sets up the database connection and runs migrations
func (dm *DatabaseManager) initialize() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if err := dm.setupDatabasePath(); err != nil {
		return err
	}
	
	if err := dm.connect(); err != nil {
		return err
	}
	
	return dm.migrate()
}

// setupDatabasePath determines and creates the database directory
func (dm *DatabaseManager) setupDatabasePath() error {
	var baseDir string
	var err error
	
	if dm.config.BaseDir != "" {
		baseDir = dm.config.BaseDir
	} else {
		baseDir, err = os.UserConfigDir()
		if err != nil {
			return &cerror.Error{
				Operation: "Setup database path", 
				Cause:     fmt.Sprintf("failed to get user config directory: %v", err),
			}
		}
	}
	
	dm.dbPath = filepath.Join(baseDir, "mault", dm.config.DatabaseName)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(dm.dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return &cerror.Error{
			Operation: "Setup database path",
			Cause:     fmt.Sprintf("failed to create database directory: %v", err),
		}
	}
	
	return nil
}

// connect establishes the database connection
func (dm *DatabaseManager) connect() error {
	config := &gorm.Config{
		Logger: logger.Default.LogMode(dm.config.LogLevel),
	}
	
	// SQLite connection string with optimizations
	dsn := fmt.Sprintf("%s?cache=shared&mode=rwc&_busy_timeout=%d&_journal_mode=WAL&_foreign_keys=on",
		dm.dbPath, int(dm.config.Timeout.Milliseconds()))
	
	db, err := gorm.Open(sqlite.Open(dsn), config)
	if err != nil {
		return &cerror.Error{
			Operation: "Connect to database",
			Cause:     fmt.Sprintf("failed to open database: %v", err),
		}
	}
	
	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return &cerror.Error{
			Operation: "Connect to database",
			Cause:     fmt.Sprintf("failed to get underlying sql.DB: %v", err),
		}
	}
	
	sqlDB.SetMaxOpenConns(dm.config.MaxConnections)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	
	dm.db = db
	return nil
}

// migrate runs database migrations
func (dm *DatabaseManager) migrate() error {
	migrations := []struct {
		name  string
		model interface{}
		table string
	}{
		{"base records", &base.Record{}, ""},
		{"secret records", &secret.Record{}, "secret"},
	}
	
	for _, migration := range migrations {
		if err := dm.runMigration(migration.model, migration.table, migration.name); err != nil {
			return err
		}
	}
	
	return nil
}

// runMigration runs a single migration
func (dm *DatabaseManager) runMigration(model interface{}, tableName, description string) error {
	var migrator gorm.Migrator
	
	if tableName != "" {
		migrator = dm.db.Table(tableName).Migrator()
	} else {
		migrator = dm.db.Migrator()
	}
	
	if !migrator.HasTable(model) {
		if err := migrator.AutoMigrate(model); err != nil {
			return &cerror.Error{
				Operation: "Initialize database",
				Cause:     fmt.Sprintf("failed to migrate %s: %v", description, err),
			}
		}
	}
	
	return nil
}

// GetDB returns the database instance with read lock
func (dm *DatabaseManager) GetDB() *gorm.DB {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.db
}

// GetDBWithContext returns the database instance with context
func (dm *DatabaseManager) GetDBWithContext(ctx context.Context) *gorm.DB {
	return dm.GetDB().WithContext(ctx)
}

// Close closes the database connection
func (dm *DatabaseManager) Close() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	
	if dm.db != nil {
		sqlDB, err := dm.db.DB()
		if err != nil {
			return &cerror.Error{
				Operation: "Close database",
				Cause:     fmt.Sprintf("failed to get underlying sql.DB: %v", err),
			}
		}
		
		if err := sqlDB.Close(); err != nil {
			return &cerror.Error{
				Operation: "Close database",
				Cause:     fmt.Sprintf("failed to close database: %v", err),
			}
		}
		
		dm.db = nil
	}
	
	return nil
}

// HealthCheck performs a basic health check on the database
func (dm *DatabaseManager) HealthCheck(ctx context.Context) error {
	db := dm.GetDBWithContext(ctx)
	
	sqlDB, err := db.DB()
	if err != nil {
		return &cerror.Error{
			Operation: "Database health check",
			Cause:     fmt.Sprintf("failed to get underlying sql.DB: %v", err),
		}
	}
	
	if err := sqlDB.PingContext(ctx); err != nil {
		return &cerror.Error{
			Operation: "Database health check",
			Cause:     fmt.Sprintf("database ping failed: %v", err),
		}
	}
	
	return nil
}

// GetDatabasePath returns the path to the database file
func (dm *DatabaseManager) GetDatabasePath() string {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return dm.dbPath
}

// Deprecated: Use GetDatabaseManager instead
// PrepareDatabase maintains backward compatibility but is deprecated
func PrepareDatabase(dir ...string) (*gorm.DB, error) {
	config := DefaultDatabaseConfig()
	if len(dir) > 0 {
		config.BaseDir = dir[0]
	}
	
	manager, err := GetDatabaseManager(config)
	if err != nil {
		return nil, err
	}
	
	return manager.GetDB(), nil
}