package cmd

import (
	"fmt"
	"mault/internal/crypto"
	"mault/internal/storage"
	"mault/internal/storage/base"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

var initC *cli.Command = &cli.Command{
	Name:  "init",
	Usage: "Initialize the mault",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.AccessDatabase()
		if err != nil {
			return fmt.Errorf("accessing the database error: %v", err)
		}

		if base.IsInitialized(db) {
			return fmt.Errorf("mault has already initialized")
		}

		return InitBase(db)
	},
}

func InitBase(db *gorm.DB) error {
	master, err := crypto.ReadPassword("master password")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}
	again, err := crypto.ReadPassword("master password again")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}

	if string(master) != string(again) {
		return fmt.Errorf("passwords don't match")
	}

	if err := base.IsStrongPassword(master); err != nil {
		return fmt.Errorf("password error: %v", err)
	}

	salt := crypto.GenerateSalt(16)
	key := crypto.GenerateDerivedKey(master, salt)
	if err := base.Init(db, key, salt); err != nil {
		return fmt.Errorf("initializion error: %v", err)
	}

	fmt.Println("mault has initialized successfully!")
	return nil
}
