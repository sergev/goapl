package apl

import (
	"math"
	"strconv"
	"strings"
)

// FormatNumber renders a number in APL notation with the default ⎕PP.
func FormatNumber(v float64) string { return formatNumber(v, defaultPP) }

// formatNumber renders a number with pp significant digits: ¯ for
// minus, E for exponent.
func formatNumber(v float64, pp int) string {
	if v == 0 {
		return "0"
	}
	if math.IsInf(v, 0) {
		if v < 0 {
			return "¯∞"
		}
		return "∞"
	}
	s := strconv.FormatFloat(v, 'g', pp, 64)
	if i := strings.IndexByte(s, 'e'); i >= 0 {
		mant, exp := s[:i], s[i+1:]
		neg := exp[0] == '-'
		exp = strings.TrimLeft(exp[1:], "0")
		if neg {
			exp = "-" + exp
		}
		s = mant + "E" + exp
	}
	return strings.ReplaceAll(s, "-", "¯")
}

// formatElem renders one element.
func formatElem(v Value, pp int) string {
	switch v := v.(type) {
	case float64:
		return formatNumber(v, pp)
	case Char:
		return string(rune(v))
	case *Array:
		return format(v, pp)
	}
	return "?"
}

// Format renders an array with the default ⎕PP.
func Format(a *Array) string { return format(a, defaultPP) }

// format renders an array as display lines with pp significant digits.
// Rows of a matrix are lines; each axis before the last two separates
// its planes by one more blank line.
func format(a *Array, pp int) string {
	r := a.Rank()
	switch r {
	case 0:
		return formatElem(a.At(0), pp)
	case 1:
		s := make([]string, a.Size())
		for i := range s {
			s[i] = formatElem(a.At(i), pp)
		}
		return strings.Join(s, " ")
	}
	cols := a.Shape[r-1]
	rows := shapeSize(a.Shape[:r-1])
	cells := make([]string, a.Size())
	width := make([]int, cols)
	for i := range cells {
		cells[i] = formatElem(a.At(i), pp)
		width[i%cols] = max(width[i%cols], len([]rune(cells[i])))
	}
	var b strings.Builder
	for row := range rows {
		if row > 0 {
			b.WriteString("\n")
			for j := range r - 2 {
				if row%shapeSize(a.Shape[j+1:r-1]) == 0 {
					b.WriteString("\n")
				}
			}
		}
		line := make([]string, cols)
		for c := range cols {
			s := cells[row*cols+c]
			line[c] = strings.Repeat(" ", width[c]-len([]rune(s))) + s
		}
		b.WriteString(strings.Join(line, " "))
	}
	return b.String()
}
