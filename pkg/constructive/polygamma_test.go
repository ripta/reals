package constructive

import (
	"context"
	"fmt"
	"maps"
	"math"
	"math/big"
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// This file holds the machinery shared by the Digamma and Polygamma tests, and the tests of the
// order-generic engine: the plan, the shift, the term count, the coefficients, the cotangent
// polynomials, the order validation, and the huge-order deadline behaviour. The values of
// Digamma are in digamma_test.go, and those of the higher orders in
// polygamma_higher_order_test.go. Both feed their tables to the runners below.

type identityTest struct {
	name     string
	build    func() Real
	expected func() Real
}

// finePrecisions are the precisions at which an identity is checked on a fresh tree each.
var finePrecisions = []int{-20, -300, -600}

// runIdentities checks each identity at 2^−60 and 2^−120 on one tree.
func runIdentities(t *testing.T, tests []identityTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.build()
			expected := test.expected()

			assertEqualAtPrecision(t, expected, result, -60)
			assertEqualAtPrecision(t, expected, result, -120)
		})
	}
}

// runIdentitiesAtPrecisions checks each identity at every given precision, with a fresh tree
// per precision, so a tree built for a coarse request does not hide a wrong choice of shift or
// term count at a fine one.
func runIdentitiesAtPrecisions(t *testing.T, tests []identityTest, precisions ...int) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, p := range precisions {
				assertEqualAtPrecision(t, test.expected(), test.build(), p)
			}
		})
	}
}

// textTest compares a tree's decimal text with the expected text at each listed digit count.
type textTest struct {
	name  string
	build func() Real
	want  map[int]string
}

// runTextTests renders each tree at every digit count of its want, coarse to fine, on one tree.
func runTextTests(t *testing.T, tests []textTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := test.build()
			for _, digits := range slices.Sorted(maps.Keys(test.want)) {
				assert.Equal(t, test.want[digits], mustText(t, node, digits, 10))
			}
		})
	}
}

// runTextTestsFresh is runTextTests with a fresh tree per digit count.
func runTextTestsFresh(t *testing.T, tests []textTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, digits := range slices.Sorted(maps.Keys(test.want)) {
				assert.Equal(t, test.want[digits], mustText(t, test.build(), digits, 10))
			}
		})
	}
}

type coarseTest struct {
	name      string
	input     Real
	precision int
}

// coarsePrecisionTests are requests coarser than one unit, at an integer, at a series argument
// of one, and at a reflected argument.
var coarsePrecisionTests = []coarseTest{
	{name: "x = 50 at 2^3", input: FromInt(50), precision: 3},
	{name: "x = 1 at 2^0", input: One(), precision: 0},
	{name: "x = −1/2 at 2^2", input: FromRat(-1, 2), precision: 2},
}

// runCoarsePrecision checks that a request coarser than one unit still returns an answer,
// without a shift or any series term.
func runCoarsePrecision(t *testing.T, fn func(Real) Real, tests []coarseTest) {
	t.Helper()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mustApproximate(t, fn(test.input), test.precision)
			fine := mustApproximate(t, fn(test.input), test.precision-20)

			// Each result is within one unit of its own precision.
			diff := new(big.Int).Sub(got, scale(fine, -20))
			assert.LessOrEqual(t, diff.CmpAbs(big.NewInt(2)), 0)
		})
	}
}

// boundaryBranchTest is an argument x = num/denom, and whether the first approximation takes
// it to the series node and not to the reflection.
type boundaryBranchTest struct {
	name   string
	num    int
	denom  int
	series bool
}

// The first approximation of an argument chooses between the series node and the reflection.
// The rough approximation rounds, so the choice flips between 3/10 and 2/5 and not at 1/2.
var boundaryBranchTests = []boundaryBranchTest{
	{name: "1/4 reflects to 3/4", num: 1, denom: 4},
	{name: "3/10 reflects to 7/10", num: 3, denom: 10},
	{name: "2/5 reaches the series below 1/2", num: 2, denom: 5, series: true},
	{name: "49/100 reaches the series below 1/2", num: 49, denom: 100, series: true},
	{name: "1/2 reaches the series", num: 1, denom: 2, series: true},
	{name: "51/100 reaches the series", num: 51, denom: 100, series: true},
}

