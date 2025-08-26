package crypto

import (
	"os"
	"os/signal"
)

type Signal struct {}
type SignalOps interface {
	Notify(c chan<- os.Signal, sig ...os.Signal)
	Stop(c chan<- os.Signal)
}

func (s *Signal) Notify(c chan<- os.Signal, sig ...os.Signal) {
	signal.Notify(c, sig...)
}

func (s *Signal) Stop(c chan<- os.Signal) {
	signal.Stop(c)
}