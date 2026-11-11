package apl

import "math"

type monadFn func(e *env, x float64) (float64, error)
type dyadFn func(e *env, a, b float64) (float64, error)

// scalarFn is a pervasive primitive; nil entries mean no such valence.
// dyad applies to numbers; any, if set, to other elements.
type scalarFn struct {
	monad monadFn
	dyad  dyadFn
	any   func(e *env, a, b Value) Value
}

var scalarFns = map[string]*scalarFn{
	"+": {
		monad: func(_ *env, x float64) (float64, error) { return x, nil },
		dyad:  func(_ *env, a, b float64) (float64, error) { return a + b, nil },
	},
	"-": {
		monad: func(_ *env, x float64) (float64, error) { return -x, nil },
		dyad:  func(_ *env, a, b float64) (float64, error) { return a - b, nil },
	},
	"×": {
		monad: signum,
		dyad:  func(_ *env, a, b float64) (float64, error) { return a * b, nil },
	},
	"÷": {
		monad: func(_ *env, x float64) (float64, error) {
			if x == 0 {
				return 0, newError(DomainError)
			}
			return 1 / x, nil
		},
		dyad: func(_ *env, a, b float64) (float64, error) {
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
		monad: func(_ *env, x float64) (float64, error) { return math.Exp(x), nil },
		dyad:  power,
	},
	"~": {
		monad: func(_ *env, x float64) (float64, error) {
			if !isBool(x) {
				return 0, newError(DomainError)
			}
			return 1 - x, nil
		},
	},
	"⌈": {
		monad: func(e *env, x float64) (float64, error) { return -floor(-x, e.ct), nil },
		dyad:  func(_ *env, a, b float64) (float64, error) { return math.Max(a, b), nil },
	},
	"⌊": {
		monad: func(e *env, x float64) (float64, error) { return floor(x, e.ct), nil },
		dyad:  func(_ *env, a, b float64) (float64, error) { return math.Min(a, b), nil },
	},
	"|": {
		monad: func(_ *env, x float64) (float64, error) { return math.Abs(x), nil },
		dyad:  residue,
	},
	"⍟": {
		monad: func(_ *env, x float64) (float64, error) {
			if x <= 0 {
				return 0, newError(DomainError)
			}
			return math.Log(x), nil
		},
		dyad: logBase,
	},
	"=": {
		dyad: compare(func(a, b, ct float64) bool { return tolEq(a, b, ct) }),
		any:  func(e *env, a, b Value) Value { return boolVal(match(a, b, e.ct)) },
	},
	"≠": {
		dyad: compare(func(a, b, ct float64) bool { return !tolEq(a, b, ct) }),
		any:  func(e *env, a, b Value) Value { return boolVal(!match(a, b, e.ct)) },
	},
	"<": {dyad: compare(func(a, b, ct float64) bool { return a < b && !tolEq(a, b, ct) })},
	"≤": {dyad: compare(func(a, b, ct float64) bool { return a < b || tolEq(a, b, ct) })},
	"≥": {dyad: compare(func(a, b, ct float64) bool { return a > b || tolEq(a, b, ct) })},
	">": {dyad: compare(func(a, b, ct float64) bool { return a > b && !tolEq(a, b, ct) })},
	"∧": {dyad: logical(func(a, b bool) bool { return a && b })},
	"∨": {dyad: logical(func(a, b bool) bool { return a || b })},
}

// identity holds identity elements for reducing an empty axis.
var identity = map[string]float64{
	"+": 0, "-": 0, "×": 1, "÷": 1, "*": 1, "|": 0,
	"⌈": -math.MaxFloat64, "⌊": math.MaxFloat64,
	"=": 1, "≠": 0, "<": 0, "≤": 1, "≥": 1, ">": 0, "∧": 1, "∨": 0,
}

func signum(_ *env, x float64) (float64, error) {
	switch {
	case x > 0:
		return 1, nil
	case x < 0:
		return -1, nil
	}
	return 0, nil
}

func power(_ *env, a, b float64) (float64, error) {
	r := math.Pow(a, b)
	if math.IsNaN(r) {
		return 0, newError(DomainError)
	}
	if math.IsInf(r, 0) && a == 0 {
		return 0, newError(DomainError)
	}
	return r, nil
}

// floor is the tolerant floor: x rounds up to an integer within
// relative tolerance ct.
func floor(x, ct float64) float64 {
	n := math.Round(x)
	if n > x && n-x > ct*math.Max(1, math.Abs(x)) {
		n--
	}
	return n
}

// residue is A|B: B modulo A with the sign of A; 0 if B÷A is a
// near-integer.
func residue(e *env, a, b float64) (float64, error) {
	if a == 0 {
		return b, nil
	}
	if q := b / a; tolEq(q, math.Round(q), e.ct) {
		return 0, nil
	}
	r := math.Mod(b, a)
	if r != 0 && (r < 0) != (a < 0) {
		r += a
	}
	return r, nil
}

// logBase is A⍟B: log of B to base A.
func logBase(_ *env, a, b float64) (float64, error) {
	if a*b < 0 {
		return 0, newError(DomainError)
	}
	la := math.Log(math.Abs(a))
	if la == 0 {
		return 0, newError(DomainError)
	}
	return math.Log(math.Abs(b)) / la, nil
}

func compare(f func(a, b, ct float64) bool) dyadFn {
	return func(e *env, a, b float64) (float64, error) { return boolVal(f(a, b, e.ct)), nil }
}

func logical(f func(a, b bool) bool) dyadFn {
	return func(_ *env, a, b float64) (float64, error) {
		if !isBool(a) || !isBool(b) {
			return 0, newError(DomainError)
		}
		return boolVal(f(a == 1, b == 1)), nil
	}
}

// tolEq is tolerant equality with relative tolerance ct.
func tolEq(a, b, ct float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) <= ct*math.Max(math.Abs(a), math.Abs(b))
}