// runBoundaryBranches checks both sides of the reflection threshold for fn, the polygamma
// function of the given order. It uses ψ⁽ⁿ⁾(x+1) − ψ⁽ⁿ⁾(x) = (−1)ⁿ·n!/xⁿ⁺¹, whose two terms
// take different branches when one side is reflected.
func runBoundaryBranches(t *testing.T, order int, fn func(Real) Real) {
	t.Helper()

	for _, test := range boundaryBranchTests {
		t.Run(test.name, func(t *testing.T) {
			node := fn(FromRat(test.num, test.denom))
			mustApproximate(t, node, -10)

			series, ok := node.(*deferredReduction).reduced.Load().root.(*polygammaSeries)
			assert.Equal(t, test.series, ok)

			if ok {
				assert.Equal(t, order, series.order)
			}

			x := FromRat(test.num, test.denom)
			lhs := Subtract(fn(FromRat(test.num+test.denom, test.denom)), fn(x))
			rhs := Multiply(factorialSigned(order), Inverse(powInt(x, uint(order+1))))
			for _, p := range []int{-60, -400} {
				assertEqualAtPrecision(t, rhs, lhs, p)
			}
		})
	}
}

// assertHonorsContext checks that a cancelled context ends the approximation of fn at a series
// argument and at a reflected one.
func assertHonorsContext(t *testing.T, fn func(Real) Real) {
	t.Helper()

	for _, arg := range []Real{FromInt(5), FromRat(-1, 2)} {
		_, err := Approximate(cancelledContext(), fn(arg), -2000)
		assert.ErrorIs(t, err, context.Canceled)
	}
}

// factorialSigned returns (−1)ⁿ·n! as an exact Real.
func factorialSigned(n int) Real {
	f := big.NewInt(1)
	for i := 2; i <= n; i++ {
		f.Mul(f, big.NewInt(int64(i)))
	}

	if n%2 == 1 {
		f.Neg(f)
	}

	return FromBigInt(f)
}

// exactlyBelow reports whether the first neglected term of order n at the integer y, with K
// terms summed, is below 2^tolExp for tolExp < 0. The term is
// |B₂ₖ₊₂|·(2K+n+1)!/((2K+2)!·y^(2K+2+n)). It uses exact integer arithmetic.
func exactlyBelow(t *testing.T, order int, y *big.Int, terms, tolExp int) bool {
	t.Helper()

	b := mustBernoulliEven(t, bernoulliCache, terms+1)
	k2 := 2*terms + 2

	// The factorial ratio (k2+n−1)!/k2! is 1/k2 for n = 0, and (k2+1)⋯(k2+n−1) otherwise.
	ratioNum, ratioDenom := big.NewInt(1), big.NewInt(1)
	if order == 0 {
		ratioDenom.SetInt64(int64(k2))
	}

	for j := 1; j < order; j++ {
		ratioNum.Mul(ratioNum, big.NewInt(int64(k2+j)))
	}

	// |num|·ratioNum / (denom·ratioDenom·y^(k2+n)) < 2^tolExp
	//   ⇔  |num|·ratioNum·2^(−tolExp) < denom·ratioDenom·y^(k2+n).
	left := new(big.Int).Lsh(new(big.Int).Abs(b.Num()), uint(-tolExp))
	left.Mul(left, ratioNum)

	right := new(big.Int).Exp(y, big.NewInt(int64(k2+order)), nil)
	right.Mul(right, ratioDenom)
	right.Mul(right, b.Denom())

	return left.Cmp(right) < 0
}

type polygammaPlanTest struct {
	name   string
	order  int
	shift  uint
	tolExp int
}

