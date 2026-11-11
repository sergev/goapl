package apl

// Operators applied to functions other than primitive scalar functions
// work on items: each element is passed to the function as an array,
// and each result becomes an element.

type dyadicFn func(a, b *Array) (*Array, error)

// dyadicOf gives f as a dyadic function that must return a value.
func (in *Interp) dyadicOf(f *fnNode) dyadicFn {
	return func(a, b *Array) (*Array, error) {
		r, err := in.applyFn(f, a, b)
		if err == nil && r == nil {
			err = newError(ValueError)
		}
		return r, err
	}
}

// fold reduces items from the right: x0 f (x1 f (... f xn)).
func fold(fn dyadicFn, items []Value) (Value, error) {
	acc := item(items[len(items)-1])
	for j := len(items) - 2; j >= 0; j-- {
		r, err := fn(item(items[j]), acc)
		if err != nil {
			return nil, err
		}
		acc = r
	}
	return itemValue(acc), nil
}

// window checks the window size A of A f/B for an axis of length n.
func window(a *Array, n int) (int, bool, error) {
	if a.Size() != 1 || a.Rank() > 1 {
		return 0, false, newError(LengthError)
	}
	w, err := toSignedInts(a)
	if err != nil {
		return 0, false, err
	}
	rev := w[0] < 0
	if abs(w[0]) > n+1 {
		return 0, false, newError(LengthError)
	}
	return abs(w[0]), rev, nil
}

// windowsGen reduces each window of w items along axis k by fn; with
// w the length of the axis it is the reduction. An empty window has
// no identity element.
func windowsGen(fn dyadicFn, x *Array, k, w int, rev bool) (*Array, error) {
	if x.IsScalar() {
		return x, nil
	}
	if w == 0 {
		return nil, newError(DomainError)
	}
	outer, n, inner := axisSplit(x.Shape, k)
	shape := append([]int(nil), x.Shape...)
	shape[k] = n - w + 1
	if w == n {
		shape = append(append([]int(nil), x.Shape[:k]...), x.Shape[k+1:]...)
	}
	var r []Value
	items := make([]Value, w)
	for i := range outer {
		for j := range n - w + 1 {
			for l := range inner {
				for m := range w {
					s := m
					if rev {
						s = w - 1 - m
					}
					items[m] = x.At((i*n+j+s)*inner + l)
				}
				v, err := fold(fn, items)
				if err != nil {
					return nil, err
				}
				r = append(r, v)
			}
		}
	}
	return General(shape, r, x.Fill()), nil
}

// scanGen is f\[k]X for any dyadic f.
func scanGen(fn dyadicFn, x *Array, k int) (*Array, error) {
	outer, n, inner := axisSplit(x.Shape, k)
	r := make([]Value, x.Size())
	for i := range outer {
		for l := range inner {
			items := make([]Value, 0, n)
			for j := range n {
				p := (i*n+j)*inner + l
				items = append(items, x.At(p))
				v, err := fold(fn, items)
				if err != nil {
					return nil, err
				}
				r[p] = v
			}
		}
	}
	return General(x.Shape, r, x.Fill()), nil
}

// outerGen is A∘.gB for any dyadic g.
func outerGen(g dyadicFn, a, b *Array) (*Array, error) {
	shape := append(append([]int(nil), a.Shape...), b.Shape...)
	r := make([]Value, 0, a.Size()*b.Size())
	for i := range a.Size() {
		for j := range b.Size() {
			v, err := g(item(a.At(i)), item(b.At(j)))
			if err != nil {
				return nil, err
			}
			r = append(r, itemValue(v))
		}
	}
	return General(shape, r, nil), nil
}

// innerGen is Af.gB for any dyadic f and g.
func innerGen(f, g dyadicFn, a, b *Array) (*Array, error) {
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
	if n == 0 {
		return nil, newError(DomainError)
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
	at := func(x *Array, i int) Value {
		if x.IsScalar() {
			return x.At(0)
		}
		return x.At(i)
	}
	r := make([]Value, 0, rows*cols)
	items := make([]Value, n)
	for i := range rows {
		for j := range cols {
			for k := range n {
				v, err := g(item(at(a, i*n+k)), item(at(b, k*cols+j)))
				if err != nil {
					return nil, err
				}
				items[k] = itemValue(v)
			}
			v, err := fold(f, items)
			if err != nil {
				return nil, err
			}
			r = append(r, v)
		}
	}
	return General(shape, r, nil), nil
}

// callOp calls a defined operator with its operands bound.
func (in *Interp) callOp(f *fnNode, a, b *Array) (*Array, error) {
	op := f.user
	if op.Arity() == 0 || a != nil && op.Arity() == 1 {
		return nil, newError(ValenceError)
	}
	frame := map[string]*fnNode{op.OpLeft: in.resolve(f.left)}
	if op.OpRight != "" {
		frame[op.OpRight] = in.resolve(f.right)
	}
	in.opFrames = append(in.opFrames, frame)
	defer func() { in.opFrames = in.opFrames[:len(in.opFrames)-1] }()
	return in.call(op, a, b)
}

// operand gives the function bound to an operand name.
func (in *Interp) operand(name string) *fnNode {
	for i := len(in.opFrames) - 1; i >= 0; i-- {
		if f, ok := in.opFrames[i][name]; ok {
			return f
		}
	}
	panic("unbound operand " + name)
}

// resolve replaces operand names in f by the functions bound to them
// now, so that f keeps its meaning inside another operator.
func (in *Interp) resolve(f *fnNode) *fnNode {
	if f == nil {
		return nil
	}
	if f.kind == fnOperand {
		return in.operand(f.glyph)
	}
	if f.left == nil && f.right == nil {
		return f
	}
	c := *f
	c.left, c.right = in.resolve(f.left), in.resolve(f.right)
	return &c
}
