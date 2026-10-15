package constructive

import (
	"context"
	"math/big"
	"sync/atomic"
)

const (
	// bernoulliMinCache is the smallest number of even-index Bernoulli numbers the cache holds
	// once anything asks for one.
	bernoulliMinCache = 32

	// bernoulliGrowthNum and bernoulliGrowthDen give the factor by which a growing cache
	// outgrows its current size. The tangent-number recurrence costs about n³ bit operations
	// and is not incremental, so every growth recomputes from scratch. A factor of 5/4 keeps a
	// run of small requests amortized, without rounding a needed size far up.
	bernoulliGrowthNum = 5
	bernoulliGrowthDen = 4
)

// bernoulliEvens caches B₀, B₂, B₄, … as exact rationals, so index k holds B₂ₖ.
//
// Readers take the published slice without locking. A reader that needs more takes the gate,
// which waits under the caller's context, so a caller whose context ends never stays behind
// another caller's long computation. Computing is cancellable, and a cancelled computation
// publishes nothing.
//
// Cached values are shared among every caller and must be treated as read-only.
type bernoulliEvens struct {
	values atomic.Pointer[[]*big.Rat]
	gate   chan struct{}
}

func newBernoulliEvens() *bernoulliEvens {
	return &bernoulliEvens{gate: make(chan struct{}, 1)}
}

var bernoulliCache = newBernoulliEvens()

// bernoulliEven returns the Bernoulli number B₂ₖ for k ≥ 0 as an exact rational. The result
// is shared; callers must not modify it. The error is the context's error when growing the
// cache is cancelled.
func bernoulliEven(ctx context.Context, k int) (*big.Rat, error) {
	return bernoulliCache.get(ctx, k)
}

func (b *bernoulliEvens) get(ctx context.Context, k int) (*big.Rat, error) {
	if values := b.values.Load(); values != nil && k < len(*values) {
		return (*values)[k], nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	select {
	case b.gate <- struct{}{}:
		defer func() { <-b.gate }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Another caller may have grown the cache while this one waited.
	var have int
	if values := b.values.Load(); values != nil {
		if k < len(*values) {
			return (*values)[k], nil
		}

		have = len(*values)
	}

	grown, err := computeBernoulliEvens(ctx, max(k+1, have*bernoulliGrowthNum/bernoulliGrowthDen, bernoulliMinCache))
	if err != nil {
		return nil, err
	}

	b.values.Store(&grown)

	return grown[k], nil
}

// computeBernoulliEvens returns B₀, B₂, …, B₂ₙ₋₂ as exact rationals, or the context's error.
//
// The numerators come from the tangent numbers Tₖ, which are integers, by
//
//	B₂ₖ = (−1)^(k−1) · 2k · Tₖ / (2^(2k) · (2^(2k) − 1)).
//
// The tangent numbers follow the recurrence of Brent and Harvey, "Fast computation of
// Bernoulli, tangent and secant numbers" (2011), algorithm TangentNumbers. It needs only
// integer arithmetic and O(n²) operations.
func computeBernoulliEvens(ctx context.Context, n int) ([]*big.Rat, error) {
	tangent := make([]*big.Int, n+1)
	for k := range tangent {
		tangent[k] = new(big.Int)
	}

	if n >= 1 {
		tangent[1].SetInt64(1)
	}

	for k := 2; k < n; k++ {
		tangent[k].Mul(tangent[k-1], big.NewInt(int64(k-1)))
	}

	var left, right big.Int
	for k := 2; k < n; k++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		for j := k; j < n; j++ {
			left.Mul(tangent[j-1], big.NewInt(int64(j-k)))
			right.Mul(tangent[j], big.NewInt(int64(j-k+2)))
			tangent[j].Add(&left, &right)
		}
	}

	values := make([]*big.Rat, n)
	values[0] = big.NewRat(1, 1)

	for k := 1; k < n; k++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		pow := new(big.Int).Lsh(big.NewInt(1), uint(2*k)) // 2^(2k)
		denom := new(big.Int).Mul(pow, new(big.Int).Sub(pow, big.NewInt(1)))
		num := new(big.Int).Mul(tangent[k], big.NewInt(int64(2*k)))
		if k%2 == 0 {
			num.Neg(num)
		}

		values[k] = new(big.Rat).SetFrac(num, denom)
	}

	return values, nil
}
