package rational

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/ripta/reals/pkg/constructive"
)

func TestNew(t *testing.T) {
	assertRationalEqual(t, Zero(), Zero())
	assertRationalEqual(t, One(), One())
	assertRationalEqual(t, New64(3, 4), New64(3, 4))
	assertRationalEqual(t, New(big.NewInt(3), big.NewInt(4)), New(big.NewInt(3), big.NewInt(4)))
	assertRationalEqual(t, New(big.NewInt(3), big.NewInt(4)), New64(3, 4))
	assertRationalEqual(t, FromRational(big.NewRat(3, 4)), New64(3, 4))
}

func TestNumber(t *testing.T) {
	assertRationalEqual(t, One(), Zero().Add(One()))
	assertRationalEqual(t, Zero(), One().Subtract(One()))
	assertRationalEqual(t, New64(3, 4), New64(1, 2).Add(New64(1, 4)))
	assertRationalEqual(t, New64(1, 4), New64(3, 4).Subtract(New64(1, 2)))
	assertRationalEqual(t, New64(3, 8), New64(3, 4).Multiply(New64(1, 2)))
	assertRationalEqual(t, New64(-3, 8), New64(3, 4).Multiply(New64(-1, 2)))
	assertRationalEqual(t, New64(-3, 8), New64(3, 8).Negate())
	assertRationalEqual(t, New64(3, 8), New64(8, 3).Inverse())
}

func assertRationalEqual(t *testing.T, expected, actual *Number) {
	if expected.r.Cmp(actual.r) != 0 {
		t.Errorf("Expected %s, got %s", expected.r.String(), actual.r.String())
	}
}

func TestNumber_Constructive(t *testing.T) {
	assertEqualAtPrecision(t, constructive.FromInt(1), One().Constructive(), -100)
	assertEqualAtPrecision(t, constructive.Pi(), New64(22, 7).Constructive(), -9)
	assertEqualAtPrecision(t, constructive.Pi(), New64(223, 71).Constructive(), -9)
	assertEqualAtPrecision(t, constructive.Pi(), New64(377, 120).Constructive(), -13)
}

func assertEqualAtPrecision(t *testing.T, a, b constructive.Real, precision int) {
	t.Helper()
	if result := mustPreciseCmp(t, a, b, precision); result != 0 {
		t.Errorf("expected [1] to be equal to [2] at precision %d\n[1]: %s\n     %#v\n[2]: %s\n     %#v", precision, mustText(t, a, -precision, 10), a, mustText(t, b, -precision, 10), b)
	}
}

func TestFloor(t *testing.T) {
	assertRationalEqual(t, New64(3, 1), New64(3, 1).Floor())
	assertRationalEqual(t, New64(2, 1), New64(5, 2).Floor())
	assertRationalEqual(t, New64(-3, 1), New64(-5, 2).Floor())
	assertRationalEqual(t, New64(2, 1), New64(7, 3).Floor())
	assertRationalEqual(t, New64(-3, 1), New64(-7, 3).Floor())
}

func TestCeil(t *testing.T) {
	assertRationalEqual(t, New64(3, 1), New64(3, 1).Ceil())
	assertRationalEqual(t, New64(3, 1), New64(5, 2).Ceil())
	assertRationalEqual(t, New64(-2, 1), New64(-5, 2).Ceil())
	assertRationalEqual(t, New64(3, 1), New64(7, 3).Ceil())
	assertRationalEqual(t, New64(-2, 1), New64(-7, 3).Ceil())
}

func TestRound(t *testing.T) {
	assertRationalEqual(t, New64(3, 1), New64(5, 2).Round())
	assertRationalEqual(t, New64(-3, 1), New64(-5, 2).Round())
	assertRationalEqual(t, New64(2, 1), New64(7, 3).Round())
	assertRationalEqual(t, New64(3, 1), New64(8, 3).Round())
	assertRationalEqual(t, New64(4, 1), New64(4, 1).Round())
}

