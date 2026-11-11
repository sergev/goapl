package apl

// Grammar, read left to right but binding right to left:
//
//	stmt    := → expr | expr
//	expr    := fn expr | name ← expr | name [index] ← expr | ⎕ ← expr
//	         | operand [fn expr]
//	         | ⍞ ← expr
//	operand := (number | string | name | niladic | ⎕ | ⍞ | ( expr )) {[index]}
//	index   := [expr] {; [expr]}
//	fn      := (prim | prim op) [axis] | prim . prim | ∘ . prim | user function
//	op      := / | ⌿ | \ | ⍀
//	axis    := [expr]

type node interface{ pos() int }

type constNode struct {
	val *Array
	at  int
}

type varNode struct {
	name string
	at   int
}

type assignNode struct {
	name  string
	at    int
	right node
}

type indexNode struct {
	target node
	idx    []node // nil item = whole axis
	at     int
}

type indexAssignNode struct {
	name  string
	idx   []node
	at    int
	right node
}

type callNode struct { // niladic user function
	fn *Function
	at int
}

type quadInNode struct { // ⎕ or ⍞ input
	at    int
	quote bool // ⍞
}

type quadOutNode struct { // ⎕← or ⍞← output
	at    int
	quote bool // ⍞
	right node
}

type branchNode struct {
	at    int
	right node
}

type monadNode struct {
	fn    *fnNode
	right node
}

type dyadNode struct {
	fn          *fnNode
	left, right node
}

type fnKind int

const (
	fnPrim   fnKind = iota // primitive function
	fnReduce               // f/ or f⌿
	fnScan                 // f\ or f⍀
	fnInner                // f.g
	fnOuter                // ∘.f
	fnUser                 // user-defined function
)

type fnNode struct {
	kind  fnKind
	glyph string // f
	g     string // g of inner product
	first bool   // operator applies along the first axis: ⌿ ⍀
	axis  node   // [k], or nil
	user  *Function
	at    int
}

func (n *constNode) pos() int  { return n.at }
func (n *varNode) pos() int    { return n.at }
func (n *assignNode) pos() int { return n.at }
func (n *indexNode) pos() int  { return n.target.pos() }
func (n *indexAssignNode) pos() int {
	return n.at
}
func (n *callNode) pos() int    { return n.at }
func (n *quadInNode) pos() int  { return n.at }
func (n *quadOutNode) pos() int { return n.at }
func (n *branchNode) pos() int  { return n.at }
func (n *monadNode) pos() int   { return n.fn.at }
func (n *dyadNode) pos() int    { return n.left.pos() }

type parser struct {
	toks []Token
	i    int
	end  int                  // rune offset of end of line
	fns  map[string]*Function // user functions
}

