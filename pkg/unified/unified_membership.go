package unified

import (
	"github.com/ripta/reals/pkg/constructive"
)

// Location re-exports constructive.Location so a caller of the unified interval
// predicates names the four regions without importing constructive. The alias
// keeps the two identical, so the inherited String method needs no restating.
type Location = constructive.Location

const (
	Below         = constructive.Below
	Inside        = constructive.Inside
	Above         = constructive.Above
	Indeterminate = constructive.Indeterminate
)

// Decision re-exports constructive.Decision for the two-valued emptiness and
// overlap predicates, on the same aliasing rationale as Location.
type Decision = constructive.Decision

const (
	Undecided = constructive.Undecided
	Yes       = constructive.Yes
	No        = constructive.No
)

// signOfDiff returns the sign of a - b and whether it was decided exactly. When
// both operands are exact rationals, cr == constructive.One(), the rationals
// compare exactly. Otherwise it falls back to the budgeted PreciseSign at
// precision p, which returns 0 when the operands are indistinguishable there.
func signOfDiff(a, b *Real, p int) (sign int, exact bool) {
	if a.cr == constructive.One() && b.cr == constructive.One() {
		return a.rr.Cmp(b.rr), true
	}
	return constructive.PreciseSign(constructive.Subtract(a.Constructive(), b.Constructive()), p), false
}

// Locate reports where x sits relative to the closed interval iv at precision p.
// It is total: it never diverges and never returns an error.
//
// Each boundary sign is decided exactly when both of its operands are exact
// rationals, falling back to the budgeted PreciseSign otherwise. An exactly-zero
// sign means x lies exactly on that endpoint, which is Inside under closed
// semantics. This is the refinement the constructive path can only report as
// Indeterminate.
func (iv Interval) Locate(x *Real, p int) Location {
	lower, lowerExact := signOfDiff(x, iv.lo, p)
	upper, upperExact := signOfDiff(iv.hi, x, p)

	switch {
	case lower < 0:
		return Below
	case upper < 0:
		return Above
	case lowerExact && lower == 0:
		return Inside
	case upperExact && upper == 0:
		return Inside
	case lower > 0 && upper > 0:
		return Inside
	default:
		return Indeterminate
	}
}

// IsEmpty reports whether the closed interval iv is empty at precision p. It is
// the sign of hi - lo, decided exactly when both endpoints are exact rationals.
// An interval is empty when hi < lo. A degenerate hi == lo is the single-point
// interval, non-empty under closed semantics, so an exact zero reports No where
// the budgeted path can only report Undecided.
func (iv Interval) IsEmpty(p int) Decision {
	sign, exact := signOfDiff(iv.hi, iv.lo, p)
	switch {
	case sign < 0:
		return Yes
	case sign > 0:
		return No
	case exact:
		return No
	default:
		return Undecided
	}
}

// Overlap reports whether iv and other share a point at precision p. It composes
// the total Intersect with the emptiness test: two intervals overlap exactly when
// their intersection is non-empty. Intersect routes through the constructive
// Min/Max core, which fuses away rational visibility, so overlap is budgeted even
// when the original endpoints are exact rationals.
func (iv Interval) Overlap(other Interval, p int) Decision {
	switch iv.Intersect(other).IsEmpty(p) {
	case Yes:
		return No
	case No:
		return Yes
	default:
		return Undecided
	}
}
