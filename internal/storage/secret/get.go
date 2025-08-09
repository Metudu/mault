package secret

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Get(db *gorm.DB, key string, masterPassword []byte) (Secret, error) {
	var secret Secret
	err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Where("key = ?", key).First(&secret).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Secret{}, fmt.Errorf("no secret with key %v found", key)
	} else if err != nil {
		return Secret{}, fmt.Errorf("could not get the secret: %v", err)
	}

	return secret, nil
}
