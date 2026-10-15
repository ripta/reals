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

// sqrtApproximation approximates √k at p by flooring √k·2^-p, which is within one
// unit as the contract requires.
func sqrtApproximation(k int, p int) *big.Int {
	if p > 0 {
		return big.NewInt(0)
	}

	return new(big.Int).Sqrt(new(big.Int).Lsh(big.NewInt(int64(k)), uint(-2*p)))
}

// countingSqrt2 returns an Approximator of √2 that counts its runs.
func countingSqrt2(calls *atomic.Int64) Approximator {
	return func(_ context.Context, p int) (*big.Int, error) {
		calls.Add(1)

		return sqrtApproximation(2, p), nil
	}
}

func sqrt2Approximator(_ context.Context, p int) (*big.Int, error) {
	return sqrtApproximation(2, p), nil
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

		return sqrtApproximation(2, p), nil
	})

	_, err := Approximate(testCtx, x, -20)
	assert.Equal(t, errLeaf, err)

	got, err := Approximate(testCtx, x, -20)
	assert.NoError(t, err)
	assert.Zero(t, got.Cmp(sqrtApproximation(2, -20)))
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

		return sqrtApproximation(2, p), nil
	})

	mustApproximate(t, x, -10)
	mustApproximate(t, x, -40)
	assert.Zero(t, nested.Cmp(sqrtApproximation(2, -10)))
}

func TestFromApproximationCopiesResult(t *testing.T) {
	buf := new(big.Int)
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		return buf.Set(sqrtApproximation(2, p)), nil
	})

	got := mustApproximate(t, x, -40)
	assert.Zero(t, got.Cmp(sqrtApproximation(2, -40)))

	buf.SetInt64(0)
	assert.Zero(t, mustApproximate(t, x, -40).Cmp(sqrtApproximation(2, -40)))
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

		return sqrtApproximation(2, p), nil
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

		diff := bigAbs(bigSub(got, sqrtApproximation(2, p)))
		assert.LessOrEqual(t, diff.Int64(), int64(1), "p=%d", p)
	}
}

