package rational

import (
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
	if result := constructive.PreciseCmp(a, b, precision); result != 0 {
		t.Errorf("expected [1] to be equal to [2] at precision %d\n[1]: %s\n     %#v\n[2]: %s\n     %#v", precision, constructive.Text(a, -precision, 10), a, constructive.Text(b, -precision, 10), b)
	}
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
