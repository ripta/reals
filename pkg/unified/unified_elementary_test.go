package unified

import (
	"errors"
	"fmt"
	"math"
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
	"github.com/stretchr/testify/assert"
)

// rat builds a rational Real with numerator a and denominator b.
func rat(a, b int64) *Real {
	return New(constructive.One(), rational.New64(a, b))
}

type unaryTest struct {
	name     string
	input    *Real
	expected *Real
}

var sinTests = []unaryTest{
	{name: "Sin(0)=0", input: rat(0, 1), expected: Zero()},
	{name: "Sin(Pi/2)=1", input: New(constructive.Pi(), rational.New64(1, 2)), expected: One()},
	{name: "Sin(Pi)=0", input: Pi(), expected: Zero()},
}

func TestSin(t *testing.T) {
	for _, test := range sinTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Sin(), -100)
		})
	}
}

var cosTests = []unaryTest{
	{name: "Cos(0)=1", input: rat(0, 1), expected: One()},
	{name: "Cos(Pi/2)=0", input: New(constructive.Pi(), rational.New64(1, 2)), expected: Zero()},
	{name: "Cos(Pi)=-1", input: Pi(), expected: NegativeOne()},
}

func TestCos(t *testing.T) {
	for _, test := range cosTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Cos(), -100)
		})
	}
}

var tanTests = []unaryTest{
	{name: "Tan(0)=0", input: rat(0, 1), expected: Zero()},
	{name: "Tan(Pi/4)=1", input: New(constructive.Pi(), rational.New64(1, 4)), expected: One()},
}

func TestTan(t *testing.T) {
	for _, test := range tanTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Tan(), -100)
		})
	}
}

var expTests = []unaryTest{
	{name: "Exp(0)=1", input: rat(0, 1), expected: One()},
	{name: "Exp(1)=e", input: One(), expected: E()},
}

func TestExp(t *testing.T) {
	for _, test := range expTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Exp(), -100)
		})
	}
}

var absTests = []unaryTest{
	{name: "Abs(-3)=3", input: rat(-3, 1), expected: rat(3, 1)},
	{name: "Abs(3)=3", input: rat(3, 1), expected: rat(3, 1)},
	{name: "Abs(0)=0", input: Zero(), expected: Zero()},
}

func TestAbs(t *testing.T) {
	for _, test := range absTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Abs(), -100)
		})
	}
}

type partialUnaryTest struct {
	name     string
	input    *Real
	expected *Real
	wantErr  error
}

var lnTests = []partialUnaryTest{
	{name: "Ln(1)=0", input: rat(1, 1), expected: Zero()},
	{name: "Ln(e)=1", input: E(), expected: One()},
	{name: "Ln(0)", input: Zero(), wantErr: ErrNonPositive},
	{name: "Ln(-1)", input: NegativeOne(), wantErr: ErrNonPositive},
}

