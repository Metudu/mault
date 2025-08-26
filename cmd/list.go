package cmd

import (
	"context"
	"fmt"
	"mault/internal/storage"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"

	"github.com/urfave/cli/v2"
)

var listC *cli.Command = &cli.Command{
	Name:    "list",
	Usage:   "List the secrets",
	Aliases: []string{"l", "ls"},
	Args:    false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.PrepareDatabase()
		if err != nil {
			return fmt.Errorf("list command failed: %w", err)
		}

		sm := secret.NewManager(db)
		if err := listSecrets(ctx.Context, sm); err != nil {
			return fmt.Errorf("listing secrets error: %w", err)
		}

		base := base.NewManager(db)
		if !base.IsInitialized() {
			fmt.Println("WARNING: You haven't initialized the mault yet!")
		}
		return nil
	},
}

func listSecrets(ctx context.Context, sm *secret.Manager) error {
	secrets, err := sm.ListSecrets(ctx)
	if err != nil {
		return fmt.Errorf("listing secrets error: %w", err)
	}
	if len(secrets) == 0 {
		fmt.Println("No secrets yet.")
		return nil
	}

	for _, secret := range secrets {
		fmt.Printf("%v\t\t********\n", secret)
	}
	return nil
}