func TestFromApproximationWaiterHonorsContext(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	x := FromApproximation("", func(_ context.Context, p int) (*big.Int, error) {
		close(entered)
		<-release

		return sqrtApproximation(2, p), nil
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

// sliceRoots are the radicands behind the outputs of the slice callbacks below.
var sliceRoots = []int{2, 3, 5}

// rootsApproximator returns a SliceApproximator whose output i is √sliceRoots[i],
// counting its runs. A nil calls counts nothing.
func rootsApproximator(calls *atomic.Int64) SliceApproximator {
	return func(_ context.Context, p int) ([]*big.Int, error) {
		if calls != nil {
			calls.Add(1)
		}

		out := make([]*big.Int, len(sliceRoots))
		for i, k := range sliceRoots {
			out[i] = sqrtApproximation(k, p)
		}

		return out, nil
	}
}

func rootsSlice(calls *atomic.Int64) []Real {
	return FromApproximationSlice("", len(sliceRoots), rootsApproximator(calls))
}

// assertOutputsEmpty checks that no output holds a cached approximation.
func assertOutputsEmpty(t *testing.T, outputs []Real) {
	t.Helper()

	for i, c := range outputs {
		assert.Nil(t, c.tracker().latest(), "output %d", i)
	}
}

func TestFromApproximationSliceComposesLikeSqrt(t *testing.T) {
	for _, test := range approximationCompositionTests {
		t.Run(test.Name, func(t *testing.T) {
			outputs := rootsSlice(nil)
			for i, k := range sliceRoots {
				got := test.Build(outputs[i])
				want := test.Build(Sqrt(FromInt(k)))

				assertEqualAtPrecision(t, got, want, -200)
			}
		})
	}
}

type approximationAcrossTest struct {
	Name  string
	Build func(outputs []Real) Real
	Want  func() Real
}

var approximationAcrossTests = []approximationAcrossTest{
	{
		Name:  "sum of two outputs",
		Build: func(o []Real) Real { return Add(o[0], o[1]) },
		Want:  func() Real { return Add(Sqrt(FromInt(2)), Sqrt(FromInt(3))) },
	},
	{
		Name:  "product of two outputs",
		Build: func(o []Real) Real { return Multiply(o[0], o[1]) },
		Want:  func() Real { return Sqrt(FromInt(6)) },
	},
	{
		Name:  "an output beside itself and a sibling",
		Build: func(o []Real) Real { return Add(Multiply(o[2], o[2]), Negate(o[0])) },
		Want:  func() Real { return Add(FromInt(5), Negate(Sqrt(FromInt(2)))) },
	},
}

func TestFromApproximationSliceComposesAcrossOutputs(t *testing.T) {
	for _, test := range approximationAcrossTests {
		t.Run(test.Name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.Build(rootsSlice(nil)), test.Want(), -200)
		})
	}
}

// sliceRead is one read of an output of a shared set, with the total number of
// callback runs expected after it.
type sliceRead struct {
	Output    int
	Precision int
	Calls     int64
}

type approximationSliceCacheTest struct {
	Name  string
	Reads []sliceRead
}

var approximationSliceCacheTests = []approximationSliceCacheTest{
	{
		Name:  "siblings at the same precision come from the cache",
		Reads: []sliceRead{{0, -100, 1}, {1, -100, 1}, {2, -100, 1}, {0, -100, 1}},
	},
	{
		Name:  "siblings at a coarser precision come from the cache",
		Reads: []sliceRead{{1, -100, 1}, {0, -50, 1}, {2, -1, 1}, {2, 10, 1}},
	},
	{
		Name:  "a finer read of any output runs the callback again",
		Reads: []sliceRead{{0, -50, 1}, {1, -100, 2}, {2, -75, 2}, {2, -150, 3}, {0, -150, 3}, {1, -120, 3}},
	},
}

func TestFromApproximationSliceCache(t *testing.T) {
	for _, test := range approximationSliceCacheTests {
		t.Run(test.Name, func(t *testing.T) {
			var calls atomic.Int64
			outputs := rootsSlice(&calls)

			for _, read := range test.Reads {
				got := mustApproximate(t, outputs[read.Output], read.Precision)
				want := mustApproximate(t, Sqrt(FromInt(sliceRoots[read.Output])), read.Precision)

				assert.LessOrEqual(t, bigAbs(bigSub(got, want)).Int64(), int64(1), "%+v", read)
				assert.Equal(t, read.Calls, calls.Load(), "%+v", read)
			}
		})
	}
}

// A new finest precision for one output refines every output. Coarser reads of the
// others are roundings of that one run, not of runs of their own.
func TestFromApproximationSliceCoarserReadsRoundOneRun(t *testing.T) {
	finest := -100
	coarser := []int{-100, -64, -17, -1, 0, 5}

	var calls atomic.Int64
	outputs := rootsSlice(&calls)
	mustApproximate(t, outputs[0], finest)

	for _, p := range coarser {
		for i, k := range sliceRoots {
			got := mustApproximate(t, outputs[i], p)
			want := scale(sqrtApproximation(k, finest), finest-p)

			assert.Zero(t, got.Cmp(want), "output %d, p=%d", i, p)
		}
	}

	assert.Equal(t, int64(1), calls.Load())
}

type approximationSliceErrorTest struct {
	Name     string
	Build    func() []Real
	Expected error
}

// slicedWith returns two outputs whose callback returns what result gives.
func slicedWith(result func() ([]*big.Int, error)) []Real {
	return FromApproximationSlice("", 2, func(context.Context, int) ([]*big.Int, error) {
		return result()
	})
}

// slicedReading returns two outputs whose callback approximates outputs[target]
// at p under ctx.
func slicedReading(label string, target int, ctx func(context.Context) context.Context) []Real {
	var outputs []Real
	outputs = FromApproximationSlice(label, 2, func(c context.Context, p int) ([]*big.Int, error) {
		v, err := Approximate(ctx(c), outputs[target], p)
		if err != nil {
			return nil, err
		}

		return []*big.Int{v, v}, nil
	})

	return outputs
}

var approximationSliceErrorTests = []approximationSliceErrorTest{
	{
		Name: "callback error",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) { return nil, errLeaf })
		},
		Expected: errLeaf,
	},
	{
		Name: "callback error beside values",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) {
				return []*big.Int{big.NewInt(1), big.NewInt(2)}, errLeaf
			})
		},
		Expected: errLeaf,
	},
	{
		Name: "nil slice",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) { return nil, nil })
		},
		Expected: ErrBadApproximation,
	},
	{
		Name: "slice too short",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) { return []*big.Int{big.NewInt(1)}, nil })
		},
		Expected: ErrBadApproximation,
	},
	{
		Name: "slice too long",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) {
				return []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}, nil
			})
		},
		Expected: ErrBadApproximation,
	},
	{
		Name: "nil first element",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) { return []*big.Int{nil, big.NewInt(2)}, nil })
		},
		Expected: ErrBadApproximation,
	},
	{
		Name: "nil last element",
		Build: func() []Real {
			return slicedWith(func() ([]*big.Int, error) { return []*big.Int{big.NewInt(1), nil}, nil })
		},
		Expected: ErrBadApproximation,
	},
	{
		Name:     "approximates itself",
		Build:    func() []Real { return slicedReading("", 0, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name:     "approximates a sibling",
		Build:    func() []Real { return slicedReading("", 1, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name:     "approximates itself under a label",
		Build:    func() []Real { return slicedReading("x", 0, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name:     "approximates a sibling under a label",
		Build:    func() []Real { return slicedReading("x", 1, sameContext) },
		Expected: ErrReentrantApproximation,
	},
	{
		Name: "approximates a sibling under a derived context",
		Build: func() []Real {
			return slicedReading("", 1, func(ctx context.Context) context.Context {
				return context.WithValue(ctx, derivedKey{}, true)
			})
		},
		Expected: ErrReentrantApproximation,
	},
}

func TestFromApproximationSliceErrors(t *testing.T) {
	for _, test := range approximationSliceErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			outputs := test.Build()

			for i, c := range outputs {
				_, err := Approximate(testCtx, c, -20)
				assert.Equal(t, test.Expected, err, "output %d", i)
				assert.ErrorIs(t, err, test.Expected, "output %d", i)
				assertOutputsEmpty(t, outputs)
			}
		})
	}
}

type approximationSliceKeepsCacheTest struct {
	Name     string
	Fail     func() ([]*big.Int, error)
	Expected error
}

var approximationSliceKeepsCacheTests = []approximationSliceKeepsCacheTest{
	{
		Name:     "callback error",
		Fail:     func() ([]*big.Int, error) { return nil, errLeaf },
		Expected: errLeaf,
	},
	{
		Name: "callback error beside values",
		Fail: func() ([]*big.Int, error) {
			return []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3)}, errLeaf
		},
		Expected: errLeaf,
	},
	{
		Name:     "slice too short",
		Fail:     func() ([]*big.Int, error) { return []*big.Int{big.NewInt(1)}, nil },
		Expected: ErrBadApproximation,
	},
	{
		Name: "slice too long",
		Fail: func() ([]*big.Int, error) {
			return []*big.Int{big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4)}, nil
		},
		Expected: ErrBadApproximation,
	},
	{
		Name: "nil element after valid ones",
		Fail: func() ([]*big.Int, error) {
			return []*big.Int{big.NewInt(1), big.NewInt(2), nil}, nil
		},
		Expected: ErrBadApproximation,
	},
}