var polygammaPlanTests = []polygammaPlanTest{
	{name: "small argument, coarse", shift: 0, tolExp: -100},
	{name: "a mid-size argument", shift: 10, tolExp: -600},
	{name: "just below the float estimate's overflow", shift: 962, tolExp: -100008},
	{name: "inside the overflow band", shift: 970, tolExp: -100008},
	{name: "deep inside the overflow band", shift: 990, tolExp: -100008},
	{name: "inside the overflow band, coarse", shift: 970, tolExp: -100},
	{name: "order 1, small argument", order: 1, shift: 0, tolExp: -100},
	{name: "order 3, small argument, fine", order: 3, shift: 0, tolExp: -600},
	{name: "order 3, a mid-size argument", order: 3, shift: 10, tolExp: -600},
	{name: "order 8, small argument", order: 8, shift: 0, tolExp: -100},
	{name: "order 20, small argument, coarse", order: 20, shift: 0, tolExp: -10},
	{name: "order 1, just below the float estimate's overflow", order: 1, shift: 962, tolExp: -100008},
	{name: "order 3, inside the overflow band", order: 3, shift: 970, tolExp: -100008},
	{name: "order 3, deep inside the overflow band", order: 3, shift: 990, tolExp: -100008},
	{name: "order 5, inside the overflow band, coarse", order: 5, shift: 970, tolExp: -100},
}

// TestPolygammaPlanBoundHoldsAtTheTrueArgument checks the truncation bound at the exact y,
// not at the estimate the plan worked from. The estimate overflows to +Inf from about 2^964,
// and the plan must then fall back to a true lower bound.
func TestPolygammaPlanBoundHoldsAtTheTrueArgument(t *testing.T) {
	for _, test := range polygammaPlanTests {
		t.Run(test.name, func(t *testing.T) {
			arg := ShiftLeft(One(), int(test.shift))

			est, err := floatEstimate(testCtx, arg)
			assert.NoError(t, err)
			assert.Equal(t, test.shift >= 970, math.IsInf(est, 1))

			m, terms, err := polygammaPlan(testCtx, test.order, est, test.tolExp)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, m, 0)

			y := new(big.Int).Lsh(big.NewInt(1), test.shift)
			y.Add(y, big.NewInt(int64(m)))

			assert.True(t, exactlyBelow(t, test.order, y, terms, test.tolExp), "K = %d", terms)
		})
	}
}

func TestPolygammaPlanTakesInfiniteEstimateAsALowerBound(t *testing.T) {
	for _, order := range []int{0, 1, 3} {
		_, terms, err := polygammaPlan(testCtx, order, math.Inf(1), -100008)
		assert.NoError(t, err)

		// Any argument in the band is at least 2^964 > 2^960.
		assert.True(t, exactlyBelow(t, order, new(big.Int).Lsh(big.NewInt(1), 964), terms, -100008))
	}
}

