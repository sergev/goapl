package apl

// structFn is a structural primitive. The axis forms take the value of
// [K], or nil without one.
type structFn struct {
	monad     func(e *env, x *Array) (*Array, error)
	dyad      func(e *env, a, b *Array) (*Array, error)
	monadAxis func(e *env, ax, x *Array) (*Array, error)
	dyadAxis  func(e *env, ax, a, b *Array) (*Array, error)
}

var structFns = map[string]*structFn{
	"⍳":  {monad: interval, dyad: indexOf},
	"⍴":  {monad: shapeOf, dyad: reshape},
	",":  {monad: ravel, dyadAxis: catenateLast},
	"⍪":  {dyadAxis: catenateFirst},
	"/":  {dyadAxis: alongAxis(compress, false)},
	"⌿":  {dyadAxis: alongAxis(compress, true)},
	"\\": {dyadAxis: alongAxis(expand, false)},
	"⍀":  {dyadAxis: alongAxis(expand, true)},
	"⌽":  {monadAxis: reverseLast, dyadAxis: rotateLast},
	"⊖":  {monadAxis: reverseFirst, dyadAxis: rotateFirst},
	"⍉":  {monad: transpose, dyad: transposeAxes},
	"↑":  {monad: first, dyad: take},
	"↓":  {dyad: drop},
	"∊":  {dyad: member},
	"~":  {dyad: without},
	"⍋":  {monad: gradeUp, dyad: gradeUpCollate},
	"⍒":  {monad: gradeDown, dyad: gradeDownCollate},
	"⊥":  {dyad: decode},
	"⊤":  {dyad: encode},
	"⌹":  {monad: inverse, dyad: matDivide},
	"?":  {dyad: deal},
}

// alongAxis adapts f, which takes an axis from 0 and a non-scalar B,
// to [K]; without [K] it works along the first or last axis.
func alongAxis(f func(e *env, a, b *Array, k int) (*Array, error), first bool) func(e *env, ax, a, b *Array) (*Array, error) {
	return func(e *env, ax, a, b *Array) (*Array, error) {
		if b.IsScalar() {
			b = &Array{Shape: []int{1}, Data: b.Data, Elems: b.Elems}
		}
		k, err := e.axisIndex(ax, b.Rank(), first)
		if err != nil {
			return nil, err
		}
		return f(e, a, b, k)
	}
}

// deal is A?B: A distinct random indices out of the first B.
func deal(e *env, a, b *Array) (*Array, error) {
	if a.Size() != 1 || b.Size() != 1 || a.Rank() > 1 || b.Rank() > 1 {
		return nil, newError(LengthError)
	}
	n, err := toInts(a)
	if err != nil {
		return nil, err
	}
	m, err := toInts(b)
	if err != nil {
		return nil, err
	}
	if n[0] > m[0] {
		return nil, newError(DomainError)
	}
	// Partial Fisher-Yates shuffle over a sparse permutation.
	perm := map[int]int{}
	at := func(i int) int {
		if v, ok := perm[i]; ok {
			return v
		}
		return i
	}
	r := make([]float64, n[0])
	for i := range r {
		j := i + int(e.random()*float64(m[0]-i))
		r[i] = float64(at(j) + e.io)
		perm[j] = at(i)
	}
	return Vector(r), nil
}

// toInts converts a numeric array to non-negative integers.
func toInts(a *Array) ([]int, error) {
	d, err := a.numeric()
	if err != nil {
		return nil, err
	}
	r := make([]int, len(d))
	for i, v := range d {
		n, ok := toInt(v)
		if !ok || n < 0 {
			return nil, newError(DomainError)
		}
		r[i] = n
	}
	return r, nil
}

// interval is ⍳N: the first N indices from ⎕IO.
func interval(e *env, x *Array) (*Array, error) {
	if x.Rank() > 1 {
		return nil, newError(RankError)
	}
	if x.Size() != 1 {
		return nil, newError(LengthError)
	}
	n, err := toInts(x)
	if err != nil {
		return nil, err
	}
	r := make([]float64, n[0])
	for i := range r {
		r[i] = float64(i + e.io)
	}
	return Vector(r), nil
}

func shapeOf(_ *env, x *Array) (*Array, error) {
	r := make([]float64, x.Rank())
	for i, d := range x.Shape {
		r[i] = float64(d)
	}
	return Vector(r), nil
}

func ravel(_ *env, x *Array) (*Array, error) {
	return &Array{Shape: []int{x.Size()}, Data: x.Data, Elems: x.Elems, Proto: x.Proto}, nil
}

// indexOf is A⍳B: index of each B in vector A, or ⎕IO+⍴A if absent.
func indexOf(e *env, a, b *Array) (*Array, error) {
	if a.Rank() != 1 {
		return nil, newError(RankError)
	}
	r := &Array{Shape: b.Shape, Data: make([]float64, b.Size())}
	for i := range r.Data {
		v := b.At(i)
		r.Data[i] = float64(a.Size() + e.io)
		for j := range a.Size() {
			if match(v, a.At(j), e.ct) {
				r.Data[i] = float64(j + e.io)
				break
			}
		}
	}
	return r, nil
}

