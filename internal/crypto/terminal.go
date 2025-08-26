package crypto

import "golang.org/x/term"

type Terminal struct {}
type TerminalOps interface {
	GetState(fd int) (*term.State, error)
	ReadPassword(fd int) ([]byte, error)
	Restore(fd int, state *term.State) error
}

func (t *Terminal) GetState(fd int) (*term.State, error) {
	return term.GetState(fd)
}

func (t *Terminal) ReadPassword(fd int) ([]byte, error) {
	return term.ReadPassword(fd)
}

func (t *Terminal) Restore(fd int, state *term.State) error {
	return term.Restore(fd, state)
}