// A run that fails at a finer precision leaves every output holding the coarser
// approximation an earlier run cached, and a read at that coarser precision still
// does not run the callback.
func TestFromApproximationSliceFailureKeepsCachedApproximations(t *testing.T) {
	const coarse, fine = -20, -100

	for _, test := range approximationSliceKeepsCacheTests {
		t.Run(test.Name, func(t *testing.T) {
			var calls atomic.Int64
			outputs := FromApproximationSlice("", len(sliceRoots), func(ctx context.Context, p int) ([]*big.Int, error) {
				calls.Add(1)
				if p < coarse {
					return test.Fail()
				}

				return rootsApproximator(nil)(ctx, p)
			})

			mustApproximate(t, outputs[0], coarse)
			assert.Equal(t, int64(1), calls.Load())

			cached := make([]*approximation, len(outputs))
			for i, c := range outputs {
				cached[i] = c.tracker().latest()
				if !assert.NotNil(t, cached[i], "output %d", i) {
					return
				}
			}

			for k, c := range outputs {
				_, err := Approximate(testCtx, c, fine)
				assert.Equal(t, test.Expected, err, "output %d", k)
				assert.ErrorIs(t, err, test.Expected, "output %d", k)

				for i, o := range outputs {
					got := o.tracker().latest()
					assert.Same(t, cached[i], got, "output %d after reading output %d", i, k)
					assert.Equal(t, coarse, got.precision, "output %d after reading output %d", i, k)
					assert.Zero(t, got.value.Cmp(sqrtApproximation(sliceRoots[i], coarse)), "output %d after reading output %d", i, k)
				}
			}

			assert.Equal(t, int64(1+len(outputs)), calls.Load())
			for i, k := range sliceRoots {
				assert.Zero(t, mustApproximate(t, outputs[i], coarse).Cmp(sqrtApproximation(k, coarse)), "output %d", i)
			}
			assert.Equal(t, int64(1+len(outputs)), calls.Load())
		})
	}
}

