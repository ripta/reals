package unified

import (
	"context"

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
// Errors from approximating the difference are returned as they are.
func signOfDiff(ctx context.Context, a, b *Real, p int) (sign int, exact bool, err error) {
	if a.cr == constructive.One() && b.cr == constructive.One() {
		return a.rr.Cmp(b.rr), true, nil
	}

	diff := constructive.Subtract(a.Constructive(), b.Constructive())

	v, err := constructive.PreciseSign(ctx, diff, p)
	if err != nil {
		return 0, false, err
	}

	return v, false, nil
}

// Locate reports where x sits relative to the closed interval iv at precision p.
// It never diverges. On an approximation error, which includes the context
// error, it returns Indeterminate beside that error.
//
// Each boundary sign is decided exactly when both of its operands are exact
// rationals, falling back to the budgeted PreciseSign otherwise. An exactly-zero
// sign means x lies exactly on that endpoint, which is Inside under closed
// semantics. This is the refinement the constructive path can only report as
// Indeterminate.
func (iv Interval) Locate(ctx context.Context, x *Real, p int) (Location, error) {
	lower, lowerExact, err := signOfDiff(ctx, x, iv.lo, p)
	if err != nil {
		return Indeterminate, err
	}

	upper, upperExact, err := signOfDiff(ctx, iv.hi, x, p)
	if err != nil {
		return Indeterminate, err
	}

	switch {
	case lower < 0:
		return Below, nil
	case upper < 0:
		return Above, nil
	case lowerExact && lower == 0:
		return Inside, nil
	case upperExact && upper == 0:
		return Inside, nil
	case lower > 0 && upper > 0:
		return Inside, nil
	default:
		return Indeterminate, nil
	}
}

// IsEmpty reports whether the closed interval iv is empty at precision p. It is
// the sign of hi - lo, decided exactly when both endpoints are exact rationals.
// An interval is empty when hi < lo. A degenerate hi == lo is the single-point
// interval, non-empty under closed semantics, so an exact zero reports No where
// the budgeted path can only report Undecided. On an approximation error it
// returns Undecided beside that error.
func (iv Interval) IsEmpty(ctx context.Context, p int) (Decision, error) {
	sign, exact, err := signOfDiff(ctx, iv.hi, iv.lo, p)
	if err != nil {
		return Undecided, err
	}

	switch {
	case sign < 0:
		return Yes, nil
	case sign > 0:
		return No, nil
	case exact:
		return No, nil
	default:
		return Undecided, nil
	}
}

// Overlap reports whether iv and other share a point at precision p. It composes
// the total Intersect with the emptiness test: two intervals overlap exactly when
// their intersection is non-empty. Intersect routes through the constructive
// Min/Max core, which fuses away rational visibility, so overlap is budgeted even
// when the original endpoints are exact rationals. On an approximation error it
// returns Undecided beside that error.
func (iv Interval) Overlap(ctx context.Context, other Interval, p int) (Decision, error) {
	empty, err := iv.Intersect(other).IsEmpty(ctx, p)
	if err != nil {
		return Undecided, err
	}

	switch empty {
	case Yes:
		return No, nil
	case No:
		return Yes, nil
	default:
		return Undecided, nil
	}
}
