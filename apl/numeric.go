package apl

import "math"

// decode is A⊥B: the values of the digits along the first axis of B in
// the radices along the last axis of A. An axis of length 1 extends.
func decode(e *env, a, b *Array) (*Array, error) {
	ad, err := a.numeric()
	if err != nil {
		return nil, err
	}
	bd, err := b.numeric()
	if err != nil {
		return nil, err
	}
	na, nb := 1, 1
	var lshape, rshape []int
	if !a.IsScalar() {
		na = a.Shape[a.Rank()-1]
		lshape = a.Shape[:a.Rank()-1]
	}
	if !b.IsScalar() {
		nb = b.Shape[0]
		rshape = b.Shape[1:]
	}
	n := max(na, nb)
	if na != nb && na != 1 && nb != 1 {
		return nil, newError(LengthError)
	}
	rows, cols := shapeSize(lshape), shapeSize(rshape)
	shape := append(append([]int(nil), lshape...), rshape...)
	r := &Array{Shape: shape, Data: make([]float64, rows*cols)}
	if na == 0 || nb == 0 {
		return r, nil
	}
	for i := range rows {
		for j := range cols {
			acc := 0.0
			for k := range n {
				radix := ad[i*na+min(k, na-1)]
				acc = acc*radix + bd[min(k, nb-1)*cols+j]
			}
			r.Data[i*cols+j] = acc
		}
	}
	return r, nil
}

// encode is A⊤B: the digits of B in the radices along the first axis
// of A. A radix 0 takes all that is left.
func encode(e *env, a, b *Array) (*Array, error) {
	ad, err := a.numeric()
	if err != nil {
		return nil, err
	}
	bd, err := b.numeric()
	if err != nil {
		return nil, err
	}
	if a.IsScalar() {
		return applyDyad(e, scalarFns["|"], a, b)
	}
	n := a.Shape[0]
	cols := a.Size() / max(n, 1)
	nb := len(bd)
	shape := append(append([]int(nil), a.Shape...), b.Shape...)
	r := &Array{Shape: shape, Data: make([]float64, len(ad)*nb)}
	for c := range cols {
		for bi, v := range bd {
			for k := n - 1; k >= 0; k-- {
				radix := ad[k*cols+c]
				d := v
				if radix != 0 {
					d, _ = residue(e, radix, v)
					v = (v - d) / radix
				} else {
					v = 0
				}
				r.Data[(k*cols+c)*nb+bi] = d
			}
		}
	}
	return r, nil
}

// matrix views x as rows×cols: a vector is one column, a scalar 1×1.
func matrix(x *Array) ([]float64, int, int, error) {
	d, err := x.numeric()
	if err != nil {
		return nil, 0, 0, err
	}
	switch x.Rank() {
	case 0:
		return d, 1, 1, nil
	case 1:
		return d, x.Shape[0], 1, nil
	case 2:
		return d, x.Shape[0], x.Shape[1], nil
	}
	return nil, 0, 0, newError(RankError)
}

// solve finds X with B+.×X closest to A in the least-squares sense,
// by Householder QR of B (m×n, m≥n); A is m×p.
func solve(a []float64, p int, b []float64, m, n int) ([]float64, error) {
	if m < n {
		return nil, newError(LengthError)
	}
	q := append([]float64(nil), b...)
	y := append([]float64(nil), a...)
	norm := 0.0
	for _, v := range q {
		norm = math.Max(norm, math.Abs(v))
	}
	for k := range n {
		s := 0.0
		for i := k; i < m; i++ {
			s += q[i*n+k] * q[i*n+k]
		}
		alpha := -math.Copysign(math.Sqrt(s), q[k*n+k])
		if math.Abs(alpha) <= 1e-13*norm || alpha == 0 {
			return nil, newError(DomainError) // singular
		}
		v := make([]float64, m)
		for i := k; i < m; i++ {
			v[i] = q[i*n+k]
		}
		v[k] -= alpha
		vv := 0.0
		for i := k; i < m; i++ {
			vv += v[i] * v[i]
		}
		reflect := func(x []float64, w, j int) {
			d := 0.0
			for i := k; i < m; i++ {
				d += v[i] * x[i*w+j]
			}
			d = 2 * d / vv
			for i := k; i < m; i++ {
				x[i*w+j] -= d * v[i]
			}
		}
		for j := k; j < n; j++ {
			reflect(q, n, j)
		}
		for j := range p {
			reflect(y, p, j)
		}
	}
	x := make([]float64, n*p)
	for j := range p {
		for i := n - 1; i >= 0; i-- {
			s := y[i*p+j]
			for k := i + 1; k < n; k++ {
				s -= q[i*n+k] * x[k*p+j]
			}
			x[i*p+j] = s / q[i*n+i]
		}
	}
	return x, nil
}

// inverse is ⌹X: the inverse of a square matrix, or the left inverse
// of a matrix with more rows than columns.
func inverse(_ *env, x *Array) (*Array, error) {
	b, m, n, err := matrix(x)
	if err != nil {
		return nil, err
	}
	id := make([]float64, m*m)
	for i := range m {
		id[i*m+i] = 1
	}
	r, err := solve(id, m, b, m, n)
	if err != nil {
		return nil, err
	}
	shape := []int{n, m}
	switch x.Rank() {
	case 0:
		shape = nil
	case 1:
		shape = []int{m}
	}
	return &Array{Shape: shape, Data: r}, nil
}

// matDivide is A⌹B: X such that B+.×X is A, or the least-squares fit.
func matDivide(_ *env, a, b *Array) (*Array, error) {
	bd, m, n, err := matrix(b)
	if err != nil {
		return nil, err
	}
	ad, ma, p, err := matrix(a)
	if err != nil {
		return nil, err
	}
	if ma != m {
		return nil, newError(LengthError)
	}
	r, err := solve(ad, p, bd, m, n)
	if err != nil {
		return nil, err
	}
	var shape []int
	if b.Rank() == 2 {
		shape = append(shape, n)
	}
	if a.Rank() == 2 {
		shape = append(shape, p)
	}
	return &Array{Shape: shape, Data: r}, nil
}
