package secret

import (
	"context"
	"mault/internal/cerror"
)

type MockRepository struct {
	Data []Record
}

func (m *MockRepository) Create(ctx context.Context, rec *Record) error {
	for _, record := range m.Data {
		if rec.Key == record.Key {
			return &cerror.Error{Operation: "Create secret", Cause: "duplicate key"}
		}
	}

	m.Data = append(m.Data, *rec)
	return nil
}

func (m *MockRepository) List(ctx context.Context) ([]string, error) {
	if len(m.Data) == 0 {
		return nil, &cerror.Error{Operation: "List secrets", Cause: "no secret found"}
	}

	keys := []string{}
	for _, rec := range m.Data {
		keys = append(keys, rec.Key)
	}

	return keys, nil
}

func (m *MockRepository) Get(ctx context.Context, key string) (*Record, error) {
	if len(m.Data) == 0 {
		return nil, &cerror.Error{Operation: "Get secret", Cause: "no secret found"}
	}
	for _, rec := range m.Data {
		if rec.Key == key {
			return &rec, nil
		}
	}

	return nil, &cerror.Error{Operation: "Get secret", Cause: "key not found"}
}

func (m *MockRepository) Update(ctx context.Context, key string) error {
	return nil
}

func (m *MockRepository) Delete(ctx context.Context, key string) error {
	if len(m.Data) == 0 {
		return &cerror.Error{Operation: "Delete secret", Cause: "no secret found"}
	}

	for i, rec := range m.Data {
		if rec.Key == key {
			m.Data = append(m.Data[:i], m.Data[i+1:]...)
			return nil
		}
	}

	return &cerror.Error{Operation: "Delete secret", Cause: "key not found"}
}
