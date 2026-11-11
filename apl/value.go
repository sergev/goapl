package apl

// MaxRank is the prototype's limit on array rank.
const MaxRank = 3

// Array is a numeric array stored in row-major order.
type Array struct {
	Shape []int
	Data  []float64
}

// Scalar returns a rank-0 array.
func Scalar(v float64) *Array { return &Array{Data: []float64{v}} }

// Vector returns a rank-1 array holding v.
func Vector(v []float64) *Array { return &Array{Shape: []int{len(v)}, Data: v} }

// Rank returns the number of axes.
func (a *Array) Rank() int { return len(a.Shape) }

// Size returns the number of elements.
func (a *Array) Size() int { return len(a.Data) }

// IsScalar reports whether a has rank 0.
func (a *Array) IsScalar() bool { return len(a.Shape) == 0 }

// sameShape reports whether a and b have identical shapes.
func sameShape(a, b *Array) bool {
	if len(a.Shape) != len(b.Shape) {
		return false
	}
	for i := range a.Shape {
		if a.Shape[i] != b.Shape[i] {
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
