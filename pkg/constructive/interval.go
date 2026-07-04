package constructive

// Interval is a range [Lo, Hi] whose endpoints are constructive reals.
// New normalizes so Lo <= Hi holds as a structural invariant.
type Interval struct {
	Lo, Hi Real
}

// New constructs an Interval from two endpoints, normalizing eagerly so that
// Lo is the smaller and Hi the larger. Min and Max route through the total
// condsign core, so where p and q are indistinguishable this yields Lo == Hi
// rather than diverging.
func New(p, q Real) Interval {
	return Interval{
		Lo: Min(p, q),
		Hi: Max(p, q),
	}
}
