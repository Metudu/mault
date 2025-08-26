package crypto

import "os"

type MockSignal struct {
	Notified bool
	Stopped  bool
	Signals  []os.Signal
}

func (m *MockSignal) Notify(c chan<- os.Signal, sig ...os.Signal) {
	m.Notified = true
	m.Signals = append(m.Signals, sig...)
}

func (m *MockSignal) Stop(c chan<- os.Signal) {
	m.Stopped = true
}