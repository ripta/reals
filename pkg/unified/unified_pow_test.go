package unified

import (
	"context"
	"errors"
	"math/big"
	"runtime/debug"
	"testing"
	"time"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
	"github.com/stretchr/testify/assert"
)

// linearPow is the reference power: repeated multiplication, one factor per unit
// of the exponent, inverted once at the end for a negative exponent.
func linearPow(u *Real, e int) *Real {
	result := One()
	for i := 0; i < e || i < -e; i++ {
		result = result.Multiply(u)
	}
	if e < 0 {
		result = result.Inverse()
	}

	return result
}

func intRat(i *big.Int) *rational.Number {
	return rational.New(i, big.NewInt(1))
}

func powInt(t *testing.T, ctx context.Context, base *Real, e int64) *Real {
	t.Helper()

	result, err := base.Pow(ctx, New(constructive.One(), rational.New64(e, 1)))
	if err != nil {
		t.Fatalf("Pow(%d): %v", e, err)
	}

	return result
}

func TestPowMatchesRepeatedMultiplication(t *testing.T) {
	bases := []struct {
		name string
		base *Real
	}{
		{"3/2", rat(3, 2)},
		{"-3/2", rat(-3, 2)},
		{"sqrt 2", irr(constructive.Sqrt2())},
		{"3 sqrt 2", New(constructive.Sqrt2(), rational.New64(3, 1))},
		{"-1/2 pi", New(constructive.Pi(), rational.New64(-1, 2))},
	}
	exponents := []int{0, 1, 2, 3, 5, 8, 13, 21, -1, -2, -5, -8}

	for _, base := range bases {
		for _, e := range exponents {
			t.Run(base.name+"^"+big.NewInt(int64(e)).String(), func(t *testing.T) {
				result := powInt(t, testCtx, base.base, int64(e))
				assertEqualAtPrecision(t, linearPow(base.base, e), result, -80)
			})
		}
	}
}

func TestPowLargeExponentIdentities(t *testing.T) {
	pow := func(b, e int64) *big.Int {
		return new(big.Int).Exp(big.NewInt(b), big.NewInt(e), nil)
	}
	threeSqrt2 := New(constructive.Sqrt2(), rational.New64(3, 1))

	// The comparison is absolute, so the precision of a small expected value must be finer than
	// its magnitude, or a wrong exponent and even zero would compare equal.
	tests := []struct {
		name      string
		base      *Real
		e         int64
		expected  *Real
		precision int
	}{
		{"sqrt 2 ^ 200 = 2^100", irr(constructive.Sqrt2()), 200, New(nil, intRat(pow(2, 100))), -40},
		{"sqrt 2 ^ 201 = 2^100 sqrt 2", irr(constructive.Sqrt2()), 201, New(constructive.Sqrt2(), intRat(pow(2, 100))), -40},
		{"sqrt 2 ^ -200 = 2^-100", irr(constructive.Sqrt2()), -200, New(nil, rational.New(big.NewInt(1), pow(2, 100))), -140},
		{
			"(3 sqrt 2)^120 = 3^120 2^60",
			threeSqrt2, 120,
			New(nil, intRat(new(big.Int).Mul(pow(3, 120), pow(2, 60)))), -40,
		},
		{
			"(3 sqrt 2)^121 = 3^121 2^60 sqrt 2",
			threeSqrt2, 121,
			New(constructive.Sqrt2(), intRat(new(big.Int).Mul(pow(3, 121), pow(2, 60)))), -40,
		},
		{
			"(3 sqrt 2)^-120 = 1 / (3^120 2^60), about 2^-250",
			threeSqrt2, -120,
			New(nil, rational.New(big.NewInt(1), new(big.Int).Mul(pow(3, 120), pow(2, 60)))), -300,
		},
		{"(-3/2)^101", rat(-3, 2), 101, New(nil, rational.New(new(big.Int).Neg(pow(3, 101)), pow(2, 101))), -40},
		{"(355/113)^-50, about 2^-83", rat(355, 113), -50, New(nil, rational.New(pow(113, 50), pow(355, 50))), -130},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, powInt(t, testCtx, test.base, test.e), test.precision)
		})
	}
}

// TestPowIdentityComparisonDiscriminates pins that the precisions above tell a
// wrong exponent and zero from the right answer.
func TestPowIdentityComparisonDiscriminates(t *testing.T) {
	expected := New(nil, rational.New(big.NewInt(1), new(big.Int).Lsh(big.NewInt(1), 100)))
	for _, wrong := range []*Real{powInt(t, testCtx, irr(constructive.Sqrt2()), -199), Zero()} {
		assert.NotEqual(t, 0, mustPreciseCmp(t, expected.Constructive(), wrong.Constructive(), -140))
	}
}

func TestPowRationalBaseStaysRational(t *testing.T) {
	bases := []*Real{rat(2, 3), rat(-5, 7), rat(4, 1), One(), NegativeOne(), rat(0, 1)}
	for _, base := range bases {
		for _, e := range []int64{0, 1, 2, 5, 100, -1, -3, -100} {
			if base.IsZero() && e < 0 {
				continue
			}
			result := powInt(t, testCtx, base, e)
			assert.True(t, result.cr == constructive.One(), "%s^%d lost its rationality", base.rr, e)

			// The rational part is exact, so it compares equal, not just close.
			expected := linearPow(base, int(e))
			assert.Equal(t, 0, expected.rr.Cmp(result.rr), "%s^%d: expected %s, got %s", base.rr, e, expected.rr, result.rr)
		}
	}
}