func isBool(x float64) bool { return x == 0 || x == 1 }

func boolVal(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

// applyMonad applies f to every element.
func applyMonad(e *env, f monadFn, x *Array) (*Array, error) {
	d, err := x.numeric()
	if err != nil {
		return nil, err
	}
	r := &Array{Shape: x.Shape, Data: make([]float64, len(d))}
	for i, v := range d {
		y, err := f(e, v)
		if err != nil {
			return nil, err
		}
		r.Data[i] = y
	}
	return r, nil
}

// elem applies the dyadic form to two elements of any type.
func (sf *scalarFn) elem(e *env, x, y Value) (Value, error) {
	a, ok1 := x.(float64)
	b, ok2 := y.(float64)
	switch {
	case ok1 && ok2:
		return sf.dyad(e, a, b)
	case sf.any != nil:
		return sf.any(e, x, y), nil
	}
	return nil, newError(DomainError)
}

// applyDyad applies sf elementwise with scalar (or one-element) extension.
func applyDyad(e *env, sf *scalarFn, a, b *Array) (*Array, error) {
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
	ai := func(i int) int {
		if a.Size() == 1 {
			return 0
		}
		return i
	}
	bi := func(i int) int {
		if b.Size() == 1 {
			return 0
		}
		return i
	}
	if a.IsNumeric() && b.IsNumeric() {
		r := &Array{Shape: shape, Data: make([]float64, n)}
		for i := range n {
			v, err := sf.dyad(e, a.Data[ai(i)], b.Data[bi(i)])
			if err != nil {
				return nil, err
			}
			r.Data[i] = v
		}
		return r, nil
	}
	r := make([]Value, n)
	for i := range n {
		v, err := sf.elem(e, a.At(ai(i)), b.At(bi(i)))
		if err != nil {
			return nil, err
		}
		r[i] = v
	}
	return General(shape, r, nil), nil
}
