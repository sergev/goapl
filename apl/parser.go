package apl

// Grammar, read left to right but binding right to left:
//
//	stmt    := → expr | expr
//	expr    := fn expr | name ← expr | name [index] ← expr | ⎕ ← expr
//	         | ⍞ ← expr | ( name {name} ) ← expr | ( select ) ← expr
//	         | strand [fn expr]
//	strand  := operand {operand}
//	operand := (number | string | name | niladic | ⎕ | ⍞ | ( expr )) {[index]}
//	index   := [expr] {; [expr]}
//	fn      := ((prim | prim op) [axis] | prim . prim | ∘ . prim | user function) {¨}
//	op      := / | ⌿ | \ | ⍀
//	axis    := [expr]
//	select  := expr of selection functions and indexing, with a name at its right

type node interface{ pos() int }

type constNode struct {
	val    *Array
	at     int
	strand bool // numeric vector constant: separate items in a strand
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

type strandNode struct { // A B (C D)
	items []node
	at    int
}

type strandAssignNode struct { // (A B)←X
	names []string
	at    int
	right node
}

type selAssignNode struct { // (select V)←X
	name  string
	sel   node // selection applied to a posNode for V
	at    int
	right node
}

type posNode struct { // positions 1..n of the elements of variable name
	name string
	at   int
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
	fnPrim    fnKind = iota // primitive function
	fnReduce                // f/ or f⌿
	fnScan                  // f\ or f⍀
	fnInner                 // f.g
	fnOuter                 // ∘.f
	fnUser                  // user-defined function
	fnEach                  // f¨
	fnSys                   // system function ⎕FX ⎕CR ⎕NC ⎕EX
	fnOperand               // operand of the operator being defined
	fnDefOp                 // function derived from a defined operator
)

type fnNode struct {
	kind  fnKind
	glyph string    // primitive; f of a scalar f/ or f\; name of fnSys, fnOperand
	first bool      // operator applies along the first axis: ⌿ ⍀
	axis  node      // [k], or nil
	left  *fnNode   // f of f/ f\ f.g f¨ and of a defined operator
	right *fnNode   // g of f.g ∘.g and of a dyadic defined operator
	user  *Function // user function or defined operator
	at    int
}

// scalarOperands reports whether the operands of an operator are all
// primitive dyadic scalar functions, for which there are fast paths.
func (f *fnNode) scalarOperands() bool {
	ok := func(g *fnNode) bool {
		return g == nil || g.kind == fnPrim && g.axis == nil && isScalarDyad(g.glyph)
	}
	return ok(f.left) && ok(f.right)
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
func (n *strandNode) pos() int  { return n.at }
func (n *strandAssignNode) pos() int {
	return n.at
}
func (n *selAssignNode) pos() int { return n.at }
func (n *posNode) pos() int       { return n.at }
func (n *dyadNode) pos() int      { return n.left.pos() }

type parser struct {
	toks []Token
	i    int
	end  int                  // rune offset of end of line
	fns  map[string]*Function // user functions and operators
	ops  map[string]bool      // operand names of the operator being run
}

// sysFns are the system functions.
var sysFns = map[string]bool{"⎕FX": true, "⎕CR": true, "⎕NC": true, "⎕EX": true, "⎕EA": true}

// parse parses one statement; returns nil if there are no tokens.
// Names of op's operands denote functions.
func (in *Interp) parse(toks []Token, end int, op *Function) (node, error) {
	if len(toks) == 0 {
		return nil, nil
	}
	for _, t := range toks {
		if t.Kind == Name && isSysName(t.Text) && !sysVars[t.Text] && !sysFns[t.Text] {
			return nil, errorAt(SyntaxError, t.Pos)
		}
		if isSpecial(t, "⋄") {
			return nil, errorAt(SyntaxError, t.Pos)
		}
	}
	p := &parser{toks: toks, end: end, fns: in.fns}
	if op != nil {
		p.ops = map[string]bool{op.OpLeft: true, op.OpRight: true}
		delete(p.ops, "")
	}
	var n node
	var err error
	if t := toks[0]; isSpecial(t, "→") && len(toks) == 1 {
		p.i++
		n = &branchNode{at: t.Pos} // escape
	} else if isSpecial(t, "→") {
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

// userFn returns the user function (not operator) named by t, if any.
func (p *parser) userFn(t Token) *Function {
	if t.Kind != Name || p.ops[t.Text] {
		return nil
	}
	if f := p.fns[t.Text]; f != nil && !f.IsOperator() {
		return f
	}
	return nil
}

// userOp returns the defined operator named by t, if any.
func (p *parser) userOp(t Token) *Function {
	if t.Kind != Name || p.ops[t.Text] {
		return nil
	}
	if f := p.fns[t.Text]; f != nil && f.IsOperator() {
		return f
	}
	return nil
}

// isFnName reports whether t names a function that takes arguments.
func (p *parser) isFnName(t Token) bool {
	if t.Kind != Name {
		return false
	}
	if f := p.userFn(t); f != nil {
		return f.Arity() > 0
	}
	return p.ops[t.Text] || sysFns[t.Text]
}

func (p *parser) atFn() bool {
	t, ok := p.peek()
	if !ok {
		return false
	}
	if p.isFnName(t) {
		return true
	}
	if isSpecial(t, "∘") {
		next, ok := p.peekAt(1)
		return ok && isSpecial(next, ".")
	}
	return t.Kind == Prim
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
	if isSpecial(t, "(") {
		if n, ok, err := p.parenAssign(); ok || err != nil {
			return n, err
		}
	}
	if p.atFn() {
		f, err := p.fn()
		if err != nil {
			return nil, err
		}
		if f.kind == fnPrim && isSlash(f.glyph) {
			return nil, errorAt(SyntaxError, f.at)
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
	left, err := p.strand()
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
	f, err := p.fnBase()
	for err == nil {
		t, ok := p.peek()
		if !ok {
			break
		}
		if t.Kind == Prim && t.Text == "¨" {
			p.i++
			f = &fnNode{kind: fnEach, left: f, at: t.Pos}
			continue
		}
		if t.Kind == Prim && isSlash(t.Text) && f.kind != fnPrim {
			p.i++
			kind := fnReduce
			if t.Text == "\\" || t.Text == "⍀" {
				kind = fnScan
			}
			f = &fnNode{kind: kind, left: f, at: t.Pos, first: t.Text == "⌿" || t.Text == "⍀"}
			continue
		}
		op := p.userOp(t)
		if op == nil {
			break
		}
		p.i++
		f = &fnNode{kind: fnDefOp, user: op, left: f, at: t.Pos}
		if op.OpRight != "" {
			f.right, err = p.operandFn()
		}
	}
	return f, err
}

// operandFn parses a primitive, user, system or operand function.
func (p *parser) operandFn() (*fnNode, error) {
	t, ok := p.peek()
	if !ok {
		return nil, errorAt(SyntaxError, p.end)
	}
	p.i++
	switch {
	case p.userFn(t) != nil && p.userFn(t).Arity() > 0:
		return &fnNode{kind: fnUser, user: p.userFn(t), at: t.Pos}, nil
	case t.Kind == Name && p.ops[t.Text]:
		return &fnNode{kind: fnOperand, glyph: t.Text, at: t.Pos}, nil
	case t.Kind == Name && sysFns[t.Text]:
		return &fnNode{kind: fnSys, glyph: t.Text, at: t.Pos}, nil
	case t.Kind == Prim && t.Text != "¨":
		return &fnNode{kind: fnPrim, glyph: t.Text, at: t.Pos}, nil
	}
	return nil, errorAt(SyntaxError, t.Pos)
}

func (p *parser) fnBase() (*fnNode, error) {
	t, _ := p.peek()
	if isSpecial(t, "∘") {
		p.i += 2 // ∘ .
		g, err := p.operandFn()
		if err != nil {
			return nil, err
		}
		return &fnNode{kind: fnOuter, glyph: g.glyph, right: g, at: t.Pos}, nil
	}
	f, err := p.operandFn()
	if err != nil {
		return nil, err
	}
	if dot, ok := p.peek(); ok && isSpecial(dot, ".") {
		p.i++
		g, err := p.operandFn()
		if err != nil {
			return nil, err
		}
		return &fnNode{kind: fnInner, glyph: f.glyph, left: f, right: g, at: t.Pos}, nil
	}
	if f.kind == fnPrim && isSlash(f.glyph) {
		f.first = f.glyph == "⌿" || f.glyph == "⍀"
	} else if op, ok := p.peek(); ok && op.Kind == Prim && isSlash(op.Text) {
		p.i++
		kind := fnReduce
		if op.Text == "\\" || op.Text == "⍀" {
			kind = fnScan
		}
		f = &fnNode{kind: kind, glyph: f.glyph, left: f, at: t.Pos,
			first: op.Text == "⌿" || op.Text == "⍀"}
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

// startsOperand reports whether an operand of a strand starts here.
func (p *parser) startsOperand() bool {
	t, ok := p.peek()
	if !ok {
		return false
	}
	next, _ := p.peekAt(1)
	switch {
	case t.Kind == Number, t.Kind == String, isSpecial(t, "("):
		return true
	case t.Kind == Name:
		return !p.isFnName(t) && p.userOp(t) == nil && !isSpecial(next, "←")
	case isSpecial(t, "⎕"), isSpecial(t, "⍞"):
		return !isSpecial(next, "←")
	}
	return false
}

// strand parses juxtaposed operands; several make a vector whose items
// are their values. The numbers of a numeric constant are separate items.
func (p *parser) strand() (node, error) {
	var items []node
	at := p.here()
	for len(items) == 0 || p.startsOperand() {
		n, err := p.operand()
		if err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	if len(items) == 1 {
		return items[0], nil
	}
	var flat []node
	for _, n := range items {
		if c, ok := n.(*constNode); ok && c.strand {
			for _, v := range c.val.Data {
				flat = append(flat, &constNode{val: Scalar(v), at: c.at})
			}
			continue
		}
		flat = append(flat, n)
	}
	return &strandNode{items: flat, at: at}, nil
}

// parenAssign parses "( names ) ← expr" and "( selection ) ← expr";
// ok is false if no such assignment starts here.
func (p *parser) parenAssign() (node, bool, error) {
	depth, m := 0, -1
	for k := p.i; k < len(p.toks) && m < 0; k++ {
		if isSpecial(p.toks[k], "(") {
			depth++
		} else if isSpecial(p.toks[k], ")") {
			if depth--; depth == 0 {
				m = k
			}
		}
	}
	if m < 0 || m+1 >= len(p.toks) || !isSpecial(p.toks[m+1], "←") {
		return nil, false, nil
	}
	start := p.toks[p.i].Pos
	var names []string
	for k := p.i + 1; k < m; k++ {
		t := p.toks[k]
		if t.Kind != Name || p.userFn(t) != nil {
			names = nil
			break
		}
		names = append(names, t.Text)
	}
	if len(names) > 0 {
		p.i = m + 2
		r, err := p.expr()
		if err != nil {
			return nil, true, err
		}
		return &strandAssignNode{names: names, at: start, right: r}, true, nil
	}
	p.i++
	sel, err := p.expr()
	if err != nil {
		return nil, true, err
	}
	if p.i != m {
		return nil, true, errorAt(SyntaxError, p.here())
	}
	sel, name, err := selection(sel)
	if err != nil {
		return nil, true, err
	}
	p.i = m + 2
	r, err := p.expr()
	if err != nil {
		return nil, true, err
	}
	return &selAssignNode{name: name, sel: sel, at: start, right: r}, true, nil
}

// selectFns may appear in a selective assignment: [monadic, dyadic].
var selectFns = map[string][2]bool{
	"↑": {true, true}, "↓": {false, true}, ",": {true, true}, "⍪": {false, true},
	"⌽": {true, true}, "⊖": {true, true}, "⍉": {true, true}, "/": {false, true},
	"⌿": {false, true}, "\\": {false, true}, "⍀": {false, true}, "⍴": {false, true},
}

// selection replaces the variable at the right end of a selective
// assignment target by its positions, and gives the variable's name.
func selection(n node) (node, string, error) {
	switch n := n.(type) {
	case *varNode:
		return &posNode{name: n.name, at: n.at}, n.name, nil
	case *indexNode:
		t, name, err := selection(n.target)
		n.target = t
		return n, name, err
	case *monadNode:
		if n.fn.kind != fnPrim || !selectFns[n.fn.glyph][0] {
			return nil, "", errorAt(SyntaxError, n.fn.at)
		}
		r, name, err := selection(n.right)
		n.right = r
		return n, name, err
	case *dyadNode:
		if n.fn.kind != fnPrim || !selectFns[n.fn.glyph][1] {
			return nil, "", errorAt(SyntaxError, n.fn.at)
		}
		r, name, err := selection(n.right)
		n.right = r
		return n, name, err
	}
	return nil, "", errorAt(SyntaxError, n.pos())
}

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
		return &constNode{val: v, at: t.Pos, strand: len(t.Nums) > 1}, nil
	case t.Kind == String:
		p.i++
		v := Chars(t.Text)
		if v.Size() == 1 {
			v.Shape = nil
		}
		return &constNode{val: v, at: t.Pos}, nil
	case t.Kind == Name:
		if p.userOp(t) != nil {
			return nil, errorAt(SyntaxError, t.Pos)
		}
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
		if c, ok := n.(*constNode); ok && c.strand {
			n = &constNode{val: c.val, at: c.at} // (1 2) is one item
		}
		return n, nil
	}
	return nil, errorAt(SyntaxError, t.Pos)
}
