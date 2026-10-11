package unified

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// ErrNonPositive indicates that a function requires a strictly positive
// argument, such as the argument of a logarithm or the base of a power with a
// non-integer exponent. It is the same value as constructive.ErrNonPositive, so
// errors.Is matches a failure raised in either layer.
var ErrNonPositive = constructive.ErrNonPositive

// ErrNegative indicates that a function requires a non-negative argument, such
// as the radicand of a square root. It is the same value as
// constructive.ErrNegative, so errors.Is matches a failure raised in either
// layer.
var ErrNegative = constructive.ErrNegative

// ErrInvalidBase indicates a logarithm base that is positive but equal to one,
// for which the logarithm is undefined.
var ErrInvalidBase = errors.New("base must not be equal to one")

// ErrOutsideUnitInterval indicates that a function requires an argument in
// [-1, 1], such as the argument of arcsine or arccosine.
var ErrOutsideUnitInterval = errors.New("argument must be in [-1, 1]")

// ErrUndefinedAtOrigin indicates that atan2 was called with both arguments
// zero, where the angle is undefined.
var ErrUndefinedAtOrigin = errors.New("atan2 is undefined at the origin")

// ErrNonPositiveInteger indicates that a function was called at a non-positive
// integer, where it is undefined. Gamma has its poles there.
var ErrNonPositiveInteger = errors.New("argument must not be a non-positive integer")

// halfPi is π/2, computed once and reused by the inverse-trig endpoints.
var halfPi = sync.OnceValue(func() *Real {
	return Pi().ShiftRight(1)
})

// ln10 is the natural logarithm of ten, computed once and reused as the
// denominator of Log10.
var ln10 = sync.OnceValue(func() *Real {
	return New(constructive.Ln(Ten().Constructive()), rational.One())
})

// Sin returns the sine of u, in radians.
func (u *Real) Sin() *Real {
	return New(constructive.Sine(u.Constructive()), rational.One())
}

// Cos returns the cosine of u, in radians.
func (u *Real) Cos() *Real {
	return New(constructive.Cosine(u.Constructive()), rational.One())
}

// Tan returns the tangent of u, in radians.
func (u *Real) Tan() *Real {
	return New(constructive.Tangent(u.Constructive()), rational.One())
}

// Exp returns e raised to the power of u.
func (u *Real) Exp() *Real {
	return New(constructive.Exp(u.Constructive()), rational.One())
}

// Abs returns the absolute value of u.
func (u *Real) Abs() *Real {
	return New(constructive.Abs(u.Constructive()), rational.One())
}

// Ln returns the natural logarithm of u. It requires a positive argument and
// returns ErrNonPositive otherwise. Deciding the sign of a non-structural zero
// runs until ctx is done, so the context error is returned in that case.
func (u *Real) Ln(ctx context.Context) (*Real, error) {
	c := u.Constructive()
	neg, err := u.nonPositive(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Ln: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Ln: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()), nil
}

// Sqrt returns the square root of u. It requires a non-negative argument and
// returns ErrNegative otherwise. Errors from deciding the sign, including the
// context error, are returned.
func (u *Real) Sqrt(ctx context.Context) (*Real, error) {
	if u.IsZero() {
		return Zero(), nil
	}
	c := u.Constructive()
	sg, err := constructive.Sign(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Sqrt: %w", err)
	}
	if sg < 0 {
		return nil, fmt.Errorf("Sqrt: %w", ErrNegative)
	}
	return New(constructive.Sqrt(c), rational.One()), nil
}

// Gamma returns the gamma function Γ(u). It is undefined at the non-positive
// integers, its poles, and returns ErrNonPositiveInteger there. Detection is
// structural: it recognizes only an exact non-positive integer on the rational
// representation, since constructive reals cannot decide equality in general. A
// near-pole irrational argument is finite and computes to a large value.
func (u *Real) Gamma() (*Real, error) {
	if u.cr == constructive.One() && u.rr.IsInteger() && u.rr.Sign() <= 0 {
		return nil, fmt.Errorf("Gamma: %w", ErrNonPositiveInteger)
	}
	return New(constructive.Gamma(u.Constructive()), rational.One()), nil
}

