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

// cellLines renders one element as lines; a nested item gets a
// leading blank.
func cellLines(v Value, pp int) []string {
	switch v := v.(type) {
	case float64:
		return []string{formatNumber(v, pp)}
	case Char:
		return []string{string(rune(v))}
	case *Array:
		lines := strings.Split(format(v, pp), "\n")
		for i, l := range lines {
			lines[i] = " " + l
		}
		return lines
	}
	return []string{"?"}
}

// Format renders an array with the default ⎕PP.
func Format(a *Array) string { return format(a, defaultPP) }

// format renders an array as display lines with pp significant digits.
// Rows of a matrix are lines; each axis before the last two separates
// its planes by one more blank line.
func format(a *Array, pp int) string {
	rows := formatRows(a, pp)
	r := a.Rank()
	var b strings.Builder
	for i, row := range rows {
		if i > 0 {
			b.WriteString("\n")
			for j := range max(r-2, 0) {
				if i%shapeSize(a.Shape[j+1:r-1]) == 0 {
					b.WriteString("\n")
				}
			}
		}
		b.WriteString(row)
	}
	return b.String()
}

// formatRows renders each row (each vector along the last axis) of an
// array; a row holding nested items may take several lines, separated
// by newlines. Columns holding numbers are right-aligned, and set off
// by a blank, as are nested items; adjacent character columns touch.
func formatRows(a *Array, pp int) []string {
	r := a.Rank()
	cols, nrows := 1, 1
	switch {
	case r == 1:
		cols = max(a.Size(), 1)
	case r > 1:
		cols, nrows = a.Shape[r-1], shapeSize(a.Shape[:r-1])
	}
	cells := make([][]string, a.Size())
	width := make([]int, cols)
	numeric := make([]bool, cols)
	nested := make([]bool, cols)
	for i := range cells {
		v := a.At(i)
		cells[i] = cellLines(v, pp)
		for _, l := range cells[i] {
			width[i%cols] = max(width[i%cols], len([]rune(l)))
		}
		switch v.(type) {
		case float64:
			numeric[i%cols] = true
		case *Array:
			nested[i%cols] = true
		}
	}
	rows := make([]string, nrows)
	if a.Size() == 0 {
		return rows
	}
	for row := range rows {
		height := 1
		for c := range cols {
			height = max(height, len(cells[row*cols+c]))
		}
		lines := make([]string, height)
		for j := range lines {
			var b strings.Builder
			for c := range cols {
				if c > 0 && (numeric[c] || numeric[c-1] || nested[c] || nested[c-1]) {
					b.WriteByte(' ')
				}
				s := ""
				if cell := cells[row*cols+c]; j < len(cell) {
					s = cell[j]
				}
				pad := strings.Repeat(" ", width[c]-len([]rune(s)))
				if numeric[c] && !nested[c] {
					b.WriteString(pad + s)
				} else {
					b.WriteString(s + pad)
				}
			}
			lines[j] = b.String()
			if !a.IsSimple() {
				lines[j] = strings.TrimRight(lines[j], " ")
			}
		}
		rows[row] = strings.Join(lines, "\n")
	}
	return rows
}

// wrap folds lines longer than width; continuation lines are indented
// by six blanks. Lines break at a blank where possible. Each piece
// holds at least one character, so any width terminates.
func wrap(text string, width int) string {
	const indent = "      "
	var out []string
	for _, line := range strings.Split(text, "\n") {
		rs := []rune(line)
		prefix, room := "", width
		for len(rs) > room {
			cut := room
			for i := room; i > room/2; i-- {
				if rs[i] == ' ' {
					cut = i
					break
				}
			}
			out = append(out, prefix+strings.TrimRight(string(rs[:cut]), " "))
			rs = []rune(strings.TrimLeft(string(rs[cut:]), " "))
			prefix, room = indent, max(width-len(indent), 1)
		}
		out = append(out, prefix+string(rs))
	}
	return strings.Join(out, "\n")
}

