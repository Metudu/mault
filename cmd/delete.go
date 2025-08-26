package cmd

import (
	"context"
	"fmt"
	"mault/internal/prompt"
	"mault/internal/storage"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"
	"os"

	"github.com/urfave/cli/v2"
)

var deleteC *cli.Command = &cli.Command{
	Name:    "delete",
	Usage:   "delete a secret",
	Aliases: []string{"del"},
	Args:    false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.PrepareDatabase()
		if err != nil {
			return fmt.Errorf("delete command failed: %w", err)
		}

		bm := base.NewManager(db)
		if !bm.IsInitialized() {
			return fmt.Errorf("you haven't initialized the mault yet")
		}

		sm := secret.NewManager(db)
		return deleteSecret(ctx.Context, bm, sm)
	},
}

func deleteSecret(ctx context.Context, bm *base.Manager, sm *secret.Manager) error {
	key, err := prompt.GetKey(os.Stdin)
	if err != nil {
		return err
	}

	_, err = bm.Authenticate(ctx)
	if err != nil {
		return err
	}

	if err := sm.DeleteSecret(ctx, key); err != nil {
		return fmt.Errorf("deleting secret error: %v", err)
	}

	fmt.Println("Secret has deleted successfully!")
	return nil
}
