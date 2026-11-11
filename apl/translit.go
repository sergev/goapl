package apl

import "strings"

// Glyphs lists the ASCII names of the APL glyphs: #rho for ⍴.
var Glyphs = map[string]rune{
	"times": '×', "div": '÷', "divide": '÷', "rho": '⍴', "iota": '⍳',
	"max": '⌈', "ceil": '⌈', "min": '⌊', "floor": '⌊', "log": '⍟',
	"circle": '○', "le": '≤', "ge": '≥', "ne": '≠', "and": '∧', "or": '∨',
	"nand": '⍲', "nor": '⍱', "take": '↑', "drop": '↓', "enclose": '⊂',
	"disclose": '⊃', "pick": '⊃', "epsilon": '∊', "in": '∊',
	"rotate": '⌽', "reverse": '⌽', "flip": '⊖', "transpose": '⍉',
	"gradeup": '⍋', "gradedown": '⍒', "decode": '⊥', "encode": '⊤',
	"domino": '⌹', "format": '⍕', "execute": '⍎', "match": '≡',
	"depth": '≡', "each": '¨', "jot": '∘', "slashbar": '⌿',
	"backslashbar": '⍀', "commabar": '⍪',
	"quad": '⎕', "quotequad": '⍞', "assign": '←', "gets": '←',
	"goto": '→', "branch": '→', "del": '∇', "delta": '∆', "comment": '⍝',
	"lamp": '⍝', "diamond": '⋄', "neg": '¯', "macron": '¯',
}

// Transliterate replaces #name by its glyph, outside strings and
// comments. Unknown names stay as they are.
func Transliterate(line string) string {
	if !strings.Contains(line, "#") {
		return line
	}
	var b strings.Builder
	rs := []rune(line)
	quoted := false
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch {
		case c == '\'':
			quoted = !quoted
		case c == '⍝' && !quoted:
			b.WriteString(string(rs[i:]))
			return b.String()
		case c == '#' && !quoted:
			if g, n := glyphAt(rs[i+1:]); n > 0 {
				b.WriteRune(g)
				i += n
				if g == '⍝' {
					b.WriteString(string(rs[i+1:]))
					return b.String()
				}
				continue
			}
		}
		b.WriteRune(c)
	}
	return b.String()
}

// glyphAt finds the longest glyph name at the start of rs; n is its
// length, 0 if there is none.
func glyphAt(rs []rune) (rune, int) {
	n := 0
	for n < len(rs) && (rs[n] >= 'a' && rs[n] <= 'z' || rs[n] >= 'A' && rs[n] <= 'Z') {
		n++
	}
	for ; n > 0; n-- {
		if g, ok := Glyphs[strings.ToLower(string(rs[:n]))]; ok {
			return g, n
		}
	}
	return 0, 0
}
