package base

import (
	"crypto/sha256"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initializes the mault by appending the hashed derived key and random generated salt to the database.
func Init(db *gorm.DB, key, salt []byte) error {
	var base Base
	hash := sha256.New()
	_, err := hash.Write(key)

	base.Hash = hash.Sum(nil)
	base.Salt = salt
	if err != nil {
		return fmt.Errorf("could not initialize mault: %v", err)
	}

	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&base).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		return fmt.Errorf("could not initialize mault: %v", err.Error())
	}

	return nil
}

// Checks if the mault is initialized before. If it is, then returns true. If not, returns false.
func IsInitialized(db *gorm.DB) bool {
	var base Base

	err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Table("bases").Select("salt").First(&base).Error
	})

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false
	}

	if string(base.Salt) != "" {
		return true
	}

	return false
}
