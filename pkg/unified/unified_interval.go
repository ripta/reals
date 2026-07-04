package unified

import (
	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// Interval is a range `[lo, hi]` whose endpoints are unified reals. Holding two
// *Real endpoints keeps each endpoint's rational factor visible, which the
// constructive layer's eager Min/Max fuses away. NewInterval normalizes so that
// lo <= hi holds as a structural invariant.
type Interval struct {
	lo, hi *Real
}

// NewInterval constructs an Interval from two endpoints, normalizing eagerly so
// that lo is the smaller and hi the larger.
//
// When both endpoints are exact rationals, ordering is exactly decidable, so it
// orders the two *Real values and preserves each one's rational factor. When an
// endpoint is genuinely irrational there is no rational visibility to preserve,
// so it falls back to the constructive Min/Max core. Both paths are total: where
// p and q are indistinguishable the constructive branches agree, yielding
// lo == hi rather than diverging.
func NewInterval(p, q *Real) Interval {
	if p.cr == constructive.One() && q.cr == constructive.One() {
		if p.rr.Cmp(q.rr) <= 0 {
			return Interval{lo: p, hi: q}
		}
		return Interval{lo: q, hi: p}
	}

	return Interval{
		lo: New(constructive.Min(p.Constructive(), q.Constructive()), rational.One()),
		hi: New(constructive.Max(p.Constructive(), q.Constructive()), rational.One()),
	}
}
