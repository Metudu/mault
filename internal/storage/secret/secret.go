package secret

import (
	"context"

	"gorm.io/gorm"
)

type Manager struct {
	repository Ops
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{repository: &DBRepository{db}}
}

func NewTestManager(repository Ops) *Manager {
	return &Manager{repository: repository}
}

type Ops interface {
	Create(context.Context, *Record) error
	List(context.Context) ([]string, error)
	Get(context.Context, string) (*Record, error)
	Update(context.Context, string) error
	Delete(context.Context, string) error
}

// Holds the information about the secrets.
// These fields can be used in order to decrypt and reveal the encrypted secret.
type Record struct {
	gorm.Model
	Key        string `gorm:"unique;not null"`
	Nonce      []byte `gorm:"not null"`
	CipherText []byte `gorm:"not null"`
}

func (sm *Manager) CreateSecret(ctx context.Context, r *Record) error {
	return sm.repository.Create(ctx, r)
}

func (sm *Manager) ListSecrets(ctx context.Context) ([]string, error) {
	return sm.repository.List(ctx)
}

func (sm *Manager) GetSecret(ctx context.Context, key string) (*Record, error) {
	return sm.repository.Get(ctx, key)
}

func (sm *Manager) UpdateSecret(ctx context.Context, key string) error {
	return sm.repository.Update(ctx, key)
}

func (sm *Manager) DeleteSecret(ctx context.Context, key string) error {
	return sm.repository.Delete(ctx, key)
}