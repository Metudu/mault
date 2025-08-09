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

// Create a secret
var createC *cli.Command = &cli.Command{
	Name: "create",
	Usage: "Create a new secret",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.AccessDatabase()
		if err != nil {
			return fmt.Errorf("accessing the database error: %v", err)
		}

		if !base.IsInitialized(db) {
			return fmt.Errorf("you haven't initialized the mault yet")
		}

		return CreateSecret(db)
	},
}

// CreateSecret function asks user the key, the secret and the master password respectively.
// While typing the secret and the master password is not visible by default.
// Using the salt and master, application creates a key in order to create nonce and ciphertext.
// Then stores the key, nonce and ciphertext in the database.
func CreateSecret(db *gorm.DB) error {
	var key string
	fmt.Print("Enter the name of the secret: ")
	_, err := fmt.Scanln(&key)
	if err != nil {
		return fmt.Errorf("could not scan the key: %v", err.Error())
	}

	secretPassword, err := crypto.ReadPassword("secret")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
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

	nonce, ciphertext, err := crypto.EncryptWithAESGCM(secretPassword, derivedKey)
	if err != nil {
		return fmt.Errorf("encryption error: %v", err)
	}

	if err := secret.Create(db, secret.Secret{
		Key:        key,
		Nonce:      nonce,
		CipherText: ciphertext,
	}); err != nil {
		return fmt.Errorf("creating secret error: %v", err)
	}

	fmt.Println("Secret has created successfully!")
	return nil
}