func TestLn(t *testing.T) {
	for _, test := range lnTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Ln(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

var sqrtTests = []partialUnaryTest{
	{name: "Sqrt(4)=2", input: rat(4, 1), expected: rat(2, 1)},
	{name: "Sqrt(0)=0", input: Zero(), expected: Zero()},
	{name: "Sqrt(2)=sqrt2", input: rat(2, 1), expected: Sqrt2()},
	{name: "Sqrt(-1)", input: NegativeOne(), wantErr: ErrNegative},
}

func TestSqrt(t *testing.T) {
	for _, test := range sqrtTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Sqrt(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

type powTest struct {
	name     string
	base     *Real
	exponent *Real
	expected *Real
	wantErr  error
}

var powTests = []powTest{
	{name: "Pow(2,3)=8", base: rat(2, 1), exponent: rat(3, 1), expected: rat(8, 1)},
	{name: "Pow(2,1)=2", base: rat(2, 1), exponent: One(), expected: rat(2, 1)},
	{name: "Pow(-2,3)=-8", base: rat(-2, 1), exponent: rat(3, 1), expected: rat(-8, 1)},
	{name: "Pow(-2,2)=4", base: rat(-2, 1), exponent: rat(2, 1), expected: rat(4, 1)},
	{name: "Pow(2,-2)=1/4", base: rat(2, 1), exponent: rat(-2, 1), expected: rat(1, 4)},
	{name: "Pow(-2,-3)=-1/8", base: rat(-2, 1), exponent: rat(-3, 1), expected: rat(-1, 8)},
	{name: "Pow(0,3)=0", base: Zero(), exponent: rat(3, 1), expected: Zero()},
	{name: "Pow(0,0)=1", base: Zero(), exponent: Zero(), expected: One()},
	{name: "Pow(9,1/2)=3", base: rat(9, 1), exponent: New(constructive.One(), rational.New64(1, 2)), expected: rat(3, 1)},
	{name: "Pow(0,-1)", base: Zero(), exponent: NegativeOne(), wantErr: ErrNonPositive},
	{name: "Pow(0,1/2)", base: Zero(), exponent: Half(), wantErr: ErrNonPositive},
	{name: "Pow(-2,1/2)", base: rat(-2, 1), exponent: Half(), wantErr: ErrNonPositive},
}

func TestPow(t *testing.T) {
	for _, test := range powTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.base.Pow(testCtx, test.exponent)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

var log10Tests = []partialUnaryTest{
	{name: "Log10(1)=0", input: rat(1, 1), expected: Zero()},
	{name: "Log10(1000)=3", input: rat(1000, 1), expected: rat(3, 1)},
	{name: "Log10(0)", input: Zero(), wantErr: ErrNonPositive},
	{name: "Log10(-1)", input: NegativeOne(), wantErr: ErrNonPositive},
}

func TestLog10(t *testing.T) {
	for _, test := range log10Tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Log10(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

var log2Tests = []partialUnaryTest{
	{name: "Log2(1)=0", input: rat(1, 1), expected: Zero()},
	{name: "Log2(8)=3", input: rat(8, 1), expected: rat(3, 1)},
	{name: "Log2(0)", input: Zero(), wantErr: ErrNonPositive},
	{name: "Log2(-1)", input: NegativeOne(), wantErr: ErrNonPositive},
}

func TestLog2(t *testing.T) {
	for _, test := range log2Tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Log2(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

type logTest struct {
	name     string
	input    *Real
	base     *Real
	expected *Real
	wantErr  error
}

var logTests = []logTest{
	{name: "Log(81,3)=4", input: rat(81, 1), base: rat(3, 1), expected: rat(4, 1)},
	{name: "Log(1000,10)=3", input: rat(1000, 1), base: rat(10, 1), expected: rat(3, 1)},
	{name: "Log(0,2)", input: Zero(), base: rat(2, 1), wantErr: ErrNonPositive},
	{name: "Log(-1,2)", input: NegativeOne(), base: rat(2, 1), wantErr: ErrNonPositive},
	{name: "Log(8,0)", input: rat(8, 1), base: Zero(), wantErr: ErrNonPositive},
	{name: "Log(8,-2)", input: rat(8, 1), base: rat(-2, 1), wantErr: ErrNonPositive},
	{name: "Log(8,1)", input: rat(8, 1), base: One(), wantErr: ErrInvalidBase},
}

func TestLog(t *testing.T) {
	for _, test := range logTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Log(testCtx, test.base)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

var sinhTests = []unaryTest{
	{name: "Sinh(0)=0", input: Zero(), expected: Zero()},
	{name: "Sinh(1)=(e-1/e)/2", input: One(), expected: E().Subtract(E().Inverse()).ShiftRight(1)},
}

func TestSinh(t *testing.T) {
	for _, test := range sinhTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Sinh(), -100)
		})
	}
}

var coshTests = []unaryTest{
	{name: "Cosh(0)=1", input: Zero(), expected: One()},
	{name: "Cosh(1)=(e+1/e)/2", input: One(), expected: E().Add(E().Inverse()).ShiftRight(1)},
}

func TestCosh(t *testing.T) {
	for _, test := range coshTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Cosh(), -100)
		})
	}
}

var tanhTests = []unaryTest{
	{name: "Tanh(0)=0", input: Zero(), expected: Zero()},
	{name: "Tanh(1)=sinh/cosh", input: One(), expected: One().Sinh().Divide(One().Cosh())},
}

func TestTanh(t *testing.T) {
	for _, test := range tanhTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Tanh(), -100)
		})
	}
}

