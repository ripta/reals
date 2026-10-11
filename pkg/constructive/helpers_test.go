package constructive

import (
	"context"
	"math/big"
	"testing"
)

// testCtx is the context for tests that have no use for cancellation.
var testCtx = context.Background()

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
