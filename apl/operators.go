package apl

// reduce applies f/ along the last axis.
func reduce(e *env, glyph string, x *Array) (*Array, error) {
	if x.IsScalar() {
		return x, nil
	}
	return reduceAxis(e, glyph, x, x.Rank()-1)
}

// reduceAxis applies f/ along axis k, folding from the right.
// An empty axis gives f's identity.
func reduceAxis(e *env, glyph string, x *Array, k int) (*Array, error) {
	sf := scalarFns[glyph]
	outer, n, inner := axisSplit(x.Shape, k)
	shape := append(append([]int(nil), x.Shape[:k]...), x.Shape[k+1:]...)
	if n == 0 {
		return identityArray(glyph, shape)
	}
	if x.IsNumeric() {
		r := &Array{Shape: shape, Data: make([]float64, outer*inner)}
		for i := range outer {
			for l := range inner {
				base := i*n*inner + l
				acc := x.Data[base+(n-1)*inner]
				for j := n - 2; j >= 0; j-- {
					v, err := sf.dyad(e, x.Data[base+j*inner], acc)
					if err != nil {
						return nil, err
					}
					acc = v
				}
				r.Data[i*inner+l] = acc
			}
		}
		return r, nil
	}
	r := make([]Value, outer*inner)
	for i := range outer {
		for l := range inner {
			base := i*n*inner + l
			acc := x.Elems[base+(n-1)*inner]
			for j := n - 2; j >= 0; j-- {
				v, err := sf.elem(e, x.Elems[base+j*inner], acc)
				if err != nil {
					return nil, err
				}
				acc = v
			}
			r[i*inner+l] = acc
		}
	}
	return General(shape, r, x.Fill()), nil
}

// associative functions can be scanned in one pass.
var associative = map[string]bool{"+": true, "×": true, "⌈": true, "⌊": true, "∧": true, "∨": true}

// scanAxis is f\[k]X: item j is the reduction of items 0 to j.
func scanAxis(e *env, glyph string, x *Array, k int) (*Array, error) {
	sf := scalarFns[glyph]
	outer, n, inner := axisSplit(x.Shape, k)
	r := make([]Value, x.Size())
	for i := range outer {
		for l := range inner {
			at := func(j int) int { return (i*n+j)*inner + l }
			for j := range n {
				if j == 0 {
					r[at(0)] = x.At(at(0))
					continue
				}
				if associative[glyph] {
					v, err := sf.elem(e, r[at(j-1)], x.At(at(j)))
					if err != nil {
						return nil, err
					}
					r[at(j)] = v
					continue
				}
				acc := x.At(at(j))
				for m := j - 1; m >= 0; m-- {
					v, err := sf.elem(e, x.At(at(m)), acc)
					if err != nil {
						return nil, err
					}
					acc = v
				}
				r[at(j)] = acc
			}
		}
	}
	return General(x.Shape, r, x.Fill()), nil
}

// nwise is N f/[k]X: f/ over each window of |N| items along axis k;
// a negative N reverses the windows.
func nwise(e *env, glyph string, a, x *Array, k int) (*Array, error) {
	if a.Size() != 1 || a.Rank() > 1 {
		return nil, newError(LengthError)
	}
	d, err := a.numeric()
	if err != nil {
		return nil, err
	}
	w, ok := toInt(d[0])
	if !ok {
		return nil, newError(DomainError)
	}
	rev := w < 0
	if rev {
		w = -w
	}
	outer, n, inner := axisSplit(x.Shape, k)
	if w > n+1 {
		return nil, newError(LengthError)
	}
	shape := append([]int(nil), x.Shape...)
	shape[k] = n - w + 1
	if w == 0 {
		return identityArray(glyph, shape)
	}
	sf := scalarFns[glyph]
	r := make([]Value, 0, shapeSize(shape))
	for i := range outer {
		for j := range n - w + 1 {
			for l := range inner {
				at := func(m int) Value {
					if rev {
						m = w - 1 - m
					}
					return x.At((i*n+j+m)*inner + l)
				}
				acc := at(w - 1)
				for m := w - 2; m >= 0; m-- {
					v, err := sf.elem(e, at(m), acc)
					if err != nil {
						return nil, err
					}
					acc = v
				}
				r = append(r, acc)
			}
		}
	}
	return General(shape, r, x.Fill()), nil
}

// identityArray is an array of f's identity element, or DOMAIN ERROR.
func identityArray(glyph string, shape []int) (*Array, error) {
	id, ok := identity[glyph]
	if !ok {
		return nil, newError(DomainError)
	}
	r := &Array{Shape: shape, Data: make([]float64, shapeSize(shape))}
	for i := range r.Data {
		r.Data[i] = id
	}
	return r, nil
}

// outer is A∘.fB.
func outer(e *env, sf *scalarFn, a, b *Array) (*Array, error) {
	shape := append(append([]int(nil), a.Shape...), b.Shape...)
	if a.IsNumeric() && b.IsNumeric() {
		r := &Array{Shape: shape, Data: make([]float64, 0, a.Size()*b.Size())}
		for _, x := range a.Data {
			for _, y := range b.Data {
				v, err := sf.dyad(e, x, y)
				if err != nil {
					return nil, err
				}
				r.Data = append(r.Data, v)
			}
		}
		return r, nil
	}
	r := make([]Value, 0, a.Size()*b.Size())
	for i := range a.Size() {
		for j := range b.Size() {
			v, err := sf.elem(e, a.At(i), b.At(j))
			if err != nil {
				return nil, err
			}
			r = append(r, v)
		}
	}
	return General(shape, r, nil), nil
}

// inner is Af.gB: f/ over g applied along the last axis of A
// and the first axis of B.
func inner(e *env, fg, gg string, a, b *Array) (*Array, error) {
	f, g := scalarFns[fg], scalarFns[gg]
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
	if n == 0 {
		return identityArray(fg, shape)
	}
	rows, cols := shapeSize(lshape), shapeSize(rshape)
	ai := func(i int) int {
		if a.IsScalar() {
			return 0
		}
		return i
	}
	bi := func(i int) int {
		if b.IsScalar() {
			return 0
		}
		return i
	}
	if a.IsNumeric() && b.IsNumeric() {
		r := &Array{Shape: shape, Data: make([]float64, rows*cols)}
		for i := range rows {
			for j := range cols {
				var acc float64
				for k := n - 1; k >= 0; k-- {
					v, err := g.dyad(e, a.Data[ai(i*n+k)], b.Data[bi(k*cols+j)])
					if err == nil && k < n-1 {
						v, err = f.dyad(e, v, acc)
					}
					if err != nil {
						return nil, err
					}
					acc = v
				}
				r.Data[i*cols+j] = acc
			}
		}
		return r, nil
	}
	r := make([]Value, rows*cols)
	for i := range rows {
		for j := range cols {
			var acc Value
			for k := n - 1; k >= 0; k-- {
				v, err := g.elem(e, a.At(ai(i*n+k)), b.At(bi(k*cols+j)))
				if err == nil && k < n-1 {
					v, err = f.elem(e, v, acc)
				}
				if err != nil {
					return nil, err
				}
				acc = v
			}
			r[i*cols+j] = acc
		}
	}
	return General(shape, r, nil), nil
}
