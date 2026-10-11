package unified

import (
	"context"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// Floor returns the greatest integer less than or equal to u. A purely rational
// value is floored exactly; otherwise the boundary is decided at precision p.
// Errors from approximating u are returned as they are.
func (u *Real) Floor(ctx context.Context, p int) (*Real, error) {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Floor()), nil
	}

	r, err := constructive.Floor(ctx, u.Constructive(), p)
	if err != nil {
		return nil, err
	}

	return New(r, rational.One()), nil
}

// Ceil returns the least integer greater than or equal to u. A purely rational
// value is ceiled exactly; otherwise the boundary is decided at precision p.
// Errors from approximating u are returned as they are.
func (u *Real) Ceil(ctx context.Context, p int) (*Real, error) {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Ceil()), nil
	}

	r, err := constructive.Ceil(ctx, u.Constructive(), p)
	if err != nil {
		return nil, err
	}

	return New(r, rational.One()), nil
}

// Round returns the nearest integer to u, rounding half away from zero. A purely
// rational value is rounded exactly; otherwise the boundary is decided at
// precision p, where an exact halfway tie is not finitely decidable. Errors from
// approximating u are returned as they are.
func (u *Real) Round(ctx context.Context, p int) (*Real, error) {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Round()), nil
	}

	r, err := constructive.Round(ctx, u.Constructive(), p)
	if err != nil {
		return nil, err
	}

	return New(r, rational.One()), nil
}

// RoundToEven returns the nearest integer to u, rounding ties to even. A purely
// rational value is rounded exactly; otherwise the boundary is decided at
// precision p, where an exact halfway tie is not finitely decidable. Errors from
// approximating u are returned as they are.
func (u *Real) RoundToEven(ctx context.Context, p int) (*Real, error) {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.RoundToEven()), nil
	}

	r, err := constructive.RoundToEven(ctx, u.Constructive(), p)
	if err != nil {
		return nil, err
	}

	return New(r, rational.One()), nil
}

// Min returns the lesser of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at precision p with PreciseCmp, so
// operands equal within p may return either operand. Errors from approximating
// either operand are returned as they are.
func (u *Real) Min(ctx context.Context, other *Real, p int) (*Real, error) {
	v, err := u.preciseCmp(ctx, other, p)
	if err != nil {
		return nil, err
	}
	if v <= 0 {
		return u, nil
	}

	return other, nil
}

// Max returns the greater of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at precision p with PreciseCmp, so
// operands equal within p may return either operand. Errors from approximating
// either operand are returned as they are.
func (u *Real) Max(ctx context.Context, other *Real, p int) (*Real, error) {
	v, err := u.preciseCmp(ctx, other, p)
	if err != nil {
		return nil, err
	}
	if v >= 0 {
		return u, nil
	}

	return other, nil
}

// preciseCmp compares u with other exactly when both are purely rational, and
// otherwise at precision p.
func (u *Real) preciseCmp(ctx context.Context, other *Real, p int) (int, error) {
	if u.cr == constructive.One() && other.cr == constructive.One() {
		return u.rr.Cmp(other.rr), nil
	}

	return constructive.PreciseCmp(ctx, u.Constructive(), other.Constructive(), p)
}
