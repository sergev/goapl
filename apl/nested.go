package apl

// enclose is ⊂X: a scalar holding X; a simple scalar stays as it is.
func enclose(_ *env, x *Array) (*Array, error) {
	if x.IsScalar() && !isNested(x.At(0)) {
		return x, nil
	}
	return &Array{Elems: []Value{x}}, nil
}

// disclose is ⊃X: the items of X as an array with the axes of X
// followed by the largest shape of the items; smaller items are padded
// with their fill elements and get leading unit axes.
func disclose(e *env, x *Array) (*Array, error) {
	if x.IsSimple() {
		return x, nil
	}
	var shape []int
	for i := range x.Size() {
		s := item(x.At(i)).Shape
		if len(s) > len(shape) {
			shape = append(ones(len(s)-len(shape)), shape...)
		}
		off := len(shape) - len(s)
		for k, n := range s {
			shape[off+k] = max(shape[off+k], n)
		}
	}
	counts := make([]float64, len(shape))
	for k, n := range shape {
		counts[k] = float64(n)
	}
	var elems []Value
	for i := range x.Size() {
		it := item(x.At(i))
		lead := ones(len(shape) - it.Rank())
		it = &Array{Shape: append(lead, it.Shape...), Data: it.Data, Elems: it.Elems, Proto: it.Proto}
		t, err := take(e, Vector(counts), it)
		if err != nil {
			return nil, err
		}
		elems = append(elems, t.elems()...)
	}
	return General(append(append([]int(nil), x.Shape...), shape...), elems, x.Fill()), nil
}

func ones(n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = 1
	}
	return r
}

// pickItem is A⊃B: each item of A selects an item at the next level of
// B; it holds one index per axis of the array at that level.
func pickItem(e *env, a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	cur := b
	for i := range a.Size() {
		idx, err := toSignedInts(item(a.At(i)))
		if err != nil {
			return nil, err
		}
		if len(idx) != cur.Rank() {
			return nil, newError(RankError)
		}
		off := 0
		for k, n := range idx {
			n -= e.io
			if n < 0 || n >= cur.Shape[k] {
				return nil, newError(IndexError)
			}
			off = off*cur.Shape[k] + n
		}
		cur = item(cur.At(off))
	}
	return cur, nil
}

// depthFn is ≡X.
func depthFn(_ *env, x *Array) (*Array, error) { return Scalar(float64(depth(x))), nil }

// matchFn is A≡B: 1 if A and B have the same shape and elements.
func matchFn(e *env, a, b *Array) (*Array, error) {
	return Scalar(boolVal(match(a, b, e.ct))), nil
}

// enlist is ∊X: the simple scalars of X, in order, as a vector.
func enlist(_ *env, x *Array) (*Array, error) {
	if x.IsNumeric() {
		return Vector(x.Data), nil
	}
	var r []Value
	var walk func(a *Array)
	walk = func(a *Array) {
		for i := range a.Size() {
			if v, ok := a.At(i).(*Array); ok {
				walk(v)
			} else {
				r = append(r, a.At(i))
			}
		}
	}
	walk(x)
	return General([]int{len(r)}, r, x.Fill()), nil
}

// nonce stands for a function form not implemented yet.
func nonce(_ *env, _, _ *Array) (*Array, error) { return nil, newError(NonceError) }
