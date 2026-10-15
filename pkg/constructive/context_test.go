package constructive

import (
	"context"
	"errors"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errLeaf = errors.New("leaf failure")

// failingReal is a leaf that always fails with err.
type failingReal struct {
	precisionTracker
	err   error
	calls int
}

func (f *failingReal) approximate(_ context.Context, _ int) (*big.Int, error) {
	f.calls++

	return nil, f.err
}

func (f *failingReal) asConstruction() string {
	return "Failing()"
}

// flakyReal is a leaf that fails its first few approximations, then succeeds.
type flakyReal struct {
	precisionTracker
	failures int
	calls    int
}

func (f *flakyReal) approximate(_ context.Context, p int) (*big.Int, error) {
	f.calls++
	if f.calls <= f.failures {
		return nil, errLeaf
	}

	return scale(big.NewInt(7), -p), nil
}

func (f *flakyReal) asConstruction() string {
	return "Flaky()"
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

// nodeTest builds one node type over the given child. HasChild is false for a
// node that never approximates another Real.
type nodeTest struct {
	Name     string
	Build    func(child Real) Real
	HasChild bool
}

var nodeTests = []nodeTest{
	{
		Name:  "integer",
		Build: func(_ Real) Real { return newInteger(big.NewInt(5)) },
	},
	{
		Name:     "addition, first operand",
		Build:    func(child Real) Real { return newAddition(child, FromInt(1)) },
		HasChild: true,
	},
	{
		Name:     "addition, second operand",
		Build:    func(child Real) Real { return newAddition(FromInt(1), child) },
		HasChild: true,
	},
	{
		Name:     "multiplication, first operand",
		Build:    func(child Real) Real { return newMultiplication(child, FromInt(2)) },
		HasChild: true,
	},
	{
		Name:     "multiplication, second operand",
		Build:    func(child Real) Real { return newMultiplication(FromInt(2), child) },
		HasChild: true,
	},
	{
		Name:     "multiplicative inverse",
		Build:    func(child Real) Real { return newMultiplicativeInverse(child) },
		HasChild: true,
	},
	{
		Name:     "shift",
		Build:    func(child Real) Real { return newShift(child, 3) },
		HasChild: true,
	},
	{
		Name:     "negation",
		Build:    func(child Real) Real { return newNegation(child) },
		HasChild: true,
	},
	{
		Name:     "condsign, selector",
		Build:    func(child Real) Real { return newCondsign(child, FromInt(1), FromInt(2)) },
		HasChild: true,
	},
	{
		Name:     "condsign, selected branch",
		Build:    func(child Real) Real { return newCondsign(FromInt(1), FromInt(2), child) },
		HasChild: true,
	},
	{
		Name:     "prescaled exponential",
		Build:    func(child Real) Real { return newPrescaledExponential(child) },
		HasChild: true,
	},
	{
		Name:     "prescaled natural log",
		Build:    func(child Real) Real { return newPrescaledNaturalLog(child) },
		HasChild: true,
	},
	{
		Name:     "integral arctangent",
		Build:    func(child Real) Real { return newIntegralArctan(child) },
		HasChild: true,
	},
	{
		Name:     "prescaled square root",
		Build:    func(child Real) Real { return newPrescaledSqrt(child) },
		HasChild: true,
	},
	{
		Name:     "prescaled arctangent",
		Build:    func(child Real) Real { return newPrescaledArctan(child) },
		HasChild: true,
	},
	{
		Name:     "prescaled cosine",
		Build:    func(child Real) Real { return newPrescaledCosine(child) },
		HasChild: true,
	},
	{
		Name:  "prime constant summation",
		Build: func(_ Real) Real { return newPrimeConstantSummation() },
	},
	{
		Name: "approximation callback",
		Build: func(_ Real) Real {
			return newApproximator(func(_ context.Context, p int) (*big.Int, error) {
				return scale(big.NewInt(5), -p), nil
			})
		},
	},
	{
		Name: "approximation callback, slice output",
		Build: func(_ Real) Real {
			return newApproximatorSlice(2, func(_ context.Context, p int) ([]*big.Int, error) {
				return []*big.Int{scale(big.NewInt(5), -p), scale(big.NewInt(7), -p)}, nil
			})[1]
		},
	},
	{
		Name:     "arithmetic-geometric mean, first operand",
		Build:    func(child Real) Real { return newArithmeticGeometricMean(child, FromInt(2)) },
		HasChild: true,
	},
	{
		Name:     "arithmetic-geometric mean, second operand",
		Build:    func(child Real) Real { return newArithmeticGeometricMean(FromInt(1), child) },
		HasChild: true,
	},
	{
		Name:     "spouge gamma",
		Build:    func(child Real) Real { return newSpougeGamma(child) },
		HasChild: true,
	},
	{
		Name:     "digamma series",
		Build:    func(child Real) Real { return newPolygammaSeries(child) },
		HasChild: true,
	},
}

func TestApproximateCancelledContext(t *testing.T) {
	for _, test := range nodeTests {
		t.Run(test.Name, func(t *testing.T) {
			node := test.Build(FromInt(3))

			_, err := Approximate(cancelledContext(), node, -10)
			assert.ErrorIs(t, err, context.Canceled)
			assert.Nil(t, node.tracker().latest())
		})
	}
}

func TestNodeChecksContextBeforeChild(t *testing.T) {
	for _, test := range nodeTests {
		if !test.HasChild {
			continue
		}

		t.Run(test.Name, func(t *testing.T) {
			leaf := &failingReal{err: errLeaf}
			node := test.Build(leaf)

			_, err := node.approximate(cancelledContext(), -10)
			assert.ErrorIs(t, err, context.Canceled)
			assert.Zero(t, leaf.calls)
		})
	}
}

func TestLeafErrorSurfacesUnchanged(t *testing.T) {
	for _, test := range nodeTests {
		if !test.HasChild {
			continue
		}

		t.Run(test.Name, func(t *testing.T) {
			leaf := &failingReal{err: errLeaf}
			tree := Negate(Add(ShiftLeft(test.Build(leaf), 1), FromInt(1)))

			_, err := Approximate(testCtx, tree, -10)
			assert.Equal(t, errLeaf, err)
			assert.ErrorIs(t, err, errLeaf)
			assert.Positive(t, leaf.calls)
		})
	}
}

// limitTest requests an approximation of a freshly built tree under a precision
// limit.
type limitTest struct {
	Name      string
	Build     func() Real
	Limit     int
	Precision int
	Expected  error
}

var limitTests = []limitTest{
	{
		Name:      "at the limit",
		Build:     func() Real { return FromInt(1) },
		Limit:     64,
		Precision: -64,
	},
	{
		Name:      "one bit past the limit",
		Build:     func() Real { return FromInt(1) },
		Limit:     64,
		Precision: -65,
		Expected:  ErrPrecisionOverflow,
	},
	{
		Name:      "negative limit has the same meaning",
		Build:     func() Real { return FromInt(1) },
		Limit:     -64,
		Precision: -65,
		Expected:  ErrPrecisionOverflow,
	},
	{
		Name:      "coarse request passes a zero limit",
		Build:     func() Real { return FromInt(1) },
		Limit:     0,
		Precision: 3,
	},
	{
		Name:      "child asked for finer precision than the caller",
		Build:     func() Real { return ShiftLeft(FromInt(1), 100) },
		Limit:     50,
		Precision: -10,
		Expected:  ErrPrecisionOverflow,
	},
	{
		Name:      "deep tree whose guard bits cross the limit",
		Build:     func() Real { return freshPi() },
		Limit:     20,
		Precision: -16,
		Expected:  ErrPrecisionOverflow,
	},
	{
		Name:      "deep tree within the limit",
		Build:     func() Real { return freshPi() },
		Limit:     2000,
		Precision: -16,
	},
}

func TestApproximatePrecisionLimit(t *testing.T) {
	for _, test := range limitTests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := WithPrecisionLimit(testCtx, test.Limit)

			_, err := Approximate(ctx, test.Build(), test.Precision)
			if test.Expected == nil {
				assert.NoError(t, err)
				return
			}

			assert.ErrorIs(t, err, test.Expected)
		})
	}
}

