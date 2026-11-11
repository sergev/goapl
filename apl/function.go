package apl

import "strings"

// MaxDepth limits nesting of user function calls.
const MaxDepth = 1000

// Function is a user-defined (∇) function.
type Function struct {
	Name   string
	Result string // explicit result name, or ""
	Left   string // left argument name (dyadic), or ""
	Right  string // right argument name (monadic, dyadic), or ""
	Locals []string
	Stmts  []*stmt
}

// Arity returns 0, 1 or 2.
func (f *Function) Arity() int {
	switch {
	case f.Left != "":
		return 2
	case f.Right != "":
		return 1
	}
	return 0
}

// stmt is one line of a function body.
type stmt struct {
	label string
	text  string
	toks  []Token
	node  node // parse cache
	gen   int  // fnGen at time of parse
}

// defineHeader parses "∇[R←][A] F [B]{;L}" and enters definition mode.
func (in *Interp) defineHeader(line string, toks []Token) error {
	toks = toks[1:] // ∇
	f := &Function{}
	var names []string
	i := 0
	name := func(code int) (string, error) {
		if i >= len(toks) || toks[i].Kind != Name {
			pos := len([]rune(line))
			if i < len(toks) {
				pos = toks[i].Pos
			}
			return "", errorAt(code, pos)
		}
		i++
		return toks[i-1].Text, nil
	}
	n, err := name(7)
	if err != nil {
		return err
	}
	if i < len(toks) && isSpecial(toks[i], "←") {
		i++
		f.Result = n
		if n, err = name(6); err != nil {
			return err
		}
	}
	names = append(names, n)
	for len(names) < 3 && i < len(toks) && toks[i].Kind == Name {
		names = append(names, toks[i].Text)
		i++
	}
	for i < len(toks) && isSpecial(toks[i], ";") {
		i++
		l, err := name(7)
		if err != nil {
			return err
		}
		f.Locals = append(f.Locals, l)
	}
	if i < len(toks) {
		return errorAt(3, toks[i].Pos)
	}
	switch len(names) {
	case 1:
		f.Name = names[0]
	case 2:
		f.Name, f.Right = names[0], names[1]
	case 3:
		f.Left, f.Name, f.Right = names[0], names[1], names[2]
	}
	if _, ok := in.fns[f.Name]; ok {
		return errorAt(5, toks[0].Pos)
	}
	in.defining = f
	return nil
}

// defineLine adds a body line or closes the definition with ∇.
func (in *Interp) defineLine(line string) error {
	f := in.defining
	toks, err := Lex(line)
	if err != nil {
		return err
	}
	closing := len(toks) > 0 && isSpecial(toks[len(toks)-1], "∇")
	if closing {
		toks = toks[:len(toks)-1]
		end := len(line)
		if i := strings.LastIndex(line, "∇"); i >= 0 {
			end = i
		}
		line = line[:end]
	}
	if len(toks) > 0 || strings.Contains(line, "⍝") {
		st := &stmt{text: line, toks: toks}
		if len(toks) >= 2 && toks[0].Kind == Name && isSpecial(toks[1], ":") {
			st.label = toks[0].Text
			st.toks = toks[2:]
		}
		f.Stmts = append(f.Stmts, st)
	}
	if closing {
		in.defining = nil
		if len(f.Stmts) == 0 {
			return newError(75)
		}
		in.fns[f.Name] = f
		in.fnGen++
	}
	return nil
}

// call runs a user function with dynamic scoping of its local names.
func (in *Interp) call(f *Function, left, right *Array) (*Array, error) {
	if in.depth >= MaxDepth {
		return nil, newError(16)
	}
	in.depth++
	defer func() { in.depth-- }()

	locals := append([]string{f.Result, f.Left, f.Right}, f.Locals...)
	for _, st := range f.Stmts {
		locals = append(locals, st.label)
	}
	saved := map[string]*Array{}
	for _, n := range locals {
		if n == "" {
			continue
		}
		if _, done := saved[n]; done {
			continue
		}
		saved[n] = in.vars[n]
		delete(in.vars, n)
	}
	defer func() {
		for n, v := range saved {
			if v == nil {
				delete(in.vars, n)
			} else {
				in.vars[n] = v
			}
		}
	}()
	for i, st := range f.Stmts {
		if st.label != "" {
			in.vars[st.label] = Scalar(float64(i + 1))
		}
	}
	if f.Left != "" {
		in.vars[f.Left] = left
	}
	if f.Right != "" {
		in.vars[f.Right] = right
	}

	for pc := 1; pc >= 1 && pc <= len(f.Stmts); {
		st := f.Stmts[pc-1]
		if st.node == nil || st.gen != in.fnGen {
			n, err := in.parse(st.toks, len([]rune(st.text)))
			if err != nil {
				return nil, locate(err, f, pc, st.text)
			}
			st.node, st.gen = n, in.fnGen
		}
		target, err := in.run(st.node)
		if err != nil {
			return nil, locate(err, f, pc, st.text)
		}
		if target >= 0 {
			pc = target
		} else {
			pc++
		}
	}
	if f.Result == "" {
		return nil, nil
	}
	r, ok := in.vars[f.Result]
	if !ok {
		return nil, newError(100)
	}
	return r, nil
}

// locate records where an error happened, unless already recorded.
func locate(err error, f *Function, pc int, text string) error {
	if e, ok := err.(*Error); ok && e.Src == "" {
		e.Src = text
		e.Where = f.Name + "[" + itoa(pc) + "]"
	}
	return err
}

func itoa(n int) string { return FormatNumber(float64(n)) }
