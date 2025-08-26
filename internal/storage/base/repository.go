package base

import (
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database for production.
type DBRepository struct {
	db *gorm.DB
}

func (d *DBRepository) Create(record *Record) error {
	return d.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(record).Error
	})
}

func (d *DBRepository) GetFirst() (*Record, error) {
	var record Record

	err := d.db.Session(&gorm.Session{
		Logger: d.db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.First(&record).Error
	})

	if err != nil {
		return nil, err
	}

	return &record, nil
}