// Pow returns u raised to the power n.
//
// An integer exponent is honored over any base via sign-aware repeated
// multiplication. A non-integer exponent requires a positive base and returns
// ErrNonPositive otherwise. The base zero is handled explicitly:
//
// - `Pow(0, k)` is 0 for positive k.
// - `Pow(0, 0)` is 1
// - A negative or non-integer exponent over a zero base returns ErrNonPositive.
//
// Errors from deciding the sign, including the context error, are returned.
func (u *Real) Pow(ctx context.Context, n *Real) (*Real, error) {
	if n.cr == constructive.One() && n.rr.IsInteger() {
		if u.IsZero() {
			switch n.rr.Sign() {
			case 1:
				return Zero(), nil
			case 0:
				return One(), nil
			default:
				return nil, fmt.Errorf("Pow: %w", ErrNonPositive)
			}
		}

		result := One()
		e := n.rr
		if e.Sign() < 0 {
			e = e.Negate()
		}
		for e.Sign() > 0 {
			result = result.Multiply(u)
			e = e.Subtract(rational.One())
		}

		if n.rr.Sign() < 0 {
			result = result.Inverse()
		}
		return result, nil
	}

	c := u.Constructive()
	neg, err := u.nonPositive(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Pow: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Pow: %w", ErrNonPositive)
	}

	return New(constructive.Pow(c, n.Constructive()), rational.One()), nil
}

// Log10 returns the base-10 logarithm of u. It requires a positive argument and
// returns ErrNonPositive otherwise. Errors from deciding the sign, including the
// context error, are returned.
func (u *Real) Log10(ctx context.Context) (*Real, error) {
	c := u.Constructive()
	neg, err := u.nonPositive(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Log10: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Log10: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()).Divide(ln10()), nil
}

// Log2 returns the base-2 logarithm of u. It requires a positive argument and
// returns ErrNonPositive otherwise. Errors from deciding the sign, including the
// context error, are returned.
func (u *Real) Log2(ctx context.Context) (*Real, error) {
	c := u.Constructive()
	neg, err := u.nonPositive(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Log2: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Log2: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()).Divide(Ln2()), nil
}

// Log returns the logarithm of u in the given base. It requires a positive
// argument and a positive base; a non-positive argument or base returns
// ErrNonPositive, and a base of one returns ErrInvalidBase. Errors from deciding
// the sign, including the context error, are returned.
func (u *Real) Log(ctx context.Context, base *Real) (*Real, error) {
	c := u.Constructive()
	neg, err := u.nonPositive(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("Log: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Log: %w", ErrNonPositive)
	}

	bc := base.Constructive()
	neg, err = base.nonPositive(ctx, bc)
	if err != nil {
		return nil, fmt.Errorf("Log: %w", err)
	}
	if neg {
		return nil, fmt.Errorf("Log: %w", ErrNonPositive)
	}
	if base.cr == constructive.One() && base.rr.Cmp(rational.One()) == 0 {
		return nil, fmt.Errorf("Log: %w", ErrInvalidBase)
	}

	num := New(constructive.Ln(c), rational.One())
	den := New(constructive.Ln(bc), rational.One())
	return num.Divide(den), nil
}

// Sinh returns the hyperbolic sine of u.
func (u *Real) Sinh() *Real {
	ex := u.Exp()
	enx := u.Negate().Exp()
	return ex.Subtract(enx).ShiftRight(1)
}

// Cosh returns the hyperbolic cosine of u.
func (u *Real) Cosh() *Real {
	ex := u.Exp()
	enx := u.Negate().Exp()
	return ex.Add(enx).ShiftRight(1)
}

// Tanh returns the hyperbolic tangent of u.
func (u *Real) Tanh() *Real {
	ex := u.Exp()
	enx := u.Negate().Exp()
	return ex.Subtract(enx).Divide(ex.Add(enx))
}

// Cbrt returns the real cube root of u. It is total: it accepts negative input,
// so Cbrt(-8) is -2, and Cbrt(0) is 0. The result is the cube root of |u|, with
// the sign of u applied by a conditional node. The sign is decided when the
// result is approximated, not here.
func (u *Real) Cbrt() *Real {
	if u.IsZero() {
		return Zero()
	}

	third := constructive.FromRat(1, 3)

	if u.cr == constructive.One() {
		magnitude := u.rr
		if magnitude.Sign() < 0 {
			magnitude = magnitude.Negate()
		}

		result := New(constructive.Pow(magnitude.Constructive(), third), rational.One())
		if u.rr.Sign() < 0 {
			return result.Negate()
		}

		return result
	}

	c := u.Constructive()
	root := constructive.Pow(constructive.Abs(c), third)

	return New(constructive.CondSign(c, constructive.Negate(root), root), rational.One())
}