var cbrtTests = []unaryTest{
	{name: "Cbrt(0)=0", input: Zero(), expected: Zero()},
	{name: "Cbrt(8)=2", input: rat(8, 1), expected: rat(2, 1)},
	{name: "Cbrt(-8)=-2", input: rat(-8, 1), expected: rat(-2, 1)},
	{name: "Cbrt(27)=3", input: rat(27, 1), expected: rat(3, 1)},
}

func TestCbrt(t *testing.T) {
	for _, test := range cbrtTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Cbrt(), -100)
		})
	}
}

var cbrtCubeTests = []unaryTest{
	{name: "positive irrational", input: New(constructive.Sqrt2(), rational.One())},
	{name: "negative irrational", input: New(constructive.Sqrt2(), rational.New64(-1, 1))},
	{name: "negative fraction", input: rat(-5, 7)},
	{name: "small positive", input: rat(1, 1000)},
	{name: "pi multiple", input: New(constructive.Pi(), rational.New64(-3, 2))},
}

func TestCbrtCubesBack(t *testing.T) {
	for _, test := range cbrtCubeTests {
		t.Run(test.name, func(t *testing.T) {
			root := test.input.Cbrt()
			cubed := root.Multiply(root).Multiply(root)
			assertEqualAtPrecision(t, test.input, cubed, -80)
		})
	}
}

func TestCbrtDecidesSignWhenApproximated(t *testing.T) {
	neg := New(constructive.Sqrt2(), rational.New64(-1, 1)).Cbrt()
	pos := New(constructive.Sqrt2(), rational.One()).Cbrt()

	assert.Equal(t, -1, mustPreciseSign(t, neg.Constructive(), -20))
	assert.Equal(t, 1, mustPreciseSign(t, pos.Constructive(), -20))
}

// piFrac builds the unified Real (a/b)·π, used for expected inverse-trig angles.
func piFrac(a, b int64) *Real {
	return New(constructive.Pi(), rational.New64(a, b))
}

// sqrt3 is √3 as a unified Real.
func sqrt3() *Real {
	return New(constructive.Sqrt(constructive.FromInt(3)), rational.One())
}

var atanTests = []unaryTest{
	{name: "Atan(0)=0", input: Zero(), expected: Zero()},
	{name: "Atan(1)=pi/4", input: One(), expected: piFrac(1, 4)},
	{name: "Atan(-1)=-pi/4", input: NegativeOne(), expected: piFrac(-1, 4)},
	{name: "Atan(sqrt3)=pi/3", input: sqrt3(), expected: piFrac(1, 3)},
	{name: "Atan(1/sqrt3)=pi/6", input: New(constructive.Sqrt(constructive.FromInt(3)), rational.New64(1, 3)), expected: piFrac(1, 6)},
}

func TestAtan(t *testing.T) {
	for _, test := range atanTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Atan(), -100)
		})
	}
}

var asinTests = []partialUnaryTest{
	{name: "Asin(0)=0", input: Zero(), expected: Zero()},
	{name: "Asin(1/2)=pi/6", input: rat(1, 2), expected: piFrac(1, 6)},
	{name: "Asin(1)=pi/2", input: One(), expected: piFrac(1, 2)},
	{name: "Asin(-1)=-pi/2", input: NegativeOne(), expected: piFrac(-1, 2)},
	{name: "Asin(2)", input: rat(2, 1), wantErr: ErrOutsideUnitInterval},
	{name: "Asin(-2)", input: rat(-2, 1), wantErr: ErrOutsideUnitInterval},
}

