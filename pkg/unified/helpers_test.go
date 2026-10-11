package unified

import (
	"context"
	"testing"

	"github.com/ripta/reals/pkg/constructive"
)

// testCtx is the context for tests that have no use for cancellation.
var testCtx = context.Background()

func mustText(t testing.TB, c constructive.Real, dec, radix int) string {
	t.Helper()

	text, err := constructive.Text(testCtx, c, dec, radix)
	if err != nil {
		t.Fatalf("Text(%d digits, radix %d): %v", dec, radix, err)
	}

	return text
}

func mustPreciseCmp(t testing.TB, a, b constructive.Real, p int) int {
	t.Helper()

	v, err := constructive.PreciseCmp(testCtx, a, b, p)
	if err != nil {
		t.Fatalf("PreciseCmp(%d): %v", p, err)
	}

	return v
}

func mustPreciseSign(t testing.TB, c constructive.Real, p int) int {
	t.Helper()

	v, err := constructive.PreciseSign(testCtx, c, p)
	if err != nil {
		t.Fatalf("PreciseSign(%d): %v", p, err)
	}

	return v
}