// Atan returns the arctangent of u, in radians. It is total.
func (u *Real) Atan() *Real {
	return New(constructive.Arctangent(u.Constructive()), rational.One())
}

// Asin returns the arcsine of u, in radians. It requires an argument in
// [-1, 1] and returns ErrOutsideUnitInterval otherwise. The endpoints ±1 are
// special-cased to ±π/2, where the derived form atan(x / sqrt(1 - x²)) would
// divide by zero. Endpoint detection is structural and so recognizes only the
// rational ±1, since constructive reals cannot decide equality in general.
// Errors from deciding the sign, including the context error, are returned.
func (u *Real) Asin(ctx context.Context) (*Real, error) {
	if s := u.unitEndpoint(); s != 0 {
		if s > 0 {
			return halfPi(), nil
		}
		return halfPi().Negate(), nil
	}

	c := u.Constructive()
	s := constructive.Subtract(constructive.One(), constructive.Square(c))
	sg, err := constructive.Sign(ctx, s)
	if err != nil {
		return nil, fmt.Errorf("Asin: %w", err)
	}
	if sg < 0 {
		return nil, fmt.Errorf("Asin: %w", ErrOutsideUnitInterval)
	}
	inner := constructive.Divide(c, constructive.Sqrt(s))
	return New(constructive.Arctangent(inner), rational.One()), nil
}

// Acos returns the arccosine of u, in radians, as π/2 - asin(u). It requires an
// argument in [-1, 1] and returns ErrOutsideUnitInterval otherwise.
func (u *Real) Acos(ctx context.Context) (*Real, error) {
	asin, err := u.Asin(ctx)
	if err != nil {
		return nil, fmt.Errorf("Acos: %w", err)
	}
	return halfPi().Subtract(asin), nil
}

// Atan2 returns the angle, in radians, of the point (x, y) measured from the
// positive x-axis, where the receiver is y. The result lies in (-π, π]. Both
// arguments zero returns ErrUndefinedAtOrigin.
func (y *Real) Atan2(ctx context.Context, x *Real) (*Real, error) {
	xZero := x.IsZero()
	yZero := y.IsZero()
	if xZero && yZero {
		return nil, fmt.Errorf("Atan2: %w", ErrUndefinedAtOrigin)
	}
	if xZero {
		ys, err := y.Sign(ctx)
		if err != nil {
			return nil, fmt.Errorf("Atan2: %w", err)
		}
		if ys > 0 {
			return halfPi(), nil
		}
		return halfPi().Negate(), nil
	}

	base := y.Divide(x).Atan()
	xs, err := x.Sign(ctx)
	if err != nil {
		return nil, fmt.Errorf("Atan2: %w", err)
	}
	if xs > 0 {
		return base, nil
	}
	ys, err := y.Sign(ctx)
	if err != nil {
		return nil, fmt.Errorf("Atan2: %w", err)
	}
	if ys >= 0 {
		return base.Add(Pi()), nil
	}
	return base.Subtract(Pi()), nil
}

// unitEndpoint reports +1 if u is exactly +1, -1 if exactly -1, and 0
// otherwise. Detection is structural: it recognizes the rational ±1, since
// constructive reals cannot decide equality in general.
func (u *Real) unitEndpoint() int {
	if u.cr != constructive.One() {
		return 0
	}
	switch {
	case u.rr.Cmp(rational.One()) == 0:
		return 1
	case u.rr.Cmp(rational.One().Negate()) == 0:
		return -1
	}
	return 0
}

// nonPositive reports whether u, whose constructive component is c, is zero or
// negative. A structural zero is decided without approximating c.
func (u *Real) nonPositive(ctx context.Context, c constructive.Real) (bool, error) {
	if u.IsZero() {
		return true, nil
	}

	sg, err := constructive.Sign(ctx, c)
	if err != nil {
		return false, err
	}

	return sg < 0, nil
}