func TestAsin(t *testing.T) {
	for _, test := range asinTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Asin(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

var acosTests = []partialUnaryTest{
	{name: "Acos(1)=0", input: One(), expected: Zero()},
	{name: "Acos(-1)=pi", input: NegativeOne(), expected: Pi()},
	{name: "Acos(0)=pi/2", input: Zero(), expected: piFrac(1, 2)},
	{name: "Acos(1/2)=pi/3", input: rat(1, 2), expected: piFrac(1, 3)},
	{name: "Acos(2)", input: rat(2, 1), wantErr: ErrOutsideUnitInterval},
}

func TestAcos(t *testing.T) {
	for _, test := range acosTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Acos(testCtx)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

type binaryPartialTest struct {
	name     string
	y, x     *Real
	expected *Real
	wantErr  error
}

var atan2Tests = []binaryPartialTest{
	{name: "Atan2(1,1)=pi/4", y: One(), x: One(), expected: piFrac(1, 4)},
	{name: "Atan2(1,-1)=3pi/4", y: One(), x: NegativeOne(), expected: piFrac(3, 4)},
	{name: "Atan2(-1,-1)=-3pi/4", y: NegativeOne(), x: NegativeOne(), expected: piFrac(-3, 4)},
	{name: "Atan2(-1,1)=-pi/4", y: NegativeOne(), x: One(), expected: piFrac(-1, 4)},
	{name: "Atan2(1,0)=pi/2", y: One(), x: Zero(), expected: piFrac(1, 2)},
	{name: "Atan2(-1,0)=-pi/2", y: NegativeOne(), x: Zero(), expected: piFrac(-1, 2)},
	{name: "Atan2(0,1)=0", y: Zero(), x: One(), expected: Zero()},
	{name: "Atan2(0,-1)=pi", y: Zero(), x: NegativeOne(), expected: Pi()},
	{name: "Atan2(0,0)", y: Zero(), x: Zero(), wantErr: ErrUndefinedAtOrigin},
}

func TestAtan2(t *testing.T) {
	for _, test := range atan2Tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.y.Atan2(testCtx, test.x)
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

// sqrtPi is √π, the value of Γ(1/2).
func sqrtPi() *Real {
	r, _ := Pi().Sqrt(testCtx)
	return r
}

var gammaTests = []partialUnaryTest{
	{name: "Gamma(1)=1", input: One(), expected: One()},
	{name: "Gamma(4)=6", input: rat(4, 1), expected: rat(6, 1)},
	{name: "Gamma(1/2)=sqrt(pi)", input: Half(), expected: sqrtPi()},
	{name: "Gamma(-1/2)=-2sqrt(pi)", input: rat(-1, 2), expected: sqrtPi().ShiftLeft(1).Negate()},
	{name: "Gamma(0)", input: Zero(), wantErr: ErrNonPositiveInteger},
	{name: "Gamma(-1)", input: NegativeOne(), wantErr: ErrNonPositiveInteger},
	{name: "Gamma(-2)", input: rat(-2, 1), wantErr: ErrNonPositiveInteger},
}

func TestGamma(t *testing.T) {
	for _, test := range gammaTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Gamma()
			if test.wantErr != nil {
				assert.Nil(t, result)
				assert.ErrorIs(t, err, test.wantErr)
				return
			}
			assert.NoError(t, err)
			assertEqualAtPrecision(t, test.expected, result, -60)
			assertEqualAtPrecision(t, test.expected, result, -120)
		})
	}
}

// TestGammaNearPoleIrrational confirms the structural pole check does not misfire
// on an irrational argument near the pole at zero: √2/100 ≈ 0.0141 is finite and
// Γ there is a large positive value, not an error.
func TestGammaNearPoleIrrational(t *testing.T) {
	x := New(constructive.Sqrt2(), rational.New64(1, 100))
	result, err := x.Gamma()
	assert.NoError(t, err)

	c := result.Constructive()
	if mustPreciseCmp(t, c, One().Constructive(), -60) <= 0 {
		t.Errorf("expected Gamma(√2/100) to exceed 1, got %s", mustText(t, c, 6, 10))
	}
}

// ln2 is the natural logarithm of two.
func ln2() *Real {
	return New(constructive.Ln(constructive.Two()), rational.One())
}

// piSquaredOver returns π²/d.
func piSquaredOver(d int64) *Real {
	return Pi().Multiply(Pi()).Divide(rat(d, 1))
}

// nearPoleIrrational returns k + √2/100, an irrational argument 0.0141 above the integer k.
// It is not structurally an integer, so no pole check applies even when k is non-positive.
func nearPoleIrrational(k int64) *Real {
	return New(constructive.Sqrt2(), rational.New64(1, 100)).Add(rat(k, 1))
}

type polygammaTest struct {
	name     string
	order    int
	input    *Real
	expected *Real

	// minus, when set, is an argument whose same-order result is subtracted from the
	// input's, so a case can check a difference with a known closed form.
	minus *Real
}

var polygammaTests = []polygammaTest{
	{name: "Digamma(1/2)-Digamma(1)=-2ln2", order: 0, input: Half(), minus: One(), expected: ln2().ShiftLeft(1).Negate()},
	{name: "Polygamma(1) at 1=pi^2/6", order: 1, input: One(), expected: piSquaredOver(6)},
	{name: "Polygamma(1) at 1/2=pi^2/2", order: 1, input: Half(), expected: piSquaredOver(2)},
	{name: "Polygamma(1) at -1/2=pi^2/2+4", order: 1, input: rat(-1, 2), expected: piSquaredOver(2).Add(rat(4, 1))},
}

// TestPolygammaValues checks Polygamma against known closed forms. Digamma(1/2) − Digamma(1)
// is −2 ln 2, so the first case compares a difference, which avoids needing the
// Euler–Mascheroni constant.
func TestPolygammaValues(t *testing.T) {
	for _, test := range polygammaTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Polygamma(test.order)
			assert.NoError(t, err)
			if test.minus != nil {
				other, err := test.minus.Polygamma(test.order)
				assert.NoError(t, err)
				result = result.Subtract(other)
			}
			assertEqualAtPrecision(t, test.expected, result, -60)
			assertEqualAtPrecision(t, test.expected, result, -120)
		})
	}
}

