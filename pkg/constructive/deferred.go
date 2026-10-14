package constructive

import (
	"context"
	"fmt"
	"math/big"
	"sync/atomic"
)

// reducer picks a range reduction for an argument from its rough approximation,
// and builds the reduced subtree. It must not approximate anything itself.
type reducer func(arg Real, rough *big.Int) (Real, error)

// deferredReduction is a function of one argument whose range reduction depends on
// the argument's magnitude. Choosing the reduction needs a rough approximation of
// the argument, which can fail or be cancelled. So the choice waits for the first
// approximation, where the caller's context is available.
//
// The reduced subtree is stored only after the reduction succeeds. A failed first
// approximation leaves the node untouched and the next call tries again.
//
// The memo is an atomic pointer, so a racing read never sees a torn interface
// value. No lock is held while approximating. Two racing first calls each build an
// equivalent subtree, and the first store wins. The precision tracker beside it is
// still unsynchronized.
type deferredReduction struct {
	precisionTracker

	name      string
	arg       Real
	roughPrec int
	reduce    reducer

	reduced atomic.Pointer[reducedTree]
}

// reducedTree boxes the interface value, because an atomic pointer needs a
// concrete type.
type reducedTree struct {
	root Real
}

func newDeferredReduction(name string, arg Real, roughPrec int, reduce reducer) Real {
	return &deferredReduction{
		name:      name,
		arg:       arg,
		roughPrec: roughPrec,
		reduce:    reduce,
	}
}

func (c *deferredReduction) approximate(ctx context.Context, p int) (*big.Int, error) {
	tree := c.reduced.Load()
	if tree == nil {
		rough, err := approx(ctx, c.arg, c.roughPrec)
		if err != nil {
			return nil, err
		}

		root, err := c.reduce(c.arg, rough)
		if err != nil {
			return nil, err
		}

		c.reduced.CompareAndSwap(nil, &reducedTree{root: root})
		tree = c.reduced.Load()
	}

	return approx(ctx, tree.root, p)
}

// asConstruction prints the unreduced form until the first successful
// approximation, and the reduced subtree afterward.
func (c *deferredReduction) asConstruction() string {
	if tree := c.reduced.Load(); tree != nil {
		return tree.root.asConstruction()
	}

	return fmt.Sprintf("%s(%s)", c.name, c.arg.asConstruction())
}
