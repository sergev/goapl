package apl

// Arrays are treated as immutable once built; results never alias
// mutable state, so values may be shared between variables.

func (in *Interp) eval(n node) (*Array, error) {
	switch n := n.(type) {
	case *constNode:
		return n.val, nil
	case *varNode:
		v, ok := in.vars[n.name]
		if !ok {
			return nil, errorAt(100, n.at)
		}
		return v, nil
	case *assignNode:
		v, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		in.vars[n.name] = v
		return v, nil
	case *indexNode:
		idx, err := in.evalIndex(n.idx)
		if err != nil {
			return nil, err
		}
		x, err := in.eval(n.target)
		if err != nil {
			return nil, err
		}
		r, err := index(x, idx)
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
		x, ok := in.vars[n.name]
		if !ok {
			return nil, errorAt(100, n.at)
		}
		r, err := indexAssign(x, idx, v)
		if err != nil {
			return nil, setPos(err, n.at)
		}
		in.vars[n.name] = r
		return v, nil
	case *monadNode:
		x, err := in.eval(n.right)
		if err != nil {
			return nil, err
		}
		r, err := monad(n.fn, x)
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
		r, err := dyad(n.fn, a, b)
		return r, setPos(err, n.fn.at)
	}
	panic("eval: unknown node")
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

// setPos attaches a position to a positionless error.
func setPos(err error, pos int) error {
	if e, ok := err.(*Error); ok && e.Pos == NoPos {
		e.Pos = pos
	}
	return err
}

func monad(f *fnNode, x *Array) (*Array, error) {
	switch f.kind {
	case fnReduce:
		return reduce(scalarFns[f.glyph].dyad, x)
	case fnInner, fnOuter:
		return nil, newError(13)
	}
	if sf := scalarFns[f.glyph]; sf != nil {
		if sf.monad == nil {
			return nil, newError(74)
		}
		return applyMonad(sf.monad, x)
	}
	if sf := structFns[f.glyph]; sf != nil && sf.monad != nil {
		return sf.monad(x)
	}
	return nil, newError(10)
}

func dyad(f *fnNode, a, b *Array) (*Array, error) {
	switch f.kind {
	case fnReduce:
		return nil, newError(73)
	case fnInner:
		return inner(scalarFns[f.glyph].dyad, scalarFns[f.g].dyad, a, b)
	case fnOuter:
		return outer(scalarFns[f.glyph].dyad, a, b)
	}
	if sf := scalarFns[f.glyph]; sf != nil && sf.dyad != nil {
		return applyDyad(sf.dyad, a, b)
	}
	if sf := structFns[f.glyph]; sf != nil && sf.dyad != nil {
		return sf.dyad(a, b)
	}
	return nil, newError(10)
}