// TestPolygammaZeroIsDigamma confirms unified Digamma and unified Polygamma(0) agree.
func TestPolygammaZeroIsDigamma(t *testing.T) {
	for _, x := range []*Real{One(), rat(-7, 3), nearPoleIrrational(-2)} {
		d, err := x.Digamma()
		assert.NoError(t, err)
		p, err := x.Polygamma(0)
		assert.NoError(t, err)
		assertEqualAtPrecision(t, d, p, -80)
	}

	_, err := Zero().Polygamma(0)
	assert.ErrorIs(t, err, ErrNonPositiveInteger)
	assert.ErrorContains(t, err, "Polygamma")
}

var polygammaPoleTests = []struct {
	name  string
	input *Real
}{
	{name: "0", input: Zero()},
	{name: "-1", input: NegativeOne()},
	{name: "-5", input: rat(-5, 1)},
	{name: "-10/2", input: rat(-10, 2)},
}

func TestDigammaPoles(t *testing.T) {
	for _, test := range polygammaPoleTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Digamma()
			assert.Nil(t, result)
			assert.ErrorIs(t, err, ErrNonPositiveInteger)
			assert.ErrorContains(t, err, "Digamma")
		})
	}
}

func TestPolygammaPoles(t *testing.T) {
	for _, test := range polygammaPoleTests {
		for _, n := range []int{1, 3} {
			t.Run(fmt.Sprintf("n=%d at %s", n, test.name), func(t *testing.T) {
				result, err := test.input.Polygamma(n)
				assert.Nil(t, result)
				assert.ErrorIs(t, err, ErrNonPositiveInteger)
				assert.ErrorContains(t, err, "Polygamma")
			})
		}
	}
}

var polygammaNegativeOrderTests = []struct {
	name  string
	order int
	input *Real
}{
	{name: "order -1 at 1", order: -1, input: One()},
	{name: "order -1 at a pole", order: -1, input: NegativeOne()},
	{name: "order -1 at zero", order: -1, input: Zero()},
	{name: "order MinInt at 1/2", order: math.MinInt, input: Half()},
	{name: "order MinInt at a pole", order: math.MinInt, input: rat(-3, 1)},
	{name: "order -2 at an irrational", order: -2, input: Pi()},
}

