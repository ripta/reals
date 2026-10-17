package constructive

import (
	"context"
	"math"
)

type precisionOverflowError struct{}

func (e precisionOverflowError) Error() string {
	return "precision overflow"
}

// ErrPrecisionOverflow is returned when a request exceeds the precision limit carried by the
// context. An approximation at a finer precision than the limit exceeds it, and so does an exact
// rational power whose estimated size is larger than the limit.
var ErrPrecisionOverflow error = precisionOverflowError{}

type precisionLimitKey struct{}

// WithoutPrecisionLimit returns a context that places no bound on precision, overriding a limit set
// by a parent.
func WithoutPrecisionLimit(parent context.Context) context.Context {
	return context.WithValue(parent, precisionLimitKey{}, math.MaxInt)
}

// WithPrecisionLimit returns a context that limits approximations to limit bits after the binary
// point. A negative limit is negated.
//
// The limit also bounds the bit size of exact rational powers. A power whose estimated size, the
// exponent times the larger bit length of the numerator and the denominator, exceeds the limit
// fails with ErrPrecisionOverflow, so a tight limit set for display precision also rejects a large
// exact power.
func WithPrecisionLimit(parent context.Context, limit int) context.Context {
	if limit < 0 {
		limit = -limit
	}
	return context.WithValue(parent, precisionLimitKey{}, limit)
}

func PrecisionLimit(ctx context.Context) (int, bool) {
	limit, ok := ctx.Value(precisionLimitKey{}).(int)
	return limit, ok
}

// CheckPrecisionOverflow returns ErrPrecisionOverflow when p is finer than the
// precision limit in the context. The limit is a count of bits after the binary
// point, and finer precision is more negative, so the check fires when p is
// below the negated limit.
func CheckPrecisionOverflow(ctx context.Context, p int) error {
	if limit, ok := PrecisionLimit(ctx); ok && limit >= 0 {
		if p < -limit {
			return ErrPrecisionOverflow
		}
	}

	return nil
}
