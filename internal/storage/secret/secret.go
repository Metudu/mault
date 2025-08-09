package secret

import "gorm.io/gorm"

// Holds the information about the secrets.
// These fields can be used in order to decrypt and reveal the encrypted secret.
type Secret struct {
	gorm.Model
	Key        string `gorm:"unique;not null"`
	Nonce      []byte `gorm:"unique;not null"`
	CipherText []byte `gorm:"unique;not null"`
}

// Basic struct in order to make things easier when listing the secrets.
type SecretList struct {
	Key string
}

// Basic struct in order to make things easier when getting a specific secret.
type SecretGet struct {
	Key      string
	Password string
}