// TestPolygammaNegativeOrder checks that the order is rejected before the argument is
// examined, so a pole argument reports the order, not the pole.
func TestPolygammaNegativeOrder(t *testing.T) {
	for _, test := range polygammaNegativeOrderTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.input.Polygamma(test.order)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, ErrNegativeOrder)
			assert.ErrorIs(t, err, constructive.ErrNegativeOrder)
			assert.False(t, errors.Is(err, ErrNonPositiveInteger))
		})
	}
}

var polygammaNonRationalTests = []struct {
	name  string
	order int
	input *Real
}{
	{name: "Digamma(sqrt2)", order: 0, input: irr(constructive.Sqrt2())},
	{name: "Digamma(pi/3)", order: 0, input: New(constructive.Pi(), rational.New64(1, 3))},
	{name: "Polygamma(2) at sqrt2", order: 2, input: irr(constructive.Sqrt2())},
	{name: "Digamma(-pi)", order: 0, input: Pi().Negate()},
	{name: "Polygamma(1) at -pi", order: 1, input: Pi().Negate()},
	{name: "Digamma near 0", order: 0, input: nearPoleIrrational(0)},
	{name: "Digamma near -1", order: 0, input: nearPoleIrrational(-1)},
	{name: "Digamma near -4", order: 0, input: nearPoleIrrational(-4)},
	{name: "Polygamma(2) near 0", order: 2, input: nearPoleIrrational(0)},
	{name: "Polygamma(2) near -1", order: 2, input: nearPoleIrrational(-1)},
	{name: "Polygamma(3) near -3", order: 3, input: nearPoleIrrational(-3)},
}

// TestPolygammaNonRationalRecurrence checks ψ⁽ⁿ⁾ on non-rational arguments, including
// irrationals just above a pole, against the recurrence
// ψ⁽ⁿ⁾(x+1) − ψ⁽ⁿ⁾(x) = (−1)ⁿ·n!/xⁿ⁺¹. The two sides take different paths through the
// reflection and the shift, so agreement shows each computes. None of the arguments is
// structurally a pole, so none returns an error.
func TestPolygammaNonRationalRecurrence(t *testing.T) {
	for _, test := range polygammaNonRationalTests {
		t.Run(test.name, func(t *testing.T) {
			x := test.input
			lo, err := x.Polygamma(test.order)
			assert.NoError(t, err)
			hi, err := x.Add(One()).Polygamma(test.order)
			assert.NoError(t, err)

			// (−1)ⁿ·n!/xⁿ⁺¹
			want := rat(1, 1)
			for i := 1; i <= test.order; i++ {
				want = want.Multiply(rat(int64(-i), 1))
			}
			for i := 0; i <= test.order; i++ {
				want = want.Divide(x)
			}
			assertEqualAtPrecision(t, want, hi.Subtract(lo), -60)
		})
	}
}

// TestPolygammaNearPoleIrrational confirms the structural pole check does not misfire on an
// irrational argument next to a pole: ψ(√2/100) is large and negative, since ψ(x) ~ −1/x.
func TestPolygammaNearPoleIrrational(t *testing.T) {
	result, err := nearPoleIrrational(0).Digamma()
	assert.NoError(t, err)

	c := result.Constructive()
	if mustPreciseCmp(t, c, rat(-50, 1).Constructive(), -60) >= 0 {
		t.Errorf("expected Digamma(√2/100) below -50, got %s", mustText(t, c, 6, 10))
	}

	result, err = nearPoleIrrational(0).Polygamma(1)
	assert.NoError(t, err)

	c = result.Constructive()
	if mustPreciseCmp(t, c, rat(1000, 1).Constructive(), -60) <= 0 {
		t.Errorf("expected Polygamma(1) at √2/100 above 1000, got %s", mustText(t, c, 6, 10))
	}
}
