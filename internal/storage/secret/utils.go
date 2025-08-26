package secret

import (
	"crypto/rand"
	"math/big"
)

var charset = "abcdefghijklmnopqrstuvwxyz" + "ABCDEFGHIJKLMNOPQRSTUVWXYZ" + "0123456789" + "!#%&*-_?"

func GenerateRandom(length int) ([]byte, error) {
	password := make([]byte, length)
	for i := range password {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return nil, err
		}
		password[i] = charset[num.Int64()]
	}
	return password, nil
}
