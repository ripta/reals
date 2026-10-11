package constructive

import (
	"math/big"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

type trackerSetTest struct {
	Name       string
	Precisions []int
	Finest     int
}

var trackerSetTests = []trackerSetTest{
	{Name: "finer replaces coarser", Precisions: []int{-10, -20}, Finest: -20},
	{Name: "coarser keeps finer", Precisions: []int{-20, -10}, Finest: -20},
	{Name: "equal keeps first", Precisions: []int{-20, -20}, Finest: -20},
	{Name: "mixed", Precisions: []int{-5, -40, -15, -30}, Finest: -40},
}

func TestTrackerSetKeepsFinest(t *testing.T) {
	for _, test := range trackerSetTests {
		t.Run(test.Name, func(t *testing.T) {
			var tr precisionTracker
			for _, p := range test.Precisions {
				tr.Set(p, big.NewInt(int64(-p)))
			}

			a := tr.latest()
			assert.Equal(t, test.Finest, a.precision)
			assert.Equal(t, int64(-test.Finest), a.value.Int64())
		})
	}
}

func TestTrackerConcurrentSetKeepsFinest(t *testing.T) {
	var tr precisionTracker
	var wg sync.WaitGroup
	for p := -1; p >= -200; p-- {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tr.Set(p, big.NewInt(int64(-p)))
		}()
	}
	wg.Wait()

	a := tr.latest()
	assert.Equal(t, -200, a.precision)
	assert.Equal(t, int64(200), a.value.Int64())
}

// sharedTree mixes eager and deferred nodes, so goroutines race on caches at
// several depths as well as on the deferred reduction.
func sharedTree() Real {
	return Add(
		Multiply(Exp(Sqrt2()), Ln(FromInt(3))),
		Divide(Pi(), Arctangent(FromRat(1, 3))),
	)
}

// TestConcurrentApproximation shares one tree across goroutines that ask for
// different precisions. Every result must be within one unit of the true value,
// as must a reference from a fresh tree, so the two are within two of each other.
// Run with -race to check the caches themselves.
func TestConcurrentApproximation(t *testing.T) {
	precisions := []int{-10, -50, -200, -30, -400, -100, -5, -300}

	want := make(map[int]*big.Int, len(precisions))
	for _, p := range precisions {
		want[p] = mustApproximate(t, sharedTree(), p)
	}

	shared := sharedTree()

	var wg sync.WaitGroup
	results := make([]*big.Int, 4*len(precisions))
	errs := make([]error, len(results))
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = Approximate(testCtx, shared, precisions[i%len(precisions)])
		}()
	}
	wg.Wait()

	for i, got := range results {
		p := precisions[i%len(precisions)]
		if !assert.NoError(t, errs[i], "p=%d", p) {
			continue
		}

		diff := bigAbs(bigSub(got, want[p]))
		assert.LessOrEqual(t, diff.Int64(), int64(2), "p=%d: got %s, want %s", p, got, want[p])
	}
}
