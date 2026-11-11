package apl

import "math"

// pick builds an array whose item j along axis k is b's item src[k][j]
// (-1 for a fill item).
func pick(b *Array, src [][]int) *Array {
	shape := make([]int, len(src))
	for k, s := range src {
		shape[k] = len(s)
	}
	pos := make([]int, shapeSize(shape))
	idx := make([]int, len(shape))
	for i := range pos {
		p := 0
		for k, j := range idx {
			s := src[k][j]
			if s < 0 {
				p = -1
				break
			}
			p = p*b.Shape[k] + s
		}
		pos[i] = p
		for k := len(idx) - 1; k >= 0; k-- {
			if idx[k]++; idx[k] < shape[k] {
				break
			}
			idx[k] = 0
		}
	}
	return b.gather(shape, pos)
}

// identitySrc is src for pick that keeps every axis of shape.
func identitySrc(shape []int) [][]int {
	src := make([][]int, len(shape))
	for k, n := range shape {
		src[k] = make([]int, n)
		for j := range n {
			src[k][j] = j
		}
	}
	return src
}

// joinRavel is a vector of a's elements followed by b's.
func joinRavel(a, b *Array) *Array {
	n := a.Size() + b.Size()
	if a.IsNumeric() && b.IsNumeric() {
		r := make([]float64, 0, n)
		return Vector(append(append(r, a.Data...), b.Data...))
	}
	r := make([]Value, 0, n)
	r = append(append(r, a.elems()...), b.elems()...)
	return General([]int{n}, r, a.Fill())
}

func catenateLast(e *env, ax, a, b *Array) (*Array, error)  { return catenate(e, ax, a, b, false) }
func catenateFirst(e *env, ax, a, b *Array) (*Array, error) { return catenate(e, ax, a, b, true) }

// catenate is A,[K]B: joins along axis K, by default the last (or
// first) axis. A fractional K laminates along a new axis. An argument
// of rank one less gets a unit axis; a scalar is extended.
func catenate(e *env, ax, a, b *Array, first bool) (*Array, error) {
	r := max(a.Rank(), b.Rank(), 1)
	if ax != nil && ax.IsNumeric() && ax.Size() == 1 {
		if x := ax.Data[0]; x != math.Round(x) {
			return laminate(e, x, a, b)
		}
	}
	if a.Rank() > 0 && b.Rank() > 0 && (a.Rank() < r-1 || b.Rank() < r-1) {
		return nil, newError(RankError)
	}
	k, err := e.axisIndex(ax, r, first)
	if err != nil {
		return nil, err
	}
	// Shapes of rank r, with axis k of a lower-rank argument set to 1.
	conform := func(x, other *Array) []int {
		switch {
		case x.Rank() == r:
			return x.Shape
		case x.IsScalar():
			s := append([]int(nil), other.Shape...)
			if len(s) < r {
				s = make([]int, r)
				for i := range s {
					s[i] = 1
				}
			}
			s[k] = 1
			return s
		}
		return append(append(append([]int(nil), x.Shape[:k]...), 1), x.Shape[k:]...)
	}
	as, bs := conform(a, b), conform(b, a)
	for i := range r {
		if i != k && as[i] != bs[i] {
			return nil, newError(LengthError)
		}
	}
	shape := append([]int(nil), as...)
	shape[k] = as[k] + bs[k]
	outer, _, inner := axisSplit(shape, k)
	src := joinRavel(a, b)
	pos := make([]int, 0, shapeSize(shape))
	add := func(x *Array, base, i, n int) {
		for j := range n * inner {
			if x.IsScalar() {
				pos = append(pos, base)
			} else {
				pos = append(pos, base+i*n*inner+j)
			}
		}
	}
	for i := range outer {
		add(a, 0, i, as[k])
		add(b, a.Size(), i, bs[k])
	}
	return src.gather(shape, pos), nil
}

