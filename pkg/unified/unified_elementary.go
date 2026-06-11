package unified

import (
	"errors"
	"fmt"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// ErrNonPositive indicates thata function requires a strictly positive
// argument, such as the argument of a logarithm or the base of a power with a
// non-integer exponent.
var ErrNonPositive = errors.New("argument must be positive")

// ErrNegative ndicates that a function requires a non-negative argument, such
// as the radicand of a square root.
var ErrNegative = errors.New("argument must be non-negative")

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
// returns ErrNonPositive otherwise.
func (u *Real) Ln() (*Real, error) {
	c := u.Constructive()
	if u.IsZero() || constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Ln: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()), nil
}

// Sqrt returns the square root of u. It requires a non-negative argument and
// returns ErrNegative otherwise.
func (u *Real) Sqrt() (*Real, error) {
	if u.IsZero() {
		return Zero(), nil
	}
	c := u.Constructive()
	if constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Sqrt: %w", ErrNegative)
	}
	return New(constructive.Sqrt(c), rational.One()), nil
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
func (u *Real) Pow(n *Real) (*Real, error) {
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
	if u.IsZero() || constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Pow: %w", ErrNonPositive)
	}

	return New(constructive.Pow(c, n.Constructive()), rational.One()), nil
}
