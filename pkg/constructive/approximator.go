package constructive

import (
	"context"
	"fmt"
	"math/big"
)

// Approximator computes a Real number to a requested precision. Given p, it returns
// an integer a such that a·2^p is within 2^p of the number. Every node in this
// package meets the same contract, and the nodes built on top of an Approximator
// rely on it to be correct in their last bits.
//
// The Real built from an Approximator calls it only for a precision finer than any
// it has computed so far. A coarser request is rounded from the finest result, so
// two coarser reads always agree with each other. Agreement between the
// Approximator's own results at different precisions is its own responsibility.
// One that is wrong at a coarse precision and right at a finer one makes
// comparisons on the number depend on the precision they reach.
//
// ctx is the caller's context, marked as running this Approximator. It should be
// honored, so that a cancelled caller is not left waiting. It must also be passed,
// or a context derived from it, to any approximation the Approximator makes. A
// nested approximation of the same Real that would run the Approximator again then
// fails with ErrReentrantApproximation. Under an unrelated context the nested call
// cannot be told from a concurrent one, so it waits for the outer run, which cannot
// finish before it, until that context ends.
//
// An error is returned to the caller unchanged and is not cached. A nil
// approximation with a nil error is reported as ErrBadApproximation. The returned
// integer is copied, so the Approximator may keep and reuse it.
//
// An Approximator is never run concurrently with itself for the same Real. Callers
// approximating that Real from several goroutines wait for the run in progress and
// share its result when it covers their precision.
type Approximator func(ctx context.Context, p int) (*big.Int, error)

// FromApproximation creates a Real number whose approximations come from fn. A
// non-empty label names the number, as the built-in constants are named, so that
// ConstructiveName returns it. It returns nil when fn is nil.
func FromApproximation(label string, fn Approximator) Real {
	if fn == nil {
		return nil
	}

	c := newApproximator(fn)
	if label == "" {
		return c
	}

	return newNamed(label, c)
}

// SliceApproximator computes several Real numbers at once to a requested precision.
// Given p, it returns one integer per output, and element i is an integer a such
// that a·2^p is within 2^p of output i. Each element meets the same contract as
// the integer an Approximator returns. The slice must have exactly the length
// given to FromApproximationSlice.
//
// The Reals built from a SliceApproximator share one computation. The callback is
// called only for a precision finer than any it has computed so far, and every
// output is then refined from that one run. A coarser request for any output is
// rounded from the finest result. Agreement between the callback's own results at
// different precisions, and between its outputs, is its own responsibility.
//
// ctx is the caller's context, marked as running this SliceApproximator. It should
// be honored, so that a cancelled caller is not left waiting. It must also be
// passed, or a context derived from it, to any approximation the callback makes. A
// nested approximation of any of the outputs that would run the callback again
// then fails with ErrReentrantApproximation. Under an unrelated context the nested
// call cannot be told from a concurrent one, so it waits for the outer run, which
// cannot finish before it, until that context ends.
//
// An error is returned to the caller unchanged, and none of the outputs is
// updated. A slice of the wrong length, or one holding a nil element, is reported
// as ErrBadApproximation and updates none of the outputs either. The returned
// integers are copied, so the callback may keep and reuse them.
//
// A SliceApproximator is never run concurrently with itself for the same outputs.
// Callers approximating any of them from several goroutines wait for the run in
// progress and share its result when it covers their precision.
type SliceApproximator func(ctx context.Context, p int) ([]*big.Int, error)

// FromApproximationSlice creates n Real numbers whose approximations all come from
// one call of fn. Output i of the result is element i of what fn returns. A
// non-empty label names every output, as for FromApproximation, so that
// ConstructiveName returns it. It returns nil when fn is nil, and an empty slice
// when n is not positive.
func FromApproximationSlice(label string, n int, fn SliceApproximator) []Real {
	if fn == nil {
		return nil
	}

	if n <= 0 {
		return []Real{}
	}

	outputs := newApproximatorSlice(n, fn)
	if label == "" {
		return outputs
	}

	for i, c := range outputs {
		outputs[i] = newNamed(label, c)
	}

	return outputs
}