func TestFromApproximationSliceRetriesAfterError(t *testing.T) {
	var calls atomic.Int64
	outputs := FromApproximationSlice("", len(sliceRoots), func(ctx context.Context, p int) ([]*big.Int, error) {
		if calls.Add(1) == 1 {
			return nil, errLeaf
		}

		return rootsApproximator(nil)(ctx, p)
	})

	_, err := Approximate(testCtx, outputs[1], -20)
	assert.Equal(t, errLeaf, err)
	assertOutputsEmpty(t, outputs)

	got, err := Approximate(testCtx, outputs[1], -20)
	assert.NoError(t, err)
	assert.Zero(t, got.Cmp(sqrtApproximation(3, -20)))
	assert.Equal(t, int64(2), calls.Load())

	for i, k := range sliceRoots {
		assert.Zero(t, mustApproximate(t, outputs[i], -20).Cmp(sqrtApproximation(k, -20)), "output %d", i)
	}
	assert.Equal(t, int64(2), calls.Load())
}

// A nested read of an output at a precision a run already covers is served from
// its cache, so it neither deadlocks nor fails as re-entrant.
func TestFromApproximationSliceCachedReadInsideCallback(t *testing.T) {
	var outputs []Real
	var nested *big.Int
	outputs = FromApproximationSlice("", 2, func(ctx context.Context, p int) ([]*big.Int, error) {
		if p < -10 {
			v, err := Approximate(ctx, outputs[1], -10)
			if err != nil {
				return nil, err
			}
			nested = v
		}

		return []*big.Int{sqrtApproximation(2, p), sqrtApproximation(3, p)}, nil
	})

	mustApproximate(t, outputs[0], -10)
	mustApproximate(t, outputs[0], -40)
	assert.Zero(t, nested.Cmp(sqrtApproximation(3, -10)))
}

