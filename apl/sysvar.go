package apl

import "math"

// env holds the system variables that primitives depend on.
type env struct {
	io int     // ⎕IO index origin
	ct float64 // ⎕CT comparison tolerance
	pp int     // ⎕PP print precision
	pw int     // ⎕PW print width
	rl int64   // ⎕RL random link
}

// Defaults of the system variables, as in APL2.
const (
	defaultCT = 1e-13
	defaultPP = 10
	defaultPW = 80
	minPW     = 30
	maxPW     = 1000
	defaultRL = 16807
	maxPP     = 17
	maxRL     = 1<<31 - 2
)

func defaultEnv() env {
	return env{io: 1, ct: defaultCT, pp: defaultPP, pw: defaultPW, rl: defaultRL}
}

// sysVars lists the system variables.
var sysVars = map[string]bool{"⎕IO": true, "⎕CT": true, "⎕PP": true, "⎕PW": true, "⎕RL": true,
	"⎕LC": true, "⎕EM": true, "⎕ET": true}

// readOnly system variables describe the interpreter's state.
var readOnly = map[string]bool{"⎕LC": true, "⎕EM": true, "⎕ET": true}

func isSysName(name string) bool { return len(name) > 0 && []rune(name)[0] == '⎕' }

func (e *env) get(name string) *Array {
	switch name {
	case "⎕IO":
		return Scalar(float64(e.io))
	case "⎕CT":
		return Scalar(e.ct)
	case "⎕PP":
		return Scalar(float64(e.pp))
	case "⎕PW":
		return Scalar(float64(e.pw))
	}
	return Scalar(float64(e.rl))
}

// set assigns a system variable, checking its domain.
func (e *env) set(name string, v *Array) error {
	if v.Size() != 1 || v.Rank() > 1 {
		return newError(LengthError)
	}
	d, err := v.numeric()
	if err != nil {
		return err
	}
	x := d[0]
	n, isInt := toInt(x)
	switch name {
	case "⎕IO":
		if !isInt || n < 0 || n > 1 {
			return newError(DomainError)
		}
		e.io = n
	case "⎕CT":
		if x < 0 || x >= 1 {
			return newError(DomainError)
		}
		e.ct = x
	case "⎕PP":
		if !isInt || n < 1 {
			return newError(DomainError)
		}
		e.pp = min(n, maxPP)
	case "⎕PW":
		if !isInt || n < minPW || n > maxPW {
			return newError(DomainError)
		}
		e.pw = n
	case "⎕RL":
		if !isInt || n < 1 || n > maxRL {
			return newError(DomainError)
		}
		e.rl = int64(n)
	}
	return nil
}

// toInt converts a near-integer to int.
func toInt(v float64) (int, bool) {
	r := math.Round(v)
	if !tolEq(v, r, defaultCT) || math.Abs(r) > math.MaxInt32 {
		return 0, false
	}
	return int(r), true
}
