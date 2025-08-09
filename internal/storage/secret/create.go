package secret

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Create(db *gorm.DB, e Secret) error {
	if err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&e).Error
	}); err != nil {
		return fmt.Errorf("could not create the secret: %v", err.Error())
	}
	return nil
}
