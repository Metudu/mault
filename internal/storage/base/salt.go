package base

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Fetches the salt which is used while creating a derived key from the master password.
// This salt can be used in order to confirm the master password.
func GetSalt(db *gorm.DB) ([]byte, error) {
	var base Base
	if err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Table("bases").Select("salt").First(&base).Error
	}); err != nil {
		return nil, fmt.Errorf("could not get the salt: %v", err.Error())
	}

	return base.Salt, nil
}
