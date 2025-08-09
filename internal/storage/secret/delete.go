package secret

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func Delete(db *gorm.DB, key string) error {
	var secret Secret
	if err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{Columns: []clause.Column{{Name: "key"}}}).Table("secrets").Where("key = ?", key).Delete(&secret).Error; err != nil {
			return err
		}
		if secret.Key == "" {
			return fmt.Errorf("no secret found with the key %v", key)
		}
		return nil
	}); err != nil {
		return fmt.Errorf("could not delete the secret: %v", err.Error())
	}
	return nil
}
