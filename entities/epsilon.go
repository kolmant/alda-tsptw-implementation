package entities

type Epsilon struct {
	Set *Set
	I   int
	T   float32
}

func (e *Epsilon) Copy() Epsilon {
	set := e.Set.Copy()
	return Epsilon{
		Set: &set,
		I:   e.I,
		T:   e.T,
	}
}