func TestPowIrrationalBaseKeepsConstructiveFactor(t *testing.T) {
	result := powInt(t, testCtx, New(constructive.Sqrt2(), rational.New64(3, 1)), 4)
	assert.False(t, result.cr == constructive.One())
	assert.Equal(t, 0, rational.New64(81, 1).Cmp(result.rr))
}

func TestPowLargeRationalExponentIsExact(t *testing.T) {
	result := powInt(t, testCtx, rat(355, 113), 1000)
	assert.True(t, result.cr == constructive.One())

	want := new(big.Rat).SetFrac(
		new(big.Int).Exp(big.NewInt(355), big.NewInt(1000), nil),
		new(big.Int).Exp(big.NewInt(113), big.NewInt(1000), nil),
	)
	assert.Equal(t, 0, rational.FromRational(want).Cmp(result.rr))
}

func TestPowWithoutDeepRecursion(t *testing.T) {
	// A linear chain of 100000 multiplications needs far more than this much
	// stack to approximate, so a regression ends in a stack overflow.
	defer debug.SetMaxStack(debug.SetMaxStack(8 << 20))

	ctx, cancel := context.WithTimeout(testCtx, time.Minute)
	defer cancel()

	result := powInt(t, ctx, irr(constructive.Sqrt2()), 100000)
	expected := New(nil, intRat(new(big.Int).Lsh(big.NewInt(1), 50000)))
	assertEqualAtPrecision(t, expected, result, -10)
}

func TestPowHugeExponentUnderPrecisionLimit(t *testing.T) {
	trillion := new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)
	exponent := New(nil, intRat(trillion))
	ctx := constructive.WithPrecisionLimit(testCtx, 1<<20)

	for _, base := range []*Real{Two(), rat(-3, 4), New(constructive.Sqrt2(), rational.New64(3, 1))} {
		result, err := base.Pow(ctx, exponent)
		assert.Nil(t, result)
		assert.ErrorIs(t, err, constructive.ErrPrecisionOverflow)
		assert.ErrorContains(t, err, "Pow:")

		result, err = base.Pow(ctx, exponent.Negate())
		assert.Nil(t, result)
		assert.ErrorIs(t, err, constructive.ErrPrecisionOverflow)
	}
}

func TestPowPrecisionLimitBoundary(t *testing.T) {
	// The estimate of 2^10 is 10 * 2 bits, because 2 has a bit length of 2.
	ten := rat(10, 1)

	result, err := Two().Pow(constructive.WithPrecisionLimit(testCtx, 20), ten)
	assert.NoError(t, err)
	assertEqualAtPrecision(t, rat(1024, 1), result, -10)

	result, err = Two().Pow(constructive.WithPrecisionLimit(testCtx, 19), ten)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, constructive.ErrPrecisionOverflow)

	// A mixed base is bounded by its rational factor, which is 3 with 2 bits.
	mixed := New(constructive.Sqrt2(), rational.New64(3, 1))
	_, err = mixed.Pow(constructive.WithPrecisionLimit(testCtx, 20), ten)
	assert.NoError(t, err)
	_, err = mixed.Pow(constructive.WithPrecisionLimit(testCtx, 19), ten)
	assert.ErrorIs(t, err, constructive.ErrPrecisionOverflow)
}

func TestPowPrecisionLimitExemptions(t *testing.T) {
	trillion := New(nil, intRat(new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)))
	ctx := constructive.WithPrecisionLimit(testCtx, 0)

	// Zero, one, and negative one are exact at any exponent.
	result, err := One().Pow(ctx, trillion)
	assert.NoError(t, err)
	assertEqualAtPrecision(t, One(), result, -10)

	result, err = NegativeOne().Pow(ctx, trillion.Add(One()))
	assert.NoError(t, err)
	assertEqualAtPrecision(t, NegativeOne(), result, -10)

	result, err = Zero().Pow(ctx, trillion)
	assert.NoError(t, err)
	assertEqualAtPrecision(t, Zero(), result, -10)

	// An irrational base has no rational factor to bound. Only a tree of about
	// 40 nodes is built, and approximating it is what the limit then governs.
	result, err = irr(constructive.Sqrt2()).Pow(ctx, trillion)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// With no limit, or an explicit lack of one, the size is unbounded.
	for _, ctx := range []context.Context{testCtx, constructive.WithoutPrecisionLimit(testCtx)} {
		result, err = Two().Pow(ctx, rat(100, 1))
		assert.NoError(t, err)
		assert.Equal(t, 0, rational.FromRational(new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), 100))).Cmp(result.rr))
	}
}

func TestPowCancelledContext(t *testing.T) {
	ctx := cancelledContext()

	for _, base := range []*Real{Two(), rat(-3, 4), irr(constructive.Sqrt2()), New(constructive.Sqrt2(), rational.New64(3, 1))} {
		result, err := base.Pow(ctx, rat(3, 1))
		assert.Nil(t, result)
		// The context error is returned as it is, not wrapped.
		assert.True(t, err == context.Canceled, "expected context.Canceled, got %v", err)
	}
}

func TestPowOverflowIsNotAContextError(t *testing.T) {
	ctx := constructive.WithPrecisionLimit(testCtx, 1)
	_, err := Two().Pow(ctx, rat(10, 1))
	assert.True(t, errors.Is(err, constructive.ErrPrecisionOverflow))
	assert.False(t, errors.Is(err, context.Canceled))
}
