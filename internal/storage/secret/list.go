package secret

import (
	"fmt"

	"gorm.io/gorm"
)

// Fetches the secrets from the database and returns.
func List(db *gorm.DB) ([]SecretList, error) {
	var secrets []SecretList

	if err := db.Transaction(func(tx *gorm.DB) error {
		return tx.Table("secrets").Find(&secrets).Error
	}); err != nil {
		return nil, fmt.Errorf("could not list the secrets: %v", err)
	}

	return secrets, nil
}
