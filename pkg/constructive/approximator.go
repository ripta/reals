package constructive

import (
	"context"
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

// approximatorNode is a Real whose approximations come from a caller's callback.
type approximatorNode struct {
	precisionTracker

	fn    Approximator
	guard *approximationGuard
}

func newApproximator(fn Approximator) Real {
	return &approximatorNode{
		fn:    fn,
		guard: newApproximationGuard(),
	}
}

func (c *approximatorNode) approximate(ctx context.Context, p int) (*big.Int, error) {
	var result *big.Int
	err := c.guard.run(ctx, func(ctx context.Context) error {
		// A run that finished while this call waited its turn may already cover p.
		if s, ok := c.Get(p); ok {
			result = s
			return nil
		}

		v, err := c.fn(ctx, p)
		if err != nil {
			return err
		}
		if v == nil {
			return ErrBadApproximation
		}

		result = new(big.Int).Set(v)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *approximatorNode) asConstruction() string {
	return "Approximation()"
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
