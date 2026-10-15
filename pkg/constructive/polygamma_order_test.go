package constructive

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// apery holds ζ(3) rounded half up to the given number of decimal digits (OEIS A002117).
var apery = map[int]string{
	10:  "1.2020569032",
	60:  "1.202056903159594285399738161511449990764986292340498881792272",
	200: "1.20205690315959428539973816151144999076498629234049888179227155534183820578631309018645587360933525814619915779526071941849199599867328321377639683720790016145394178294936006671919157552224249424396156",
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

// recurrenceArgument is an argument x for the recurrence ψ⁽ⁿ⁾(x+1) − ψ⁽ⁿ⁾(x) = (−1)ⁿ·n!/xⁿ⁺¹.
type recurrenceArgument struct {
	name  string
	build func() Real
}

// The arguments reach each branch: x + 1 in the series and x reflected, both reflected, both
// in the series, and x beyond a pole.
var recurrenceArguments = []recurrenceArgument{
	{name: "√2", build: Sqrt2},
	{name: "−√2", build: func() Real { return Negate(Sqrt2()) }},
	{name: "−3 − √2", build: func() Real { return Negate(Add(FromInt(3), Sqrt2())) }},
	{name: "1/3", build: func() Real { return FromRat(1, 3) }},
	{name: "−3/5", build: func() Real { return FromRat(-3, 5) }},
	{name: "−7/2", build: func() Real { return FromRat(-7, 2) }},
	{name: "11/2", build: func() Real { return FromRat(11, 2) }},
}

func polygammaRecurrenceTests() []digammaIdentityTest {
	var tests []digammaIdentityTest

	for _, order := range []int{1, 2, 3, 4} {
		for _, arg := range recurrenceArguments {
			tests = append(tests, digammaIdentityTest{
				name: fmt.Sprintf("ψ⁽%d⁾(x+1) − ψ⁽%d⁾(x) = (−1)^%d·%d!/x^%d at x = %s",
					order, order, order, order, order+1, arg.name),
				build: func() Real {
					x := arg.build()

					return Subtract(Polygamma(order, Add(x, One())), Polygamma(order, x))
				},
				expected: func() Real {
					return Multiply(factorialSigned(order), Inverse(powInt(arg.build(), uint(order+1))))
				},
			})
		}
	}

	return tests
}

// cot returns cot(πx) as the reflection builds it.
func cot(x Real) Real {
	px := Multiply(Pi(), x)

	return Divide(Cosine(px), Sine(px))
}

var polygammaIdentityTests = append([]digammaIdentityTest{
	{
		name:     "ψ⁽¹⁾(1) = π²/6",
		build:    func() Real { return Polygamma(1, One()) },
		expected: func() Real { return Divide(Square(Pi()), FromInt(6)) },
	},
	{
		name:     "ψ⁽¹⁾(1/2) = π²/2",
		build:    func() Real { return Polygamma(1, FromRat(1, 2)) },
		expected: func() Real { return Divide(Square(Pi()), Two()) },
	},
	{
		name:     "ψ⁽¹⁾(7/2) = π²/2 − 4(1 + 1/9 + 1/25)",
		build:    func() Real { return Polygamma(1, FromRat(7, 2)) },
		expected: func() Real { return Subtract(Divide(Square(Pi()), Two()), FromRat(4*(225+25+9), 225)) },
	},
	{
		name:     "ψ⁽¹⁾(−1/2) = π²/2 + 4",
		build:    func() Real { return Polygamma(1, FromRat(-1, 2)) },
		expected: func() Real { return Add(Divide(Square(Pi()), Two()), FromInt(4)) },
	},
	{
		name:     "ψ⁽³⁾(1) = π⁴/15",
		build:    func() Real { return Polygamma(3, One()) },
		expected: func() Real { return Divide(Square(Square(Pi())), FromInt(15)) },
	},
	{
		name:     "ψ⁽³⁾(1/2) = π⁴",
		build:    func() Real { return Polygamma(3, FromRat(1, 2)) },
		expected: func() Real { return Square(Square(Pi())) },
	},
	{
		name:     "ψ⁽³⁾(−1/2) = π⁴ + 96",
		build:    func() Real { return Polygamma(3, FromRat(-1, 2)) },
		expected: func() Real { return Add(Square(Square(Pi())), FromInt(96)) },
	},
	{
		name: "ψ⁽²⁾(−1/2) − ψ⁽²⁾(1/2) = 16",
		build: func() Real {
			return Subtract(Polygamma(2, FromRat(-1, 2)), Polygamma(2, FromRat(1, 2)))
		},
		expected: func() Real { return FromInt(16) },
	},
	{
		// Both arguments reach the series, so the reflection polynomial is checked against a
		// closed form that does not use it: ψ'(x) + ψ'(1−x) = π²/sin²(πx).
		name: "ψ⁽¹⁾(2/5) + ψ⁽¹⁾(3/5) = π²/sin²(2π/5)",
		build: func() Real {
			return Add(Polygamma(1, FromRat(2, 5)), Polygamma(1, FromRat(3, 5)))
		},
		expected: func() Real {
			return Divide(Square(Pi()), Square(Sine(Multiply(Pi(), FromRat(2, 5)))))
		},
	},
	{
		// P₂(t) = 2t + 2t³, written out here and not taken from cotangentPolynomial.
		name: "ψ⁽²⁾(3/5) − ψ⁽²⁾(2/5) = π³·(2t + 2t³)",
		build: func() Real {
			return Subtract(Polygamma(2, FromRat(3, 5)), Polygamma(2, FromRat(2, 5)))
		},
		expected: func() Real {
			t := cot(FromRat(2, 5))
			poly := Add(Multiply(Two(), t), Multiply(Two(), powInt(t, 3)))

			return Multiply(powInt(Pi(), 3), poly)
		},
	},
	{
		// P₃(t) = −2 − 8t² − 6t⁴.
		name: "ψ⁽³⁾(2/5) + ψ⁽³⁾(3/5) = π⁴·(2 + 8t² + 6t⁴)",
		build: func() Real {
			return Add(Polygamma(3, FromRat(2, 5)), Polygamma(3, FromRat(3, 5)))
		},
		expected: func() Real {
			t := cot(FromRat(2, 5))
			poly := Add(FromInt(2), Add(Multiply(FromInt(8), powInt(t, 2)), Multiply(FromInt(6), powInt(t, 4))))

			return Multiply(powInt(Pi(), 4), poly)
		},
	},
	{
		// P₄(t) = 16t + 40t³ + 24t⁵.
		name: "ψ⁽⁴⁾(3/5) − ψ⁽⁴⁾(2/5) = π⁵·(16t + 40t³ + 24t⁵)",
		build: func() Real {
			return Subtract(Polygamma(4, FromRat(3, 5)), Polygamma(4, FromRat(2, 5)))
		},
		expected: func() Real {
			t := cot(FromRat(2, 5))
			poly := Add(Multiply(FromInt(16), t),
				Add(Multiply(FromInt(40), powInt(t, 3)), Multiply(FromInt(24), powInt(t, 5))))

			return Multiply(powInt(Pi(), 5), poly)
		},
	},
	{
		name: "x·ψ⁽¹⁾(x) = 1 inside the float estimate's overflow band",
		build: func() Real {
			x := ShiftLeft(One(), 970)

			return Multiply(x, Polygamma(1, x))
		},
		expected: One,
	},
	{
		name: "x²·ψ⁽²⁾(x) = −1 inside the float estimate's overflow band",
		build: func() Real {
			x := ShiftLeft(One(), 970)

			return Multiply(Square(x), Polygamma(2, x))
		},
		expected: func() Real { return FromInt(-1) },
	},
	{
		name: "x³·ψ⁽³⁾(x) = 2 beyond float64 range",
		build: func() Real {
			x := Pow10(FromInt(400))

			return Multiply(powInt(x, 3), Polygamma(3, x))
		},
		expected: Two,
	},
}, polygammaRecurrenceTests()...)

func TestPolygammaIdentities(t *testing.T) {
	for _, test := range polygammaIdentityTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.build()
			expected := test.expected()

			assertEqualAtPrecision(t, expected, result, -60)
			assertEqualAtPrecision(t, expected, result, -120)
		})
	}
}

