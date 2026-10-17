package constructive

import (
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

// harmonic returns 1/1 + 1/2 + … + 1/n.
func harmonic(n int) Real {
	terms := make([]Real, n)
	for i := range terms {
		terms[i] = FromRat(1, i+1)
	}

	return sumBalanced(terms)
}

var digammaIdentityTests = []identityTest{
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
	runIdentitiesAtPrecisions(t, []identityTest{{
		name:     "ψ(10⁶) against the leading asymptotic terms",
		build:    func() Real { return Digamma(FromInt(1000000)) },
		expected: func() Real { return asymptoticDigamma(1000000) },
	}}, -20, -60, -80)
}

func TestDigammaIdentities(t *testing.T) {
	runIdentities(t, digammaIdentityTests)
}

// TestDigammaIdentitiesAtFinePrecision checks the same identities with a fresh
// tree per precision, so a tree built for a coarse request does not hide a wrong
// choice of shift at a fine one.
func TestDigammaIdentitiesAtFinePrecision(t *testing.T) {
	runIdentitiesAtPrecisions(t, digammaIdentityTests, finePrecisions...)
}

// TestDigammaEulerMascheroni compares against γ from OEIS A001620:
// ψ(1) = −γ and ψ(1/2) = −γ − 2 ln 2.
func TestDigammaEulerMascheroni(t *testing.T) {
	runTextTests(t, []textTest{
		{
			name:  "ψ(1) = −γ",
			build: func() Real { return Negate(Digamma(One())) },
			want:  eulerMascheroni,
		},
		{
			name:  "ψ(1/2) = −γ − 2 ln 2",
			build: func() Real { return Negate(Add(Digamma(FromRat(1, 2)), Multiply(Two(), Ln2()))) },
			want:  eulerMascheroni,
		},
		{
			name:  "ψ(2) = 1 − γ",
			build: func() Real { return Subtract(One(), Digamma(Two())) },
			want:  eulerMascheroni,
		},
		{
			name:  "ψ(−1/2) = 2 − γ − 2 ln 2",
			build: func() Real { return Negate(Subtract(Add(Digamma(FromRat(-1, 2)), Multiply(Two(), Ln2())), Two())) },
			want:  eulerMascheroni,
		},
	})
}

func TestDigammaValues(t *testing.T) {
	want := map[int]string{10: "-0.5772156649"}

	runTextTests(t, []textTest{
		{name: "ψ(1)", build: func() Real { return Digamma(One()) }, want: want},
		{name: "ψ(2/2)", build: func() Real { return Digamma(FromRat(2, 2)) }, want: want},
	})
}

func TestDigammaCoarsePrecision(t *testing.T) {
	runCoarsePrecision(t, Digamma, coarsePrecisionTests)
}

func TestDigammaAsConstruction(t *testing.T) {
	assert.Equal(t, "Digamma(Int(3))", Digamma(FromInt(3)).asConstruction())
	assert.Equal(t, "Digamma(Int(3))", newPolygammaSeries(0, FromInt(3)).asConstruction())
}

func TestDigammaHonorsContext(t *testing.T) {
	assertHonorsContext(t, Digamma)
}

// TestDigammaAtThousandDigits exercises the shift and the Bernoulli cache at the scale where
// they matter: ψ(100) − ψ(1) = H₉₉ at about 1000 digits.
func TestDigammaAtThousandDigits(t *testing.T) {
	runIdentitiesAtPrecisions(t, []identityTest{{
		name:     "ψ(100) − ψ(1) = H₉₉",
		build:    func() Real { return Subtract(Digamma(FromInt(100)), Digamma(One())) },
		expected: func() Real { return harmonic(99) },
	}}, -3400)
}

// TestDigammaBoundaryBranches checks both sides of the reflection threshold, using
// ψ(x+1) − ψ(x) = 1/x, whose two digammas take different branches when one side is reflected.
func TestDigammaBoundaryBranches(t *testing.T) {
	runBoundaryBranches(t, 0, Digamma)
}
