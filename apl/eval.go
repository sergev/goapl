package apl

import "strings"

// Arrays are treated as immutable once built; results never alias
// mutable state, so values may be shared between variables.

// eval evaluates n, which must produce a value.
func (in *Interp) eval(n node) (*Array, error) {
	v, err := in.evalNode(n)
	if err == nil && v == nil {
		err = errorAt(ValueError, n.pos())
	}
	return v, err
}

// evalNode evaluates n; a call of a function without result gives nil.
func (in *Interp) evalNode(n node) (*Array, error) {
	switch n := n.(type) {
	case *callNode:
		r, err := in.call(n.fn, nil, nil)
		return r, setPos(err, n.at)
	case *quadInNode:
		if n.quote {
			return in.quoteInput()
		}
		return in.quadInput()
	case *quadOutNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		if n.quote {
			return v, in.quoteOutput(v)
		}
		return v, in.print(v)
	case *constNode:
		return n.val, nil
	case *varNode:
		v, ok := in.getVar(n.name)
		if !ok {
			return nil, errorAt(ValueError, n.at)
		}
		return v, nil
	case *assignNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		return v, setPos(in.setVar(n.name, v), n.at)
	case *indexNode:
		idx, err := in.evalIndex(n.idx)
		if err != nil {
			return nil, err
		}
		x, err := in.eval(n.target)
		if err != nil {
			return nil, err
		}
		r, err := index(&in.env, x, idx)
		return r, setPos(err, n.at)
	case *indexAssignNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		idx, err := in.evalIndex(n.idx)
		if err != nil {
			return nil, err
		}
		x, ok := in.getVar(n.name)
		if !ok {
			return nil, errorAt(ValueError, n.at)
		}
		r, err := indexAssign(&in.env, x, idx, v)
		if err == nil {
			err = in.setVar(n.name, r)
		}
		return v, setPos(err, n.at)
	case *monadNode:
		x, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		r, err := in.applyFn(n.fn, nil, x)
		return r, setPos(err, n.fn.at)
	case *dyadNode:
		b, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		a, err := in.eval(n.left)
		if err != nil {
			return nil, err
		}
		r, err := in.applyFn(n.fn, a, b)
		return r, setPos(err, n.fn.at)
	case *strandNode:
		items := make([]Value, len(n.items))
		for i := len(items) - 1; i >= 0; i-- {
			v, err := in.eval(n.items[i])
			if err != nil {
				return nil, err
			}
			items[i] = itemValue(v)
		}
		return General([]int{len(items)}, items, items[0]), nil
	case *strandAssignNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		if v.Rank() > 1 {
			return nil, errorAt(RankError, n.at)
		}
		if v.Size() != 1 && v.Size() != len(n.names) {
			return nil, errorAt(LengthError, n.at)
		}
		for i, name := range n.names {
			x := item(v.At(min(i, v.Size()-1)))
			if err := in.setVar(name, x); err != nil {
				return nil, setPos(err, n.at)
			}
		}
		return v, nil
	case *selAssignNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		return v, setPos(in.selectiveAssign(n, v), n.at)
	case *posNode:
		x, ok := in.getVar(n.name)
		if !ok {
			return nil, errorAt(ValueError, n.at)
		}
		r := &Array{Shape: x.Shape, Data: make([]float64, x.Size())}
		for i := range r.Data {
			r.Data[i] = float64(i + 1)
		}
		return r, nil
	}
	panic("eval: unknown node")
}

// selectiveAssign is (sel V)←X: the elements of V that sel selects
// get the items of X. Fill positions of the selection are skipped.
func (in *Interp) selectiveAssign(n *selAssignNode, v *Array) error {
	pos, err := in.eval(n.sel)
	if err != nil {
		return err
	}
	if !pos.IsNumeric() {
		return newError(DomainError)
	}
	if v.Size() != 1 && !sameShape(v, pos) {
		if v.Rank() != pos.Rank() {
			return newError(RankError)
		}
		return newError(LengthError)
	}
	x, _ := in.getVar(n.name)
	elems := append([]Value(nil), x.elems()...)
	for i, p := range pos.Data {
		if p >= 1 {
			elems[int(p)-1] = v.At(min(i, v.Size()-1))
		}
	}
	return in.setVar(n.name, General(x.Shape, elems, x.Fill()))
}