// reshape is A⍴B, cycling B; an empty B gives its fill element.
func reshape(_ *env, a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	shape, err := toInts(a)
	if err != nil {
		return nil, err
	}
	pos := make([]int, shapeSize(shape))
	for i := range pos {
		pos[i] = -1
		if b.Size() > 0 {
			pos[i] = i % b.Size()
		}
	}
	return b.gather(shape, pos), nil
}

// indexPositions returns flat offsets selected by idx (nil = whole axis,
// indices from io) and the shape of the selection.
func indexPositions(io int, shape []int, idx []*Array) ([]int, []int, error) {
	if len(idx) != len(shape) {
		return nil, nil, newError(RankError)
	}
	pos := []int{0}
	var rshape []int
	for k, ix := range idx {
		var sel []int
		if ix == nil {
			sel = make([]int, shape[k])
			for i := range sel {
				sel[i] = i
			}
			rshape = append(rshape, shape[k])
		} else {
			d, err := ix.numeric()
			if err != nil {
				return nil, nil, err
			}
			sel = make([]int, len(d))
			for i, v := range d {
				n, ok := toInt(v)
				if !ok {
					return nil, nil, newError(DomainError)
				}
				n -= io
				if n < 0 || n >= shape[k] {
					return nil, nil, newError(IndexError)
				}
				sel[i] = n
			}
			rshape = append(rshape, ix.Shape...)
		}
		next := make([]int, 0, len(pos)*len(sel))
		for _, p := range pos {
			for _, s := range sel {
				next = append(next, p*shape[k]+s)
			}
		}
		pos = next
	}
	return pos, rshape, nil
}

// index is X[I;J;...].
func index(e *env, x *Array, idx []*Array) (*Array, error) {
	pos, shape, err := indexPositions(e.io, x.Shape, idx)
	if err != nil {
		return nil, err
	}
	return x.gather(shape, pos), nil
}

// indexAssign returns a copy of x with X[I;J;...]←v applied.
func indexAssign(e *env, x *Array, idx []*Array, v *Array) (*Array, error) {
	pos, shape, err := indexPositions(e.io, x.Shape, idx)
	if err != nil {
		return nil, err
	}
	if v.Size() != 1 && !equalShape(v.Shape, shape) {
		return nil, newError(LengthError)
	}
	elem := func(i int) int {
		if v.Size() == 1 {
			return 0
		}
		return i
	}
	if x.IsNumeric() && v.IsNumeric() {
		r := &Array{Shape: x.Shape, Data: append([]float64(nil), x.Data...)}
		for i, p := range pos {
			r.Data[p] = v.Data[elem(i)]
		}
		return r, nil
	}
	r := append([]Value(nil), x.elems()...)
	for i, p := range pos {
		r[p] = v.At(elem(i))
	}
	return General(x.Shape, r, x.Fill()), nil
}

// compress is A/[k]B: A gives a repeat count for each item along
// axis k. A single count applies to every item, and a single item is
// repeated to match A.
func compress(_ *env, a, b *Array, k int) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	counts, err := toInts(a)
	if err != nil {
		return nil, err
	}
	outer, n, inner := axisSplit(b.Shape, k)
	src := make([]int, 0, len(counts))
	switch {
	case len(counts) == 1:
		for j := range n {
			for range counts[0] {
				src = append(src, j)
			}
		}
	case n == 1 || len(counts) == n:
		for j, c := range counts {
			for range c {
				src = append(src, min(j, n-1))
			}
		}
	default:
		return nil, newError(LengthError)
	}
	return selectAxis(b, k, outer, n, inner, src), nil
}

// expand is A\[k]B: B's items along axis k go where A is 1, and fill
// elements where A is 0. A single item of B is repeated.
func expand(_ *env, a, b *Array, k int) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	bits, err := toInts(a)
	if err != nil {
		return nil, err
	}
	outer, n, inner := axisSplit(b.Shape, k)
	ones := 0
	for _, v := range bits {
		if v > 1 {
			return nil, newError(DomainError)
		}
		ones += v
	}
	if ones != n && n != 1 {
		return nil, newError(LengthError)
	}
	src := make([]int, len(bits))
	j := 0
	for i, v := range bits {
		src[i] = -1
		if v == 1 {
			src[i] = min(j, n-1)
			j++
		}
	}
	return selectAxis(b, k, outer, n, inner, src), nil
}

// selectAxis builds an array from b whose items along axis k are b's
// items src (-1 for a fill item).
func selectAxis(b *Array, k, outer, n, inner int, src []int) *Array {
	shape := append([]int(nil), b.Shape...)
	shape[k] = len(src)
	pos := make([]int, 0, outer*len(src)*inner)
	for i := range outer {
		for _, j := range src {
			for l := range inner {
				if j < 0 {
					pos = append(pos, -1)
				} else {
					pos = append(pos, (i*n+j)*inner+l)
				}
			}
		}
	}
	return b.gather(shape, pos)
}
