package crypto

import "golang.org/x/term"

type MockTerminal struct {
	Password []byte
	GetStateError error
	ReadPasswordError error
	RestoreError error
}

func (t *MockTerminal) GetState(fd int) (*term.State, error) {
	if t.GetStateError != nil {
		return nil, t.GetStateError
	}
	return &term.State{}, nil 
}

func (t *MockTerminal) ReadPassword(fd int) ([]byte, error) { return t.Password, t.ReadPasswordError }
func (t *MockTerminal) Restore(fd int, state *term.State) error { return t.RestoreError }