// laminate is A,[K]B for fractional K: joins A and B along a new axis
// of length 2 before axis ⌈K.
func laminate(e *env, x float64, a, b *Array) (*Array, error) {
	shape := a.Shape
	switch {
	case a.IsScalar():
		shape = b.Shape
	case !b.IsScalar() && !sameShape(a, b):
		if a.Rank() != b.Rank() {
			return nil, newError(RankError)
		}
		return nil, newError(LengthError)
	}
	p := int(math.Ceil(x)) - e.io
	if p < 0 || p > len(shape) {
		return nil, newError(AxisError)
	}
	rshape := append(append(append([]int(nil), shape[:p]...), 2), shape[p:]...)
	outer, inner := shapeSize(shape[:p]), shapeSize(shape[p:])
	src := joinRavel(a, b)
	pos := make([]int, 0, 2*outer*inner)
	for i := range outer {
		for l := range inner {
			if a.IsScalar() {
				pos = append(pos, 0)
			} else {
				pos = append(pos, i*inner+l)
			}
		}
		for l := range inner {
			if b.IsScalar() {
				pos = append(pos, a.Size())
			} else {
				pos = append(pos, a.Size()+i*inner+l)
			}
		}
	}
	return src.gather(rshape, pos), nil
}

func reverseLast(e *env, ax, x *Array) (*Array, error)  { return reverse(e, ax, x, false) }
func reverseFirst(e *env, ax, x *Array) (*Array, error) { return reverse(e, ax, x, true) }

// reverse is ⌽[K]X.
func reverse(e *env, ax, x *Array, first bool) (*Array, error) {
	if x.IsScalar() && ax == nil {
		return x, nil
	}
	k, err := e.axisIndex(ax, x.Rank(), first)
	if err != nil {
		return nil, err
	}
	src := identitySrc(x.Shape)
	n := x.Shape[k]
	for j := range n {
		src[k][j] = n - 1 - j
	}
	return pick(x, src), nil
}

func rotateLast(e *env, ax, a, b *Array) (*Array, error)  { return rotate(e, ax, a, b, false) }
func rotateFirst(e *env, ax, a, b *Array) (*Array, error) { return rotate(e, ax, a, b, true) }

// rotate is A⌽[K]B: rotates each line along axis K by A, a scalar or
// an array of the shape of B without axis K.
func rotate(e *env, ax, a, b *Array, first bool) (*Array, error) {
	if b.IsScalar() {
		if ax != nil {
			return nil, newError(AxisError)
		}
		_, err := toSignedInts(a)
		return b, err
	}
	k, err := e.axisIndex(ax, b.Rank(), first)
	if err != nil {
		return nil, err
	}
	amounts, err := toSignedInts(a)
	if err != nil {
		return nil, err
	}
	frame := append(append([]int(nil), b.Shape[:k]...), b.Shape[k+1:]...)
	if a.Size() != 1 && !equalShape(a.Shape, frame) {
		if a.Rank() != len(frame) {
			return nil, newError(RankError)
		}
		return nil, newError(LengthError)
	}
	outer, n, inner := axisSplit(b.Shape, k)
	pos := make([]int, b.Size())
	for i := range outer {
		for l := range inner {
			s := amounts[0]
			if len(amounts) > 1 {
				s = amounts[i*inner+l]
			}
			for j := range n {
				src := ((j+s)%n + n) % n
				pos[(i*n+j)*inner+l] = (i*n+src)*inner + l
			}
		}
	}
	return b.gather(b.Shape, pos), nil
}

// toSignedInts converts a numeric array to integers.
func toSignedInts(a *Array) ([]int, error) {
	d, err := a.numeric()
	if err != nil {
		return nil, err
	}
	r := make([]int, len(d))
	for i, v := range d {
		n, ok := toInt(v)
		if !ok {
			return nil, newError(DomainError)
		}
		r[i] = n
	}
	return r, nil
}

