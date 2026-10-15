package constructive

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type bernoulliTest struct {
	index    int
	expected *big.Rat
}

var bernoulliTests = []bernoulliTest{
	{index: 0, expected: big.NewRat(1, 1)},
	{index: 1, expected: big.NewRat(1, 6)},
	{index: 2, expected: big.NewRat(-1, 30)},
	{index: 3, expected: big.NewRat(1, 42)},
	{index: 4, expected: big.NewRat(-1, 30)},
	{index: 5, expected: big.NewRat(5, 66)},
	{index: 6, expected: big.NewRat(-691, 2730)},
	{index: 7, expected: big.NewRat(7, 6)},
	{index: 8, expected: big.NewRat(-3617, 510)},
	{index: 10, expected: big.NewRat(-174611, 330)},
}

func mustBernoulliEven(t testing.TB, b *bernoulliEvens, k int) *big.Rat {
	t.Helper()

	r, err := b.get(testCtx, k)
	if err != nil {
		t.Fatalf("B_%d: %v", 2*k, err)
	}

	return r
}

// TestBernoulliEven indexes by k, so index k holds B₂ₖ.
func TestBernoulliEven(t *testing.T) {
	for _, test := range bernoulliTests {
		got := mustBernoulliEven(t, bernoulliCache, test.index)
		assert.Zero(t, test.expected.Cmp(got), "B_%d = %s, want %s", 2*test.index, got, test.expected)
	}
}

// TestBernoulliEvenSatisfiesDefiningSum checks Σ_{j=0}^{m} C(m+1, j)·B_j = 0 for
// every odd m = 2k−1 up to a few hundred, which rests on no table of values.
// Odd-index Bernoulli numbers other than B₁ = −1/2 vanish.
func TestBernoulliEvenSatisfiesDefiningSum(t *testing.T) {
	for k := 1; k <= 150; k++ {
		m := 2*k - 1

		sum := new(big.Rat)
		for j := 0; j <= m; j++ {
			var b *big.Rat
			switch {
			case j == 1:
				b = big.NewRat(-1, 2)
			case j%2 == 1:
				continue
			default:
				b = mustBernoulliEven(t, bernoulliCache, j/2)
			}

			binom := new(big.Int).Binomial(int64(m+1), int64(j))
			sum.Add(sum, new(big.Rat).Mul(new(big.Rat).SetInt(binom), b))
		}

		assert.Zero(t, sum.Sign(), "m = %d", m)
	}
}

func TestBernoulliCacheGrowthKeepsValues(t *testing.T) {
	cache := newBernoulliEvens()

	first := mustBernoulliEven(t, cache, 3)
	assert.Zero(t, mustBernoulliEven(t, bernoulliCache, 500).Cmp(mustBernoulliEven(t, cache, 500)))
	assert.Zero(t, first.Cmp(mustBernoulliEven(t, cache, 3)))
	assert.Greater(t, len(*cache.values.Load()), 500)
}

// TestBernoulliCacheGrowsToTheNeededSize checks that a request does not round up to a power
// of two, and that a run of small requests grows geometrically instead of one entry at a time.
func TestBernoulliCacheGrowsToTheNeededSize(t *testing.T) {
	cache := newBernoulliEvens()

	mustBernoulliEven(t, cache, 300)
	assert.Equal(t, 301, len(*cache.values.Load()))

	growths := 0
	size := len(*cache.values.Load())
	for k := 301; k < 600; k++ {
		mustBernoulliEven(t, cache, k)
		if now := len(*cache.values.Load()); now != size {
			growths++
			size = now
		}
	}

	assert.LessOrEqual(t, growths, 5)
}

func TestBernoulliCancelledGrowthLeavesCacheUsable(t *testing.T) {
	cache := newBernoulliEvens()
	small := mustBernoulliEven(t, cache, 5)
	published := cache.values.Load()

	_, err := cache.get(cancelledContext(), 4000)
	assert.ErrorIs(t, err, context.Canceled)
	assert.Same(t, published, cache.values.Load())

	// A cached index still answers under a cancelled context, because it needs no growth.
	got, err := cache.get(cancelledContext(), 5)
	assert.NoError(t, err)
	assert.Same(t, small, got)

	// A computation cancelled part way publishes nothing either.
	ctx, cancel := context.WithTimeout(testCtx, 20*time.Millisecond)
	defer cancel()

	_, err = cache.get(ctx, 6000)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Same(t, published, cache.values.Load())

	assert.Zero(t, mustBernoulliEven(t, bernoulliCache, 40).Cmp(mustBernoulliEven(t, cache, 40)))
	assert.False(t, len(cache.gate) > 0, "the gate was released")
}

// TestBernoulliWaiterHonorsItsContext checks that a caller queued behind a long computation
// gives up when its own context ends.
func TestBernoulliWaiterHonorsItsContext(t *testing.T) {
	cache := newBernoulliEvens()
	cache.gate <- struct{}{} // another caller holds the gate

	ctx, cancel := context.WithTimeout(testCtx, 20*time.Millisecond)
	defer cancel()

	_, err := cache.get(ctx, 10)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	<-cache.gate
	assert.NotNil(t, mustBernoulliEven(t, cache, 10))
}

func TestBernoulliCacheIsSafeForConcurrentUse(t *testing.T) {
	cache := newBernoulliEvens()

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for k := g; k < 200; k += 7 {
				got, err := cache.get(testCtx, k)
				assert.NoError(t, err)
				assert.Zero(t, mustBernoulliEven(t, bernoulliCache, k).Cmp(got))
			}
		}()
	}

	wg.Wait()
}
