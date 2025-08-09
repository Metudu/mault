package cmd

import (
	"fmt"
	"mault/internal/crypto"
	"mault/internal/storage"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

var deleteC *cli.Command = &cli.Command{
	Name:    "delete",
	Usage:   "delete a secret",
	Aliases: []string{"del"},
	Args:    false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.AccessDatabase()
		if err != nil {
			return fmt.Errorf("accessing the database error: %v", err)
		}

		if !base.IsInitialized(db) {
			return fmt.Errorf("you haven't initialized the mault yet")
		}

		return DeleteSecret(db)
	},
}

func DeleteSecret(db *gorm.DB) error {
	var key string
	fmt.Print("Enter the name of the secret: ")
	_, err := fmt.Scanln(&key)
	if err != nil {
		return fmt.Errorf("could not scan the key: %v", err.Error())
	}

	master, err := crypto.ReadPassword("master password")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}

	salt, err := base.GetSalt(db)
	if err != nil {
		return fmt.Errorf("getting salt error: %v", err)
	}

	derivedKey := crypto.GenerateDerivedKey(master, salt)

	if !base.IsMatch(db, derivedKey) {
		return fmt.Errorf("authentication failed")
	}

	if err := secret.Delete(db, key); err != nil {
		return fmt.Errorf("deleting secret error: %v", err)
	}

	fmt.Println("Secret has deleted successfully!")
	return nil
}
