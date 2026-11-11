package apl

import "testing"

// chars returns a character vector.
func chars(s string) *Array {
	var e []Value
	for _, c := range s {
		e = append(e, Char(c))
	}
	return General([]int{len(e)}, e, Char(' '))
}

func TestGeneral(t *testing.T) {
	if a := General([]int{2}, []Value{1.0, 2.0}, nil); !a.IsNumeric() || a.Data[1] != 2 {
		t.Errorf("all numbers: %+v", a)
	}
	if a := General([]int{0}, nil, nil); !a.IsNumeric() || a.Fill() != 0.0 {
		t.Errorf("empty numeric: %+v", a)
	}
	a := General([]int{0}, nil, Char('x'))
	if a.IsNumeric() || a.Fill() != Char(' ') || a.Size() != 0 {
		t.Errorf("empty character: %+v", a)
	}
	m := General([]int{2}, []Value{1.0, Char('a')}, nil)
	if m.IsNumeric() || m.At(0) != 1.0 || m.Fill() != 0.0 {
		t.Errorf("mixed: %+v", m)
	}
	n := General([]int{2}, []Value{Vector([]float64{1, 2}), chars("ab")}, nil)
	p, ok := n.Fill().(*Array)
	if !ok || !match(p, Vector([]float64{0, 0})) {
		t.Errorf("nested prototype: %v", n.Fill())
	}
	if p := prototype(chars("ab")).(*Array); !match(p, chars("  ")) {
		t.Errorf("character array prototype: %v", p)
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		x, y Value
		want bool
	}{
		{1.0, 1.0 + 1e-15, true},
		{1.0, 2.0, false},
		{Char('a'), Char('a'), true},
		{Char('a'), 97.0, false},
		{chars("ab"), chars("ab"), true},
		{chars("ab"), chars("abc"), false},
		{chars("ab"), Char('a'), false},
		{Vector([]float64{1, 2}), Vector([]float64{1, 3}), false},
		{nil, nil, false},
	}
	for _, c := range cases {
		if got := match(c.x, c.y); got != c.want {
			t.Errorf("match(%v, %v) = %v", c.x, c.y, got)
		}
	}
}

func TestAxisSplit(t *testing.T) {
	o, n, i := axisSplit([]int{2, 3, 4, 5}, 1)
	if o != 2 || n != 3 || i != 20 {
		t.Errorf("got %d %d %d", o, n, i)
	}
	x := &Array{Shape: []int{2, 3}, Data: []float64{1, 2, 3, 4, 5, 6}}
	r, err := reduceAxis("+", x, 0)
	if err != nil || Format(r) != "5 7 9" {
		t.Errorf("+ along axis 0: %v %v", r, err)
	}
	r, err = reduceAxis("-", x, 1)
	if err != nil || Format(r) != "2 5" {
		t.Errorf("- along axis 1: %v %v", r, err)
	}
}

// Structural functions work on any element type.
func TestGeneralStructural(t *testing.T) {
	abc := chars("abc")
	check := func(name string, r *Array, err error, want string) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", name, err)
		} else if got := Format(r); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	r, err := reshape(Vector([]float64{2, 2}), abc)
	check("2 2⍴'abc'", r, err, "a b\nc a")
	r, err = reshape(Scalar(3), chars(""))
	check("3⍴''", r, err, "     ")
	if r.IsNumeric() {
		t.Errorf("3⍴'' is numeric")
	}
	r, err = catenate(Scalar(1), abc)
	check("1,'abc'", r, err, "1 a b c")
	r, err = catenate(chars(""), Vector(nil))
	if err != nil || r.IsNumeric() || r.Fill() != Char(' ') {
		t.Errorf("'',⍳0: %+v", r)
	}
	r, err = ravel(&Array{Shape: []int{1, 3}, Elems: abc.Elems})
	check(",1 3⍴'abc'", r, err, "a b c")
	r, err = index(abc, []*Array{Vector([]float64{3, 1})})
	check("'abc'[3 1]", r, err, "c a")
	r, err = index(General([]int{2}, []Value{Char('a'), 5.0}, nil), []*Array{Scalar(2)})
	if err != nil || !r.IsNumeric() {
		t.Errorf("numeric item of a mixed vector: %+v %v", r, err)
	}
	_, err = index(abc, []*Array{chars("a")})
	if e, ok := err.(*Error); !ok || e.Class != DomainError {
		t.Errorf("character index: %v", err)
	}
	r, err = indexAssign(Vector([]float64{1, 2, 3}), []*Array{Scalar(2)}, Scalar(0).gather(nil, []int{0}))
	check("numeric ← number", r, err, "1 0 3")
	r, err = indexAssign(Vector([]float64{1, 2, 3}), []*Array{Scalar(2)}, chars("x").gather(nil, []int{0}))
	check("numeric ← character", r, err, "1 x 3")
	r, err = indexAssign(abc, []*Array{nil}, Scalar(7))
	check("all ← number", r, err, "7 7 7")
	if !r.IsNumeric() {
		t.Errorf("all-numeric result is general")
	}
	r, err = compress(Vector([]float64{1, 0, 2}), abc)
	check("1 0 2/'abc'", r, err, "a c c")
	r, err = compress(Scalar(2), abc.gather(nil, []int{0}))
	check("2/'a'", r, err, "a a")
	r, err = compress(Scalar(1), Vector(nil))
	check("1/⍳0", r, err, "")
	r, err = indexOf(abc, chars("cz"))
	check("'abc'⍳'cz'", r, err, "3 4")
	r, err = shapeOf(abc)
	check("⍴'abc'", r, err, "3")
}

// Numeric functions reject characters.
func TestGeneralDomain(t *testing.T) {
	abc := chars("abc")
	plus := scalarFns["+"]
	_, e1 := applyMonad(plus.monad, abc)
	_, e2 := applyDyad(plus.dyad, Scalar(1), abc)
	_, e3 := reduce("+", abc)
	_, e4 := outer(plus.dyad, abc, Scalar(1))
	_, e5 := outer(plus.dyad, Scalar(1), abc)
	_, e6 := inner("+", "×", abc, abc)
	_, e7 := inner("+", "×", Scalar(1), abc)
	_, e8 := interval(chars("a"))
	_, e9 := reshape(chars("a"), Scalar(1))
	for i, err := range []error{e1, e2, e3, e4, e5, e6, e7, e8, e9} {
		if e, ok := err.(*Error); !ok || e.Class != DomainError {
			t.Errorf("case %d: got %v", i+1, err)
		}
	}
}

func TestFormatGeneral(t *testing.T) {
	m := General([]int{2, 2}, []Value{Char('a'), 10.0, Char('b'), Char('c')}, nil)
	if got := Format(m); got != "a 10\nb  c" {
		t.Errorf("got %q", got)
	}
	n := General(nil, []Value{chars("ab")}, nil)
	if got := Format(n); got != "a b" {
		t.Errorf("nested: %q", got)
	}
	if got := formatElem(nil); got != "?" {
		t.Errorf("unknown element: %q", got)
	}
}
