package base

import "gorm.io/gorm"

// Base stores the master password under the hood.
type Base struct {
	gorm.Model
	Hash []byte `db:"hash"`
	Salt []byte `db:"salt"`
}
