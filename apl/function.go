package apl

import "strings"

// MaxDepth limits nesting of user function calls.
const MaxDepth = 1000

// Function is a user-defined (∇) function or operator.
type Function struct {
	Name    string
	Result  string // explicit result name, or ""
	Left    string // left argument name (dyadic), or ""
	Right   string // right argument name (monadic, dyadic), or ""
	OpLeft  string // left operand name of an operator, or ""
	OpRight string // right operand name of a dyadic operator, or ""
	Locals  []string
	Header  string // header text without ∇
	Stmts   []*stmt
}

// Arity returns 0, 1 or 2: the valence of the function, or of the
// functions an operator derives.
func (f *Function) Arity() int {
	switch {
	case f.Left != "":
		return 2
	case f.Right != "":
		return 1
	}
	return 0
}

// IsOperator reports whether f is a defined operator.
func (f *Function) IsOperator() bool { return f.OpLeft != "" }

// stmt is one line of a function body.
type stmt struct {
	label string
	text  string
	toks  []Token
	nodes []node // parse cache: statements separated by ⋄
	gen   int    // fnGen at time of parse
}

// Lines gives the header and the body lines, as ⎕CR shows them.
func (f *Function) Lines() []string {
	r := []string{f.Header}
	for _, st := range f.Stmts {
		r = append(r, st.text)
	}
	return r
}