// applyFn applies f to B, or to A and B; a is nil for a monadic call.
func (in *Interp) applyFn(f *fnNode, a, b *Array) (*Array, error) {
	switch f.kind {
	case fnUser:
		if f.user.Arity() == 0 || a != nil && f.user.Arity() == 1 {
			return nil, newError(ValenceError)
		}
		return in.call(f.user, a, b)
	case fnEach:
		return in.each(f.left, a, b)
	case fnSys:
		if a != nil {
			return nil, newError(ValenceError)
		}
		return in.sysFn(f.glyph, b)
	case fnOperand:
		return in.applyFn(in.operand(f.glyph), a, b)
	case fnDefOp:
		return in.callOp(f, a, b)
	}
	ax, err := in.evalAxis(f)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return in.monad(f, ax, b)
	}
	return in.dyad(f, ax, a, b)
}

// each is f¨B or A f¨B: f applied to each item, or to corresponding
// items with a scalar or one-item argument extended.
func (in *Interp) each(f *fnNode, a, b *Array) (*Array, error) {
	shape, n := b.Shape, b.Size()
	if a != nil {
		switch {
		case sameShape(a, b):
		case a.Size() == 1 && (a.IsScalar() || !b.IsScalar()):
		case b.Size() == 1:
			shape, n = a.Shape, a.Size()
		case a.Rank() != b.Rank():
			return nil, newError(RankError)
		default:
			return nil, newError(LengthError)
		}
	}
	at := func(x *Array, i int) *Array {
		if x.Size() == 1 {
			return item(x.At(0))
		}
		return item(x.At(i))
	}
	r := make([]Value, n)
	for i := range r {
		var x *Array
		if a != nil {
			x = at(a, i)
		}
		v, err := in.applyFn(f, x, at(b, i))
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, newError(ValueError)
		}
		r[i] = itemValue(v)
	}
	return General(shape, r, b.Fill()), nil
}

// run executes a statement, printing its value unless it is an
// assignment. It returns the branch target, or -1 for no branch.
func (in *Interp) run(n node) (int, error) {
	if n == nil {
		return -1, nil
	}
	if b, ok := n.(*branchNode); ok {
		v, err := in.eval(b.right)
		if err != nil {
			return -1, err
		}
		if v.Size() == 0 {
			return -1, nil
		}
		d, err := v.numeric()
		if err != nil {
			return -1, setPos(err, b.at)
		}
		t, ok := toInt(d[0])
		if !ok {
			return -1, errorAt(DomainError, b.at)
		}
		return max(t, 0), nil
	}
	v, err := in.evalNode(n)
	if err != nil || v == nil {
		return -1, err
	}
	switch n.(type) {
	case *assignNode, *indexAssignNode, *quadOutNode, *strandAssignNode, *selAssignNode:
		return -1, nil
	}
	return -1, in.print(v)
}

// evalIndex evaluates index items right to left; nil stays nil.
func (in *Interp) evalIndex(items []node) ([]*Array, error) {
	idx := make([]*Array, len(items))
	for i := len(items) - 1; i >= 0; i-- {
		if items[i] == nil {
			continue
		}
		v, err := in.eval(items[i])
		if err != nil {
			return nil, err
		}
		idx[i] = v
	}
	return idx, nil
}

// setPos attaches a position to an error not yet located in a line.
func setPos(err error, pos int) error {
	if e, ok := err.(*Error); ok && e.Pos == NoPos && e.Src == "" {
		e.Pos = pos
	}
	return err
}

// evalAxis evaluates the axis of f, or gives nil.
func (in *Interp) evalAxis(f *fnNode) (*Array, error) {
	if f.axis == nil {
		return nil, nil
	}
	return in.eval(f.axis)
}

// axisIndex converts axis ax (from ⎕IO) of an array of rank r to an
// index from 0; without ax it gives the first or the last axis.
func (e *env) axisIndex(ax *Array, r int, first bool) (int, error) {
	if ax == nil {
		if first {
			return 0, nil
		}
		return r - 1, nil
	}
	if ax.Size() != 1 || ax.Rank() > 1 || !ax.IsNumeric() {
		return 0, newError(AxisError)
	}
	k, ok := toInt(ax.Data[0])
	k -= e.io
	if !ok || k < 0 || k >= r {
		return 0, newError(AxisError)
	}
	return k, nil
}

