package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"mault/internal/storage/secret"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

// Generates a random salt
func GenerateSalt(bytes int) []byte {
	salt := make([]byte, bytes)
	rand.Read(salt)

	return salt
}

// Reads password from stdin and returns it with a nil error value if no error occures.
// Can be exited immediately by Ctrl+C combination.
func ReadPassword(passwordType string) ([]byte, error) {
	oldState, err := term.GetState(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	if err != nil {
		return nil, fmt.Errorf("could not get the current terminal state: %v", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		term.Restore(int(os.Stdin.Fd()), oldState) // The error this function may return is ignored for now.
		fmt.Println("\nAborted.")
		os.Exit(1)
	}()

	fmt.Printf("Enter the %v: ", passwordType)

	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	signal.Stop(c)

	if err != nil {
		return nil, fmt.Errorf("could not read the %v: %v", passwordType, err)
	}

	return password, nil
}

// Creates a key using the password and salt provided
func GenerateDerivedKey(masterPassword, salt []byte) []byte {
	return argon2.IDKey(masterPassword, salt, 1, 64*1024, 4, 32)
}

// Creates a nonce. The size of the nonce is equal to the size of the GCM.
func generateNonce(gcm cipher.AEAD) []byte {
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	return nonce
}

// Encrypts the password with the derived key.
func EncryptWithAESGCM(password, key []byte) ([]byte, []byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("could not encrypt the password: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("could not encrypt the password: %v", err)
	}

	nonce := generateNonce(gcm)

	cipherText := gcm.Seal(nil, nonce, password, nil)

	return nonce, cipherText, nil
}

// Decrypts the secret using the salt and master password provided. If the password is incorrect, secret won't be decrypted.
func DecryptWithAESGCM(secret secret.Secret, master, salt []byte) (string, error) {
	block, err := aes.NewCipher(GenerateDerivedKey(master, salt))
	if err != nil {
		return "", fmt.Errorf("couldn't decrypt the secret: %v", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("couldn't decrypt the secret: %v", err)
	}

	plainText, err := gcm.Open(nil, secret.Nonce, secret.CipherText, nil)
	if err != nil {
		return "", fmt.Errorf("wrong password or corrupted data: %v", err.Error())
	}

	return string(plainText), nil
}
