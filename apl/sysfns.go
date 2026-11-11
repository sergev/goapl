package apl

import "strings"

// sysFn applies the system function name to x.
func (in *Interp) sysFn(name string, x *Array) (*Array, error) {
	switch name {
	case "⎕FX":
		return in.quadFX(x)
	case "⎕CR":
		return in.quadCR(x)
	}
	names, err := nameList(x)
	if err != nil {
		return nil, err
	}
	r := make([]float64, len(names))
	for i, n := range names {
		if name == "⎕NC" {
			r[i] = float64(in.nameClass(n))
		} else {
			r[i] = in.erase(n)
		}
	}
	if x.Rank() < 2 {
		return Scalar(r[0]), nil
	}
	return Vector(r), nil
}

// textLines gives the rows of a character matrix (without trailing
// blanks), the items of a vector of strings, or a string as one line.
func textLines(x *Array) ([]string, error) {
	str := func(a *Array) (string, error) {
		if a.Size() > 0 && !a.IsCharacter() || a.Rank() > 1 {
			return "", newError(DomainError)
		}
		var b strings.Builder
		for i := range a.Size() {
			b.WriteRune(rune(a.At(i).(Char)))
		}
		return strings.TrimRight(b.String(), " "), nil
	}
	switch {
	case x.Rank() == 2 && x.IsCharacter():
		var lines []string
		w := x.Shape[1]
		for r := range x.Shape[0] {
			s, _ := str(x.gather([]int{w}, rangeFrom(r*w, w)))
			lines = append(lines, s)
		}
		return lines, nil
	case x.Rank() == 1 && !x.IsSimple():
		var lines []string
		for i := range x.Size() {
			s, err := str(item(x.At(i)))
			if err != nil {
				return nil, err
			}
			lines = append(lines, s)
		}
		return lines, nil
	case x.Rank() <= 1:
		s, err := str(x)
		return []string{s}, err
	}
	return nil, newError(RankError)
}

func rangeFrom(start, n int) []int {
	r := make([]int, n)
	for i := range r {
		r[i] = start + i
	}
	return r
}

// nameList gives the names in a string or the rows of a matrix.
func nameList(x *Array) ([]string, error) {
	if x.Rank() > 2 || x.Size() > 0 && !x.IsCharacter() {
		return nil, newError(DomainError)
	}
	lines, err := textLines(x)
	for i, l := range lines {
		lines[i] = strings.TrimSpace(l)
	}
	return lines, err
}

// quadFX is ⎕FX: defines a function from its lines and gives its name,
// or the index (from ⎕IO) of the first line in error.
func (in *Interp) quadFX(x *Array) (*Array, error) {
	lines, err := textLines(x)
	if err != nil {
		return nil, err
	}
	f, line, err := fix(lines)
	if err == nil {
		if _, ok := in.vars[f.Name]; ok {
			err = newError(DefnError)
		}
	}
	if err != nil {
		return Scalar(float64(line + in.io)), nil
	}
	in.fns[f.Name] = f
	in.fnGen++
	return Chars(f.Name), nil
}

// quadCR is ⎕CR: the lines of a function as a character matrix, or an
// empty matrix if the name is not a function.
func (in *Interp) quadCR(x *Array) (*Array, error) {
	names, err := nameList(x)
	if err != nil {
		return nil, err
	}
	if len(names) != 1 {
		return nil, newError(LengthError)
	}
	f := in.fns[names[0]]
	if f == nil {
		return &Array{Shape: []int{0, 0}, Elems: []Value{}, Proto: Char(' ')}, nil
	}
	r := toChars(f.Lines(), []int{len(f.Lines())})
	return r, nil
}

// validName reports whether s is a name of a variable or function.
func validName(s string) bool {
	toks, err := Lex(s)
	return err == nil && len(toks) == 1 && toks[0].Kind == Name
}

// nameClass is the ⎕NC of a name: 0 unused, 2 variable, 3 function,
// 4 operator, ¯1 not a valid name.
func (in *Interp) nameClass(n string) int {
	switch {
	case !validName(n):
		return -1
	case isSysName(n):
		if sysVars[n] {
			return 2
		}
		if sysFns[n] {
			return 3
		}
		return -1
	case in.fns[n] != nil && in.fns[n].IsOperator():
		return 4
	case in.fns[n] != nil:
		return 3
	case in.vars[n] != nil:
		return 2
	}
	return 0
}

// erase is ⎕EX of a name: 1 if the name is now unused.
func (in *Interp) erase(n string) float64 {
	if !validName(n) || isSysName(n) {
		return 0
	}
	if in.fns[n] != nil {
		delete(in.fns, n)
		in.fnGen++
	}
	delete(in.vars, n)
	return 1
}
