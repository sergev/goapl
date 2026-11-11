// Package apl implements an APL interpreter.
package apl

import (
	"strconv"
	"strings"
	"unicode"
)

// Kind classifies a token.
type Kind int

const (
	Number  Kind = iota // numeric vector constant
	Name                // identifier
	Prim                // primitive function or operator glyph
	Special             // ← → ∇ ⎕ ⍞ ( ) [ ] ; : . ∘
	String              // character constant
)

// Token is a lexical unit of an APL line.
type Token struct {
	Kind Kind
	Pos  int       // rune offset in line
	Text string    // name, glyph or characters of a string
	Nums []float64 // values for Number
}

func (t Token) String() string {
	switch t.Kind {
	case Number:
		s := make([]string, len(t.Nums))
		for i, v := range t.Nums {
			s[i] = strconv.FormatFloat(v, 'g', -1, 64)
		}
		return "NUM(" + strings.Join(s, " ") + ")"
	case Name:
		return "NAME(" + t.Text + ")"
	case Prim:
		return "PRIM(" + t.Text + ")"
	case String:
		return "STR(" + t.Text + ")"
	}
	return "SPEC(" + t.Text + ")"
}

const (
	primGlyphs    = "+-×÷*⍳⍴,⍪~=≠<≤≥>∧∨⍲⍱⌈⌊○/⌿\\⍀?↑↓⊥⊤|ε∊⌽⊖⍉⍋⍒!⍟⌹⍕⍎⊂⊃≡¨"
	specialGlyphs = "←→∇⎕⍞()[];:.∘⋄"
)

// Lex splits a line into tokens.
func Lex(line string) ([]Token, error) {
	src := []rune(line)
	var toks []Token
	i := 0
	for i < len(src) {
		c := src[i]
		switch {
		case c == ' ' || c == '\t':
			i++
		case c == '⍝':
			i = len(src) // comment
		case c == '\'':
			start := i
			var b strings.Builder
			for {
				i++
				if i >= len(src) {
					return nil, errorAt(SyntaxError, start)
				}
				if src[i] == '\'' {
					if i+1 < len(src) && src[i+1] == '\'' {
						i++ // doubled quote
					} else {
						break
					}
				}
				b.WriteRune(src[i])
			}
			i++
			toks = append(toks, Token{Kind: String, Pos: start, Text: b.String()})
		case isDigit(c) || c == '¯' || (c == '.' && i+1 < len(src) && isDigit(src[i+1])):
			start := i
			var nums []float64
			for {
				v, n, err := lexNumber(src, i)
				if err != nil {
					return nil, err
				}
				nums = append(nums, v)
				i = n
				j := i
				for j < len(src) && src[j] == ' ' {
					j++
				}
				if j < len(src) && (isDigit(src[j]) || src[j] == '¯' ||
					(src[j] == '.' && j+1 < len(src) && isDigit(src[j+1]))) {
					i = j
					continue
				}
				break
			}
			toks = append(toks, Token{Kind: Number, Pos: start, Nums: nums})
		case isLetter(c) || c == '⎕' && i+1 < len(src) && isLetter(src[i+1]):
			start := i
			i++
			for i < len(src) && (isLetter(src[i]) || isDigit(src[i])) {
				i++
			}
			toks = append(toks, Token{Kind: Name, Pos: start, Text: string(src[start:i])})
		case strings.ContainsRune(primGlyphs, c):
			if c == 'ε' {
				c = '∊'
			}
			toks = append(toks, Token{Kind: Prim, Pos: i, Text: string(c)})
			i++
		case strings.ContainsRune(specialGlyphs, c):
			toks = append(toks, Token{Kind: Special, Pos: i, Text: string(c)})
			i++
		default:
			return nil, errorAt(SyntaxError, i)
		}
	}
	return toks, nil
}

// lexNumber reads one number at src[i]; returns value and next index.
func lexNumber(src []rune, i int) (float64, int, error) {
	var b strings.Builder
	if src[i] == '¯' {
		b.WriteByte('-')
		i++
		if i >= len(src) || !(isDigit(src[i]) || src[i] == '.') {
			return 0, i, errorAt(SyntaxError, i)
		}
	}
	for i < len(src) && isDigit(src[i]) {
		b.WriteRune(src[i])
		i++
	}
	if i < len(src) && src[i] == '.' {
		b.WriteByte('.')
		i++
		if i >= len(src) || !isDigit(src[i]) {
			return 0, i, errorAt(SyntaxError, i)
		}
		for i < len(src) && isDigit(src[i]) {
			b.WriteRune(src[i])
			i++
		}
	}
	if i < len(src) && src[i] == 'E' && i+1 < len(src) &&
		(isDigit(src[i+1]) || src[i+1] == '¯') {
		b.WriteByte('e')
		i++
		if src[i] == '¯' {
			b.WriteByte('-')
			i++
		}
		for i < len(src) && isDigit(src[i]) {
			b.WriteRune(src[i])
			i++
		}
	}
	v, err := strconv.ParseFloat(b.String(), 64)
	if err != nil {
		return 0, i, errorAt(DomainError, i)
	}
	return v, i, nil
}

func isDigit(c rune) bool { return c >= '0' && c <= '9' }
func isLetter(c rune) bool {
	return c == '∆' || c == '_' || unicode.IsLetter(c) && c != 'ε' && c != '∊'
}
