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

// constructive projects the endpoints to the constructive layer, keeping their
// order. The endpoints are already normalized, so the result needs no
// renormalization.
func (iv Interval) constructive() constructive.Interval {
	return constructive.Interval{Lo: iv.lo.Constructive(), Hi: iv.hi.Constructive()}
}

// Width computes `hi - lo`, the length of the interval.
func (iv Interval) Width() *Real {
	return New(constructive.Width(iv.constructive()), rational.One())
}

// Mid computes the midpoint `(lo + hi) / 2`. The halving is an exact
// power-of-two shift, so no precision is lost to general division.
func (iv Interval) Mid() *Real {
	return New(constructive.Mid(iv.constructive()), rational.One())
}

// Clamp confines x to the interval, returning lo below the band, hi above it,
// and x within. It routes through the total constructive core, so it never
// decides where x sits relative to a boundary.
func (iv Interval) Clamp(x *Real) *Real {
	return New(constructive.Clamp(iv.constructive(), x.Constructive()), rational.One())
}

// Hull computes the convex hull of iv and other, the smallest interval
// containing both. Both inputs are already normalized, so the result is ordered
// without renormalizing.
func (iv Interval) Hull(other Interval) Interval {
	h := constructive.Hull(iv.constructive(), other.constructive())
	return Interval{
		lo: New(h.Lo, rational.One()),
		hi: New(h.Hi, rational.One()),
	}
}

// Intersect computes the overlap of iv and other. The result is a total value
// even when the inputs are disjoint: it then yields lo > hi to signal
// emptiness, so it builds the struct directly rather than renormalizing through
// NewInterval. Testing that emptiness is a separate, semi-decidable concern.
func (iv Interval) Intersect(other Interval) Interval {
	r := constructive.Intersect(iv.constructive(), other.constructive())
	return Interval{
		lo: New(r.Lo, rational.One()),
		hi: New(r.Hi, rational.One()),
	}
}
