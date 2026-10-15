package constructive

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

// eulerMascheroni holds γ rounded half up to the given number of decimal digits
// (OEIS A001620).
var eulerMascheroni = map[int]string{
	10:  "0.5772156649",
	60:  "0.577215664901532860606512090082402431042159335939923598805767",
	200: "0.57721566490153286060651209008240243104215933593992359880576723488486772677766467093694706329174674951463144724980708248096050401448654283622417399764492353625350033374293733773767394279259525824709492",
}

type digammaIdentityTest struct {
	name     string
	build    func() Real
	expected func() Real
}

// harmonic returns 1/1 + 1/2 + … + 1/n.
func harmonic(n int) Real {
	terms := make([]Real, n)
	for i := range terms {
		terms[i] = FromRat(1, i+1)
	}

	return sumBalanced(terms)
}

var digammaIdentityTests = []digammaIdentityTest{
	{
		name:     "ψ(1/2) − ψ(1) = −2 ln 2",
		build:    func() Real { return Subtract(Digamma(FromRat(1, 2)), Digamma(One())) },
		expected: func() Real { return Negate(Multiply(Two(), Ln2())) },
	},
	{
		name:     "ψ(1/4) − ψ(1) = −π/2 − 3 ln 2",
		build:    func() Real { return Subtract(Digamma(FromRat(1, 4)), Digamma(One())) },
		expected: func() Real { return Negate(Add(Divide(Pi(), Two()), Multiply(FromInt(3), Ln2()))) },
	},
	{
		name:     "ψ(3/4) − ψ(1/4) = π",
		build:    func() Real { return Subtract(Digamma(FromRat(3, 4)), Digamma(FromRat(1, 4))) },
		expected: Pi,
	},
	{
		name:     "ψ(3/2) − ψ(1/2) = 2",
		build:    func() Real { return Subtract(Digamma(FromRat(3, 2)), Digamma(FromRat(1, 2))) },
		expected: func() Real { return FromInt(2) },
	},
	{
		name:     "ψ(−1/2) − ψ(1/2) = 2",
		build:    func() Real { return Subtract(Digamma(FromRat(-1, 2)), Digamma(FromRat(1, 2))) },
		expected: func() Real { return FromInt(2) },
	},
	{
		name:     "ψ(−5/2) − ψ(1/2) = 46/15",
		build:    func() Real { return Subtract(Digamma(FromRat(-5, 2)), Digamma(FromRat(1, 2))) },
		expected: func() Real { return FromRat(46, 15) },
	},
	{
		name:     "ψ(−1/10) − ψ(9/10) = 10",
		build:    func() Real { return Subtract(Digamma(FromRat(-1, 10)), Digamma(FromRat(9, 10))) },
		expected: func() Real { return FromInt(10) },
	},
	{
		name:     "ψ(100) − ψ(1) = H₉₉",
		build:    func() Real { return Subtract(Digamma(FromInt(100)), Digamma(One())) },
		expected: func() Real { return harmonic(99) },
	},
	{
		name:     "ψ(√2 + 1) − ψ(√2) = 1/√2",
		build:    func() Real { return Subtract(Digamma(Add(Sqrt2(), One())), Digamma(Sqrt2())) },
		expected: func() Real { return Inverse(Sqrt2()) },
	},
	{
		name:     "ψ(1 − √2) − ψ(−√2) = −1/√2",
		build:    func() Real { return Subtract(Digamma(Subtract(One(), Sqrt2())), Digamma(Negate(Sqrt2()))) },
		expected: func() Real { return Negate(Inverse(Sqrt2())) },
	},
	{
		name:     "ψ(1 − x) − ψ(x) = π·cot(πx) at x = 1/3",
		build:    func() Real { return Subtract(Digamma(FromRat(2, 3)), Digamma(FromRat(1, 3))) },
		expected: func() Real { return Divide(Pi(), Sqrt(FromInt(3))) },
	},
	{
		name:     "an argument inside the float estimate's overflow band",
		build:    func() Real { return Subtract(Digamma(ShiftLeft(One(), 970)), Multiply(FromInt(970), Ln2())) },
		expected: func() Real { return FromInt(0) },
	},
	{
		name:     "an argument beyond float64 range",
		build:    func() Real { return Subtract(Digamma(Pow10(FromInt(400))), Ln(Pow10(FromInt(400)))) },
		expected: func() Real { return FromInt(0) },
	},
}

