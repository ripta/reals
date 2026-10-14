package constructive

import (
	"math/big"
	"sync/atomic"
)

// approximation is one cached result: value·2^precision is within 2^precision of
// the real number. It is never modified once stored, so a reader holding one sees
// a precision and a value that belong together.
type approximation struct {
	precision int
	value     *big.Int
}

// precisionTracker caches the finest approximation of a Real number computed so
// far. The tracker is used by embedding in a struct that implements the Real
// interface, on which the `tracker` function can be called.
//
// A Real may be approximated from several goroutines at once. The cache is a
// single atomic pointer to an immutable approximation, so a reader never sees a
// precision from one result paired with the value of another.
type precisionTracker struct {
	finest atomic.Pointer[approximation]
}

// latest returns the finest approximation cached so far, or nil when there is none.
func (t *precisionTracker) latest() *approximation {
	return t.finest.Load()
}

func (t *precisionTracker) Get(p int) (*big.Int, bool) {
	a := t.finest.Load()
	if a != nil && p >= a.precision {
		return scale(a.value, a.precision-p), true
	}

	return nil, false
}

// Set caches i as the approximation at precision p, unless a racing approximation
// already cached one at least as fine. It returns i either way, since i is a valid
// answer for p.
func (t *precisionTracker) Set(p int, i *big.Int) *big.Int {
	next := &approximation{precision: p, value: i}
	for {
		cur := t.finest.Load()
		if cur != nil && cur.precision <= p {
			return i
		}

		if t.finest.CompareAndSwap(cur, next) {
			return i
		}
	}
}

func (t *precisionTracker) tracker() *precisionTracker {
	return t
}