func (in *Interp) monad(f *fnNode, ax, x *Array) (*Array, error) {
	switch f.kind {
	case fnReduce, fnScan:
		if x.IsScalar() && ax == nil {
			return x, nil
		}
		k, err := in.axisIndex(ax, x.Rank(), f.first)
		if err != nil {
			return nil, err
		}
		if !f.scalarOperands() {
			fn := in.dyadicOf(f.left)
			if f.kind == fnScan {
				return scanGen(fn, x, k)
			}
			return windowsGen(fn, x, k, x.Shape[k], false)
		}
		if f.kind == fnScan {
			return scanAxis(&in.env, f.glyph, x, k)
		}
		return reduceAxis(&in.env, f.glyph, x, k)
	case fnInner, fnOuter:
		return nil, newError(SyntaxError)
	}
	if f.glyph == "⍎" && ax == nil {
		return in.execute(x)
	}
	sf, st := scalarFns[f.glyph], structFns[f.glyph]
	if st != nil && st.monadAxis != nil {
		return st.monadAxis(&in.env, ax, x)
	}
	if ax != nil {
		return nil, newError(AxisError)
	}
	switch {
	case sf != nil && sf.monad != nil:
		return applyMonad(&in.env, sf.monad, x)
	case st != nil && st.monad != nil:
		return st.monad(&in.env, x)
	case sf != nil || st != nil:
		return nil, newError(ValenceError)
	}
	return nil, newError(NonceError)
}

func (in *Interp) dyad(f *fnNode, ax, a, b *Array) (*Array, error) {
	switch f.kind {
	case fnReduce:
		if b.IsScalar() {
			b = &Array{Shape: []int{1}, Data: b.Data, Elems: b.Elems}
		}
		k, err := in.axisIndex(ax, b.Rank(), f.first)
		if err != nil {
			return nil, err
		}
		if !f.scalarOperands() {
			w, rev, err := window(a, b.Shape[k])
			if err != nil {
				return nil, err
			}
			return windowsGen(in.dyadicOf(f.left), b, k, w, rev)
		}
		return nwise(&in.env, f.glyph, a, b, k)
	case fnScan:
		return nil, newError(ValenceError)
	case fnInner:
		if !f.scalarOperands() {
			return innerGen(in.dyadicOf(f.left), in.dyadicOf(f.right), a, b)
		}
		return inner(&in.env, f.left.glyph, f.right.glyph, a, b)
	case fnOuter:
		if !f.scalarOperands() {
			return outerGen(in.dyadicOf(f.right), a, b)
		}
		return outer(&in.env, scalarFns[f.right.glyph], a, b)
	}
	if f.glyph == "⍎" {
		return nil, newError(ValenceError)
	}
	sf, st := scalarFns[f.glyph], structFns[f.glyph]
	if st != nil && st.dyadAxis != nil {
		return st.dyadAxis(&in.env, ax, a, b)
	}
	if ax != nil {
		return nil, newError(AxisError)
	}
	switch {
	case sf != nil && sf.dyad != nil:
		return applyDyad(&in.env, sf, a, b)
	case st != nil && st.dyad != nil:
		return st.dyad(&in.env, a, b)
	case sf != nil || st != nil:
		return nil, newError(ValenceError)
	}
	return nil, newError(NonceError)
}

// execute is ⍎X: runs the character vector X as a statement and gives
// its value, or nil if it has none.
func (in *Interp) execute(x *Array) (*Array, error) {
	if x.Rank() > 1 {
		return nil, newError(RankError)
	}
	if x.Size() > 0 && !x.IsCharacter() {
		return nil, newError(DomainError)
	}
	var b strings.Builder
	for i := range x.Size() {
		b.WriteRune(rune(x.At(i).(Char)))
	}
	text := b.String()
	toks, err := Lex(text)
	var nodes []node
	if err == nil {
		nodes, err = in.parseLine(toks, len([]rune(text)), nil)
	}
	var v *Array
	for i, n := range nodes {
		if _, ok := n.(*branchNode); ok {
			err = errorAt(DomainError, 0)
		}
		if err != nil {
			break
		}
		if i < len(nodes)-1 {
			_, err = in.run(n)
		} else if n != nil {
			v, err = in.evalNode(n)
		}
	}
	if e, ok := err.(*Error); ok && e.Src == "" {
		e.Src, e.Where = text, "⍎"
	}
	return v, err
}