// TestPolygammaIdentitiesAtFinePrecision checks the same identities with a fresh tree per
// precision, so a tree built for a coarse request does not hide a wrong choice of shift or term
// count at a fine one.
func TestPolygammaIdentitiesAtFinePrecision(t *testing.T) {
	for _, test := range polygammaIdentityTests {
		t.Run(test.name, func(t *testing.T) {
			for _, p := range []int{-20, -300, -600} {
				assertEqualAtPrecision(t, test.expected(), test.build(), p)
			}
		})
	}
}

// TestPolygammaApery compares against ζ(3) from OEIS A002117, using ψ⁽²⁾(1) = −2ζ(3),
// ψ⁽²⁾(1/2) = −14ζ(3), and ψ⁽²⁾(−1/2) = 16 − 14ζ(3).
func TestPolygammaApery(t *testing.T) {
	tests := []digammaTextTest{
		{
			name:  "ψ⁽²⁾(1) = −2ζ(3)",
			build: func() Real { return Negate(ShiftRight(Polygamma(2, One()), 1)) },
		},
		{
			name:  "ψ⁽²⁾(1/2) = −14ζ(3)",
			build: func() Real { return Divide(Polygamma(2, FromRat(1, 2)), FromInt(-14)) },
		},
		{
			name: "ψ⁽²⁾(−1/2) = 16 − 14ζ(3)",
			build: func() Real {
				return Divide(Subtract(FromInt(16), Polygamma(2, FromRat(-1, 2))), FromInt(14))
			},
		},
		{
			name:  "ψ⁽²⁾(2) = 2 − 2ζ(3)",
			build: func() Real { return Subtract(One(), ShiftRight(Polygamma(2, Two()), 1)) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := test.build()
			for _, digits := range []int{10, 60, 200} {
				assert.Equal(t, apery[digits], mustText(t, node, digits, 10))
			}
		})
	}
}

