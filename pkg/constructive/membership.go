package constructive

import "fmt"

// Location names where a point sits relative to a closed interval `[Lo, Hi]`.
// It is the result of the semi-decidable membership predicate Locate.
//
// Below, Inside, and Above take the values -1, 0, and 1 so that on a degenerate
// interval `[c, c]` a decided result mirrors PreciseSign(x - c, p). Undecided
// sits off that ordinal line at 2. It reports that the point is indistinguishable
// from a boundary at the given precision, not that it is out of domain.
type Location int

const (
	Below     Location = iota - 1 // -1
	Inside                        // 0
	Above                         // 1
	Undecided                     // 2
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
	case Undecided:
		return "Undecided"
	default:
		return fmt.Sprintf("Location(%d)", int(l))
	}
}

// Locate reports where x sits relative to the closed interval iv at precision p.
// It is total: it never diverges and never returns an error.
//
// The two boundary signs `x - Lo` and `Hi - x` decide the four regions. A
// negative first sign is Below. A negative second sign is Above. Two positive
// signs are Inside. A zero in either sign is Undecided, since the point cannot
// be distinguished from that boundary at p. An exact-boundary point therefore
// reports Undecided here; the exact-rational path in unified resolves it to
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
		return Undecided
	}
}
