package secret

import (
	"context"
	"errors"
	"fmt"
	"mault/internal/cerror"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBRepository struct {
	db *gorm.DB
}

func NewDBRepository(db *gorm.DB) *DBRepository {
	return &DBRepository{db: db}
}

func (d *DBRepository) Create(ctx context.Context, record *Record) error {
	if err := d.db.WithContext(ctx).Session(&gorm.Session{
		Logger: d.db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Create(&record).Error
	}); err != nil {
		return &cerror.Error{Operation: "Create secret", Cause: err.Error()}
	}
	return nil
}

func (d *DBRepository) List(ctx context.Context) ([]string, error) {
	var keys []string

	if err := d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.Table("secrets").Select("key").Pluck("key", &keys).Error
	}); err != nil {
		return nil, &cerror.Error{Operation: "List secrets", Cause: err.Error()}
	}

	return keys, nil
}

func (d *DBRepository) Get(ctx context.Context, key string) (*Record, error) {
	var record Record
	err := d.db.WithContext(ctx).Session(&gorm.Session{
		Logger: d.db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Where("key = ?", key).First(&record).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, &cerror.Error{Operation: "Get secret", Cause: fmt.Sprintf("no secret with key %s found", key)}
	} else if err != nil {
		return nil, &cerror.Error{Operation: "Get secret", Cause: err.Error()}
	}

	return &record, nil
}

func (d *DBRepository) Update(ctx context.Context, key string) error {
	return nil
}

func (d *DBRepository) Delete(ctx context.Context, key string) error {
	result := d.db.WithContext(ctx).Where("key = ?", key).Delete(&Record{})

	if result.Error != nil {
		return &cerror.Error{Operation: "Delete secret", Cause: result.Error.Error()}
	}

	if result.RowsAffected == 0 {
		return &cerror.Error{Operation: "Delete secret", Cause: fmt.Sprintf("no secret found with the key %s", key)}
	}

	return nil
}