// transpose is ⍉X: reverses the order of the axes.
func transpose(e *env, x *Array) (*Array, error) {
	perm := make([]float64, x.Rank())
	for i := range perm {
		perm[i] = float64(x.Rank() - 1 - i + e.io)
	}
	return transposeAxes(e, Vector(perm), x)
}

// transposeAxes is A⍉B: axis I of B becomes axis A[I] of the result.
// Axes mapped to the same result axis take their diagonal.
func transposeAxes(e *env, a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	if a.Size() != b.Rank() {
		return nil, newError(LengthError)
	}
	to, err := toSignedInts(a)
	if err != nil {
		return nil, err
	}
	r := 0
	for i := range to {
		to[i] -= e.io
		if to[i] < 0 || to[i] >= b.Rank() {
			return nil, newError(DomainError)
		}
		r = max(r, to[i]+1)
	}
	shape := make([]int, r)
	for j := range shape {
		shape[j] = -1
	}
	for i, j := range to {
		if shape[j] < 0 || b.Shape[i] < shape[j] {
			shape[j] = b.Shape[i]
		}
	}
	for _, n := range shape {
		if n < 0 {
			return nil, newError(DomainError)
		}
	}
	pos := make([]int, shapeSize(shape))
	idx := make([]int, r)
	for p := range pos {
		off := 0
		for i, j := range to {
			off = off*b.Shape[i] + idx[j]
		}
		pos[p] = off
		for j := r - 1; j >= 0; j-- {
			if idx[j]++; idx[j] < shape[j] {
				break
			}
			idx[j] = 0
		}
	}
	return b.gather(shape, pos), nil
}

// first is ↑X: the first element of X, or its fill element if empty.
func first(_ *env, x *Array) (*Array, error) {
	v := x.Fill()
	if x.Size() > 0 {
		v = x.At(0)
	}
	if a, ok := v.(*Array); ok {
		return a, nil
	}
	return General(nil, []Value{v}, nil), nil
}

// takeDropArgs checks A of A↑B or A↓B and gives B with at least as
// many axes as A has elements.
func takeDropArgs(a, b *Array) ([]int, *Array, error) {
	if a.Rank() > 1 {
		return nil, nil, newError(RankError)
	}
	n, err := toSignedInts(a)
	if err != nil {
		return nil, nil, err
	}
	if b.IsScalar() {
		shape := make([]int, len(n))
		for i := range shape {
			shape[i] = 1
		}
		b = &Array{Shape: shape, Data: b.Data, Elems: b.Elems, Proto: b.Proto}
	}
	if len(n) > b.Rank() {
		return nil, nil, newError(RankError)
	}
	return n, b, nil
}

// take is A↑B: the first (or, for negative A, last) A[K] items along
// each axis K; beyond B there are fill elements.
func take(_ *env, a, b *Array) (*Array, error) {
	counts, b, err := takeDropArgs(a, b)
	if err != nil {
		return nil, err
	}
	src := identitySrc(b.Shape)
	for k, c := range counts {
		n := b.Shape[k]
		m := abs(c)
		src[k] = make([]int, m)
		for j := range m {
			s := j
			if c < 0 {
				s = j + n - m
			}
			if s < 0 || s >= n {
				s = -1
			}
			src[k][j] = s
		}
	}
	return pick(b, src), nil
}

// drop is A↓B: without the first (or, for negative A, last) A[K] items
// along each axis K.
func drop(_ *env, a, b *Array) (*Array, error) {
	counts, b, err := takeDropArgs(a, b)
	if err != nil {
		return nil, err
	}
	src := identitySrc(b.Shape)
	for k, c := range counts {
		n := b.Shape[k]
		m := max(0, n-abs(c))
		src[k] = make([]int, m)
		for j := range m {
			src[k][j] = j + max(c, 0)
		}
	}
	return pick(b, src), nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
