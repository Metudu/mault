package crypto

import "os"

type System struct {}
type SystemOps interface {
	Exit(int)
	Fd() int
}

func (s *System) Fd() int {
	return int(os.Stdin.Fd())
}

func (s *System) Exit(code int) {
	os.Exit(code)
}