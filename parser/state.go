package parser

import "github.com/okneniz/parsec"

type State struct {
	GroupIndex int
}

var _ parsec.Recoverable[*State] = &State{}

func (s *State) Clone() *State {
	return &State{
		GroupIndex: s.GroupIndex,
	}
}
func (s *State) Recover(clone *State) {
	s.GroupIndex = clone.GroupIndex
}