// TestPolygammaAperyAtFinePrecision rebuilds the tree per precision, for the same reason as
// TestPolygammaIdentitiesAtFinePrecision. ψ⁽²⁾(1) = −2ζ(3) at 200 digits.
func TestPolygammaAperyAtFinePrecision(t *testing.T) {
	for _, digits := range []int{10, 60, 200} {
		node := Negate(ShiftRight(Polygamma(2, One()), 1))
		assert.Equal(t, apery[digits], mustText(t, node, digits, 10))
	}
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

type polygammaValueTest struct {
	name     string
	build    func() Real
	expected map[int]string
}

// TestPolygammaValues checks decimal text against closed forms in π. The values of ψ⁽ⁿ⁾(7/2)
// and ψ⁽²⁾(−1/2), which came from an independent Euler–Maclaurin sum for ζ(n+1, x), sit with
// the other lazily reduced functions in reductionValueTests.
func TestPolygammaValues(t *testing.T) {
	tests := []polygammaValueTest{
		{
			name:  "ψ⁽¹⁾(−1/2)",
			build: func() Real { return Polygamma(1, FromRat(-1, 2)) },
			expected: map[int]string{
				10: "8.9348022005",
				60: "8.934802200544679309417245499938075567656849703620395313206675",
			},
		},
		{
			name:  "ψ⁽³⁾(−1/2)",
			build: func() Real { return Polygamma(3, FromRat(-1, 2)) },
			expected: map[int]string{
				10: "193.4090910340",
				60: "193.409091034002437236440332688705111249727585672685421691467859",
			},
		},
		{
			name:  "ψ⁽³⁾(1)",
			build: func() Real { return Polygamma(3, One()) },
			expected: map[int]string{
				10: "6.4939394023",
				60: "6.493939402266829149096022179247007416648505711512361446097857",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			node := test.build()
			for _, digits := range []int{10, 60} {
				assert.Equal(t, test.expected[digits], mustText(t, node, digits, 10))
			}
		})
	}
}

// asymptoticPolygamma3 returns the first four terms of the expansion of ψ⁽³⁾(x) for an integer
// x: 2/x³ + 3/x⁴ + 2/x⁵ − 1/x⁷, which is within 2/x⁹ of the function.
func asymptoticPolygamma3(x int) Real {
	inv := Inverse(FromInt(x))
	terms := []Real{
		Multiply(Two(), powInt(inv, 3)),
		Multiply(FromInt(3), powInt(inv, 4)),
		Multiply(Two(), powInt(inv, 5)),
		Negate(powInt(inv, 7)),
	}

	return sumBalanced(terms)
}

// TestPolygammaLargeArgument compares with the leading asymptotic terms, which are good to
// 2^−170 here and no finer.
func TestPolygammaLargeArgument(t *testing.T) {
	for _, p := range []int{-20, -60, -120, -160} {
		assertEqualAtPrecision(t, asymptoticPolygamma3(1000000), Polygamma(3, FromInt(1000000)), p)
	}
}

func TestPolygammaCoarsePrecision(t *testing.T) {
	tests := []digammaCoarseTest{
		{name: "ψ⁽¹⁾(50) at 2^3", input: FromInt(50), precision: 3},
		{name: "ψ⁽¹⁾(1) at 2^0", input: One(), precision: 0},
		{name: "ψ⁽¹⁾(−1/2) at 2^2", input: FromRat(-1, 2), precision: 2},
	}

	for _, order := range []int{1, 3, 8} {
		for _, test := range tests {
			t.Run(fmt.Sprintf("order %d, %s", order, test.name), func(t *testing.T) {
				got := mustApproximate(t, Polygamma(order, test.input), test.precision)
				fine := mustApproximate(t, Polygamma(order, test.input), test.precision-20)

				// Each result is within one unit of its own precision.
				diff := new(big.Int).Sub(got, scale(fine, -20))
				assert.LessOrEqual(t, diff.CmpAbs(big.NewInt(2)), 0)
			})
		}
	}
}

func TestPolygammaAsConstruction(t *testing.T) {
	assert.Equal(t, "Polygamma(3, Int(2))", Polygamma(3, FromInt(2)).asConstruction())
	assert.Equal(t, "Polygamma(3, Int(2))", newPolygammaSeries(3, FromInt(2)).asConstruction())
	assert.Equal(t, "Polygamma(-1, Int(2))", Polygamma(-1, FromInt(2)).asConstruction())
	assert.Equal(t, "Digamma(Int(2))", Polygamma(0, FromInt(2)).asConstruction())

	reflected := Polygamma(2, FromRat(-1, 2))
	assert.Equal(t, "Polygamma(2, Multiply(Int(-1), Inverse(Int(2))))", reflected.asConstruction())

	mustApproximate(t, reflected, -10)
	assert.Contains(t, reflected.asConstruction(), `Polygamma(2, Add(Named("1", Int(1)), `)
}

func TestPolygammaHonorsContext(t *testing.T) {
	for _, arg := range []Real{FromInt(5), FromRat(-1, 2)} {
		_, err := Approximate(cancelledContext(), Polygamma(3, arg), -2000)
		assert.ErrorIs(t, err, context.Canceled)
	}
}

// TestPolygammaAtThousandDigits exercises the shift and the Bernoulli cache at the scale where
// they matter: ψ⁽¹⁾(100) = ψ⁽¹⁾(1) − Σ_{k<100} 1/k² = π²/6 − Σ_{k<100} 1/k² at about 1000
// digits.
func TestPolygammaAtThousandDigits(t *testing.T) {
	terms := make([]Real, 99)
	for i := range terms {
		terms[i] = FromRat(1, (i+1)*(i+1))
	}

	expected := Subtract(Divide(Square(Pi()), FromInt(6)), sumBalanced(terms))

	assertEqualAtPrecision(t, expected, Polygamma(1, FromInt(100)), -3400)
}

// TestPolygammaHighOrderAtHighPrecision checks ψ⁽³⁾(20) = π⁴/15 − 6·Σ_{k<20} 1/k⁴.
func TestPolygammaHighOrderAtHighPrecision(t *testing.T) {
	terms := make([]Real, 19)
	for i := range terms {
		terms[i] = Multiply(FromInt(6), Inverse(powInt(FromInt(i+1), 4)))
	}

	expected := Subtract(Divide(Square(Square(Pi())), FromInt(15)), sumBalanced(terms))

	assertEqualAtPrecision(t, expected, Polygamma(3, FromInt(20)), -1500)
}

type polygammaBoundaryTest struct {
	name   string
	order  int
	num    int
	denom  int
	series bool
}

func polygammaBoundaryTests() []polygammaBoundaryTest {
	var tests []polygammaBoundaryTest

	for _, order := range []int{1, 2, 3} {
		for _, boundary := range digammaBoundaryTests {
			tests = append(tests, polygammaBoundaryTest{
				name:   fmt.Sprintf("order %d, %s", order, boundary.name),
				order:  order,
				num:    boundary.num,
				denom:  boundary.denom,
				series: boundary.series,
			})
		}
	}

	return tests
}

// TestPolygammaBoundaryBranches checks both sides of the reflection threshold at orders that
// differ in parity, using ψ⁽ⁿ⁾(x+1) − ψ⁽ⁿ⁾(x) = (−1)ⁿ·n!/xⁿ⁺¹, whose two polygammas take
// different branches when one side is reflected.
func TestPolygammaBoundaryBranches(t *testing.T) {
	for _, test := range polygammaBoundaryTests() {
		t.Run(test.name, func(t *testing.T) {
			node := Polygamma(test.order, FromRat(test.num, test.denom))
			mustApproximate(t, node, -10)

			series, ok := node.(*deferredReduction).reduced.Load().root.(*polygammaSeries)
			assert.Equal(t, test.series, ok)

			if ok {
				assert.Equal(t, test.order, series.order)
			}

			x := FromRat(test.num, test.denom)
			lhs := Subtract(Polygamma(test.order, Add(x, One())), Polygamma(test.order, x))
			rhs := Multiply(factorialSigned(test.order), Inverse(powInt(x, uint(test.order+1))))
			for _, p := range []int{-60, -400} {
				assertEqualAtPrecision(t, rhs, lhs, p)
			}
		})
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

// TestPolygammaHighOrderRecurrenceAtHighPrecision checks ψ⁽³⁰⁾(x+1) − ψ⁽³⁰⁾(x) = 30!/x³¹ at
// about 450 digits. The arguments reach the series on both sides, the reflection on both sides,
// and the reflection beside the series, so the reflection polynomial P₃₀ is checked against
// the series with exact integer coefficients of about 130 bits.
func TestPolygammaHighOrderRecurrenceAtHighPrecision(t *testing.T) {
	const order = 30

	args := []recurrenceArgument{
		{name: "√2", build: Sqrt2},
		{name: "−3/5", build: func() Real { return FromRat(-3, 5) }},
		{name: "−√2", build: func() Real { return Negate(Sqrt2()) }},
	}

	for _, arg := range args {
		t.Run(arg.name, func(t *testing.T) {
			x := arg.build()
			lhs := Subtract(Polygamma(order, Add(x, One())), Polygamma(order, x))
			rhs := Multiply(factorialSigned(order), Inverse(powInt(arg.build(), order+1)))

			assertEqualAtPrecision(t, rhs, lhs, -1500)
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
