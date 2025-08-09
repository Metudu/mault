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

// getC is the command for decrypting the secret.
// First, user enters the key and master password. Then a database query runs and fetches the nonce and ciphertext.
// The using these values, decrpytion process starts and if everything is in order, the secret is being printed to the console.
var getC *cli.Command = &cli.Command{
	Name:  "get",
	Usage: "Reveal a secret",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.AccessDatabase()
		if err != nil {
			return fmt.Errorf("accessing the database error: %v", err)
		}

		if !base.IsInitialized(db) {
			return fmt.Errorf("you haven't initialized the mault yet")
		}

		return GetSecret(db)
	},
}

func GetSecret(db *gorm.DB) error {
	var key string
	fmt.Print("Enter the key you want to get: ")
	fmt.Scanln(&key)

	master, err := crypto.ReadPassword("master")
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

	secret, err := secret.Get(db, key, master)
	if err != nil {
		return fmt.Errorf("getting secret error: %v", err)
	}

	plainText, err := crypto.DecryptWithAESGCM(secret, master, salt)
	if err != nil {
		return fmt.Errorf("decryption error: %v", err)
	}

	fmt.Println(plainText)
	return nil
}
