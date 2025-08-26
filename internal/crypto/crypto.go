package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"mault/internal/cerror"
	"mault/internal/storage/secret"
	"os"
	"regexp"
	"syscall"

	"golang.org/x/crypto/argon2"
)

// Generates a random salt
func GenerateSalt(bytes int) []byte {
	salt := make([]byte, bytes)
	rand.Read(salt)

	return salt
}

func ReadPassword(passwordType string) ([]byte, error) {
	return read(passwordType, &Terminal{}, &System{}, &Signal{})
}

// Reads password from stdin and returns it with a nil error value if no error occures.
// Can be exited immediately by Ctrl+C combination.
func read(passwordType string, t TerminalOps, s SystemOps, sig SignalOps) ([]byte, error) {
	oldState, err := t.GetState(s.Fd())
	defer t.Restore(s.Fd(), oldState)
	if err != nil {
		return nil, &cerror.Error{Operation: "Get terminal state", Cause: err.Error()}
	}

	c := make(chan os.Signal, 1)
	sig.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		t.Restore(s.Fd(), oldState) // The error this function may return is ignored for now.
		fmt.Println("\nAborted.")
		s.Exit(1)
	}()

	fmt.Printf("Enter the %v: ", passwordType)

	password, err := t.ReadPassword(s.Fd())
	fmt.Println()
	sig.Stop(c)

	if err != nil {
		return nil, &cerror.Error{Operation: "Read password", Cause: err.Error()}
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
		return nil, nil, &cerror.Error{Operation: "Encrypt", Cause: err.Error()}
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, &cerror.Error{Operation: "Encrypt", Cause: err.Error()}
	}

	nonce := generateNonce(gcm)

	cipherText := gcm.Seal(nil, nonce, password, nil)

	return nonce, cipherText, nil
}

// Decrypts the secret using the salt and master password provided. If the password is incorrect, secret won't be decrypted.
func DecryptWithAESGCM(secret secret.Record, master, salt []byte) (string, error) {
	block, err := aes.NewCipher(GenerateDerivedKey(master, salt))
	if err != nil {
		return "", &cerror.Error{Operation: "Decrypt", Cause: err.Error()}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &cerror.Error{Operation: "Decrypt", Cause: err.Error()}
	}

	plainText, err := gcm.Open(nil, secret.Nonce, secret.CipherText, nil)
	if err != nil {
		return "", &cerror.Error{Operation: "Decrypt", Cause: err.Error()}
	}

	return string(plainText), nil
}


var (
	lowerRe   = regexp.MustCompile(`[a-z]`)
	upperRe   = regexp.MustCompile(`[A-Z]`)
	digitRe   = regexp.MustCompile(`\d`)
	specialRe = regexp.MustCompile(`[^A-Za-z0-9]`)
)

func IsStrong(pwd []byte) error {
	if len(pwd) < 8 {
		return fmt.Errorf("master password must be longer than 8 characters")
	}

	if !lowerRe.Match(pwd) {
		return fmt.Errorf("master password must include at least one lowercase character")
	}
	if !upperRe.Match(pwd) {
		return fmt.Errorf("master password must include at least one uppercase character")
	}
	if !digitRe.Match(pwd) {
		return fmt.Errorf("master password must include at least one digit")
	}
	if !specialRe.Match(pwd) {
		return fmt.Errorf("master password must include at least one special character")
	}

	return nil
}