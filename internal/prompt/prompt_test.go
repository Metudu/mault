package prompt

import (
	"mault/internal/cerror"
	"os"
	"testing"
)

func Test_GetKey(t *testing.T) {
	tests := []struct{
		key string
		expectedError error
	}{
		{"", cerror.ErrScanKey},
		{"\n", cerror.ErrEmptyKey},
		{"test_key\n", nil},
	}

	for _, test := range tests {
		r, w, err := os.Pipe()
		if err != nil {
			t.Errorf("could not create a pipe for testing prompt")
		}
		w.Write([]byte(test.key))
		w.Close()

		_, err = GetKey(r)
		if err != test.expectedError {
			t.Errorf("expected error: %v, but got: %v", test.expectedError, err)
		}

		r.Close()
	}
}