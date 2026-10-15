package constructive

import (
	"context"
	"math/big"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// sqrt2Approximation approximates √2 at p by flooring √2·2^-p, which is within one
// unit as the contract requires.
func sqrt2Approximation(p int) *big.Int {
	if p > 0 {
		return big.NewInt(0)
	}

	return new(big.Int).Sqrt(new(big.Int).Lsh(big.NewInt(2), uint(-2*p)))
}

// countingSqrt2 returns an Approximator of √2 that counts its runs.
func countingSqrt2(calls *atomic.Int64) Approximator {
	return func(_ context.Context, p int) (*big.Int, error) {
		calls.Add(1)

		return sqrt2Approximation(p), nil
	}
}

func sqrt2Approximator(_ context.Context, p int) (*big.Int, error) {
	return sqrt2Approximation(p), nil
}

type approximationCompositionTest struct {
	Name  string
	Build func(x Real) Real
}

var approximationCompositionTests = []approximationCompositionTest{
	{
		Name:  "alone",
		Build: func(x Real) Real { return x },
	},
	{
		Name:  "added to one",
		Build: func(x Real) Real { return Add(x, FromInt(1)) },
	},
	{
		Name:  "squared",
		Build: func(x Real) Real { return Multiply(x, x) },
	},
	{
		Name:  "times pi",
		Build: func(x Real) Real { return Multiply(Pi(), x) },
	},
	{
		Name:  "inverted",
		Build: func(x Real) Real { return Inverse(x) },
	},
	{
		Name:  "under exp",
		Build: func(x Real) Real { return Exp(x) },
	},
}

func TestFromApproximationComposesLikeSqrt(t *testing.T) {
	for _, test := range approximationCompositionTests {
		t.Run(test.Name, func(t *testing.T) {
			got := test.Build(FromApproximation("", sqrt2Approximator))
			want := test.Build(Sqrt(FromInt(2)))

			assertEqualAtPrecision(t, got, want, -200)
		})
	}
}

// approximationRead is one read of a shared node, with the total number of
// callback runs expected after it.
type approximationRead struct {
	Precision int
	Calls     int64
}

type approximationCacheTest struct {
	Name  string
	Reads []approximationRead
}

var approximationCacheTests = []approximationCacheTest{
	{
		Name:  "coarser reads come from the cache",
		Reads: []approximationRead{{-100, 1}, {-50, 1}, {-100, 1}, {-1, 1}, {10, 1}},
	},
	{
		Name:  "a finer read runs the callback again",
		Reads: []approximationRead{{-50, 1}, {-100, 2}, {-75, 2}, {-150, 3}},
	},
}

func TestFromApproximationCache(t *testing.T) {
	for _, test := range approximationCacheTests {
		t.Run(test.Name, func(t *testing.T) {
			var calls atomic.Int64
			x := FromApproximation("", countingSqrt2(&calls))

			for _, read := range test.Reads {
				got := mustApproximate(t, x, read.Precision)
				want := mustApproximate(t, Sqrt(FromInt(2)), read.Precision)

				assert.LessOrEqual(t, bigAbs(bigSub(got, want)).Int64(), int64(1), "p=%d", read.Precision)
				assert.Equal(t, read.Calls, calls.Load(), "p=%d", read.Precision)
			}
		})
	}
}

// selfApproximating builds a node whose callback asks for wrap(self) under ctx.
func selfApproximating(label string, wrap func(self Real) Real, ctx func(context.Context) context.Context) Real {
	var self Real
	self = FromApproximation(label, func(c context.Context, p int) (*big.Int, error) {
		return Approximate(ctx(c), wrap(self), p)
	})

	return self
}

func sameContext(ctx context.Context) context.Context {
	return ctx
}

func identity(x Real) Real {
	return x
}

type derivedKey struct{}

type approximationErrorTest struct {
	Name     string
	Build    func() Real
	Expected error
}

var approximationErrorTests = []approximationErrorTest{
	{
		Name: "callback error",
		Build: func() Real {
			return FromApproximation("", func(context.Context, int) (*big.Int, error) {
				return nil, errLeaf
			})
		},
		Expected: errLeaf,
	},
	{
		Name: "callback error beside a value",
		Build: func() Real {
			return FromApproximation("", func(context.Context, int) (*big.Int, error) {
				return big.NewInt(1), errLeaf
			})
		},
		Expected: errLeaf,
	},
	{
		Name: "nil approximation",
		Build: func() Real {
			return FromApproximation("", func(context.Context, int) (*big.Int, error) {
				return nil, nil
			})
		},
		Expected: ErrBadApproximation,
	},
	{
		Name:     "approximates itself",
		Build:    func() Real { return selfApproximating("", identity, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name:     "approximates itself under a label",
		Build:    func() Real { return selfApproximating("√2", identity, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name: "approximates a tree containing itself",
		Build: func() Real {
			return selfApproximating("", func(self Real) Real { return Add(self, FromInt(1)) }, sameContext)
		},
		Expected: ErrReentrantApproximation,
	},
	{
		Name: "approximates itself under a derived context",
		Build: func() Real {
			return selfApproximating("", identity, func(ctx context.Context) context.Context {
				return context.WithValue(ctx, derivedKey{}, true)
			})
		},
		Expected: ErrReentrantApproximation,
	},
	{
		Name: "two callbacks approximate each other",
		Build: func() Real {
			var a, b Real
			a = FromApproximation("", func(ctx context.Context, p int) (*big.Int, error) {
				return Approximate(ctx, b, p)
			})
			b = FromApproximation("", func(ctx context.Context, p int) (*big.Int, error) {
				return Approximate(ctx, a, p)
			})

			return a
		},
		Expected: ErrReentrantApproximation,
	},
}

func TestFromApproximationErrors(t *testing.T) {
	for _, test := range approximationErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			x := test.Build()

			_, err := Approximate(testCtx, x, -20)
			assert.Equal(t, test.Expected, err)
			assert.ErrorIs(t, err, test.Expected)
			assert.Nil(t, x.tracker().latest())
		})
	}
}

func TestFromApproximationRetriesAfterError(t *testing.T) {
	var calls atomic.Int64
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		if calls.Add(1) == 1 {
			return nil, errLeaf
		}

		return sqrt2Approximation(p), nil
	})

	_, err := Approximate(testCtx, x, -20)
	assert.Equal(t, errLeaf, err)

	got, err := Approximate(testCtx, x, -20)
	assert.NoError(t, err)
	assert.Zero(t, got.Cmp(sqrt2Approximation(-20)))
	assert.Equal(t, int64(2), calls.Load())
}

