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
}

// toInt converts a near-integer to int.
func toInt(v float64) (int, bool) {
	r := math.Round(v)
	if !tolEq(v, r) || math.Abs(r) > math.MaxInt32 {
		return 0, false
	}
	return int(r), true
}

// interval is ⍳N: 1 2 ... N.
func interval(x *Array) (*Array, error) {
	if x.Size() != 1 || x.Rank() > 1 {
		return nil, newError(21)
	}
	if x.Data[0] < 0 {
		return nil, newError(22)
	}
	n, ok := toInt(x.Data[0])
	if !ok {
		return nil, newError(23)
	}
	r := make([]float64, n)
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

func ravel(x *Array) (*Array, error) { return Vector(x.Data), nil }

// indexOf is A⍳B: position of each B in vector A, or 1+⍴A if absent.
func indexOf(a, b *Array) (*Array, error) {
	if a.Rank() != 1 {
		return nil, newError(29)
	}
	r := &Array{Shape: b.Shape, Data: make([]float64, b.Size())}
	for i, v := range b.Data {
		r.Data[i] = float64(a.Size() + 1)
		for j, w := range a.Data {
			if tolEq(v, w) {
				r.Data[i] = float64(j + 1)
				break
			}
		}
	}
	return r, nil
}

// reshape is A⍴B, cycling B; an empty B fills with zeros.
func reshape(a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(56)
	}
	shape := make([]int, a.Size())
	for i, v := range a.Data {
		if v < 0 {
			return nil, newError(22)
		}
		n, ok := toInt(v)
		if !ok {
			return nil, newError(23)
		}
		shape[i] = n
	}
	if len(shape) > MaxRank {
		return nil, newError(60)
	}
	r := &Array{Shape: shape, Data: make([]float64, shapeSize(shape))}
	if b.Size() > 0 {
		for i := range r.Data {
			r.Data[i] = b.Data[i%b.Size()]
		}
	}
	return r, nil
}

// catenate is A,B for scalars and vectors.
func catenate(a, b *Array) (*Array, error) {
	if a.Rank() > 1 || b.Rank() > 1 {
		return nil, newError(53)
	}
	r := make([]float64, 0, a.Size()+b.Size())
	r = append(r, a.Data...)
	return Vector(append(r, b.Data...)), nil
}

// indexPositions returns flat offsets selected by idx (nil = whole axis)
// and the shape of the selection.
func indexPositions(shape []int, idx []*Array) ([]int, []int, error) {
	if len(idx) != len(shape) {
		return nil, nil, newError(35)
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
			sel = make([]int, ix.Size())
			for i, v := range ix.Data {
				n, ok := toInt(v)
				if !ok {
					return nil, nil, newError(37)
				}
				if n < 1 || n > shape[k] {
					return nil, nil, newError(38)
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
	if len(shape) > MaxRank {
		return nil, newError(60)
	}
	r := &Array{Shape: shape, Data: make([]float64, len(pos))}
	for i, p := range pos {
		r.Data[i] = x.Data[p]
	}
	return r, nil
}

// indexAssign returns a copy of x with X[I;J;...]←v applied.
func indexAssign(x *Array, idx []*Array, v *Array) (*Array, error) {
	pos, shape, err := indexPositions(x.Shape, idx)
	if err != nil {
		return nil, err
	}
	if v.Size() != 1 && !sameShape(v, &Array{Shape: shape}) {
		return nil, newError(36)
	}
	r := &Array{Shape: x.Shape, Data: append([]float64(nil), x.Data...)}
	for i, p := range pos {
		if v.Size() == 1 {
			r.Data[p] = v.Data[0]
		} else {
			r.Data[p] = v.Data[i]
		}
	}
	return r, nil
}
