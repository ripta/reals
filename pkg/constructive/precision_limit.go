package constructive

import (
	"context"
	"math"
)

type precisionOverflowError struct{}

func (e precisionOverflowError) Error() string {
	return "precision overflow"
}

// ErrPrecisionOverflow is returned when an approximation is requested at a finer
// precision than the limit carried by the context.
var ErrPrecisionOverflow error = precisionOverflowError{}

type precisionLimitKey struct{}

func WithoutPrecisionLimit(parent context.Context) context.Context {
	return context.WithValue(parent, precisionLimitKey{}, math.MaxInt)
}

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
