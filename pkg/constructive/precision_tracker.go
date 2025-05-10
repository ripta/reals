package constructive

import "math/big"

// precisionTracker tracks the minimum precision (more negative is more precise)
// and the maximum approximation of a Real number. The tracker is used
// by embedding in a struct that implements the Real interface, on which
// the `tracker` function can be called.
type precisionTracker struct {
	IsValid bool

	MaxApproximation *big.Int
	MinPrecision     int
}

func (t *precisionTracker) Get(p int) (*big.Int, bool) {
	if t.IsValid && p >= t.MinPrecision {
		return scale(t.MaxApproximation, t.MinPrecision-p), true
	}

	return nil, false
}

func (t *precisionTracker) Set(p int, i *big.Int) *big.Int {
	t.IsValid = true
	t.MaxApproximation = i
	t.MinPrecision = p

	return i
}

func (t *precisionTracker) tracker() *precisionTracker {
	return t
}
