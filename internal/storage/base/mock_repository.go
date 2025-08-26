package base

import "errors"

// Base for testing purposes. Acts as a mault base
type MockRepository struct {
	Data  *Record
	Error error
}

func (m *MockRepository) Create(record *Record) error {
	m.Data = record
	m.Error = nil

	return nil
}

func (m *MockRepository) GetFirst() (*Record, error) {
	if m.Data == nil {
		return nil, errors.New("empty record")
	}
	return m.Data, nil
}