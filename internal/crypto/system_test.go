package crypto

type MockSystem struct {
	fd int
	exit int
}

func (s *MockSystem) Fd() int { return s.fd }
func (s *MockSystem) Exit(code int) { s.exit = code }