// asymptoticDigamma returns ln x − 1/(2x) − 1/(12x²) for an integer x, which is
// within 1/(120x⁴) of ψ(x).
func asymptoticDigamma(x int) Real {
	return Subtract(
		Subtract(Ln(FromInt(x)), FromRat(1, 2*x)),
		Divide(One(), Multiply(FromInt(12), Square(FromInt(x)))),
	)
}

// TestDigammaLargeArgument compares with the leading asymptotic terms, which are
// good to 2^−80 here and no finer.
func TestDigammaLargeArgument(t *testing.T) {
	for _, p := range []int{-20, -60, -80} {
		assertEqualAtPrecision(t, asymptoticDigamma(1000000), Digamma(FromInt(1000000)), p)
	}
}

func TestDigammaIdentities(t *testing.T) {
	for _, test := range digammaIdentityTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.build()
			expected := test.expected()

			assertEqualAtPrecision(t, expected, result, -60)
			assertEqualAtPrecision(t, expected, result, -120)
		})
	}
}

// TestDigammaIdentitiesAtFinePrecision checks the same identities with a fresh
// tree per precision, so a tree built for a coarse request does not hide a wrong
// choice of shift at a fine one.
func TestDigammaIdentitiesAtFinePrecision(t *testing.T) {
	for _, test := range digammaIdentityTests {
		t.Run(test.name, func(t *testing.T) {
			for _, p := range []int{-20, -300, -600} {
				assertEqualAtPrecision(t, test.expected(), test.build(), p)
			}
		})
	}
}

type digammaTextTest struct {
	name  string
	build func() Real
}

// TestDigammaEulerMascheroni compares against γ from OEIS A001620:
// ψ(1) = −γ and ψ(1/2) = −γ − 2 ln 2.
func TestDigammaEulerMascheroni(t *testing.T) {
	tests := []digammaTextTest{
		{
			name:  "ψ(1) = −γ",
			build: func() Real { return Negate(Digamma(One())) },
		},
		{
			name:  "ψ(1/2) = −γ − 2 ln 2",
			build: func() Real { return Negate(Add(Digamma(FromRat(1, 2)), Multiply(Two(), Ln2()))) },
		},
		{
			name:  "ψ(2) = 1 − γ",
			build: func() Real { return Subtract(One(), Digamma(Two())) },
		},
		{
			name:  "ψ(−1/2) = 2 − γ − 2 ln 2",
			build: func() Real { return Negate(Subtract(Add(Digamma(FromRat(-1, 2)), Multiply(Two(), Ln2())), Two())) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := test.build()
			for _, digits := range []int{10, 60, 200} {
				assert.Equal(t, eulerMascheroni[digits], mustText(t, node, digits, 10))
			}
		})
	}
}

func TestDigammaValues(t *testing.T) {
	assert.Equal(t, "-0.5772156649", mustText(t, Digamma(One()), 10, 10))
	assert.Equal(t, "-0.5772156649", mustText(t, Digamma(FromRat(2, 2)), 10, 10))
}

type digammaCoarseTest struct {
	name      string
	input     Real
	precision int
}

// TestDigammaCoarsePrecision checks that a request coarser than one unit still
// returns an answer, without a shift or any series term.
func TestDigammaCoarsePrecision(t *testing.T) {
	tests := []digammaCoarseTest{
		{name: "ψ(50) at 2^3", input: FromInt(50), precision: 3},
		{name: "ψ(1) at 2^0", input: One(), precision: 0},
		{name: "ψ(−1/2) at 2^2", input: FromRat(-1, 2), precision: 2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := mustApproximate(t, Digamma(test.input), test.precision)
			fine := mustApproximate(t, Digamma(test.input), test.precision-20)

			// Each result is within one unit of its own precision.
			diff := new(big.Int).Sub(got, scale(fine, -20))
			assert.LessOrEqual(t, diff.CmpAbs(big.NewInt(2)), 0)
		})
	}
}

func TestDigammaAsConstruction(t *testing.T) {
	assert.Equal(t, "Digamma(Int(3))", Digamma(FromInt(3)).asConstruction())
	assert.Equal(t, "Digamma(Int(3))", newPolygammaSeries(FromInt(3)).asConstruction())
}

