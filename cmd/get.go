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

// getC is the command for decrypting the secret.
// First, user enters the key and master password. Then a database query runs and fetches the nonce and ciphertext.
// The using these values, decrpytion process starts and if everything is in order, the secret is being printed to the console.
var getC *cli.Command = &cli.Command{
	Name:  "get",
	Usage: "Reveal a secret",
	Args:  false,
	Action: func(ctx *cli.Context) error {
		db, err := storage.PrepareDatabase()
		if err != nil {
			return fmt.Errorf("get command failed: %w", err)
		}

		bm := base.NewManager(db)
		if !bm.IsInitialized() {
			return fmt.Errorf("you haven't initialized the mault yet")
		}

		sm := secret.NewManager(db)
		return getSecret(ctx.Context, bm, sm)
	},
}

func getSecret(ctx context.Context, bm *base.Manager, sm *secret.Manager) error {
	key, err := prompt.GetKey(os.Stdin)
	if err != nil {
		return err
	}

	result, err := bm.Authenticate(ctx)
	if err != nil {
		return err
	}

	secret, err := sm.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("getting secret error: %v", err)
	}

	plainText, err := crypto.DecryptWithAESGCM(*secret, result.Master, result.Salt)
	if err != nil {
		return fmt.Errorf("decryption error: %v", err)
	}

	fmt.Println(plainText)
	return nil
}
