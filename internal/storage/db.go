package storage

import (
	"fmt"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"
	"os"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Accesses the database located under the default config folder of the operating system. Then connects to the database and returns it.
func AccessDatabase() (*gorm.DB, error) {
	config, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("could not access the config directory: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(path.Join(config, "mault", "mault.db")))
	if err != nil {
		return nil, fmt.Errorf("could not open database: %v", err)
	}

	if err := db.AutoMigrate(&base.Base{}, &secret.Secret{}); err != nil {
		return nil, fmt.Errorf("could not add schema to database: %v", err)
	}

	return db, nil
}
