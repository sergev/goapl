package apl

// reduce applies f/ along the last axis, folding from the right.
// A scalar is returned as is; an empty axis gives f's identity.
func reduce(glyph string, x *Array) (*Array, error) {
	if x.IsScalar() {
		return x, nil
	}
	f := scalarFns[glyph].dyad
	n := x.Shape[x.Rank()-1]
	shape := append([]int(nil), x.Shape[:x.Rank()-1]...)
	rows := shapeSize(shape)
	r := &Array{Shape: shape, Data: make([]float64, rows)}
	if n == 0 {
		id, ok := identity[glyph]
		if !ok {
			return nil, newError(DomainError)
		}
		for i := range r.Data {
			r.Data[i] = id
		}
		return r, nil
	}
	for i := range rows {
		row := x.Data[i*n : (i+1)*n]
		acc := row[n-1]
		for j := n - 2; j >= 0; j-- {
			v, err := f(row[j], acc)
			if err != nil {
				return nil, err
			}
			acc = v
		}
		r.Data[i] = acc
	}
	return r, nil
}

// outer is A∘.fB.
func outer(f dyadFn, a, b *Array) (*Array, error) {
	shape := append(append([]int(nil), a.Shape...), b.Shape...)
	if len(shape) > MaxRank {
		return nil, newError(SystemLimit)
	}
	r := &Array{Shape: shape, Data: make([]float64, 0, a.Size()*b.Size())}
	for _, x := range a.Data {
		for _, y := range b.Data {
			v, err := f(x, y)
			if err != nil {
				return nil, err
			}
			r.Data = append(r.Data, v)
		}
	}
	return r, nil
}

// inner is Af.gB: f/ over g applied along the last axis of A
// and the first axis of B.
func inner(fg, gg string, a, b *Array) (*Array, error) {
	f, g := scalarFns[fg].dyad, scalarFns[gg].dyad
	n := 1
	switch {
	case !a.IsScalar() && !b.IsScalar():
		n = a.Shape[a.Rank()-1]
		if n != b.Shape[0] {
			return nil, newError(LengthError)
		}
	case !a.IsScalar():
		n = a.Shape[a.Rank()-1]
	case !b.IsScalar():
		n = b.Shape[0]
	}
	var lshape, rshape []int
	if !a.IsScalar() {
		lshape = a.Shape[:a.Rank()-1]
	}
	if !b.IsScalar() {
		rshape = b.Shape[1:]
	}
	shape := append(append([]int(nil), lshape...), rshape...)
	if len(shape) > MaxRank {
		return nil, newError(SystemLimit)
	}
	rows, cols := shapeSize(lshape), shapeSize(rshape)
	elem := func(x *Array, i int) float64 {
		if x.IsScalar() {
			return x.Data[0]
		}
		return x.Data[i]
	}
	r := &Array{Shape: shape, Data: make([]float64, rows*cols)}
	if n == 0 {
		id, ok := identity[fg]
		if !ok {
			return nil, newError(DomainError)
		}
		for i := range r.Data {
			r.Data[i] = id
		}
		return r, nil
	}
	for i := range rows {
		for j := range cols {
			var acc float64
			for k := n - 1; k >= 0; k-- {
				v, err := g(elem(a, i*n+k), elem(b, k*cols+j))
				if err != nil {
					return nil, err
				}
				if k < n-1 {
					if v, err = f(v, acc); err != nil {
						return nil, err
					}
				}
				acc = v
			}
			r.Data[i*cols+j] = acc
		}
	}
	return r, nil
}
