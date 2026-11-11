package apl

import (
	"math"
	"strconv"
	"strings"
)

// PrintPrecision is the number of significant digits shown (⎕PP).
const PrintPrecision = 10

// FormatNumber renders a number in APL notation: ¯ for minus, E for exponent.
func FormatNumber(v float64) string {
	if v == 0 {
		return "0"
	}
	if math.IsInf(v, 0) {
		if v < 0 {
			return "¯∞"
		}
		return "∞"
	}
	s := strconv.FormatFloat(v, 'g', PrintPrecision, 64)
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

// Format renders an array as display lines.
func Format(a *Array) string {
	switch a.Rank() {
	case 0:
		return FormatNumber(a.Data[0])
	case 1:
		s := make([]string, len(a.Data))
		for i, v := range a.Data {
			s[i] = FormatNumber(v)
		}
		return strings.Join(s, " ")
	}
	cols := a.Shape[a.Rank()-1]
	rows := a.Shape[a.Rank()-2]
	cells := make([]string, len(a.Data))
	width := make([]int, cols)
	for i, v := range a.Data {
		cells[i] = FormatNumber(v)
		width[i%cols] = max(width[i%cols], len([]rune(cells[i])))
	}
	var b strings.Builder
	plane := rows * cols
	planes := shapeSize(a.Shape[:a.Rank()-2])
	for p := range planes {
		if p > 0 {
			b.WriteString("\n")
		}
		for r := range rows {
			if p > 0 || r > 0 {
				b.WriteString("\n")
			}
			line := make([]string, cols)
			for c := range cols {
				s := cells[p*plane+r*cols+c]
				line[c] = strings.Repeat(" ", width[c]-len([]rune(s))) + s
			}
			b.WriteString(strings.Join(line, " "))
		}
	}
	return b.String()
}
