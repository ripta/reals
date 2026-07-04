package unified

import (
	"errors"
	"fmt"
	"iter"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// ErrNegativeCount indicates that a sequence generator was given a negative
// point count, which has no valid sequence.
var ErrNegativeCount = errors.New("sample count must be non-negative")

// LinSpace returns a sequence of n evenly-spaced points across the closed
// interval [start, stop]. The endpoints are both included: the first point is
// start, the last is stop, and the width divides into n - 1 gaps.
//
// Degenerate counts follow NumPy. An n of zero yields an empty sequence, and an
// n of one yields a single point at start. A negative n returns ErrNegativeCount.
//
// The count is fixed by n, so LinSpace never compares against an endpoint
// mid-loop. When start and stop reduce to rationals the spacing is exact, so the
// emitted points are exact and the last is stop exactly. The final point yields
// stop directly, which keeps the closed-interval contract exact even for
// irrational endpoints.
func LinSpace(start, stop *Real, n int) (iter.Seq[*Real], error) {
	if n < 0 {
		return nil, fmt.Errorf("LinSpace: %w", ErrNegativeCount)
	}
	if n == 0 {
		return func(yield func(*Real) bool) {}, nil
	}
	if n == 1 {
		return func(yield func(*Real) bool) { yield(start) }, nil
	}

	delta := stop.Subtract(start).Divide(intReal(n - 1))
	return func(yield func(*Real) bool) {
		for i := 0; i < n-1; i++ {
			if !yield(start.Add(intReal(i).Multiply(delta))) {
				return
			}
		}
		yield(stop)
	}, nil
}

// intReal builds a purely-rational Real holding the integer i, the form the
// package constants use for exact rational values.
func intReal(i int) *Real {
	return New(constructive.One(), rational.New64(int64(i), 1))
}
