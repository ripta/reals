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