func TestRoundToEven(t *testing.T) {
	assertRationalEqual(t, New64(2, 1), New64(5, 2).RoundToEven())
	assertRationalEqual(t, New64(4, 1), New64(7, 2).RoundToEven())
	assertRationalEqual(t, New64(-2, 1), New64(-5, 2).RoundToEven())
	assertRationalEqual(t, New64(2, 1), New64(7, 3).RoundToEven())
}

func TestShift(t *testing.T) {
	// Shifting left positive
	assertRationalEqual(t, New64(12, 4), New64(3, 4).ShiftLeft(2)) // 3/4 * 4 = 12/4
	assertRationalEqual(t, New64(6, 4), New64(3, 4).ShiftLeft(1))  // 3/4 * 2 = 6/4
	assertRationalEqual(t, New64(24, 4), New64(3, 4).ShiftLeft(3)) // 3/4 * 8 = 24/4 = 6
	assertRationalEqual(t, New64(6, 1), New64(3, 4).ShiftLeft(3))  // 3/4 * 8 = 24/4 = 6

	// Shifting left negative
	assertRationalEqual(t, New64(3, 16), New64(3, 4).ShiftLeft(-2))
	assertRationalEqual(t, New64(3, 8), New64(3, 4).ShiftLeft(-1))

	// Shifting right positive
	assertRationalEqual(t, New64(3, 16), New64(3, 4).ShiftRight(2)) // 3/4 / 4 = 3/16
	assertRationalEqual(t, New64(3, 8), New64(3, 4).ShiftRight(1))  // 3/4 / 2 = 3/8
	assertRationalEqual(t, New64(3, 32), New64(3, 4).ShiftRight(3)) // 3/4 / 8 = 3/32

	// Shifting right negative
	assertRationalEqual(t, New64(12, 4), New64(3, 4).ShiftRight(-2))
	assertRationalEqual(t, New64(6, 4), New64(3, 4).ShiftRight(-1))

	// Shifting by zero
	assertRationalEqual(t, New64(3, 4), New64(3, 4).ShiftLeft(0))
	assertRationalEqual(t, New64(3, 4), New64(3, 4).ShiftRight(0))

	// Shifting zero
	assertRationalEqual(t, Zero(), Zero().ShiftLeft(5))
	assertRationalEqual(t, Zero(), Zero().ShiftRight(5))
	assertRationalEqual(t, Zero(), Zero().ShiftLeft(-5))
	assertRationalEqual(t, Zero(), Zero().ShiftRight(-5))

	// Shifting one
	assertRationalEqual(t, New64(8, 1), One().ShiftLeft(3))  // 1 * 8 = 8
	assertRationalEqual(t, New64(1, 8), One().ShiftRight(3)) // 1 / 8 = 1/8

	// Large shifts
	assertRationalEqual(t, New64(3072, 4), New64(3, 4).ShiftLeft(10))  // 3/4 * 1024 = 3072/4
	assertRationalEqual(t, New64(3, 4096), New64(3, 4).ShiftRight(10)) // 3/4 / 1024 = 3/4096
}

// normalizedPow is the reference power: SetFrac normalizes the raised numerator
// and denominator, and a negative exponent swaps them first.
func normalizedPow(a, b, e int64) *big.Rat {
	n, d := big.NewInt(a), big.NewInt(b)
	if e < 0 {
		n, d = d, n
		e = -e
	}
	exp := big.NewInt(e)
	n = new(big.Int).Exp(n, exp, nil)
	d = new(big.Int).Exp(d, exp, nil)

	return new(big.Rat).SetFrac(n, d)
}

func assertRatParts(t *testing.T, name string, expected *big.Rat, actual *Number) {
	t.Helper()
	if expected.Num().Cmp(actual.r.Num()) != 0 || expected.Denom().Cmp(actual.r.Denom()) != 0 {
		t.Errorf("%s: expected %s, got %s", name, expected.String(), actual.r.String())
	}
}

