package base

import (
	"crypto/sha256"
	"testing"
)

var (
	mockKey  []byte = []byte("supersecretkey1234!")
	mockSalt []byte = []byte("supersecretsalt123")
)

func Test_Manager_Init(t *testing.T) {
	m := NewTestManager(&MockRepository{})

	if err := m.Init([]byte(mockKey), []byte(mockSalt)); err != nil {
		t.Errorf("base initialization has failed: %v", err)
	}

	rec, err := m.repository.GetFirst()
	if err != nil {
		t.Errorf("fetching data from base has failed: %v", err)
	}

	if string(mockSalt) != string(rec.Salt) && sha256.Sum256([]byte(mockKey)) != [32]byte(rec.Hash) {
		t.Errorf("mismatched key and salt: expected salt: %s, but got: %s | expected hashed key: %s, but got %s", mockSalt, rec.Salt, sha256.Sum256([]byte(mockKey)), rec.Hash)
	}

	if string(mockSalt) != string(rec.Salt) {
		t.Errorf("mismatched salt: expected salt: %s, but got: %s", mockSalt, rec.Salt)
	}

	if sha256.Sum256([]byte(mockKey)) != [32]byte(rec.Hash) {
		t.Errorf("mismatched key: expected hashed key: %s, but got: %s", sha256.Sum256([]byte(mockKey)), rec.Hash)
	}
}

func Test_Manager_IsInitialized(t *testing.T) {
	m := NewTestManager(&MockRepository{})
	if m.IsInitialized() {
		t.Errorf("base is not initialized yet but IsInitialized function returns true")
	}

	hash := sha256.Sum256(mockKey)
	m.repository.Create(&Record{
		Hash: hash[:],
		Salt: mockSalt,
	})

	if !m.IsInitialized() {
		t.Errorf("base is initialized but IsInitialized function returns false")
	}
}

func Test_Manager_Match(t *testing.T) {
	hash := sha256.Sum256(mockKey)
	m := NewTestManager(&MockRepository{
		Data: &Record{
			Hash: hash[:],
			Salt: mockSalt,
		},
	})

	if !m.Match(mockKey) {
		t.Errorf("expected true for key match, but got false")
	}

	if m.Match([]byte("wrongkey")) {
		t.Errorf("expected false for key match, but got true")
	}
}

func Test_Manager_Get(t *testing.T) {
	hash := sha256.Sum256(mockKey)
	m := NewTestManager(&MockRepository{
		Data: &Record{
			Hash: hash[:],
			Salt: mockSalt,
		},
	})

	testSalt, err := m.Get()
	if err != nil {
		t.Errorf("expected to get the record, but got an error : %v", err)
	}
	if string(testSalt) != string(mockSalt) {
		t.Errorf("expected same salt, but got two different salts -> %s | %s", testSalt, mockSalt)
	}
}