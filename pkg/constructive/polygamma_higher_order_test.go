package constructive

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These are the value tests of the polygamma functions of order one and above, and the tests
// of Polygamma that need an order, as Digamma's are in digamma_test.go. They feed their tables
// to the runners in polygamma_test.go, over a few orders each.

// apery holds ζ(3) rounded half up to the given number of decimal digits (OEIS A002117).
var apery = map[int]string{
	10:  "1.2020569032",
	60:  "1.202056903159594285399738161511449990764986292340498881792272",
	200: "1.20205690315959428539973816151144999076498629234049888179227155534183820578631309018645587360933525814619915779526071941849199599867328321377639683720790016145394178294936006671919157552224249424396156",
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

func polygammaRecurrenceTests() []identityTest {
	var tests []identityTest

	for _, order := range []int{1, 2, 3, 4} {
		for _, arg := range recurrenceArguments {
			tests = append(tests, identityTest{
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

var polygammaIdentityTests = append([]identityTest{
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
	runIdentities(t, polygammaIdentityTests)
}

// TestPolygammaIdentitiesAtFinePrecision checks the same identities with a fresh tree per
// precision, so a tree built for a coarse request does not hide a wrong choice of shift or term
// count at a fine one.
func TestPolygammaIdentitiesAtFinePrecision(t *testing.T) {
	runIdentitiesAtPrecisions(t, polygammaIdentityTests, finePrecisions...)
}

// TestPolygammaApery compares against ζ(3) from OEIS A002117, using ψ⁽²⁾(1) = −2ζ(3),
// ψ⁽²⁾(1/2) = −14ζ(3), and ψ⁽²⁾(−1/2) = 16 − 14ζ(3).
func TestPolygammaApery(t *testing.T) {
	runTextTests(t, []textTest{
		{
			name:  "ψ⁽²⁾(1) = −2ζ(3)",
			build: func() Real { return Negate(ShiftRight(Polygamma(2, One()), 1)) },
			want:  apery,
		},
		{
			name:  "ψ⁽²⁾(1/2) = −14ζ(3)",
			build: func() Real { return Divide(Polygamma(2, FromRat(1, 2)), FromInt(-14)) },
			want:  apery,
		},
		{
			name: "ψ⁽²⁾(−1/2) = 16 − 14ζ(3)",
			build: func() Real {
				return Divide(Subtract(FromInt(16), Polygamma(2, FromRat(-1, 2))), FromInt(14))
			},
			want: apery,
		},
		{
			name:  "ψ⁽²⁾(2) = 2 − 2ζ(3)",
			build: func() Real { return Subtract(One(), ShiftRight(Polygamma(2, Two()), 1)) },
			want:  apery,
		},
	})
}

// TestPolygammaAperyAtFinePrecision rebuilds the tree per precision, for the same reason as
// TestPolygammaIdentitiesAtFinePrecision. ψ⁽²⁾(1) = −2ζ(3) at 200 digits.
func TestPolygammaAperyAtFinePrecision(t *testing.T) {
	runTextTestsFresh(t, []textTest{{
		name:  "ψ⁽²⁾(1) = −2ζ(3)",
		build: func() Real { return Negate(ShiftRight(Polygamma(2, One()), 1)) },
		want:  apery,
	}})
}

// TestPolygammaValues checks decimal text against closed forms in π. The values of ψ⁽ⁿ⁾(7/2)
// and ψ⁽²⁾(−1/2), which came from an independent Euler–Maclaurin sum for ζ(n+1, x), sit with
// the other lazily reduced functions in reductionValueTests.
func TestPolygammaValues(t *testing.T) {
	runTextTests(t, []textTest{
		{
			name:  "ψ⁽¹⁾(−1/2)",
			build: func() Real { return Polygamma(1, FromRat(-1, 2)) },
			want: map[int]string{
				10: "8.9348022005",
				60: "8.934802200544679309417245499938075567656849703620395313206675",
			},
		},
		{
			name:  "ψ⁽³⁾(−1/2)",
			build: func() Real { return Polygamma(3, FromRat(-1, 2)) },
			want: map[int]string{
				10: "193.4090910340",
				60: "193.409091034002437236440332688705111249727585672685421691467859",
			},
		},
		{
			name:  "ψ⁽³⁾(1)",
			build: func() Real { return Polygamma(3, One()) },
			want: map[int]string{
				10: "6.4939394023",
				60: "6.493939402266829149096022179247007416648505711512361446097857",
			},
		},
	})
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
	runIdentitiesAtPrecisions(t, []identityTest{{
		name:     "ψ⁽³⁾(10⁶) against the leading asymptotic terms",
		build:    func() Real { return Polygamma(3, FromInt(1000000)) },
		expected: func() Real { return asymptoticPolygamma3(1000000) },
	}}, -20, -60, -120, -160)
}

func TestPolygammaCoarsePrecision(t *testing.T) {
	for _, order := range []int{1, 3, 8} {
		t.Run(fmt.Sprintf("order %d", order), func(t *testing.T) {
			runCoarsePrecision(t, func(x Real) Real { return Polygamma(order, x) }, coarsePrecisionTests)
		})
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
	assertHonorsContext(t, func(x Real) Real { return Polygamma(3, x) })
}

// TestPolygammaAtThousandDigits exercises the shift and the Bernoulli cache at the scale where
// they matter: ψ⁽¹⁾(100) = ψ⁽¹⁾(1) − Σ_{k<100} 1/k² = π²/6 − Σ_{k<100} 1/k² at about 1000
// digits.
func TestPolygammaAtThousandDigits(t *testing.T) {
	runIdentitiesAtPrecisions(t, []identityTest{{
		name:  "ψ⁽¹⁾(100) = π²/6 − Σ_{k<100} 1/k²",
		build: func() Real { return Polygamma(1, FromInt(100)) },
		expected: func() Real {
			terms := make([]Real, 99)
			for i := range terms {
				terms[i] = FromRat(1, (i+1)*(i+1))
			}

			return Subtract(Divide(Square(Pi()), FromInt(6)), sumBalanced(terms))
		},
	}}, -3400)
}

// TestPolygammaHighOrderAtHighPrecision checks ψ⁽³⁾(20) = π⁴/15 − 6·Σ_{k<20} 1/k⁴.
func TestPolygammaHighOrderAtHighPrecision(t *testing.T) {
	runIdentitiesAtPrecisions(t, []identityTest{{
		name:  "ψ⁽³⁾(20) = π⁴/15 − 6·Σ_{k<20} 1/k⁴",
		build: func() Real { return Polygamma(3, FromInt(20)) },
		expected: func() Real {
			terms := make([]Real, 19)
			for i := range terms {
				terms[i] = Multiply(FromInt(6), Inverse(powInt(FromInt(i+1), 4)))
			}

			return Subtract(Divide(Square(Square(Pi())), FromInt(15)), sumBalanced(terms))
		},
	}}, -1500)
}

// TestPolygammaBoundaryBranches checks both sides of the reflection threshold at orders that
// differ in parity.
func TestPolygammaBoundaryBranches(t *testing.T) {
	for _, order := range []int{1, 2, 3} {
		t.Run(fmt.Sprintf("order %d", order), func(t *testing.T) {
			runBoundaryBranches(t, order, func(x Real) Real { return Polygamma(order, x) })
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