func TestDigammaHonorsContext(t *testing.T) {
	_, err := Approximate(cancelledContext(), Digamma(FromInt(5)), -2000)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = Approximate(cancelledContext(), Digamma(FromRat(-1, 2)), -2000)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestDigammaAtThousandDigits exercises the shift and the Bernoulli cache at the scale where
// they matter: ψ(100) − ψ(1) = H₉₉ at about 1000 digits.
func TestDigammaAtThousandDigits(t *testing.T) {
	assertEqualAtPrecision(t, harmonic(99), Subtract(Digamma(FromInt(100)), Digamma(One())), -3400)
}

type digammaBoundaryTest struct {
	name   string
	num    int
	denom  int
	series bool
}

// The first approximation of an argument chooses between the series node and the reflection.
// The rough approximation rounds, so the choice flips between 3/10 and 2/5 and not at 1/2.
var digammaBoundaryTests = []digammaBoundaryTest{
	{name: "1/4 reflects to 3/4", num: 1, denom: 4},
	{name: "3/10 reflects to 7/10", num: 3, denom: 10},
	{name: "2/5 reaches the series below 1/2", num: 2, denom: 5, series: true},
	{name: "49/100 reaches the series below 1/2", num: 49, denom: 100, series: true},
	{name: "1/2 reaches the series", num: 1, denom: 2, series: true},
	{name: "51/100 reaches the series", num: 51, denom: 100, series: true},
}

// TestDigammaBoundaryBranches checks both sides of the reflection threshold, using
// ψ(x+1) − ψ(x) = 1/x, whose two digammas take different branches when one side is reflected.
func TestDigammaBoundaryBranches(t *testing.T) {
	for _, test := range digammaBoundaryTests {
		t.Run(test.name, func(t *testing.T) {
			node := Digamma(FromRat(test.num, test.denom))
			mustApproximate(t, node, -10)

			_, series := node.(*deferredReduction).reduced.Load().root.(*polygammaSeries)
			assert.Equal(t, test.series, series)

			lhs := Subtract(Digamma(FromRat(test.num+test.denom, test.denom)), Digamma(FromRat(test.num, test.denom)))
			rhs := FromRat(test.denom, test.num)
			for _, p := range []int{-60, -400} {
				assertEqualAtPrecision(t, rhs, lhs, p)
			}
		})
	}
}

// exactlyBelow reports whether the first neglected term at the integer y, with K terms
// summed, is below 2^tolExp for tolExp < 0. It uses exact integer arithmetic.
func exactlyBelow(t *testing.T, y *big.Int, terms, tolExp int) bool {
	t.Helper()

	b := mustBernoulliEven(t, bernoulliCache, terms+1)
	k2 := 2*terms + 2

	// |B| / (k2·y^k2) < 2^tolExp  ⇔  |num|·2^(−tolExp) < denom·k2·y^k2.
	left := new(big.Int).Lsh(new(big.Int).Abs(b.Num()), uint(-tolExp))
	right := new(big.Int).Exp(y, big.NewInt(int64(k2)), nil)
	right.Mul(right, big.NewInt(int64(k2)))
	right.Mul(right, b.Denom())

	return left.Cmp(right) < 0
}

type polygammaPlanTest struct {
	name   string
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

			m, terms, err := polygammaPlan(testCtx, est, test.tolExp)
			assert.NoError(t, err)
			assert.GreaterOrEqual(t, m, 0)

			y := new(big.Int).Lsh(big.NewInt(1), test.shift)
			y.Add(y, big.NewInt(int64(m)))

			assert.True(t, exactlyBelow(t, y, terms, test.tolExp), "K = %d", terms)
		})
	}
}

func TestPolygammaPlanTakesInfiniteEstimateAsALowerBound(t *testing.T) {
	_, terms, err := polygammaPlan(testCtx, math.Inf(1), -100008)
	assert.NoError(t, err)

	// Any argument in the band is at least 2^964 > 2^960.
	assert.True(t, exactlyBelow(t, new(big.Int).Lsh(big.NewInt(1), 964), terms, -100008))
}

func TestPolygammaPlanHonorsContext(t *testing.T) {
	_, _, err := polygammaPlan(cancelledContext(), 5, -100)
	assert.ErrorIs(t, err, context.Canceled)
}