func TestFromApproximationSliceCopiesResults(t *testing.T) {
	shared := new(big.Int)
	bufs := []*big.Int{new(big.Int), new(big.Int)}
	outputs := FromApproximationSlice("", 3, func(_ context.Context, p int) ([]*big.Int, error) {
		shared.Set(sqrtApproximation(2, p))
		bufs[0].Set(sqrtApproximation(3, p))
		bufs[1].Set(sqrtApproximation(5, p))

		// Two outputs hand back the same integer.
		return []*big.Int{shared, shared, bufs[1]}, nil
	})

	first := mustApproximate(t, outputs[0], -40)
	second := mustApproximate(t, outputs[1], -40)
	third := mustApproximate(t, outputs[2], -40)
	assert.NotSame(t, shared, first)
	assert.NotSame(t, first, second)
	assert.NotSame(t, shared, third)
	assert.NotSame(t, bufs[1], third)

	shared.SetInt64(0)
	bufs[1].SetInt64(0)
	assert.Zero(t, mustApproximate(t, outputs[0], -40).Cmp(sqrtApproximation(2, -40)))
	assert.Zero(t, mustApproximate(t, outputs[1], -40).Cmp(sqrtApproximation(2, -40)))
	assert.Zero(t, mustApproximate(t, outputs[2], -40).Cmp(sqrtApproximation(5, -40)))
}

// TestFromApproximationSliceConcurrentReaders shares one set of outputs across
// goroutines that ask for different outputs at different precisions. None is
// rejected as re-entrant, and the callback never runs concurrently with itself.
// Run with -race.
func TestFromApproximationSliceConcurrentReaders(t *testing.T) {
	precisions := []int{-10, -50, -200, -30, -400, -100, -5, -300}

	var active, overlaps atomic.Int64
	outputs := FromApproximationSlice("", len(sliceRoots), func(ctx context.Context, p int) ([]*big.Int, error) {
		if active.Add(1) > 1 {
			overlaps.Add(1)
		}
		defer active.Add(-1)

		runtime.Gosched()

		return rootsApproximator(nil)(ctx, p)
	})

	var wg sync.WaitGroup
	results := make([]*big.Int, 4*len(precisions)*len(sliceRoots))
	errs := make([]error, len(results))
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = Approximate(testCtx, outputs[i%len(sliceRoots)], precisions[(i/len(sliceRoots))%len(precisions)])
		}()
	}
	wg.Wait()

	assert.Zero(t, overlaps.Load())
	for i, got := range results {
		k := sliceRoots[i%len(sliceRoots)]
		p := precisions[(i/len(sliceRoots))%len(precisions)]
		if !assert.NoError(t, errs[i], "√%d, p=%d", k, p) {
			continue
		}

		diff := bigAbs(bigSub(got, sqrtApproximation(k, p)))
		assert.LessOrEqual(t, diff.Int64(), int64(1), "√%d, p=%d", k, p)
	}
}

// A reader of one output that waited for a run started by another output finds its
// own output already written by that run.
//
// The waiter signals just before it calls Approximate, and the test releases the run
// only after that signal and a short pause. A waiter that has not yet reached the
// guard by then reads the finished run through the cache fast path instead. The
// test still passes, but it no longer exercises the re-check under the guard. That
// remaining window is accepted, because nothing in the package can report that a
// caller is blocked on the guard.
func TestFromApproximationSliceWaiterOnSiblingSharesRun(t *testing.T) {
	var calls atomic.Int64
	entered := make(chan struct{})
	release := make(chan struct{})
	outputs := FromApproximationSlice("", len(sliceRoots), func(ctx context.Context, p int) ([]*big.Int, error) {
		if calls.Add(1) == 1 {
			close(entered)
			<-release
		}

		return rootsApproximator(nil)(ctx, p)
	})

	done := make(chan error, 2)
	go func() {
		_, err := Approximate(testCtx, outputs[0], -40)
		done <- err
	}()
	<-entered

	waiting := make(chan struct{})
	go func() {
		close(waiting)
		_, err := Approximate(testCtx, outputs[1], -40)
		done <- err
	}()
	<-waiting
	time.Sleep(10 * time.Millisecond)
	close(release)

	assert.NoError(t, <-done)
	assert.NoError(t, <-done)
	assert.Equal(t, int64(1), calls.Load())
}

