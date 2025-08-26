package cmd

import (
	"fmt"
	"mault/internal/crypto"
	"mault/internal/storage"
	"mault/internal/storage/base"

	"github.com/urfave/cli/v2"
)

// Init command is the entry point of mault. It initializes the database file for the secrets. 
// It also asks for a master password, which will be used to get data from the database.
// User needs to authenticate with their master password in order to get data from the database.
var initC *cli.Command = &cli.Command{
	Name:  "init",
	Usage: "Initialize the mault",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.PrepareDatabase()
		if err != nil {
			return fmt.Errorf("init command failed: %w", err)
		}

		bm := base.NewManager(db)
		if bm.IsInitialized() {
			return fmt.Errorf("mault has already initialized")
		}

		return initBase(bm)
	},
}

// Function for initialize the base,
// which covers the action of defining and inserting the master password in the database.
func initBase(bm *base.Manager) error {
	master, err := crypto.ReadPassword("master password")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}

	// Reading master password twice is important because user might enter the password incorrectly.
	again, err := crypto.ReadPassword("master password again")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}

	// Checking if two passwords match
	if string(master) != string(again) {
		return fmt.Errorf("passwords don't match")
	}

	// Checking if the master password is strong enough.
	// The password must be at least 8 characters and involve at least one uppercase, one lowercase,
	// one digit and one special character.
	if err := crypto.IsStrong(master); err != nil {
		return fmt.Errorf("password error: %v", err)
	}

	// Encryption process
	salt := crypto.GenerateSalt(16)
	key := crypto.GenerateDerivedKey(master, salt)
	if err := bm.Init(key, salt); err != nil {
		return fmt.Errorf("initializion error: %v", err)
	}

	fmt.Println("mault has initialized successfully!")
	return nil
}
