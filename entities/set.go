package entities

import (
	"fmt"
	"strings"

	"github.com/yourbasic/bit"
)

type Set struct {
	Mapa   *bit.Set
	strRep string
	N      int
	Elems  int
	Camino string
}

func (e *Set) Copy() Set {
	copiedMap := new(bit.Set)

	e.Mapa.Visit(func(n int) (skip bool) {
		copiedMap.Add(n)

		return false
	})

	return Set{
		Mapa:   copiedMap,
		N:      e.N,
		Elems:  e.Elems,
		Camino: e.Camino,
	}
}

func (s *Set) GetStrRep() string {
	if s.strRep == "" {
		s.GenerateStrRep()
	}

	return s.strRep
}
func (s *Set) GetLargeStrRep() string {
	var sb strings.Builder
	var i int
	for i = 0; i < s.N; i++ {
		fmt.Fprintf(&sb, "%t_", s.Mapa.Contains(i))
	}
	return sb.String()
}

func (s *Set) GenerateStrRep() {
	s.strRep = s.Mapa.String()
}

func (s Set) Remove(i int) (Set, bool) {
	ok := s.Mapa.Contains(i)
	if ok {
		s.Mapa.Delete(i)
		s.GenerateStrRep()
		s.Elems--
	}

	return s, ok
}

func (s Set) Add(i int) (Set, bool) {
	ok := s.Mapa.Contains(i)
	newSet := s.Mapa
	if !ok {
		newSet = s.Mapa.Add(i)
		s.GenerateStrRep()
		s.Elems++
		s.Camino = fmt.Sprintf("%s%d ", s.Camino, i)
	}
	s.Mapa = newSet

	return s, !ok
}