func TestFromApproximationSliceWaiterHonorsContext(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	outputs := FromApproximationSlice("", len(sliceRoots), func(ctx context.Context, p int) ([]*big.Int, error) {
		close(entered)
		<-release

		return rootsApproximator(nil)(ctx, p)
	})

	done := make(chan error)
	go func() {
		_, err := Approximate(testCtx, outputs[0], -20)
		done <- err
	}()
	<-entered

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := Approximate(ctx, outputs[1], -40)
	assert.ErrorIs(t, err, context.DeadlineExceeded)

	close(release)
	assert.NoError(t, <-done)
}

type approximationSliceConstructionTest struct {
	Name          string
	Label         string
	Constructions []string
	HasName       bool
}

var approximationSliceConstructionTests = []approximationSliceConstructionTest{
	{
		Name:          "labelled",
		Label:         "x",
		Constructions: []string{`Named("x", Approximation(0, 3))`, `Named("x", Approximation(1, 3))`, `Named("x", Approximation(2, 3))`},
		HasName:       true,
	},
	{
		Name:          "unlabelled",
		Constructions: []string{"Approximation(0, 3)", "Approximation(1, 3)", "Approximation(2, 3)"},
	},
}

func TestFromApproximationSliceConstruction(t *testing.T) {
	for _, test := range approximationSliceConstructionTests {
		t.Run(test.Name, func(t *testing.T) {
			outputs := FromApproximationSlice(test.Label, len(test.Constructions), rootsApproximator(nil))
			if !assert.Len(t, outputs, len(test.Constructions)) {
				return
			}

			for i, c := range outputs {
				assert.Equal(t, test.Constructions[i], AsConstruction(c))

				name, ok := ConstructiveName(c)
				assert.Equal(t, test.HasName, ok)
				assert.Equal(t, test.Label, name)
			}
		})
	}
}

type approximationSliceDegenerateTest struct {
	Name  string
	Label string
	N     int
	Fn    SliceApproximator
	IsNil bool
}

var approximationSliceDegenerateTests = []approximationSliceDegenerateTest{
	{Name: "nil callback", N: 3, IsNil: true},
	{Name: "nil callback with a label", Label: "x", N: 3, IsNil: true},
	{Name: "nil callback and zero outputs", N: 0, IsNil: true},
	{Name: "nil callback and negative outputs", Label: "x", N: -1, IsNil: true},
	{Name: "zero outputs", Fn: rootsApproximator(nil), N: 0},
	{Name: "zero outputs with a label", Label: "x", Fn: rootsApproximator(nil), N: 0},
	{Name: "negative outputs", Fn: rootsApproximator(nil), N: -2},
	{Name: "negative outputs with a label", Label: "x", Fn: rootsApproximator(nil), N: -2},
}

func TestFromApproximationSliceDegenerateArguments(t *testing.T) {
	for _, test := range approximationSliceDegenerateTests {
		t.Run(test.Name, func(t *testing.T) {
			outputs := FromApproximationSlice(test.Label, test.N, test.Fn)
			if test.IsNil {
				assert.Nil(t, outputs)
				return
			}

			assert.NotNil(t, outputs)
			assert.Empty(t, outputs)
		})
	}
}

// One output is the single form's case n = 1: both build the same kind of node.
func TestFromApproximationSliceOfOne(t *testing.T) {
	outputs := FromApproximationSlice("", 1, func(ctx context.Context, p int) ([]*big.Int, error) {
		return []*big.Int{sqrtApproximation(2, p)}, nil
	})
	if !assert.Len(t, outputs, 1) {
		return
	}

	assertEqualAtPrecision(t, outputs[0], Sqrt(FromInt(2)), -200)
	assert.Equal(t, "Approximation(0, 1)", AsConstruction(outputs[0]))
}
