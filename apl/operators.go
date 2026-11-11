package apl

// reduce applies f/ along the last axis.
func reduce(glyph string, x *Array) (*Array, error) {
	if x.IsScalar() {
		return x, nil
	}
	return reduceAxis(glyph, x, x.Rank()-1)
}

// reduceAxis applies f/ along axis k, folding from the right.
// An empty axis gives f's identity.
func reduceAxis(glyph string, x *Array, k int) (*Array, error) {
	d, err := x.numeric()
	if err != nil {
		return nil, err
	}
	f := scalarFns[glyph].dyad
	outer, n, inner := axisSplit(x.Shape, k)
	shape := append(append([]int(nil), x.Shape[:k]...), x.Shape[k+1:]...)
	r := &Array{Shape: shape, Data: make([]float64, outer*inner)}
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
	for i := range outer {
		for l := range inner {
			base := i*n*inner + l
			acc := d[base+(n-1)*inner]
			for j := n - 2; j >= 0; j-- {
				if acc, err = f(d[base+j*inner], acc); err != nil {
					return nil, err
				}
			}
			r.Data[i*inner+l] = acc
		}
	}
	return r, nil
}

// outer is A∘.fB.
func outer(f dyadFn, a, b *Array) (*Array, error) {
	ad, err := a.numeric()
	if err != nil {
		return nil, err
	}
	bd, err := b.numeric()
	if err != nil {
		return nil, err
	}
	shape := append(append([]int(nil), a.Shape...), b.Shape...)
	r := &Array{Shape: shape, Data: make([]float64, 0, len(ad)*len(bd))}
	for _, x := range ad {
		for _, y := range bd {
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
	ad, err := a.numeric()
	if err != nil {
		return nil, err
	}
	bd, err := b.numeric()
	if err != nil {
		return nil, err
	}
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
	rows, cols := shapeSize(lshape), shapeSize(rshape)
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
	elem := func(x *Array, d []float64, i int) float64 {
		if x.IsScalar() {
			return d[0]
		}
		return d[i]
	}
	for i := range rows {
		for j := range cols {
			var acc float64
			for k := n - 1; k >= 0; k-- {
				v, err := g(elem(a, ad, i*n+k), elem(b, bd, k*cols+j))
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