func TestNumber_Pow(t *testing.T) {
	bases := []struct{ a, b int64 }{
		{2, 1}, {-2, 1}, {3, 4}, {-3, 4}, {1, 7}, {-1, 7}, {355, 113}, {-6, 4}, {10, 1},
		{1, 1}, {-1, 1}, {0, 1},
	}
	exponents := []int64{0, 1, 2, 3, 4, 7, 10, 33, 64, -1, -2, -3, -10, -33}
	for _, base := range bases {
		for _, e := range exponents {
			if base.a == 0 && e < 0 {
				continue
			}
			name := New64(base.a, base.b).String() + "^" + big.NewInt(e).String()
			got, err := New64(base.a, base.b).Pow(testCtx, big.NewInt(e))
			if err != nil {
				t.Errorf("%s: unexpected error %v", name, err)
				continue
			}
			assertRatParts(t, name, normalizedPow(base.a, base.b, e), got)
		}
	}
}

func TestNumber_PowResultIsUsable(t *testing.T) {
	got, err := New64(-3, 4).Pow(testCtx, big.NewInt(5))
	if err != nil {
		t.Fatal(err)
	}
	assertRationalEqual(t, New64(-243, 1024), got)
	assertRationalEqual(t, New64(-243, 1024).Add(One()), got.Add(One()))
	if got.IsInteger() || got.Sign() != -1 {
		t.Errorf("expected a negative non-integer, got %s", got)
	}
}

func TestNumber_PowZeroBase(t *testing.T) {
	got, err := Zero().Pow(testCtx, big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}
	assertRationalEqual(t, One(), got)

	got, err = Zero().Pow(testCtx, big.NewInt(5))
	if err != nil {
		t.Fatal(err)
	}
	assertRationalEqual(t, Zero(), got)

	got, err = Zero().Pow(testCtx, big.NewInt(-1))
	if !errors.Is(err, constructive.ErrNonPositive) || got != nil {
		t.Errorf("expected ErrNonPositive and no result, got %v and %v", got, err)
	}
}

func TestNumber_PowDoesNotMutate(t *testing.T) {
	base, e := New64(3, 4), big.NewInt(5)
	if _, err := base.Pow(testCtx, e); err != nil {
		t.Fatal(err)
	}
	assertRationalEqual(t, New64(3, 4), base)
	if e.Cmp(big.NewInt(5)) != 0 {
		t.Errorf("expected the exponent to be unchanged, got %s", e)
	}
}

func TestNumber_Num(t *testing.T) {
	n := New64(-6, 4)
	got := n.Num()
	if got.Cmp(big.NewInt(-3)) != 0 {
		t.Errorf("expected -3, got %s", got)
	}
	got.SetInt64(99)
	assertRationalEqual(t, New64(-3, 2), n)
}

func TestNumber_PowPrecisionLimit(t *testing.T) {
	// The estimate of 2^10 is 10 * 2 bits, because 2 has a bit length of 2.
	e := big.NewInt(10)
	exact, err := New64(2, 1).Pow(testCtx, e)
	if err != nil {
		t.Fatal(err)
	}

	got, err := New64(2, 1).Pow(constructive.WithPrecisionLimit(testCtx, 20), e)
	if err != nil {
		t.Fatalf("expected an estimate equal to the limit to succeed, got %v", err)
	}
	assertRationalEqual(t, exact, got)

	got, err = New64(2, 1).Pow(constructive.WithPrecisionLimit(testCtx, 19), e)
	if !errors.Is(err, constructive.ErrPrecisionOverflow) || got != nil {
		t.Errorf("expected ErrPrecisionOverflow and no result, got %v and %v", got, err)
	}

	// The estimate uses the larger of the numerator and denominator bit lengths,
	// and the magnitude of a negative exponent.
	got, err = New64(1, 7).Pow(constructive.WithPrecisionLimit(testCtx, 30), big.NewInt(-10))
	if err != nil {
		t.Errorf("expected 10 * 3 bits to fit a limit of 30, got %v", err)
	}
	assertRationalEqual(t, New64(282475249, 1), got)
	if _, err := New64(1, 7).Pow(constructive.WithPrecisionLimit(testCtx, 29), big.NewInt(-10)); !errors.Is(err, constructive.ErrPrecisionOverflow) {
		t.Errorf("expected ErrPrecisionOverflow, got %v", err)
	}

	// A negative limit is negated, as WithPrecisionLimit documents.
	if _, err := New64(2, 1).Pow(constructive.WithPrecisionLimit(testCtx, -20), e); err != nil {
		t.Errorf("expected a limit of -20 to act as 20, got %v", err)
	}
}