// A nested call under a context unrelated to the run cannot be told from a
// concurrent caller, so it waits until its own context ends.
func TestFromApproximationReentryUnderUnrelatedContextWaits(t *testing.T) {
	unrelated, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	x := selfApproximating("", identity, func(context.Context) context.Context {
		return unrelated
	})

	_, err := Approximate(testCtx, x, -20)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, x.tracker().latest())
}

func TestFromApproximationCachedReadInsideCallback(t *testing.T) {
	var x Real
	var nested *big.Int
	x = FromApproximation("", func(ctx context.Context, p int) (*big.Int, error) {
		if p < -10 {
			v, err := Approximate(ctx, x, -10)
			if err != nil {
				return nil, err
			}
			nested = v
		}

		return sqrt2Approximation(p), nil
	})

	mustApproximate(t, x, -10)
	mustApproximate(t, x, -40)
	assert.Zero(t, nested.Cmp(sqrt2Approximation(-10)))
}

func TestFromApproximationCopiesResult(t *testing.T) {
	buf := new(big.Int)
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		return buf.Set(sqrt2Approximation(p)), nil
	})

	got := mustApproximate(t, x, -40)
	assert.Zero(t, got.Cmp(sqrt2Approximation(-40)))

	buf.SetInt64(0)
	assert.Zero(t, mustApproximate(t, x, -40).Cmp(sqrt2Approximation(-40)))
}

// TestFromApproximationConcurrentReaders shares one node across goroutines that
// ask for different precisions. None is rejected as re-entrant, and the callback
// never runs concurrently with itself. Run with -race.
func TestFromApproximationConcurrentReaders(t *testing.T) {
	precisions := []int{-10, -50, -200, -30, -400, -100, -5, -300}

	var active, overlaps atomic.Int64
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		if active.Add(1) > 1 {
			overlaps.Add(1)
		}
		defer active.Add(-1)

		runtime.Gosched()

		return sqrt2Approximation(p), nil
	})

	var wg sync.WaitGroup
	results := make([]*big.Int, 4*len(precisions))
	errs := make([]error, len(results))
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = Approximate(testCtx, x, precisions[i%len(precisions)])
		}()
	}
	wg.Wait()

	assert.Zero(t, overlaps.Load())
	for i, got := range results {
		p := precisions[i%len(precisions)]
		if !assert.NoError(t, errs[i], "p=%d", p) {
			continue
		}

		diff := bigAbs(bigSub(got, sqrt2Approximation(p)))
		assert.LessOrEqual(t, diff.Int64(), int64(1), "p=%d", p)
	}
}

func TestFromApproximationWaiterHonorsContext(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		close(entered)
		<-release

		return sqrt2Approximation(p), nil
	})

	done := make(chan error)
	go func() {
		_, err := Approximate(testCtx, x, -20)
		done <- err
	}()
	<-entered

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := Approximate(ctx, x, -40)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	assert.NoError(t, <-done)
}

type approximationConstructionTest struct {
	Name         string
	Label        string
	Construction string
	HasName      bool
}

var approximationConstructionTests = []approximationConstructionTest{
	{
		Name:         "labelled",
		Label:        "√2",
		Construction: `Named("√2", Approximation())`,
		HasName:      true,
	},
	{
		Name:         "unlabelled",
		Construction: "Approximation()",
	},
}

func TestFromApproximationConstruction(t *testing.T) {
	for _, test := range approximationConstructionTests {
		t.Run(test.Name, func(t *testing.T) {
			x := FromApproximation(test.Label, sqrt2Approximator)
			assert.Equal(t, test.Construction, AsConstruction(x))

			name, ok := ConstructiveName(x)
			assert.Equal(t, test.HasName, ok)
			assert.Equal(t, test.Label, name)
		})
	}
}

func TestFromApproximationNilCallback(t *testing.T) {
	assert.Nil(t, FromApproximation("", nil))
	assert.Nil(t, FromApproximation("√2", nil))
}
