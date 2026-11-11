package apl

// reduce applies f/ along the last axis, folding from the right.
func reduce(f dyadFn, x *Array) (*Array, error) {
	// Prototype: scalar or empty → 24, one-element vector → 51.
	if x.IsScalar() || x.Size() == 0 {
		return nil, newError(24)
	}
	if x.Rank() == 1 && x.Shape[0] == 1 {
		return nil, newError(51)
	}
	n := x.Shape[x.Rank()-1]
	shape := append([]int(nil), x.Shape[:x.Rank()-1]...)
	rows := shapeSize(shape)
	r := &Array{Shape: shape, Data: make([]float64, rows)}
	for i := range rows {
		row := x.Data[i*n : (i+1)*n]
		acc := row[n-1]
		for j := n - 2; j >= 0; j-- {
			v, err := f(row[j], acc)
			if err != nil {
				return nil, err
			}
			acc = v
		}
		r.Data[i] = acc
	}
	return r, nil
}
