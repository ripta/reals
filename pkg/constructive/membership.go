package constructive

import (
	"context"
	"fmt"
)

// Location names where a point sits relative to a closed interval `[Lo, Hi]`.
// It is the result of the semi-decidable membership predicate Locate.
//
// Below, Inside, and Above take the values -1, 0, and 1 so that on a degenerate
// interval `[c, c]` a decided result mirrors PreciseSign(x - c, p). Indeterminate
// sits off that ordinal line at 2. It reports that the point is indistinguishable
// from a boundary at the given precision, not that it is out of domain.
type Location int

const (
	Below         Location = iota - 1 // -1
	Inside                            // 0
	Above                             // 1
	Indeterminate                     // 2
)

// String returns the constant name.
func (l Location) String() string {
	switch l {
	case Below:
		return "Below"
	case Inside:
		return "Inside"
	case Above:
		return "Above"
	case Indeterminate:
		return "Indeterminate"
	default:
		return fmt.Sprintf("Location(%d)", int(l))
	}
}

// Locate reports where x sits relative to the closed interval iv at precision p.
// It never diverges. Its only errors are those of approximating x or the bounds,
// returned as they are with Indeterminate.
//
// The two boundary signs `x - Lo` and `Hi - x` decide the four regions. A
// negative first sign is Below. A negative second sign is Above. Two positive
// signs are Inside. A zero in either sign is Indeterminate, since the point cannot
// be distinguished from that boundary at p. An exact-boundary point therefore
// reports Indeterminate here; the exact-rational path in unified resolves it to
// Inside under the closed-interval semantics.
func Locate(ctx context.Context, iv Interval, x Real, p int) (Location, error) {
	lower, err := PreciseSign(ctx, Subtract(x, iv.Lo), p)
	if err != nil {
		return Indeterminate, err
	}

	upper, err := PreciseSign(ctx, Subtract(iv.Hi, x), p)
	if err != nil {
		return Indeterminate, err
	}

	switch {
	case lower < 0:
		return Below, nil
	case upper < 0:
		return Above, nil
	case lower > 0 && upper > 0:
		return Inside, nil
	default:
		return Indeterminate, nil
	}
}

// Decision is the tri-state result of a semi-decidable two-valued interval
// predicate: emptiness and overlap. Unlike Location, which orders three spatial
// regions, Decision carries no ordering. Undecided is the zero value, so a zero
// Decision reads as "don't know" rather than a false No.
type Decision int

const (
	Undecided Decision = iota // 0
	Yes
	No
)

// String returns the constant name.
func (d Decision) String() string {
	switch d {
	case Undecided:
		return "Undecided"
	case Yes:
		return "Yes"
	case No:
		return "No"
	default:
		return fmt.Sprintf("Decision(%d)", int(d))
	}
}

// IsEmpty reports whether the closed interval iv is empty at precision p. An
// interval is empty when Hi < Lo, the state Intersect yields for disjoint inputs.
// It is the precision-bounded sign of Hi - Lo: negative is empty (Yes), positive
// is non-empty (No), and zero is Undecided, since a touching-at-a-point interval
// cannot be distinguished from empty at p. Errors from approximation are returned
// as they are with Undecided.
func IsEmpty(ctx context.Context, iv Interval, p int) (Decision, error) {
	sign, err := PreciseSign(ctx, Subtract(iv.Hi, iv.Lo), p)
	if err != nil {
		return Undecided, err
	}

	switch {
	case sign < 0:
		return Yes, nil
	case sign > 0:
		return No, nil
	default:
		return Undecided, nil
	}
}

// Overlap reports whether a and b share a point at precision p. It composes the
// total Intersect with the emptiness test: two intervals overlap exactly when
// their intersection is non-empty.
func Overlap(ctx context.Context, a, b Interval, p int) (Decision, error) {
	d, err := IsEmpty(ctx, Intersect(a, b), p)
	if err != nil {
		return Undecided, err
	}

	switch d {
	case Yes:
		return No, nil
	case No:
		return Yes, nil
	default:
		return Undecided, nil
	}
}
