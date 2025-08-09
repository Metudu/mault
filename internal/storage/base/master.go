package base

import (
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"regexp"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	lowerRe   = regexp.MustCompile(`[a-z]`)
	upperRe   = regexp.MustCompile(`[A-Z]`)
	digitRe   = regexp.MustCompile(`\d`)
	specialRe = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func IsStrongPassword(password []byte) error {
	if len(password) < 8 {
		return fmt.Errorf("master password must be longer than 8 characters")
	}

	if !lowerRe.Match(password) {
		return fmt.Errorf("master password must include at least one lowercase character")
	}
	if !upperRe.Match(password) {
		return fmt.Errorf("master password must include at least one uppercase character")
	}
	if !digitRe.Match(password) {
		return fmt.Errorf("master password must include at least one digit")
	}
	if !specialRe.Match(password) {
		return fmt.Errorf("master password must include at least one special character")
	}

	return nil
}

func IsMatch(db *gorm.DB, key []byte) bool {
	var base Base
	err := db.Session(&gorm.Session{
		Logger: db.Logger.LogMode(logger.Silent),
	}).Transaction(func(tx *gorm.DB) error {
		return tx.Table("bases").First(&base).Error
	})

	if err != nil {
		return false
	}

	// Encoding the key based on the user input
	h := sha256.New()
	h.Write(key)
	password := h.Sum(nil)

	return subtle.ConstantTimeCompare(base.Hash[:], password[:]) == 1
}
