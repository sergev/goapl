package apl

import "slices"

// member is A∊B: 1 for each element of A found in B.
func member(e *env, a, b *Array) (*Array, error) {
	r := &Array{Shape: a.Shape, Data: make([]float64, a.Size())}
	for i := range r.Data {
		for j := range b.Size() {
			if match(a.At(i), b.At(j), e.ct) {
				r.Data[i] = 1
				break
			}
		}
	}
	return r, nil
}

// without is A~B: the elements of vector A not found in B.
func without(e *env, a, b *Array) (*Array, error) {
	if a.Rank() > 1 {
		return nil, newError(RankError)
	}
	in, err := member(e, a, b)
	if err != nil {
		return nil, err
	}
	var pos []int
	for i, v := range in.Data {
		if v == 0 {
			pos = append(pos, i)
		}
	}
	return a.gather([]int{len(pos)}, pos), nil
}

// sortKey orders elements for grading: numbers by value, characters by
// position in a collating sequence (or by code without one).
type sortKey struct {
	collate *Array // collating sequence, or nil
}

func (s sortKey) key(v Value) (float64, error) {
	switch v := v.(type) {
	case float64:
		if s.collate != nil {
			return 0, newError(DomainError)
		}
		return v, nil
	case Char:
		if s.collate == nil {
			return float64(v), nil
		}
		for i := range s.collate.Size() {
			if s.collate.At(i) == v {
				return float64(i), nil
			}
		}
		return float64(s.collate.Size()), nil
	}
	return 0, newError(DomainError)
}

// grade gives the indices (from ⎕IO) that order the items of x along
// its first axis, ascending or descending; equal items keep their
// order.
func grade(e *env, x *Array, s sortKey, down bool) (*Array, error) {
	if x.IsScalar() {
		return nil, newError(RankError)
	}
	n := x.Shape[0]
	w := 0
	if n > 0 {
		w = x.Size() / n
	}
	keys := make([]float64, x.Size())
	chars := false
	for i := range keys {
		v := x.At(i)
		_, isChar := v.(Char)
		if i == 0 {
			chars = isChar
		} else if isChar != chars {
			return nil, newError(DomainError)
		}
		k, err := s.key(v)
		if err != nil {
			return nil, err
		}
		keys[i] = k
	}
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	slices.SortStableFunc(idx, func(i, j int) int {
		for k := range w {
			a, b := keys[i*w+k], keys[j*w+k]
			if a != b {
				if (a < b) != down {
					return -1
				}
				return 1
			}
		}
		return 0
	})
	r := make([]float64, n)
	for i, j := range idx {
		r[i] = float64(j + e.io)
	}
	return Vector(r), nil
}

func gradeUp(e *env, x *Array) (*Array, error)   { return grade(e, x, sortKey{}, false) }
func gradeDown(e *env, x *Array) (*Array, error) { return grade(e, x, sortKey{}, true) }

// gradeUpCollate is A⍋B: grades characters by their position in A.
func gradeUpCollate(e *env, a, b *Array) (*Array, error) {
	if a.IsNumeric() {
		return nil, newError(DomainError)
	}
	return grade(e, b, sortKey{collate: a}, false)
}

func gradeDownCollate(e *env, a, b *Array) (*Array, error) {
	if a.IsNumeric() {
		return nil, newError(DomainError)
	}
	return grade(e, b, sortKey{collate: a}, true)
}
