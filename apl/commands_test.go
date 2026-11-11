package apl

import (
	"bytes"
	"testing"
)

// literal gives expressions that recreate their values exactly.
func TestLiteral(t *testing.T) {
	cases := []string{
		"5", "¯2.5E¯20", "1E300", "0.1", "÷3", "'Q'", "''''", "⍳0", "''", ",5", ",'A'",
		"1 2 3", "'IT''S'", "2 3⍴⍳6", "2 0⍴0", "0 3⍴''", "2 2 2⍴'ABCDEFGH'",
		"1 'A' 2.5", "1 (2 3) ('AB' (4 5))", "⊂1 2", "⊂⊂1 2", ",⊂1 2",
		"0⍴⊂1 2", "2 2⍴(1 2) 'X' 3 (⍳0)", "(⊂'AB'),⊂⍳0",
	}
	for _, c := range cases {
		in := New(&bytes.Buffer{})
		if err := in.Exec("V←" + c); err != nil {
			t.Fatalf("%s: %v", c, err)
		}
		v := in.vars["V"]
		lit := literal(v)
		if err := in.Exec("W←" + lit); err != nil {
			t.Errorf("%s: literal %s: %v", c, lit, err)
			continue
		}
		w := in.vars["W"]
		if !sameShape(v, w) || !match(v, w, 0) || depth(v) != depth(w) ||
			v.Size() == 0 && !match(v.Fill(), w.Fill(), 0) {
			t.Errorf("%s: literal %s gives %s", c, lit, Format(w))
		}
	}
}