func TestPolygammaPlanHonorsContext(t *testing.T) {
	_, _, err := polygammaPlan(cancelledContext(), 0, 5, -100)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestPolygammaZeroIsDigamma checks that order zero is Digamma, in value and in construction.
func TestPolygammaZeroIsDigamma(t *testing.T) {
	assert.Equal(t, Digamma(FromInt(3)).asConstruction(), Polygamma(0, FromInt(3)).asConstruction())

	for _, arg := range []func() Real{
		func() Real { return FromInt(3) },
		func() Real { return FromRat(-1, 2) },
		Sqrt2,
	} {
		assertEqualAtPrecision(t, Digamma(arg()), Polygamma(0, arg()), -200)
	}
}

type negativeOrderTest struct {
	name  string
	order int
}

var negativeOrderTests = []negativeOrderTest{
	{name: "−1", order: -1},
	{name: "−2", order: -2},
	{name: "−1000", order: -1000},
	{name: "the smallest int", order: math.MinInt},
}

func TestPolygammaNegativeOrder(t *testing.T) {
	for _, test := range negativeOrderTests {
		t.Run(test.name, func(t *testing.T) {
			// A leaf that fails when it is approximated shows that neither construction nor the
			// failing approximation touches the argument.
			leaf := &failingReal{err: errLeaf}

			node := Polygamma(test.order, leaf)
			assert.NotNil(t, node)
			assert.Zero(t, leaf.calls)
			assert.Nil(t, node.tracker().latest())
			assert.Equal(t, fmt.Sprintf("Polygamma(%d, Failing())", test.order), AsConstruction(node))

			for i := 0; i < 2; i++ {
				_, err := Approximate(testCtx, node, -20)
				assert.ErrorIs(t, err, ErrNegativeOrder)
				assert.Nil(t, node.tracker().latest())
			}

			_, err := Text(testCtx, node, 10, 10)
			assert.ErrorIs(t, err, ErrNegativeOrder)
			assert.Zero(t, leaf.calls)
		})
	}
}

func TestPolygammaNegativeOrderOfAValue(t *testing.T) {
	for _, arg := range []Real{FromInt(3), FromRat(-1, 2), Sqrt2(), FromInt(0)} {
		_, err := Approximate(testCtx, Polygamma(-1, arg), -20)
		assert.ErrorIs(t, err, ErrNegativeOrder)
	}

	// The failure also surfaces through a tree.
	_, err := Approximate(testCtx, Add(One(), Polygamma(-3, Two())), -20)
	assert.ErrorIs(t, err, ErrNegativeOrder)
}

func TestPolygammaNegativeOrderHonorsContext(t *testing.T) {
	_, err := Approximate(cancelledContext(), Polygamma(-1, Two()), -20)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, ErrNegativeOrder)
}

type cotangentPolynomialTest struct {
	order  int
	coeffs []int64
}

// Each row holds the coefficients of Pₙ, lowest degree first, derived by hand from
// P₀(t) = t and Pₖ₊₁(t) = −(1 + t²)·Pₖ'(t).
var cotangentPolynomialTests = []cotangentPolynomialTest{
	{order: 0, coeffs: []int64{0, 1}},
	{order: 1, coeffs: []int64{-1, 0, -1}},
	{order: 2, coeffs: []int64{0, 2, 0, 2}},
	{order: 3, coeffs: []int64{-2, 0, -8, 0, -6}},
	{order: 4, coeffs: []int64{0, 16, 0, 40, 0, 24}},
}

func TestCotangentPolynomial(t *testing.T) {
	for _, test := range cotangentPolynomialTests {
		t.Run(fmt.Sprintf("P_%d", test.order), func(t *testing.T) {
			got, err := cotangentPolynomial(testCtx, test.order)
			assert.NoError(t, err)
			assert.Len(t, got, len(test.coeffs))

			for i, want := range test.coeffs {
				assert.Zero(t, got[i].Cmp(big.NewInt(want)), "coefficient of t^%d is %s", i, got[i])
			}
		})
	}
}

// TestCotangentPolynomialShape checks what the recurrence guarantees at any order: degree n+1,
// the parity of n+1, and a nonzero leading coefficient.
func TestCotangentPolynomialShape(t *testing.T) {
	for _, order := range []int{0, 1, 2, 5, 12, 40} {
		p, err := cotangentPolynomial(testCtx, order)
		assert.NoError(t, err)
		assert.Len(t, p, order+2)
		assert.NotZero(t, p[order+1].Sign())

		for i, c := range p {
			if (i+order+1)%2 == 1 {
				assert.Zero(t, c.Sign(), "P_%d has a t^%d term", order, i)
			}
		}
	}
}

