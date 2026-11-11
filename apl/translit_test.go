package apl

import "testing"

func TestTransliterate(t *testing.T) {
	cases := []struct{ in, want string }{
		{"#rho 2 3", "⍴ 2 3"},
		{"2 3#RHO#iota6", "2 3⍴⍳6"},
		{"+/#iota10", "+/⍳10"},
		{"X#gets'#rho' #comment #rho", "X←'#rho' ⍝ #rho"},
		{"#ROTATE#rhoX", "⌽⍴X"},
		{"#nosuch #", "#nosuch #"},
		{"1#ne2", "1≠2"},
		{"no hash", "no hash"},
	}
	for _, c := range cases {
		if got := Transliterate(c.in); got != c.want {
			t.Errorf("Transliterate(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// Every glyph name can be typed.
func TestGlyphNames(t *testing.T) {
	for name, g := range Glyphs {
		if got := Transliterate("#" + name); got != string(g) {
			t.Errorf("#%s gives %q", name, got)
		}
	}
}
