package apl

// reduce applies f/ along the last axis, folding from the right.
func reduce(f dyadFn, x *Array) (*Array, error) {
	// Prototype: scalar or empty → 24, one-element vector → 51.
	if x.IsScalar() || x.Size() == 0 {
		return nil, newError(24)
	}
	if x.Rank() == 1 && x.Shape[0] == 1 {
		return nil, newError(51)
	}
	n := x.Shape[x.Rank()-1]
	shape := append([]int(nil), x.Shape[:x.Rank()-1]...)
	rows := shapeSize(shape)
	r := &Array{Shape: shape, Data: make([]float64, rows)}
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
		return nil, newError(60)
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
func inner(f, g dyadFn, a, b *Array) (*Array, error) {
	n := 1
	switch {
	case !a.IsScalar() && !b.IsScalar():
		n = a.Shape[a.Rank()-1]
		if n != b.Shape[0] {
			return nil, newError(52)
		}
	case !a.IsScalar():
		n = a.Shape[a.Rank()-1]
	case !b.IsScalar():
		n = b.Shape[0]
	}
	if n == 0 {
		return nil, newError(24)
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
		return nil, newError(60)
	}
	rows, cols := shapeSize(lshape), shapeSize(rshape)
	elem := func(x *Array, i int) float64 {
		if x.IsScalar() {
			return x.Data[0]
		}
		return x.Data[i]
	}
	r := &Array{Shape: shape, Data: make([]float64, rows*cols)}
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