// TestCotangentDerivativeValue checks the tree against the polynomial evaluated at a rational.
func TestCotangentDerivativeValue(t *testing.T) {
	for _, order := range []int{0, 1, 2, 3, 6} {
		t.Run(fmt.Sprintf("P_%d", order), func(t *testing.T) {
			x := big.NewRat(3, 7)

			want := new(big.Rat)
			power := big.NewRat(1, 1)
			poly, err := cotangentPolynomial(testCtx, order)
			assert.NoError(t, err)

			for _, c := range poly {
				want.Add(want, new(big.Rat).Mul(new(big.Rat).SetInt(c), power))
				power.Mul(power, x)
			}

			expected := Divide(FromBigInt(want.Num()), FromBigInt(want.Denom()))

			got, err := cotangentDerivative(testCtx, order, FromRat(3, 7))
			assert.NoError(t, err)
			assertEqualAtPrecision(t, expected, got, -100)
		})
	}
}

type polygammaCoefficientTest struct {
	order int
	k     int
}

// TestPolygammaCoefficient checks c_k = B₂ₖ·(2k+n−1)!/(2k)! against an independent exact
// computation with factorials.
func TestPolygammaCoefficient(t *testing.T) {
	var tests []polygammaCoefficientTest
	for _, order := range []int{0, 1, 2, 3, 7} {
		for _, k := range []int{1, 2, 3, 10, 25} {
			tests = append(tests, polygammaCoefficientTest{order: order, k: k})
		}
	}

	factorial := func(n int) *big.Int { return new(big.Int).MulRange(1, int64(n)) }

	for _, test := range tests {
		t.Run(fmt.Sprintf("n = %d, k = %d", test.order, test.k), func(t *testing.T) {
			b := mustBernoulliEven(t, bernoulliCache, test.k)

			// (2k+n−1)! / (2k)!, with 0! = 1 so that n = 0 gives (2k−1)!/(2k)!.
			want := new(big.Rat).SetFrac(factorial(2*test.k+test.order-1), factorial(2*test.k))
			want.Mul(want, b)

			got, err := polygammaCoefficient(testCtx, test.order, test.k)
			assert.NoError(t, err)

			expected := Divide(FromBigInt(want.Num()), FromBigInt(want.Denom()))
			assertEqualAtPrecision(t, expected, got, -200)
		})
	}
}

// TestPolygammaShiftRetriesWhenTheTermsTurnAround drives the retry in the shift search. A
// starting target of one is too small for a high order, so the first round's terms turn around
// before the tolerance, and the search shifts further until they do not.
func TestPolygammaShiftRetriesWhenTheTermsTurnAround(t *testing.T) {
	tests := []struct {
		name   string
		order  int
		tolExp int
	}{
		{name: "order 1", order: 1, tolExp: -300},
		{name: "order 3", order: 3, tolExp: -600},
		{name: "order 20", order: 20, tolExp: -600},
		{name: "order 60", order: 60, tolExp: -200},
	}

	const lower = 1 - 0x1p-20 // a lower bound of the argument 1

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// The first round, with the starting target, must fail for the test to mean anything.
			m0 := 1
			_, ok, err := polygammaTerms(testCtx, test.order, lower+float64(m0), test.tolExp)
			assert.NoError(t, err)
			assert.False(t, ok)

			m, terms, err := polygammaShift(testCtx, test.order, lower, test.tolExp, 1)
			assert.NoError(t, err)
			assert.Greater(t, m, m0)

			y := big.NewInt(int64(1 + m))
			assert.True(t, exactlyBelow(t, test.order, y, terms, test.tolExp), "K = %d", terms)
		})
	}
}

func TestPolygammaShiftBeyondTheIntegerRange(t *testing.T) {
	_, _, err := polygammaShift(testCtx, 0, 0.5, -100, 0x1p70)
	assert.Equal(t, ErrInvalidPrecision, err)
}

// TestPowIntExponentBeyondInt checks that the exponent n+1 of the largest int order does not
// wrap: the power of one is one at any exponent, and the squaring recursion is at most 64 deep.
func TestPowIntExponentBeyondInt(t *testing.T) {
	for _, k := range []uint{1, 2, 3, uint(math.MaxInt), uint(math.MaxInt) + 1, math.MaxUint} {
		assertEqualAtPrecision(t, One(), powInt(One(), k), -20)
	}
}