func TestApproximateWithoutPrecisionLimit(t *testing.T) {
	ctx := WithoutPrecisionLimit(testCtx)

	_, err := Approximate(ctx, ShiftLeft(FromInt(1), 100), -10)
	assert.NoError(t, err)
}

type checkOverflowTest struct {
	Name      string
	Limit     int
	Precision int
	Expected  error
}

var checkOverflowTests = []checkOverflowTest{
	{Name: "finer than the limit", Limit: 10, Precision: -11, Expected: ErrPrecisionOverflow},
	{Name: "exactly the limit", Limit: 10, Precision: -10},
	{Name: "coarser than the limit", Limit: 10, Precision: -9},
	{Name: "positive precision", Limit: 10, Precision: 500},
	{Name: "zero limit rejects any fraction bit", Limit: 0, Precision: -1, Expected: ErrPrecisionOverflow},
}

func TestCheckPrecisionOverflow(t *testing.T) {
	for _, test := range checkOverflowTests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := WithPrecisionLimit(testCtx, test.Limit)
			assert.Equal(t, test.Expected, CheckPrecisionOverflow(ctx, test.Precision))
		})
	}

	t.Run("no limit in the context", func(t *testing.T) {
		assert.NoError(t, CheckPrecisionOverflow(testCtx, math.MinInt+1))
	})
}

type failedRetryTest struct {
	Name  string
	Build func(leaf Real) Real
}

var failedRetryTests = []failedRetryTest{
	{
		Name:  "leaf itself",
		Build: func(leaf Real) Real { return leaf },
	},
	{
		Name:  "negation of the leaf",
		Build: func(leaf Real) Real { return Negate(leaf) },
	},
	{
		Name:  "deeper composite",
		Build: func(leaf Real) Real { return Add(ShiftLeft(Negate(leaf), 2), FromInt(1)) },
	},
}

