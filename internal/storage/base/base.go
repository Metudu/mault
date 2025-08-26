package base

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"mault/internal/cerror"
	"mault/internal/crypto"

	"gorm.io/gorm"
)

// The testable base
type Manager struct {
	repository Ops
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{repository: &DBRepository{db: db}}
}

func NewTestManager(repo Ops) *Manager {
	return &Manager{repository: repo}
}

type Ops interface {
	Create(base *Record) error
	GetFirst() (*Record, error)
}

// record
type Record struct {
	gorm.Model
	Hash []byte `db:"hash"`
	Salt []byte `db:"salt"`
}

type AuthResult struct {
	Master     []byte
	Salt       []byte
	DerivedKey []byte
}

// Initializes the mault by appending the hashed derived key and random generated salt to the database.
func (b *Manager) Init(key, salt []byte) error {
	hash := sha256.Sum256(key) // Simpler and correct

	record := &Record{
		Hash: hash[:],
		Salt: salt,
	}

	if err := b.repository.Create(record); err != nil {
		return &cerror.Error{Operation: "Initialize base", Cause: err.Error()}
	}
	return nil
}

func (b *Manager) IsInitialized() bool {
	_, err := b.repository.GetFirst()
	return err == nil
}

func (b *Manager) Match(key []byte) bool {
	record, err := b.repository.GetFirst()
	if err != nil {
		return false
	}

	hash := sha256.Sum256(key)
	return subtle.ConstantTimeCompare(record.Hash, hash[:]) == 1
}

func (b *Manager) Get() ([]byte, error) {
	rec, err := b.repository.GetFirst()
	if err != nil {
		return nil, &cerror.Error{Operation: "Get master"}
	}

	return rec.Salt, nil
}

func (bm *Manager) Authenticate(ctx context.Context) (*AuthResult, error) {
	if ctx.Err() != nil {
		return nil, &cerror.Error{Operation: "Authenticate", Cause: ctx.Err().Error()}
	}

	master, err := crypto.ReadPassword("master password")
	if err != nil {
		return nil, err
	}

	if ctx.Err() != nil {
		return nil, &cerror.Error{Operation: "Authenticate", Cause: ctx.Err().Error()}
	}

	salt, err := bm.Get()
	if err != nil {
		return nil, err
	}

	derivedKey := crypto.GenerateDerivedKey(master, salt)
	if !bm.Match(derivedKey) {
		return nil, &cerror.Error{Operation: "Authenticate", Cause: "master password does not match"}
	}

	return &AuthResult{Master: master, Salt: salt, DerivedKey: derivedKey}, nil
}