// toChars turns display rows into a character array of shape s.
func toChars(rows []string, shape []int) *Array {
	if strings.Contains(strings.Join(rows, ""), "\n") {
		rows = strings.Split(strings.Join(rows, "\n"), "\n")
		shape = []int{len(rows)}
	}
	w := 0
	for _, r := range rows {
		w = max(w, len([]rune(r)))
	}
	shape = append(append([]int(nil), shape...), w)
	e := make([]Value, 0, len(rows)*w)
	for _, r := range rows {
		rs := []rune(r)
		for i := range w {
			c := ' '
			if i < len(rs) {
				c = rs[i]
			}
			e = append(e, Char(c))
		}
	}
	return &Array{Shape: shape, Elems: e, Proto: Char(' ')}
}

// formatFn is ⍕X: the character array that displays as X.
func formatFn(e *env, x *Array) (*Array, error) {
	if x.Size() > 0 && x.IsCharacter() {
		return x, nil
	}
	var lead []int
	if x.Rank() > 0 {
		lead = x.Shape[:x.Rank()-1]
	}
	return toChars(formatRows(x, e.pp), lead), nil
}

// formatWidth is A⍕B: B in fixed format. A holds a width and a number of
// decimals for all columns or for each, or just the decimals. Negative
// decimals give that many significant digits in E notation. A width of
// 0 fits the column with a blank before it; a number that does not fit
// fills its field with *.
func formatWidth(e *env, a, b *Array) (*Array, error) {
	spec, err := toSignedInts(a)
	if err != nil || a.Rank() > 1 {
		if err == nil {
			err = newError(RankError)
		}
		return nil, err
	}
	d, err := b.numeric()
	if err != nil {
		return nil, err
	}
	cols := 1
	if b.Rank() > 0 {
		cols = b.Shape[b.Rank()-1]
	}
	var lead []int
	if b.Rank() > 0 {
		lead = b.Shape[:b.Rank()-1]
	}
	widths, decs := make([]int, cols), make([]int, cols)
	switch {
	case len(spec) == 1:
		for c := range cols {
			decs[c] = spec[0]
		}
	case len(spec) == 2:
		for c := range cols {
			widths[c], decs[c] = spec[0], spec[1]
		}
	case len(spec) == 2*cols:
		for c := range cols {
			widths[c], decs[c] = spec[2*c], spec[2*c+1]
		}
	default:
		return nil, newError(LengthError)
	}
	cells := make([]string, len(d))
	for i, v := range d {
		cells[i] = fixed(v, decs[i%cols])
	}
	for c := range cols {
		if widths[c] == 0 {
			for i := c; i < len(cells); i += cols {
				widths[c] = max(widths[c], len([]rune(cells[i]))+1)
			}
		}
	}
	nrows := len(d) / max(cols, 1)
	if b.Rank() == 0 {
		nrows = 1
	}
	rows := make([]string, nrows)
	for row := range rows {
		var sb strings.Builder
		for c := range cols {
			s := cells[row*cols+c]
			n := len([]rune(s))
			if n > widths[c] {
				s, n = strings.Repeat("*", widths[c]), widths[c]
			}
			sb.WriteString(strings.Repeat(" ", widths[c]-n) + s)
		}
		rows[row] = sb.String()
	}
	return toChars(rows, lead), nil
}

// fixed renders v with p decimals, or with -p significant digits in E
// notation if p is negative.
func fixed(v float64, p int) string {
	if v == 0 {
		v = 0 // no negative zero
	}
	var s string
	if p >= 0 {
		if p <= 15 { // round halves away from zero
			m := math.Pow(10, float64(p))
			if r := math.Round(v*m) / m; !math.IsInf(r, 0) {
				v = r
			}
		}
		s = strconv.FormatFloat(v, 'f', p, 64)
	} else {
		s = strconv.FormatFloat(v, 'e', -p-1, 64)
		i := strings.IndexByte(s, 'e')
		exp := strings.TrimLeft(s[i+2:], "0")
		if exp == "" {
			exp = "0"
		}
		if s[i+1] == '-' {
			exp = "-" + exp
		}
		s = s[:i] + "E" + exp
	}
	return strings.ReplaceAll(s, "-", "¯")
}
