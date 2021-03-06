package entities

import "fmt"

type FType struct {
	Mapa   map[string]float32
	N      int
	Result map[string][]TType
}

func (f *FType) Set(s *Set, i int, t float32, value float32) {
	f.Mapa[fmt.Sprintf("%s_%d_%f", s.GetStrRep(), i, t)] = value
	if s.Elems == f.N {
		key := fmt.Sprintf("%s_%d", s.GetLargeStrRep(), i)
		res, ok := f.Result[key]
		if !ok {
			res = make([]TType, 0)
		}
		res = append(res, TType{T: t, Value: value, Camino: s.Camino})
		f.Result[key] = res
	}
}

func (f *FType) Get(s *Set, i int, t float32) float32 {
	return f.Mapa[fmt.Sprintf("%s_%d_%f", s.GetStrRep(), i, t)]
}
