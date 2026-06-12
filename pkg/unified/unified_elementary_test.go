package unified

import (
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
			result, err := test.input.Ln()
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
			result, err := test.input.Sqrt()
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
			result, err := test.base.Pow(test.exponent)
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
			result, err := test.input.Log10()
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
			result, err := test.input.Log2()
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
			result, err := test.input.Log(test.base)
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
			result, err := test.input.Asin()
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
			result, err := test.input.Acos()
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
			result, err := test.y.Atan2(test.x)
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
