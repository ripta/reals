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

// ErrZeroStep indicates that Range was given a zero step, which cannot advance
// toward the stop.
var ErrZeroStep = errors.New("step must be non-zero")

// ErrDirectionMismatch indicates that Range's step points away from stop, so the
// half-open range would never reach it.
var ErrDirectionMismatch = errors.New("step must advance from start toward stop")

// ErrInexactOperand indicates that Range was given an operand that is not an
// exact rational. Range is exact-only: sampling an irrational span is LinSpace's
// job.
var ErrInexactOperand = errors.New("operand must be an exact rational")

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

// Range returns the half-open sequence [start, stop) advancing by step. The
// first point is start, each subsequent point adds step, and the last point is
// the greatest start + i*step that stays strictly below stop.
//
// Range is exact-only. Every operand must be an exact rational, so the point
// count is derived exactly and no comparison against stop happens mid-loop. An
// operand that is not an exact rational returns ErrInexactOperand; sampling an
// irrational span is LinSpace's job. A zero step returns ErrZeroStep, and a step
// pointing away from stop returns ErrDirectionMismatch. A start equal to stop
// yields an empty sequence.
//
// The count is Ceil((stop - start) / step), computed on the rational factors
// directly. Divide would route through the constructive layer and lose the exact
// rational form, so Range reads the factors instead.
func Range(start, stop, step *Real) (iter.Seq[*Real], error) {
	if start.cr != constructive.One() || stop.cr != constructive.One() || step.cr != constructive.One() {
		return nil, fmt.Errorf("Range: %w", ErrInexactOperand)
	}
	if step.IsZero() {
		return nil, fmt.Errorf("Range: %w", ErrZeroStep)
	}

	q := stop.rr.Subtract(start.rr).Divide(step.rr)
	switch q.Sign() {
	case -1:
		return nil, fmt.Errorf("Range: %w", ErrDirectionMismatch)
	case 0:
		return func(yield func(*Real) bool) {}, nil
	}

	count := q.Ceil()
	return func(yield func(*Real) bool) {
		for i := rational.Zero(); i.Cmp(count) < 0; i = i.Add(rational.One()) {
			if !yield(New(nil, start.rr.Add(i.Multiply(step.rr)))) {
				return
			}
		}
	}, nil
}