func TestFailedApproximationIsNotCached(t *testing.T) {
	for _, test := range failedRetryTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &flakyReal{failures: 1}
			tree := test.Build(leaf)

			_, err := Approximate(testCtx, tree, -20)
			assert.ErrorIs(t, err, errLeaf)
			assert.Nil(t, leaf.tracker().latest())
			assert.Nil(t, tree.tracker().latest())

			first, err := Approximate(testCtx, tree, -20)
			assert.NoError(t, err)
			assert.NotNil(t, tree.tracker().latest())

			calls := leaf.calls
			second, err := Approximate(testCtx, tree, -20)
			assert.NoError(t, err)
			assert.Equal(t, calls, leaf.calls)
			assert.Zero(t, first.Cmp(second))
		})
	}
}

func TestCancelledApproximationDoesNotPoisonValue(t *testing.T) {
	pi := freshPi()

	_, err := Approximate(cancelledContext(), pi, -40)
	assert.ErrorIs(t, err, context.Canceled)

	text, err := Text(testCtx, pi, 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, "3.1415926536", text)
}

type sentinelTest struct {
	Name      string
	Build     func() Real
	Precision int
	Expected  error
}

var sentinelTests = []sentinelTest{
	{
		Name:      "nil real",
		Build:     func() Real { return nil },
		Precision: -10,
		Expected:  ErrNilReal,
	},
	{
		Name:      "precision beyond the integer width",
		Build:     func() Real { return FromInt(1) },
		Precision: math.MaxInt,
		Expected:  ErrInvalidPrecision,
	},
	{
		Name:      "square root of a negative operand",
		Build:     func() Real { return newPrescaledSqrt(FromInt(-4)) },
		Precision: -10,
		Expected:  ErrNegative,
	},
	{
		Name:      "inverse of zero",
		Build:     func() Real { return Inverse(FromInt(0)) },
		Precision: -10,
		Expected:  ErrUndecided,
	},
	{
		Name:      "operand nil inside a composite",
		Build:     func() Real { return Add(FromInt(1), nil) },
		Precision: -10,
		Expected:  ErrNilReal,
	},
	{
		Name: "approximation callback returning nil inside a composite",
		Build: func() Real {
			return Add(FromInt(1), FromApproximation("", func(context.Context, int) (*big.Int, error) {
				return nil, nil
			}))
		},
		Precision: -10,
		Expected:  ErrBadApproximation,
	},
	{
		Name: "approximation slice of the wrong length inside a composite",
		Build: func() Real {
			outputs := FromApproximationSlice("", 2, func(context.Context, int) ([]*big.Int, error) {
				return []*big.Int{big.NewInt(1)}, nil
			})

			return Add(FromInt(1), outputs[1])
		},
		Precision: -10,
		Expected:  ErrBadApproximation,
	},
	{
		Name: "approximation slice with a nil element inside a composite",
		Build: func() Real {
			outputs := FromApproximationSlice("", 2, func(context.Context, int) ([]*big.Int, error) {
				return []*big.Int{big.NewInt(1), nil}, nil
			})

			return Add(outputs[0], outputs[1])
		},
		Precision: -10,
		Expected:  ErrBadApproximation,
	},
}

func TestApproximateSentinels(t *testing.T) {
	for _, test := range sentinelTests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := Approximate(testCtx, test.Build(), test.Precision)
			assert.ErrorIs(t, err, test.Expected)
		})
	}
}

func TestTextErrors(t *testing.T) {
	t.Run("nil real", func(t *testing.T) {
		_, err := Text(testCtx, nil, 5, 10)
		assert.ErrorIs(t, err, ErrNilReal)
	})

	t.Run("cancelled context", func(t *testing.T) {
		_, err := Text(cancelledContext(), Pi(), 5, 10)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("leaf error", func(t *testing.T) {
		_, err := Text(testCtx, Add(FromInt(1), &failingReal{err: errLeaf}), 5, 10)
		assert.Equal(t, errLeaf, err)
	})
}

type signLoopTest struct {
	Name     string
	Value    Real
	Context  func() context.Context
	Expected int
	Err      error
}

var signLoopTests = []signLoopTest{
	{
		Name:     "positive value",
		Value:    FromInt(3),
		Context:  func() context.Context { return testCtx },
		Expected: 1,
	},
	{
		Name:     "negative irrational value",
		Value:    Negate(Sqrt2()),
		Context:  func() context.Context { return testCtx },
		Expected: -1,
	},
	{
		Name:    "zero stops at the precision-validity bound",
		Value:   FromInt(0),
		Context: func() context.Context { return testCtx },
		Err:     ErrUndecided,
	},
	{
		Name:    "cancelled before the first doubling",
		Value:   FromInt(0),
		Context: cancelledContext,
		Err:     context.Canceled,
	},
}

func TestSignLoop(t *testing.T) {
	for _, test := range signLoopTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := Sign(test.Context(), test.Value)
			assert.ErrorIs(t, err, test.Err)
			assert.Equal(t, test.Expected, got)
		})
	}
}
