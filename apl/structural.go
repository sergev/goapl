package apl

import "math"

type structFn struct {
	monad func(x *Array) (*Array, error)
	dyad  func(a, b *Array) (*Array, error)
}

var structFns = map[string]*structFn{
	"⍳": {monad: interval, dyad: indexOf},
	"⍴": {monad: shapeOf, dyad: reshape},
	",": {monad: ravel, dyad: catenate},
	"/": {dyad: compress},
}

// toInt converts a near-integer to int.
func toInt(v float64) (int, bool) {
	r := math.Round(v)
	if !tolEq(v, r) || math.Abs(r) > math.MaxInt32 {
		return 0, false
	}
	return int(r), true
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

// interval is ⍳N: 1 2 ... N.
func interval(x *Array) (*Array, error) {
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
		r[i] = float64(i + 1)
	}
	return Vector(r), nil
}

func shapeOf(x *Array) (*Array, error) {
	r := make([]float64, x.Rank())
	for i, d := range x.Shape {
		r[i] = float64(d)
	}
	return Vector(r), nil
}

func ravel(x *Array) (*Array, error) {
	return &Array{Shape: []int{x.Size()}, Data: x.Data, Elems: x.Elems, Proto: x.Proto}, nil
}

// indexOf is A⍳B: position of each B in vector A, or 1+⍴A if absent.
func indexOf(a, b *Array) (*Array, error) {
	if a.Rank() != 1 {
		return nil, newError(RankError)
	}
	r := &Array{Shape: b.Shape, Data: make([]float64, b.Size())}
	for i := range r.Data {
		v := b.At(i)
		r.Data[i] = float64(a.Size() + 1)
		for j := range a.Size() {
			if match(v, a.At(j)) {
				r.Data[i] = float64(j + 1)
				break
			}
		}
	}
	return r, nil
}

// reshape is A⍴B, cycling B; an empty B gives its fill element.
func reshape(a, b *Array) (*Array, error) {
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

// catenate is A,B for scalars and vectors.
func catenate(a, b *Array) (*Array, error) {
	if a.Rank() > 1 || b.Rank() > 1 {
		return nil, newError(RankError)
	}
	n := a.Size() + b.Size()
	if a.IsNumeric() && b.IsNumeric() {
		r := make([]float64, 0, n)
		r = append(r, a.Data...)
		return Vector(append(r, b.Data...)), nil
	}
	r := make([]Value, 0, n)
	r = append(r, a.elems()...)
	return General([]int{n}, append(r, b.elems()...), a.Fill()), nil
}

// indexPositions returns flat offsets selected by idx (nil = whole axis)
// and the shape of the selection.
func indexPositions(shape []int, idx []*Array) ([]int, []int, error) {
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
				if n < 1 || n > shape[k] {
					return nil, nil, newError(IndexError)
				}
				sel[i] = n - 1
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
func index(x *Array, idx []*Array) (*Array, error) {
	pos, shape, err := indexPositions(x.Shape, idx)
	if err != nil {
		return nil, err
	}
	return x.gather(shape, pos), nil
}

// indexAssign returns a copy of x with X[I;J;...]←v applied.
func indexAssign(x *Array, idx []*Array, v *Array) (*Array, error) {
	pos, shape, err := indexPositions(x.Shape, idx)
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

// compress is A/B along the last axis; A gives a repeat count per item.
func compress(a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	if b.IsScalar() {
		b = &Array{Shape: []int{1}, Data: b.Data, Elems: b.Elems}
	}
	k := b.Rank() - 1
	outer, n, _ := axisSplit(b.Shape, k)
	counts, err := toInts(a)
	if err != nil {
		return nil, err
	}
	if len(counts) == 1 && n != 1 {
		c := counts[0]
		counts = make([]int, n)
		for i := range counts {
			counts[i] = c
		}
	} else if len(counts) != n {
		return nil, newError(LengthError)
	}
	total := 0
	for _, c := range counts {
		total += c
	}
	shape := append(append([]int(nil), b.Shape[:k]...), total)
	pos := make([]int, 0, outer*total)
	for i := range outer {
		for j, c := range counts {
			for range c {
				pos = append(pos, i*n+j)
			}
		}
	}
	return b.gather(shape, pos), nil
}
