package constructive

// Interval is a range `[Lo, Hi]` whose endpoints are constructive reals.
// New normalizes so `Lo <= Hi` holds as a structural invariant.
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

// Width computes `Hi - Lo`, the length of the interval.
func Width(iv Interval) Real {
	return Subtract(iv.Hi, iv.Lo)
}

// Mid computes the midpoint `(Lo + Hi) / 2`. The halving is an exact power-of-two
// shift, so no precision is lost to general division.
func Mid(iv Interval) Real {
	return ShiftRight(Add(iv.Lo, iv.Hi), 1)
}

// Clamp confines x to the interval, returning Lo below the band, Hi above it,
// and x within. It routes through the total Min and Max core, so it never
// decides where x sits relative to a boundary.
func Clamp(iv Interval, x Real) Real {
	return Max(iv.Lo, Min(x, iv.Hi))
}

// Hull computes the convex hull of a and b, the smallest interval containing
// both. Both inputs are already normalized, so the result is ordered without
// renormalizing.
func Hull(a, b Interval) Interval {
	return Interval{
		Lo: Min(a.Lo, b.Lo),
		Hi: Max(a.Hi, b.Hi),
	}
}

// Intersect computes the overlap of a and b. The result is a total value even
// when the inputs are disjoint: it then yields Lo > Hi to signal emptiness,
// which is why it builds the struct directly rather than renormalizing through
// New. Testing that emptiness is a separate, semi-decidable concern.
func Intersect(a, b Interval) Interval {
	return Interval{
		Lo: Max(a.Lo, b.Lo),
		Hi: Min(a.Hi, b.Hi),
	}
}
