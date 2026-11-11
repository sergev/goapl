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
	"⍲": {dyad: logical(func(a, b bool) bool { return !(a && b) })},
	"⍱": {dyad: logical(func(a, b bool) bool { return !(a || b) })},
	"○": {
		monad: func(_ *env, x float64) (float64, error) { return math.Pi * x, nil },
		dyad:  circle,
	},
	"!": {monad: factorial, dyad: binomial},
	"?": {monad: roll},
}

// identity holds identity elements for reducing an empty axis.
var identity = map[string]float64{
	"+": 0, "-": 0, "×": 1, "÷": 1, "*": 1, "|": 0,
	"⌈": -math.MaxFloat64, "⌊": math.MaxFloat64, "!": 1,
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

// circle is A○B: the circular, hyperbolic and Pythagorean functions.
func circle(_ *env, a, b float64) (float64, error) {
	k, ok := toInt(a)
	if !ok || k < -7 || k > 7 {
		return 0, newError(DomainError)
	}
	var r float64
	switch k {
	case 0:
		r = math.Sqrt(1 - b*b)
	case 1:
		r = math.Sin(b)
	case 2:
		r = math.Cos(b)
	case 3:
		r = math.Tan(b)
	case 4:
		r = math.Sqrt(1 + b*b)
	case 5:
		r = math.Sinh(b)
	case 6:
		r = math.Cosh(b)
	case 7:
		r = math.Tanh(b)
	case -1:
		r = math.Asin(b)
	case -2:
		r = math.Acos(b)
	case -3:
		r = math.Atan(b)
	case -4:
		r = math.Copysign(math.Sqrt(b*b-1), b)
		if math.Abs(b) < 1 {
			r = math.NaN()
		}
	case -5:
		r = math.Asinh(b)
	case -6:
		r = math.Acosh(b)
	case -7:
		r = math.Atanh(b)
	}
	if math.IsNaN(r) || math.IsInf(r, 0) {
		return 0, newError(DomainError)
	}
	return r, nil
}

// gamma is Γ(x), or DOMAIN ERROR at a pole or on overflow.
func gamma(x float64) (float64, error) {
	if n, ok := toInt(x); ok && n <= 0 {
		return 0, newError(DomainError)
	}
	r := math.Gamma(x)
	if math.IsNaN(r) || math.IsInf(r, 0) {
		return 0, newError(DomainError)
	}
	return r, nil
}

// factorial is !X: Γ(X+1).
func factorial(_ *env, x float64) (float64, error) {
	if n, ok := toInt(x); ok && n >= 0 && n <= 170 {
		r := 1.0
		for i := 2; i <= n; i++ {
			r *= float64(i)
		}
		return r, nil
	}
	return gamma(x + 1)
}

// binomial is A!B: the number of ways to choose A of B, extended to
// negative integers as in APL2 and to other numbers by Γ.
func binomial(e *env, a, b float64) (float64, error) {
	ia, aok := toInt(a)
	ib, bok := toInt(b)
	if !aok || !bok {
		ga, err := gamma(a + 1)
		if err != nil {
			return 0, err
		}
		gb, err := gamma(b + 1)
		if err != nil {
			return 0, err
		}
		gc, err := gamma(b - a + 1)
		if err != nil {
			return 0, err
		}
		return gb / (ga * gc), nil
	}
	switch {
	case ia >= 0 && ib >= 0:
		if ia > ib {
			return 0, nil
		}
		r := 1.0
		for i := 1; i <= min(ia, ib-ia); i++ {
			r = r * float64(ib-min(ia, ib-ia)+i) / float64(i)
		}
		return math.Round(r), nil
	case ia >= 0: // B < 0
		r, err := binomial(e, a, a-b-1)
		return r * sign(ia), err
	case ib >= 0: // A < 0 ≤ B
		return 0, nil
	case ib >= ia: // A ≤ B < 0
		r, err := binomial(e, float64(-ib-1), float64(-ia-1))
		return r * sign(ib-ia), err
	}
	return 0, nil
}

// sign is ¯1*n.
func sign(n int) float64 {
	if n%2 != 0 {
		return -1
	}
	return 1
}

// Random numbers: the Lehmer generator of APL2, seeded by ⎕RL.
const rlModulus = 1<<31 - 1

// random advances ⎕RL and returns a number in [0, 1).
func (e *env) random() float64 {
	e.rl = e.rl * 16807 % rlModulus
	return float64(e.rl-1) / float64(rlModulus-1)
}

// roll is ?N: a random integer among the first N indices.
func roll(e *env, x float64) (float64, error) {
	n, ok := toInt(x)
	if !ok || n < 1 {
		return 0, newError(DomainError)
	}
	return float64(e.io + int(e.random()*float64(n))), nil
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