// parse parses one statement; returns nil if there are no tokens.
func (in *Interp) parse(toks []Token, end int) (node, error) {
	if len(toks) == 0 {
		return nil, nil
	}
	for _, t := range toks {
		if t.Kind == Name && isSysName(t.Text) && !sysVars[t.Text] {
			return nil, errorAt(SyntaxError, t.Pos)
		}
	}
	p := &parser{toks: toks, end: end, fns: in.fns}
	var n node
	var err error
	if t := toks[0]; isSpecial(t, "→") {
		p.i++
		var r node
		if r, err = p.expr(); err == nil {
			n = &branchNode{at: t.Pos, right: r}
		}
	} else {
		n, err = p.expr()
	}
	if err != nil {
		return nil, err
	}
	if t, ok := p.peek(); ok {
		if t.Kind == Special && t.Text == ")" {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		return nil, errorAt(SyntaxError, t.Pos)
	}
	return n, nil
}

func (p *parser) peek() (Token, bool) {
	if p.i < len(p.toks) {
		return p.toks[p.i], true
	}
	return Token{}, false
}

func (p *parser) peekAt(k int) (Token, bool) {
	if p.i+k < len(p.toks) {
		return p.toks[p.i+k], true
	}
	return Token{}, false
}

// here returns the position of the next token, or end of line.
func (p *parser) here() int {
	if t, ok := p.peek(); ok {
		return t.Pos
	}
	return p.end
}

func isSpecial(t Token, s string) bool { return t.Kind == Special && t.Text == s }

// atExprEnd reports whether the current expression ends here.
func (p *parser) atExprEnd() bool {
	t, ok := p.peek()
	return !ok || isSpecial(t, ")") || isSpecial(t, "]") || isSpecial(t, ";")
}

// userFn returns the user function named by t, if any.
func (p *parser) userFn(t Token) *Function {
	if t.Kind != Name {
		return nil
	}
	return p.fns[t.Text]
}

func (p *parser) atFn() bool {
	t, ok := p.peek()
	if f := p.userFn(t); f != nil {
		return f.Arity() > 0
	}
	if ok && isSpecial(t, "∘") {
		next, ok := p.peekAt(1)
		return ok && isSpecial(next, ".")
	}
	return ok && t.Kind == Prim
}

// isScalarDyad reports whether g names a dyadic scalar function.
func isScalarDyad(g string) bool {
	sf := scalarFns[g]
	return sf != nil && sf.dyad != nil
}

func (p *parser) expr() (node, error) {
	if p.atExprEnd() {
		return nil, errorAt(SyntaxError, p.here())
	}
	t, _ := p.peek()
	if next, ok := p.peekAt(1); ok && t.Kind == Name && isSpecial(next, "←") {
		if p.userFn(t) != nil {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		p.i += 2
		r, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &assignNode{name: t.Text, at: t.Pos, right: r}, nil
	}
	if p.atFn() {
		f, err := p.fn()
		if err != nil {
			return nil, err
		}
		if f.kind == fnPrim && isSlash(f.glyph) {
			return nil, errorAt(SyntaxError, f.at)
		}
		if f.kind == fnUser && f.user.Arity() == 2 {
			return nil, errorAt(ValenceError, f.at)
		}
		r, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &monadNode{fn: f, right: r}, nil
	}
	if next, ok := p.peekAt(1); ok && (isSpecial(t, "⎕") || isSpecial(t, "⍞")) && isSpecial(next, "←") {
		p.i += 2
		r, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &quadOutNode{at: t.Pos, quote: t.Text == "⍞", right: r}, nil
	}
	if t.Kind == Name && p.userFn(t) == nil && p.indexAssignAhead() {
		p.i += 2 // name [
		idx, err := p.index()
		if err != nil {
			return nil, err
		}
		p.i++ // ←
		r, err := p.expr()
		if err != nil {
			return nil, err
		}
		return &indexAssignNode{name: t.Text, idx: idx, at: t.Pos, right: r}, nil
	}
	left, err := p.operand()
	if err != nil {
		return nil, err
	}
	if p.atExprEnd() {
		return left, nil
	}
	if !p.atFn() {
		return nil, errorAt(SyntaxError, p.here())
	}
	f, err := p.fn()
	if err != nil {
		return nil, err
	}
	if f.kind == fnUser && f.user.Arity() != 2 {
		return nil, errorAt(ValenceError, f.at)
	}
	r, err := p.expr()
	if err != nil {
		return nil, err
	}
	return &dyadNode{fn: f, left: left, right: r}, nil
}

// indexAssignAhead reports whether "name [ ... ] ←" starts here.
func (p *parser) indexAssignAhead() bool {
	if t, ok := p.peekAt(1); !ok || !isSpecial(t, "[") {
		return false
	}
	depth := 0
	for k := p.i + 1; k < len(p.toks); k++ {
		switch t := p.toks[k]; {
		case isSpecial(t, "["):
			depth++
		case isSpecial(t, "]"):
			depth--
			if depth == 0 {
				return k+1 < len(p.toks) && isSpecial(p.toks[k+1], "←")
			}
		}
	}
	return false
}

func (p *parser) fn() (*fnNode, error) {
	t, _ := p.peek()
	p.i++
	if f := p.userFn(t); f != nil {
		return &fnNode{kind: fnUser, user: f, at: t.Pos}, nil
	}
	if isSpecial(t, "∘") {
		p.i++ // .
		f, ok := p.peek()
		if !ok || f.Kind != Prim || !isScalarDyad(f.Text) {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		p.i++
		return &fnNode{kind: fnOuter, glyph: f.Text, at: t.Pos}, nil
	}
	if dot, ok := p.peek(); ok && isSpecial(dot, ".") {
		g, ok := p.peekAt(1)
		if !ok || g.Kind != Prim || !isScalarDyad(t.Text) || !isScalarDyad(g.Text) {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		p.i += 2
		return &fnNode{kind: fnInner, glyph: t.Text, g: g.Text, at: t.Pos}, nil
	}
	f := &fnNode{kind: fnPrim, glyph: t.Text, at: t.Pos}
	if op, ok := p.peek(); ok && op.Kind == Prim && isSlash(op.Text) && !isSlash(t.Text) {
		p.i++
		if !isScalarDyad(t.Text) {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		f.kind = fnReduce
		if op.Text == "\\" || op.Text == "⍀" {
			f.kind = fnScan
		}
		f.first = op.Text == "⌿" || op.Text == "⍀"
	} else if isSlash(t.Text) {
		f.first = t.Text == "⌿" || t.Text == "⍀"
	}
	if b, ok := p.peek(); ok && isSpecial(b, "[") {
		p.i++
		idx, err := p.index()
		if err != nil {
			return nil, err
		}
		if len(idx) != 1 || idx[0] == nil {
			return nil, errorAt(AxisError, b.Pos)
		}
		f.axis = idx[0]
	}
	return f, nil
}

// isSlash reports whether g is one of the operators / ⌿ \ ⍀, which are
// also the functions compress and expand.
func isSlash(g string) bool { return g == "/" || g == "⌿" || g == "\\" || g == "⍀" }

func (p *parser) operand() (node, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		t, ok := p.peek()
		if !ok || !isSpecial(t, "[") {
			return n, nil
		}
		p.i++
		idx, err := p.index()
		if err != nil {
			return nil, err
		}
		n = &indexNode{target: n, idx: idx, at: t.Pos}
	}
}

// index parses index items after "[" up to and including "]".
func (p *parser) index() ([]node, error) {
	var idx []node
	for {
		t, ok := p.peek()
		if !ok {
			return nil, errorAt(SyntaxError, p.end)
		}
		var item node
		if !isSpecial(t, ";") && !isSpecial(t, "]") {
			var err error
			if item, err = p.expr(); err != nil {
				return nil, err
			}
			if t, ok = p.peek(); !ok {
				return nil, errorAt(SyntaxError, p.end)
			}
		}
		idx = append(idx, item)
		p.i++
		switch {
		case isSpecial(t, "]"):
			return idx, nil
		case !isSpecial(t, ";"):
			return nil, errorAt(SyntaxError, t.Pos)
		}
	}
}

func (p *parser) primary() (node, error) {
	t, _ := p.peek()
	switch {
	case t.Kind == Number:
		p.i++
		v := Vector(t.Nums)
		if len(t.Nums) == 1 {
			v = Scalar(t.Nums[0])
		}
		return &constNode{val: v, at: t.Pos}, nil
	case t.Kind == String:
		p.i++
		v := Chars(t.Text)
		if v.Size() == 1 {
			v.Shape = nil
		}
		return &constNode{val: v, at: t.Pos}, nil
	case t.Kind == Name:
		p.i++
		if f := p.userFn(t); f != nil { // niladic; others are taken by atFn
			return &callNode{fn: f, at: t.Pos}, nil
		}
		return &varNode{name: t.Text, at: t.Pos}, nil
	case isSpecial(t, "⎕") || isSpecial(t, "⍞"):
		p.i++
		return &quadInNode{at: t.Pos, quote: t.Text == "⍞"}, nil
	case isSpecial(t, "("):
		p.i++
		n, err := p.expr()
		if err != nil {
			return nil, err
		}
		if c, ok := p.peek(); !ok || !isSpecial(c, ")") {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		p.i++
		return n, nil
	}
	return nil, errorAt(SyntaxError, t.Pos)
}
