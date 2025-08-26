package crypto

import (
	"testing"
)

func Test_GenerateSalt(t *testing.T) {
	var bytes = 16
	salt := GenerateSalt(bytes)

	if len(salt) != 16 {
		t.Errorf("Expected 16 bytes, but got %d", len(salt))
	}
}

func Test_read(t *testing.T) {
}