// parseHeader parses "[R←][A] F [B]{;L}", where F may be "(F OP [G])"
// for an operator.
func parseHeader(text string) (*Function, error) {
	toks, err := Lex(text)
	if err != nil {
		return nil, err
	}
	f := &Function{Header: strings.TrimSpace(text)}
	end := len([]rune(text))
	i := 0
	bad := func() error {
		if i < len(toks) {
			return errorAt(DefnError, toks[i].Pos)
		}
		return errorAt(DefnError, end)
	}
	// name reads a name; sys allows a system variable.
	name := func(sys bool) (string, error) {
		if i >= len(toks) || toks[i].Kind != Name ||
			isSysName(toks[i].Text) && !(sys && sysVars[toks[i].Text] && !readOnly[toks[i].Text]) {
			return "", bad()
		}
		i++
		return toks[i-1].Text, nil
	}
	if len(toks) > 1 && toks[0].Kind == Name && isSpecial(toks[1], "←") {
		if f.Result, err = name(false); err != nil {
			return nil, err
		}
		i++
	}
	// Up to three items: names, and at most one "(F OP [G])".
	var items []string
	group := -1
	for len(items) < 3 && i < len(toks) && !isSpecial(toks[i], ";") {
		if isSpecial(toks[i], "(") {
			if group >= 0 {
				return nil, bad()
			}
			i++
			var g []string
			for len(g) < 3 && i < len(toks) && toks[i].Kind == Name {
				n, _ := name(false)
				g = append(g, n)
			}
			if len(g) < 2 || i >= len(toks) || !isSpecial(toks[i], ")") {
				return nil, bad()
			}
			i++
			group = len(items)
			f.OpLeft, f.Name = g[0], g[1]
			if len(g) == 3 {
				f.OpRight = g[2]
			}
			items = append(items, f.Name)
			continue
		}
		n, err := name(false)
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	for i < len(toks) && isSpecial(toks[i], ";") {
		i++
		l, err := name(true)
		if err != nil {
			return nil, err
		}
		f.Locals = append(f.Locals, l)
	}
	if i < len(toks) || len(items) == 0 {
		return nil, bad()
	}
	fn := 0 // index of the function name among items
	if len(items) == 3 {
		fn = 1
	}
	if group >= 0 && group != fn {
		return nil, errorAt(DefnError, 0)
	}
	switch len(items) {
	case 1:
		f.Name = items[0]
	case 2:
		f.Name, f.Right = items[0], items[1]
	case 3:
		f.Left, f.Name, f.Right = items[0], items[1], items[2]
	}
	return f, nil
}

// fix builds a function from its header and body lines. A bad line
// gives a DEFN ERROR; line is its index (0 for the header).
func fix(lines []string) (f *Function, line int, err error) {
	if len(lines) == 0 {
		return nil, 0, newError(DefnError)
	}
	f, err = parseHeader(lines[0])
	if err != nil {
		return nil, 0, err
	}
	labels := map[string]bool{}
	for k, text := range lines[1:] {
		toks, err := Lex(text)
		if err != nil {
			return nil, k + 1, err
		}
		st := &stmt{text: text, toks: toks}
		if len(toks) >= 2 && toks[0].Kind == Name && isSpecial(toks[1], ":") {
			st.label = toks[0].Text
			st.toks = toks[2:]
			if labels[st.label] || isSysName(st.label) {
				return nil, k + 1, errorAt(DefnError, toks[0].Pos)
			}
			labels[st.label] = true
		}
		f.Stmts = append(f.Stmts, st)
	}
	return f, 0, nil
}

// parseLine parses the statements of a line, separated by ⋄; end is
// the rune length of the line.
func (in *Interp) parseLine(toks []Token, end int, op *Function) ([]node, error) {
	var nodes []node
	start := 0
	for i := 0; i <= len(toks); i++ {
		e := end
		if i < len(toks) {
			if !isSpecial(toks[i], "⋄") {
				continue
			}
			e = toks[i].Pos
		}
		n, err := in.parse(toks[start:i], e, op)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
		start = i + 1
	}
	return nodes, nil
}

// runLine runs the statements of a line in order until a branch.
func (in *Interp) runLine(nodes []node) (int, error) {
	for _, n := range nodes {
		target, err := in.run(n)
		if err != nil || target >= 0 {
			return target, err
		}
	}
	return -1, nil
}

// call runs a user function with dynamic scoping of its local names.
// A dyadic function called with one argument leaves its left argument
// without a value.
func (in *Interp) call(f *Function, left, right *Array) (*Array, error) {
	if in.depth >= MaxDepth {
		return nil, newError(SystemLimit)
	}
	in.depth++
	defer func() { in.depth-- }()

	locals := append([]string{f.Result, f.Left, f.Right}, f.Locals...)
	for _, st := range f.Stmts {
		locals = append(locals, st.label)
	}
	saved := map[string]*Array{} // nil: had no value
	for _, n := range locals {
		if n == "" {
			continue
		}
		if _, done := saved[n]; done {
			continue
		}
		saved[n], _ = in.getVar(n)
		if !isSysName(n) {
			delete(in.vars, n) // a localized system variable keeps its value
		}
	}
	defer func() {
		for n, v := range saved {
			switch {
			case isSysName(n):
				in.env.set(n, v)
			case v == nil:
				delete(in.vars, n)
			default:
				in.vars[n] = v
			}
		}
	}()
	for i, st := range f.Stmts {
		if st.label != "" {
			in.vars[st.label] = Scalar(float64(i + 1))
		}
	}
	if f.Left != "" && left != nil {
		in.vars[f.Left] = left
	}
	if f.Right != "" {
		in.vars[f.Right] = right
	}
	var op *Function
	if f.IsOperator() {
		op = f
	}
	fr := &frame{f: f}
	in.stack = append(in.stack, fr)
	defer func() { in.stack = in.stack[:len(in.stack)-1] }()
	for pc := 1; pc >= 1 && pc <= len(f.Stmts); {
		fr.pc = pc
		st := f.Stmts[pc-1]
		target, err := -1, error(nil)
		if in.interrupted.Swap(false) {
			err = newError(Interrupt)
		} else if st.nodes == nil || st.gen != in.fnGen {
			st.nodes, err = in.parseLine(st.toks, len([]rune(st.text)), op)
			st.gen = in.fnGen
			if err != nil {
				st.nodes = nil
			}
		}
		if err == nil {
			target, err = in.runLine(st.nodes)
		}
		if err != nil {
			err = locate(err, f, pc, st.text)
			if !in.canSuspend(err) {
				return nil, err
			}
			if target, err = in.suspend(err); err != nil {
				return nil, err
			}
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
		return nil, newError(ValueError)
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
