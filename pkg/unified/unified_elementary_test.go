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
