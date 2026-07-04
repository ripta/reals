package constructive

import "fmt"

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
// It is total: it never diverges and never returns an error.
//
// The two boundary signs `x - Lo` and `Hi - x` decide the four regions. A
// negative first sign is Below. A negative second sign is Above. Two positive
// signs are Inside. A zero in either sign is Indeterminate, since the point cannot
// be distinguished from that boundary at p. An exact-boundary point therefore
// reports Indeterminate here; the exact-rational path in unified resolves it to
// Inside under the closed-interval semantics.
func Locate(iv Interval, x Real, p int) Location {
	lower := PreciseSign(Subtract(x, iv.Lo), p)
	upper := PreciseSign(Subtract(iv.Hi, x), p)

	switch {
	case lower < 0:
		return Below
	case upper < 0:
		return Above
	case lower > 0 && upper > 0:
		return Inside
	default:
		return Indeterminate
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
// cannot be distinguished from empty at p.
func IsEmpty(iv Interval, p int) Decision {
	switch sign := PreciseSign(Subtract(iv.Hi, iv.Lo), p); {
	case sign < 0:
		return Yes
	case sign > 0:
		return No
	default:
		return Undecided
	}
}

// Overlap reports whether a and b share a point at precision p. It composes the
// total Intersect with the emptiness test: two intervals overlap exactly when
// their intersection is non-empty.
func Overlap(a, b Interval, p int) Decision {
	switch IsEmpty(Intersect(a, b), p) {
	case Yes:
		return No
	case No:
		return Yes
	default:
		return Undecided
	}
}