func TestNumber_PowHugeExponent(t *testing.T) {
	trillion := new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)
	ctx := constructive.WithPrecisionLimit(testCtx, 1<<20)
	for _, base := range []*Number{New64(2, 1), New64(-3, 4), New64(1, 3)} {
		got, err := base.Pow(ctx, trillion)
		if !errors.Is(err, constructive.ErrPrecisionOverflow) || got != nil {
			t.Errorf("%s: expected ErrPrecisionOverflow and no result, got %v and %v", base, got, err)
		}
		if _, err := base.Pow(ctx, new(big.Int).Neg(trillion)); !errors.Is(err, constructive.ErrPrecisionOverflow) {
			t.Errorf("%s: expected ErrPrecisionOverflow for the reciprocal, got %v", base, err)
		}
	}
}

func TestNumber_PowExactBasesAreNotBounded(t *testing.T) {
	trillion := new(big.Int).Exp(big.NewInt(10), big.NewInt(12), nil)
	ctx := constructive.WithPrecisionLimit(testCtx, 0)

	tests := []struct {
		name     string
		base     *Number
		e        *big.Int
		expected *Number
	}{
		{"1^huge", One(), trillion, One()},
		{"1^-huge", One(), new(big.Int).Neg(trillion), One()},
		{"(-1)^even", New64(-1, 1), trillion, One()},
		{"(-1)^odd", New64(-1, 1), new(big.Int).Add(trillion, big.NewInt(1)), New64(-1, 1)},
		{"(-1)^-odd", New64(-1, 1), new(big.Int).Neg(new(big.Int).Add(trillion, big.NewInt(1))), New64(-1, 1)},
		{"0^huge", Zero(), trillion, Zero()},
		{"2^0", New64(2, 1), big.NewInt(0), One()},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.base.Pow(ctx, test.e)
			if err != nil {
				t.Fatal(err)
			}
			assertRationalEqual(t, test.expected, got)
		})
	}
}

func TestNumber_PowWithoutLimit(t *testing.T) {
	// With no limit in the context, and with an explicit lack of one, the size is unbounded.
	for _, ctx := range []context.Context{testCtx, constructive.WithoutPrecisionLimit(testCtx)} {
		got, err := New64(2, 1).Pow(ctx, big.NewInt(1000))
		if err != nil {
			t.Fatal(err)
		}
		assertRatParts(t, "2^1000", normalizedPow(2, 1, 1000), got)
	}

	// An explicit lack of a limit overrides a limit set by a parent.
	ctx := constructive.WithoutPrecisionLimit(constructive.WithPrecisionLimit(testCtx, 1))
	if _, err := New64(2, 1).Pow(ctx, big.NewInt(1000)); err != nil {
		t.Errorf("expected no bound, got %v", err)
	}
}

// failingContext reports no error for its first ok calls to Err, then reports
// context.Canceled.
type failingContext struct {
	context.Context
	ok int
}

func (c *failingContext) Err() error {
	if c.ok > 0 {
		c.ok--
		return nil
	}
	return context.Canceled
}

func TestNumber_PowCancelled(t *testing.T) {
	cancelled, cancel := context.WithCancel(testCtx)
	cancel()

	got, err := New64(2, 1).Pow(cancelled, big.NewInt(3))
	if err != context.Canceled || got != nil {
		t.Errorf("expected context.Canceled unwrapped and no result, got %v and %v", got, err)
	}

	// The check on entry passes, so the error comes from between squaring steps.
	got, err = New64(3, 1).Pow(&failingContext{Context: testCtx, ok: 1}, big.NewInt(100))
	if err != context.Canceled || got != nil {
		t.Errorf("expected context.Canceled from the squaring loop, got %v and %v", got, err)
	}

	// Bases that need no squaring still honor a context that is already done.
	if _, err := One().Pow(cancelled, big.NewInt(3)); err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