type polygammaDeadlineTest struct {
	name      string
	order     int
	arg       func() Real
	reflected bool // the reduction itself is cancelled, so nothing is memoized
}

// The orders are far past what finishes in the deadline. The series arguments take the shift,
// the Bernoulli coefficients and the factorials. The reflected ones start with the
// cotangent-derivative polynomial. MaxInt also checks that n+1 and 2k+n cannot wrap.
var polygammaDeadlineTests = []polygammaDeadlineTest{
	{name: "series, n = 20000", order: 20000, arg: func() Real { return FromInt(3) }},
	{name: "reflected, n = 5000", order: 5000, arg: func() Real { return FromRat(-1, 2) }, reflected: true},
	{name: "reflected, n = 50000", order: 50000, arg: func() Real { return FromRat(-1, 2) }, reflected: true},
	{name: "series, n = 10⁹", order: 1000000000, arg: func() Real { return FromRat(7, 2) }},
	{name: "reflected, n = 10⁹", order: 1000000000, arg: func() Real { return FromRat(-1, 2) }, reflected: true},
	{name: "series, n = MaxInt", order: math.MaxInt, arg: func() Real { return FromInt(3) }},
	{name: "reflected, n = MaxInt", order: math.MaxInt, arg: func() Real { return FromRat(-1, 2) }, reflected: true},
	{name: "series, n = MaxInt, argument 2/5", order: math.MaxInt, arg: func() Real { return FromRat(2, 5) }},
}

// TestPolygammaHugeOrderHonorsDeadline checks that a very large order is slow and not stuck:
// the context's error comes back soon after the deadline, from every loop that scales with the
// order, and a later approximation with a live context is not poisoned by the failure.
func TestPolygammaHugeOrderHonorsDeadline(t *testing.T) {
	const deadline = 100 * time.Millisecond

	for _, test := range polygammaDeadlineTests {
		t.Run(test.name, func(t *testing.T) {
			node := Polygamma(test.order, test.arg())

			ctx, cancel := context.WithTimeout(testCtx, deadline)
			defer cancel()

			start := time.Now()
			_, err := Approximate(ctx, node, -20)
			elapsed := time.Since(start)

			assert.ErrorIs(t, err, context.DeadlineExceeded)
			assert.Less(t, elapsed, 5*time.Second)
			assert.Nil(t, node.tracker().latest())

			// A cancelled reduction memoizes nothing. A series order is reduced at once, and only
			// its series is cancelled.
			tree := node.(*deferredReduction)
			assert.Equal(t, !test.reflected, tree.reduced.Load() != nil)
		})
	}
}

func TestPolygammaHugeOrderCancelledContext(t *testing.T) {
	for _, test := range polygammaDeadlineTests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Approximate(cancelledContext(), Polygamma(test.order, test.arg()), -20)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

// TestPolygammaLoopsFollowTheContext calls the helpers that scale with the order directly, so
// a loop that stopped checking the context would show up here by name.
func TestPolygammaLoopsFollowTheContext(t *testing.T) {
	t.Run("cotangentPolynomial", func(t *testing.T) {
		_, err := cotangentPolynomial(cancelledContext(), 30)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("cotangentDerivative", func(t *testing.T) {
		_, err := cotangentDerivative(cancelledContext(), 30, Two())
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("reducePolygamma", func(t *testing.T) {
		_, err := reducePolygamma(cancelledContext(), 30, FromRat(-1, 2), big.NewInt(-2))
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("log2FactorialRatio, running", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(testCtx, 50*time.Millisecond)
		defer cancel()

		_, err := log2FactorialRatio(ctx, math.MaxInt, 1)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("polygammaCoefficient, running", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(testCtx, 50*time.Millisecond)
		defer cancel()

		_, err := polygammaCoefficient(ctx, math.MaxInt, 1)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("expectedPolygammaTerms", func(t *testing.T) {
		_, err := expectedPolygammaTerms(cancelledContext(), 0, 20, -100)
		assert.ErrorIs(t, err, context.Canceled)
	})
}
