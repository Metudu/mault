package storage

import (
	"context"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm/logger"
)

// TestDatabaseManager_Initialize tests the database initialization process
func TestDatabaseManager_Initialize(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "test_mault.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	manager, err := newDatabaseManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()
	
	// Test database file exists
	expectedPath := filepath.Join(tempDir, "mault", "test_mault.db")
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Fatalf("Database file was not created at expected path: %s", expectedPath)
	}
	
	// Test database connection
	db := manager.GetDB()
	if db == nil {
		t.Fatal("Database connection is nil")
	}
	
	// Test health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := manager.HealthCheck(ctx); err != nil {
		t.Fatalf("Health check failed: %v", err)
	}
}

// TestDatabaseManager_Migrations tests database migrations
func TestDatabaseManager_Migrations(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "test_migrations.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	manager, err := newDatabaseManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()
	
	db := manager.GetDB()
	
	// Test that base.Record table exists
	if !db.Migrator().HasTable(&base.Record{}) {
		t.Error("base.Record table was not created")
	}
	
	// Test that secret.Record table exists with custom table name
	if !db.Table("secret").Migrator().HasTable(&secret.Record{}) {
		t.Error("secret.Record table was not created with custom table name")
	}
	
	// Test table creation by trying to insert a record
	baseRecord := &base.Record{}
	if err := db.Create(baseRecord).Error; err != nil {
		t.Errorf("Failed to create base record: %v", err)
	}
	
	secretRecord := &secret.Record{
		Key: "test_key",
		Nonce: []byte("test_nonce"),
		CipherText: []byte("test_ciphertext"),
	}
	if err := db.Table("secret").Create(secretRecord).Error; err != nil {
		t.Errorf("Failed to create secret record: %v", err)
	}
}

// TestDatabaseManager_Singleton tests singleton behavior
func TestDatabaseManager_Singleton(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "test_singleton.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	// Reset singleton for testing
	globalManager = nil
	managerOnce = sync.Once{}
	
	manager1, err := GetDatabaseManager(config)
	if err != nil {
		t.Fatalf("Failed to get first database manager: %v", err)
	}
	
	manager2, err := GetDatabaseManager(config)
	if err != nil {
		t.Fatalf("Failed to get second database manager: %v", err)
	}
	
	// Should be the same instance
	if manager1 != manager2 {
		t.Error("GetDatabaseManager did not return the same instance")
	}
	
	// Clean up
	manager1.Close()
}

// TestDatabaseManager_Configuration tests different configurations
func TestDatabaseManager_Configuration(t *testing.T) {
	tests := []struct {
		name   string
		config *DatabaseConfig
	}{
		{
			name:   "Default configuration",
			config: DefaultDatabaseConfig(),
		},
		{
			name: "Custom configuration",
			config: &DatabaseConfig{
				BaseDir:        t.TempDir(),
				DatabaseName:   "custom_test.db",
				MaxConnections: 1,
				LogLevel:       logger.Info,
				Timeout:        10 * time.Second,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton for each test
			globalManager = nil
			managerOnce = sync.Once{}
			
			if tt.config.BaseDir == "" {
				tt.config.BaseDir = t.TempDir()
			}
			
			manager, err := GetDatabaseManager(tt.config)
			if err != nil {
				t.Fatalf("Failed to create database manager: %v", err)
			}
			defer manager.Close()
			
			if manager.config.DatabaseName != tt.config.DatabaseName {
				t.Errorf("Expected database name %s, got %s", tt.config.DatabaseName, manager.config.DatabaseName)
			}
			
			if manager.config.MaxConnections != tt.config.MaxConnections {
				t.Errorf("Expected max connections %d, got %d", tt.config.MaxConnections, manager.config.MaxConnections)
			}
		})
	}
}

// TestDatabaseManager_ErrorHandling tests error handling scenarios
func TestDatabaseManager_ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		config      *DatabaseConfig
		expectError bool
	}{
		{
			name: "Invalid directory permissions",
			config: &DatabaseConfig{
				BaseDir:        "/invalid/readonly/path",
				DatabaseName:   "test.db",
				MaxConnections: 1,
				LogLevel:       logger.Silent,
				Timeout:        5 * time.Second,
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset singleton for each test
			globalManager = nil
			managerOnce = sync.Once{}
			
			_, err := GetDatabaseManager(tt.config)
			
			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestDatabaseManager_ConcurrentAccess tests concurrent access to the database
func TestDatabaseManager_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "test_concurrent.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	// Reset singleton for testing
	globalManager = nil
	managerOnce = sync.Once{}
	
	manager, err := GetDatabaseManager(config)
	if err != nil {
		t.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()
	
	// Test concurrent database access
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			
			db := manager.GetDBWithContext(ctx)
			if db == nil {
				t.Errorf("Goroutine %d: Got nil database", id)
				return
			}
			
			// Perform a simple operation
			var count int64
			if err := db.Model(&base.Record{}).Count(&count).Error; err != nil {
				t.Errorf("Goroutine %d: Database operation failed: %v", id, err)
				return
			}
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations to complete")
		}
	}
}

// TestPrepareDatabase_BackwardCompatibility tests backward compatibility
func TestPrepareDatabase_BackwardCompatibility(t *testing.T) {
	tempDir := t.TempDir()
	
	// Reset singleton for testing
	globalManager = nil
	managerOnce = sync.Once{}
	
	db, err := PrepareDatabase(tempDir)
	if err != nil {
		t.Fatalf("PrepareDatabase failed: %v", err)
	}
	
	if db == nil {
		t.Fatal("PrepareDatabase returned nil database")
	}
	
	// Test that migrations ran
	if !db.Migrator().HasTable(&base.Record{}) {
		t.Error("base.Record table was not created")
	}
	
	if !db.Table("secret").Migrator().HasTable(&secret.Record{}) {
		t.Error("secret.Record table was not created")
	}
}

// BenchmarkDatabaseManager_GetDB benchmarks database access
func BenchmarkDatabaseManager_GetDB(b *testing.B) {
	tempDir := b.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "bench_test.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	// Reset singleton for benchmarking
	globalManager = nil
	managerOnce = sync.Once{}
	
	manager, err := GetDatabaseManager(config)
	if err != nil {
		b.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db := manager.GetDB()
			if db == nil {
				b.Fatal("Got nil database")
			}
		}
	})
}

// BenchmarkDatabaseManager_HealthCheck benchmarks health checks
func BenchmarkDatabaseManager_HealthCheck(b *testing.B) {
	tempDir := b.TempDir()
	
	config := &DatabaseConfig{
		BaseDir:        tempDir,
		DatabaseName:   "bench_health.db",
		MaxConnections: 1,
		LogLevel:       logger.Silent,
		Timeout:        5 * time.Second,
	}
	
	// Reset singleton for benchmarking
	globalManager = nil
	managerOnce = sync.Once{}
	
	manager, err := GetDatabaseManager(config)
	if err != nil {
		b.Fatalf("Failed to create database manager: %v", err)
	}
	defer manager.Close()
	
	ctx := context.Background()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		if err := manager.HealthCheck(ctx); err != nil {
			b.Fatalf("Health check failed: %v", err)
		}
	}
}