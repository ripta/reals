package constructive

import (
	"context"
	"math"
)

type precisionOverflowError struct{}

func (e precisionOverflowError) Error() string {
	return "precision overflow"
}

var PrecisionOverflow error = precisionOverflowError{}

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

func CheckPrecisionOverflow(ctx context.Context, p int) error {
	if limit, ok := PrecisionLimit(ctx); ok && limit >= 0 {
		if p > limit {
			return PrecisionOverflow
		}
	}

	return nil
}
