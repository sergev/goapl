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
	}
	if sf := scalarFns[f.glyph]; sf != nil {
		if sf.monad == nil {
			return nil, newError(74)
		}
		return applyMonad(sf.monad, x)
	}
	return nil, newError(10)
}

func dyad(f *fnNode, a, b *Array) (*Array, error) {
	if f.kind == fnReduce {
		return nil, newError(73)
	}
	if sf := scalarFns[f.glyph]; sf != nil && sf.dyad != nil {
		return applyDyad(sf.dyad, a, b)
	}
	return nil, newError(10)
}
