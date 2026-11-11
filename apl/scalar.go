package apl

import "math"

// Comparison tolerance (⎕CT).
const compTol = 1e-13

type monadFn func(x float64) (float64, error)
type dyadFn func(a, b float64) (float64, error)

// scalarFn is a pervasive primitive; nil entries mean no such valence.
type scalarFn struct {
	monad monadFn
	dyad  dyadFn
}

var scalarFns = map[string]*scalarFn{
	"+": {
		monad: func(x float64) (float64, error) { return x, nil },
		dyad:  func(a, b float64) (float64, error) { return a + b, nil },
	},
	"-": {
		monad: func(x float64) (float64, error) { return -x, nil },
		dyad:  func(a, b float64) (float64, error) { return a - b, nil },
	},
	"×": {
		monad: signum,
		dyad:  func(a, b float64) (float64, error) { return a * b, nil },
	},
	"÷": {
		monad: func(x float64) (float64, error) {
			if x == 0 {
				return 0, newError(DomainError)
			}
			return 1 / x, nil
		},
		dyad: func(a, b float64) (float64, error) {
			if b == 0 {
				if a == 0 {
					return 1, nil
				}
				return 0, newError(DomainError)
			}
			return a / b, nil
		},
	},
	"*": {
		monad: func(x float64) (float64, error) { return math.Exp(x), nil },
		dyad:  power,
	},
	"~": {
		monad: func(x float64) (float64, error) {
			if !isBool(x) {
				return 0, newError(DomainError)
			}
			return 1 - x, nil
		},
	},
	"⌈": {
		monad: func(x float64) (float64, error) { return math.Ceil(x), nil },
		dyad:  func(a, b float64) (float64, error) { return math.Max(a, b), nil },
	},
	"⌊": {
		monad: func(x float64) (float64, error) { return math.Floor(x), nil },
		dyad:  func(a, b float64) (float64, error) { return math.Min(a, b), nil },
	},
	"|": {
		monad: func(x float64) (float64, error) { return math.Abs(x), nil },
		dyad:  residue,
	},
	"⍟": {
		monad: func(x float64) (float64, error) {
			if x <= 0 {
				return 0, newError(DomainError)
			}
			return math.Log(x), nil
		},
		dyad: logBase,
	},
	"=": {dyad: compare(func(a, b float64) bool { return tolEq(a, b) })},
	"≠": {dyad: compare(func(a, b float64) bool { return !tolEq(a, b) })},
	"<": {dyad: compare(func(a, b float64) bool { return a < b && !tolEq(a, b) })},
	"≤": {dyad: compare(func(a, b float64) bool { return a < b || tolEq(a, b) })},
	"≥": {dyad: compare(func(a, b float64) bool { return a > b || tolEq(a, b) })},
	">": {dyad: compare(func(a, b float64) bool { return a > b && !tolEq(a, b) })},
	"∧": {dyad: logical(func(a, b bool) bool { return a && b })},
	"∨": {dyad: logical(func(a, b bool) bool { return a || b })},
}

// identity holds identity elements for reducing an empty axis.
var identity = map[string]float64{
	"+": 0, "-": 0, "×": 1, "÷": 1, "*": 1, "|": 0,
	"⌈": -math.MaxFloat64, "⌊": math.MaxFloat64,
	"=": 1, "≠": 0, "<": 0, "≤": 1, "≥": 1, ">": 0, "∧": 1, "∨": 0,
}

func signum(x float64) (float64, error) {
	switch {
	case x > 0:
		return 1, nil
	case x < 0:
		return -1, nil
	}
	return 0, nil
}

func power(a, b float64) (float64, error) {
	r := math.Pow(a, b)
	if math.IsNaN(r) {
		return 0, newError(DomainError)
	}
	if math.IsInf(r, 0) && a == 0 {
		return 0, newError(DomainError)
	}
	return r, nil
}

// residue is A|B: B modulo A with the sign of A.
func residue(a, b float64) (float64, error) {
	if a == 0 {
		return b, nil
	}
	r := math.Mod(b, a)
	if r != 0 && (r < 0) != (a < 0) {
		r += a
	}
	return r, nil
}

// logBase is A⍟B: log of B to base A.
func logBase(a, b float64) (float64, error) {
	if a*b < 0 {
		return 0, newError(DomainError)
	}
	la := math.Log(math.Abs(a))
	if la == 0 {
		return 0, newError(DomainError)
	}
	return math.Log(math.Abs(b)) / la, nil
}

func compare(f func(a, b float64) bool) dyadFn {
	return func(a, b float64) (float64, error) { return boolVal(f(a, b)), nil }
}

func logical(f func(a, b bool) bool) dyadFn {
	return func(a, b float64) (float64, error) {
		if !isBool(a) || !isBool(b) {
			return 0, newError(DomainError)
		}
		return boolVal(f(a == 1, b == 1)), nil
	}
}

func tolEq(a, b float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) <= compTol*math.Max(math.Abs(a), math.Abs(b))
}

func isBool(x float64) bool { return x == 0 || x == 1 }

func boolVal(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// applyMonad applies f to every element.
func applyMonad(f monadFn, x *Array) (*Array, error) {
	r := &Array{Shape: x.Shape, Data: make([]float64, len(x.Data))}
	for i, v := range x.Data {
		y, err := f(v)
		if err != nil {
			return nil, err
		}
		r.Data[i] = y
	}
	return r, nil
}

// applyDyad applies f elementwise with scalar (or one-element) extension.
func applyDyad(f dyadFn, a, b *Array) (*Array, error) {
	var shape []int
	switch {
	case sameShape(a, b):
		shape = a.Shape
	case a.IsScalar() || (a.Size() == 1 && !b.IsScalar()):
		shape = b.Shape
	case b.IsScalar() || b.Size() == 1:
		shape = a.Shape
	case a.Rank() != b.Rank():
		return nil, newError(RankError)
	default:
		return nil, newError(LengthError)
	}
	n := shapeSize(shape)
	r := &Array{Shape: shape, Data: make([]float64, n)}
	for i := range n {
		x := a.Data[0]
		if a.Size() > 1 {
			x = a.Data[i]
		}
		y := b.Data[0]
		if b.Size() > 1 {
			y = b.Data[i]
		}
		v, err := f(x, y)
		if err != nil {
			return nil, err
		}
		r.Data[i] = v
	}
	return r, nil
}
