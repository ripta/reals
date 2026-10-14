package constructive

import (
	"context"
	"math/big"
	"testing"
)

// testCtx is the context for tests that have no use for cancellation.
var testCtx = context.Background()

func mustPreciseCmp(t testing.TB, a, b Real, p int) int {
	t.Helper()

	v, err := PreciseCmp(testCtx, a, b, p)
	if err != nil {
		t.Fatalf("PreciseCmp(%d): %v", p, err)
	}

	return v
}

func mustFloor(t testing.TB, c Real, p int) Real {
	t.Helper()

	r, err := Floor(testCtx, c, p)
	if err != nil {
		t.Fatalf("Floor(%d): %v", p, err)
	}

	return r
}

func mustCeil(t testing.TB, c Real, p int) Real {
	t.Helper()

	r, err := Ceil(testCtx, c, p)
	if err != nil {
		t.Fatalf("Ceil(%d): %v", p, err)
	}

	return r
}

func mustRound(t testing.TB, c Real, p int) Real {
	t.Helper()

	r, err := Round(testCtx, c, p)
	if err != nil {
		t.Fatalf("Round(%d): %v", p, err)
	}

	return r
}

func mustRoundToEven(t testing.TB, c Real, p int) Real {
	t.Helper()

	r, err := RoundToEven(testCtx, c, p)
	if err != nil {
		t.Fatalf("RoundToEven(%d): %v", p, err)
	}

	return r
}

func mustText(t testing.TB, c Real, dec, radix int) string {
	t.Helper()

	text, err := Text(testCtx, c, dec, radix)
	if err != nil {
		t.Fatalf("Text(%d digits, radix %d): %v", dec, radix, err)
	}

	return text
}

func mustApproximate(t testing.TB, c Real, p int) *big.Int {
	t.Helper()

	i, err := Approximate(testCtx, c, p)
	if err != nil {
		t.Fatalf("Approximate(%d): %v", p, err)
	}

	return i
}
