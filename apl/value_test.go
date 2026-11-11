package apl

import "testing"

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
	n := General([]int{2}, []Value{Vector([]float64{1, 2}), Chars("ab")}, nil)
	p, ok := n.Fill().(*Array)
	if !ok || !match(p, Vector([]float64{0, 0}), defaultCT) {
		t.Errorf("nested prototype: %v", n.Fill())
	}
	if p := prototype(Chars("ab")).(*Array); !match(p, Chars("  "), defaultCT) {
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
		{Chars("ab"), Chars("ab"), true},
		{Chars("ab"), Chars("abc"), false},
		{Chars("ab"), Char('a'), false},
		{Vector([]float64{1, 2}), Vector([]float64{1, 3}), false},
		{nil, nil, false},
	}
	for _, c := range cases {
		if got := match(c.x, c.y, defaultCT); got != c.want {
			t.Errorf("match(%v, %v) = %v", c.x, c.y, got)
		}
	}
}

func TestAxisSplit(t *testing.T) {
	e := &env{io: 1, ct: defaultCT}
	o, n, i := axisSplit([]int{2, 3, 4, 5}, 1)
	if o != 2 || n != 3 || i != 20 {
		t.Errorf("got %d %d %d", o, n, i)
	}
	x := &Array{Shape: []int{2, 3}, Data: []float64{1, 2, 3, 4, 5, 6}}
	r, err := reduceAxis(e, "+", x, 0)
	if err != nil || Format(r) != "5 7 9" {
		t.Errorf("+ along axis 0: %v %v", r, err)
	}
	r, err = reduceAxis(e, "-", x, 1)
	if err != nil || Format(r) != "2 5" {
		t.Errorf("- along axis 1: %v %v", r, err)
	}
}

// Structural functions work on any element type.
func TestGeneralStructural(t *testing.T) {
	e := &env{io: 1, ct: defaultCT}
	abc := Chars("abc")
	check := func(name string, r *Array, err error, want string) {
		t.Helper()
		if err != nil {
			t.Errorf("%s: %v", name, err)
		} else if got := Format(r); got != want {
			t.Errorf("%s = %q, want %q", name, got, want)
		}
	}
	r, err := reshape(e, Vector([]float64{2, 2}), abc)
	check("2 2⍴'abc'", r, err, "ab\nca")
	r, err = reshape(e, Scalar(3), Chars(""))
	check("3⍴''", r, err, "   ")
	if r.IsNumeric() {
		t.Errorf("3⍴'' is numeric")
	}
	r, err = catenate(e, nil, Scalar(1), abc, false)
	check("1,'abc'", r, err, "1 abc")
	r, err = catenate(e, nil, Chars(""), Vector(nil), false)
	if err != nil || r.IsNumeric() || r.Fill() != Char(' ') {
		t.Errorf("'',⍳0: %+v", r)
	}
	r, err = ravel(e, &Array{Shape: []int{1, 3}, Elems: abc.Elems})
	check(",1 3⍴'abc'", r, err, "abc")
	r, err = index(e, abc, []*Array{Vector([]float64{3, 1})})
	check("'abc'[3 1]", r, err, "ca")
	r, err = index(e, General([]int{2}, []Value{Char('a'), 5.0}, nil), []*Array{Scalar(2)})
	if err != nil || !r.IsNumeric() {
		t.Errorf("numeric item of a mixed vector: %+v %v", r, err)
	}
	_, err = index(e, abc, []*Array{Chars("a")})
	if e, ok := err.(*Error); !ok || e.Class != DomainError {
		t.Errorf("character index: %v", err)
	}
	r, err = indexAssign(e, Vector([]float64{1, 2, 3}), []*Array{Scalar(2)}, Scalar(0).gather(nil, []int{0}))
	check("numeric ← number", r, err, "1 0 3")
	r, err = indexAssign(e, Vector([]float64{1, 2, 3}), []*Array{Scalar(2)}, Chars("x").gather(nil, []int{0}))
	check("numeric ← character", r, err, "1 x 3")
	r, err = indexAssign(e, abc, []*Array{nil}, Scalar(7))
	check("all ← number", r, err, "7 7 7")
	if !r.IsNumeric() {
		t.Errorf("all-numeric result is general")
	}
	r, err = compress(e, Vector([]float64{1, 0, 2}), abc, 0)
	check("1 0 2/'abc'", r, err, "acc")
	r, err = compress(e, Scalar(2), abc.gather([]int{1}, []int{0}), 0)
	check("2/'a'", r, err, "aa")
	r, err = compress(e, Scalar(1), Vector(nil), 0)
	check("1/⍳0", r, err, "")
	r, err = indexOf(e, abc, Chars("cz"))
	check("'abc'⍳'cz'", r, err, "3 4")
	r, err = shapeOf(e, abc)
	check("⍴'abc'", r, err, "3")
}

// Numeric functions reject characters.
func TestGeneralDomain(t *testing.T) {
	e := &env{io: 1, ct: defaultCT}
	abc := Chars("abc")
	plus := scalarFns["+"]
	_, e1 := applyMonad(e, plus.monad, abc)
	_, e2 := applyDyad(e, plus, Scalar(1), abc)
	_, e3 := reduce(e, "+", abc)
	_, e4 := outer(e, plus, abc, Scalar(1))
	_, e5 := outer(e, plus, Scalar(1), abc)
	_, e6 := inner(e, "+", "×", abc, abc)
	_, e7 := inner(e, "+", "×", Scalar(1), abc)
	_, e8 := interval(e, Chars("a"))
	_, e9 := reshape(e, Chars("a"), Scalar(1))
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
	n := General(nil, []Value{Chars("ab")}, nil)
	if got := Format(n); got != " ab" {
		t.Errorf("nested: %q", got)
	}
	if got := cellLines(nil, defaultPP)[0]; got != "?" {
		t.Errorf("unknown element: %q", got)
	}
}

func TestWrap(t *testing.T) {
	cases := []struct {
		in    string
		width int
		want  string
	}{
		{"1 2 3", 30, "1 2 3"},
		{"aaaa bbbb cccc", 10, "aaaa bbbb\n      cccc"},
		{"abcdefghij", 8, "abcdefgh\n      ij"},
		{"abcdefghij", 4, "abcd\n      e\n      f\n      g\n      h\n      i\n      j"},
		{"12 34\n56 78", 8, "12 34\n56 78"},
		{"12 34 56", 6, "12 34\n      5\n      6"},
	}
	for _, c := range cases {
		if got := wrap(c.in, c.width); got != c.want {
			t.Errorf("wrap(%q, %d) = %q, want %q", c.in, c.width, got, c.want)
		}
	}
}
