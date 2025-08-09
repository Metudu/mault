package cmd

import (
	"fmt"
	"mault/internal/storage"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"

	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
)

var listC *cli.Command = &cli.Command{
	Name:    "list",
	Usage:   "List the secrets",
	Aliases: []string{"l", "ls"},
	Args:    false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.AccessDatabase()
		if err != nil {
			return fmt.Errorf("accessing the database error: %v", err)
		}
		ListSecrets(db)

		if !base.IsInitialized(db) {
			fmt.Println("WARNING: You haven't initialized the mault yet!")
		}
		return nil
	},
}

func ListSecrets(db *gorm.DB) error {
	secrets, err := secret.List(db)
	if err != nil {
		return fmt.Errorf("listing secrets error: %v", err)
	}
	if len(secrets) == 0 {
		fmt.Println("No secrets yet.")
		return nil
	}

	for _, secret := range secrets {
		fmt.Printf("%v\t\t********\n", secret.Key)
	}
	return nil
}
