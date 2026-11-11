package apl

// Char is a character element.
type Char rune

// Value is an element of a general array: float64, Char or *Array.
type Value any

// Array is an array of any rank stored in row-major order.
// A numeric array keeps its elements in Data, with Elems nil;
// any other array keeps them in Elems.
type Array struct {
	Shape []int
	Data  []float64 // elements of a numeric array
	Elems []Value   // elements of a general array, non-nil if general
	Proto Value     // prototype of an empty general array
}

// Scalar returns a numeric rank-0 array.
func Scalar(v float64) *Array { return &Array{Data: []float64{v}} }

// Vector returns a numeric rank-1 array holding v.
func Vector(v []float64) *Array { return &Array{Shape: []int{len(v)}, Data: v} }

// General returns an array of elems, numeric if they all are numbers.
// proto is the prototype used when elems is empty.
func General(shape []int, elems []Value, proto Value) *Array {
	nums := make([]float64, len(elems))
	for i, e := range elems {
		v, ok := e.(float64)
		if !ok {
			return &Array{Shape: shape, Elems: elems, Proto: prototype(proto)}
		}
		nums[i] = v
	}
	if len(elems) == 0 && proto != nil {
		if _, ok := proto.(float64); !ok {
			return &Array{Shape: shape, Elems: []Value{}, Proto: prototype(proto)}
		}
	}
	return &Array{Shape: shape, Data: nums}
}

// Rank returns the number of axes.
func (a *Array) Rank() int { return len(a.Shape) }

// Size returns the number of elements.
func (a *Array) Size() int {
	if a.Elems != nil {
		return len(a.Elems)
	}
	return len(a.Data)
}

// IsScalar reports whether a has rank 0.
func (a *Array) IsScalar() bool { return len(a.Shape) == 0 }

// IsNumeric reports whether all elements of a are numbers.
func (a *Array) IsNumeric() bool { return a.Elems == nil }

// At returns element i.
func (a *Array) At(i int) Value {
	if a.Elems != nil {
		return a.Elems[i]
	}
	return a.Data[i]
}

// Fill returns the prototype of a: the fill element for take, expand
// and reshape of an empty array.
func (a *Array) Fill() Value {
	switch {
	case a.Elems == nil:
		return 0.0
	case len(a.Elems) > 0:
		return prototype(a.Elems[0])
	}
	return a.Proto
}

// prototype maps numbers to 0 and characters to blank, recursively.
func prototype(v Value) Value {
	switch v := v.(type) {
	case Char:
		return Char(' ')
	case *Array:
		if v.Elems == nil {
			return &Array{Shape: v.Shape, Data: make([]float64, len(v.Data))}
		}
		e := make([]Value, len(v.Elems))
		for i, x := range v.Elems {
			e[i] = prototype(x)
		}
		return &Array{Shape: v.Shape, Elems: e, Proto: v.Proto}
	}
	return 0.0
}

// gather returns an array of shape whose elements are a's at pos;
// a negative position gives the fill element.
func (a *Array) gather(shape []int, pos []int) *Array {
	if a.Elems == nil {
		r := make([]float64, len(pos))
		for i, p := range pos {
			if p >= 0 {
				r[i] = a.Data[p]
			}
		}
		return &Array{Shape: shape, Data: r}
	}
	fill := a.Fill()
	r := make([]Value, len(pos))
	for i, p := range pos {
		if p >= 0 {
			r[i] = a.Elems[p]
		} else {
			r[i] = fill
		}
	}
	return General(shape, r, fill)
}

// elems returns the elements of a as values.
func (a *Array) elems() []Value {
	if a.Elems != nil {
		return a.Elems
	}
	r := make([]Value, len(a.Data))
	for i, v := range a.Data {
		r[i] = v
	}
	return r
}

// numeric returns a's numbers, or DOMAIN ERROR for a general array.
func (a *Array) numeric() ([]float64, error) {
	if a.Elems != nil {
		return nil, newError(DomainError)
	}
	return a.Data, nil
}

// sameShape reports whether a and b have identical shapes.
func sameShape(a, b *Array) bool { return equalShape(a.Shape, b.Shape) }

func equalShape(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// shapeSize returns the product of dimensions.
func shapeSize(shape []int) int {
	n := 1
	for _, d := range shape {
		n *= d
	}
	return n
}

// axisSplit views shape as frames along axis k: element (i, j, l) with
// i < outer, j < n, l < inner is at offset (i*n+j)*inner+l.
func axisSplit(shape []int, k int) (outer, n, inner int) {
	return shapeSize(shape[:k]), shape[k], shapeSize(shape[k+1:])
}

// match reports whether two elements are equal, with tolerance ct for
// numbers and recursively for nested arrays.
func match(x, y Value, ct float64) bool {
	switch x := x.(type) {
	case float64:
		y, ok := y.(float64)
		return ok && tolEq(x, y, ct)
	case Char:
		y, ok := y.(Char)
		return ok && x == y
	case *Array:
		y, ok := y.(*Array)
		if !ok || !sameShape(x, y) {
			return false
		}
		for i := range x.Size() {
			if !match(x.At(i), y.At(i), ct) {
				return false
			}
		}
		return true
	}
	return false
}