// approximationSet is the computation shared by the outputs of one callback. It
// owns the one guard, so that a run started by any output keeps every sibling out
// until it finishes, and a nested read of any of them is detected as re-entrant.
type approximationSet struct {
	fn      SliceApproximator
	guard   *approximationGuard
	outputs []*approximatorNode
}

// approximatorNode is one output of an approximationSet. It is the Real whose
// approximations come from a caller's callback.
type approximatorNode struct {
	precisionTracker

	set          *approximationSet
	index        int
	construction string
}

// newApproximator builds the single output of a callback that returns one integer.
func newApproximator(fn Approximator) Real {
	set := newApproximationSet(1, func(ctx context.Context, p int) ([]*big.Int, error) {
		v, err := fn(ctx, p)
		if err != nil {
			return nil, err
		}

		return []*big.Int{v}, nil
	})
	set.outputs[0].construction = "Approximation()"

	return set.outputs[0]
}

// newApproximatorSlice builds the n outputs of a callback that returns n integers.
func newApproximatorSlice(n int, fn SliceApproximator) []Real {
	set := newApproximationSet(n, fn)

	outputs := make([]Real, n)
	for i, c := range set.outputs {
		c.construction = fmt.Sprintf("Approximation(%d, %d)", i, n)
		outputs[i] = c
	}

	return outputs
}

func newApproximationSet(n int, fn SliceApproximator) *approximationSet {
	set := &approximationSet{
		fn:      fn,
		guard:   newApproximationGuard(),
		outputs: make([]*approximatorNode, n),
	}
	for i := range set.outputs {
		set.outputs[i] = &approximatorNode{set: set, index: i}
	}

	return set
}

func (c *approximatorNode) approximate(ctx context.Context, p int) (*big.Int, error) {
	return c.set.approximate(ctx, c.index, p)
}

func (c *approximatorNode) asConstruction() string {
	return c.construction
}

// approximate returns output i at p, running the callback when no run has covered
// p yet. Every output is written from a run before the guard is released, so a
// caller waiting on any sibling finds it in its own cache.
func (s *approximationSet) approximate(ctx context.Context, i, p int) (*big.Int, error) {
	var result *big.Int
	err := s.guard.run(ctx, func(ctx context.Context) error {
		// A run that finished while this call waited its turn may already cover p.
		if v, ok := s.outputs[i].Get(p); ok {
			result = v
			return nil
		}

		values, err := s.fn(ctx, p)
		if err != nil {
			return err
		}

		if len(values) != len(s.outputs) {
			return ErrBadApproximation
		}
		for _, v := range values {
			if v == nil {
				return ErrBadApproximation
			}
		}

		for j, v := range values {
			// Each output gets its own copy, so that no two outputs, and not the
			// callback, share an integer.
			c := s.outputs[j].Set(p, new(big.Int).Set(v))
			if j == i {
				result = c
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

// approximationGuard serializes the runs of one approximation callback and detects
// a run nested inside itself.
//
// Go has no goroutine identity, so a waiting call cannot tell a nested run from a
// concurrent one. Each run marks the context it passes to the callback instead. A
// call that arrives carrying the mark comes from inside the run, and is rejected
// rather than left waiting on a run that cannot finish before it.
type approximationGuard struct {
	turn chan struct{}
}

// inProgressKey marks a context as belonging to a run of one guard.
type inProgressKey struct {
	guard *approximationGuard
}

func newApproximationGuard() *approximationGuard {
	return &approximationGuard{turn: make(chan struct{}, 1)}
}

// run calls fn with a marked context once no other run is in progress. It returns
// ErrReentrantApproximation without waiting when ctx already carries this guard's
// mark, and the context's error when ctx ends while waiting.
func (g *approximationGuard) run(ctx context.Context, fn func(context.Context) error) error {
	key := inProgressKey{guard: g}
	if ctx.Value(key) != nil {
		return ErrReentrantApproximation
	}

	select {
	case g.turn <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-g.turn }()

	return fn(context.WithValue(ctx, key, struct{}{}))
}
