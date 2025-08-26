package cmd

import (
	"context"
	"fmt"
	"mault/internal/crypto"
	"mault/internal/prompt"
	"mault/internal/storage"
	"mault/internal/storage/base"
	"mault/internal/storage/secret"
	"os"

	"github.com/urfave/cli/v2"
)

// Create a secret
var createC *cli.Command = &cli.Command{
	Name:  "create",
	Usage: "Create a new secret",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		// Accessing the database
		manager, err := storage.GetDatabaseManager(nil)
		if err != nil {
			return fmt.Errorf("create command failed: %w", err)
		}

		// Perform health check to ensure database is accessible
		if err := manager.HealthCheck(ctx.Context); err != nil {
			return fmt.Errorf("create command failed database health check: %w", err)
		}

		// Get database connection with context
		db := manager.GetDBWithContext(ctx.Context)

		// Accessing the base, which holds the master password data
		bm := base.NewManager(db)
		if !bm.IsInitialized() {
			return fmt.Errorf("you haven't initialized the base yet")
		}

		// Accessing the secrets
		sm := secret.NewManager(db)
		return createSecret(ctx.Context, bm, sm)
	},
}

// CreateSecret function asks user the key, the secret and the master password respectively.
// While typing the secret and the master password is not visible by default.
// Using the salt and master, application creates a key in order to create nonce and ciphertext.
// Then stores the key, nonce and ciphertext in the database.
func createSecret(ctx context.Context, bm *base.Manager, sm *secret.Manager) error {
	key, err := prompt.GetKey(os.Stdin)
	if err != nil {
		return err
	}

	secretPassword, err := crypto.ReadPassword("secret")
	if err != nil {
		return fmt.Errorf("reading password error: %v", err)
	}

	result, err := bm.Authenticate(ctx)
	if err != nil {
		return err
	}

	nonce, ciphertext, err := crypto.EncryptWithAESGCM(secretPassword, result.DerivedKey)
	if err != nil {
		return fmt.Errorf("encryption error: %v", err)
	}

	if err := sm.CreateSecret(ctx, &secret.Record{
		Key:        key,
		Nonce:      nonce,
		CipherText: ciphertext,
	}); err != nil {
		return fmt.Errorf("creating secret error: %v", err)
	}

	fmt.Println("Secret has created successfully!")
	return nil
}
