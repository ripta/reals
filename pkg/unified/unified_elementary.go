package unified

import (
	"errors"
	"fmt"
	"sync"

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

// ErrInvalidBase indicates a logarithm base that is positive but equal to one,
// for which the logarithm is undefined.
var ErrInvalidBase = errors.New("base must not be equal to one")

// ErrOutsideUnitInterval indicates that a function requires an argument in
// [-1, 1], such as the argument of arcsine or arccosine.
var ErrOutsideUnitInterval = errors.New("argument must be in [-1, 1]")

// ErrUndefinedAtOrigin indicates that atan2 was called with both arguments
// zero, where the angle is undefined.
var ErrUndefinedAtOrigin = errors.New("atan2 is undefined at the origin")

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

// Log10 returns the base-10 logarithm of u. It requires a positive argument and
// returns ErrNonPositive otherwise.
func (u *Real) Log10() (*Real, error) {
	c := u.Constructive()
	if u.IsZero() || constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Log10: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()).Divide(ln10()), nil
}

// Log2 returns the base-2 logarithm of u. It requires a positive argument and
// returns ErrNonPositive otherwise.
func (u *Real) Log2() (*Real, error) {
	c := u.Constructive()
	if u.IsZero() || constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Log2: %w", ErrNonPositive)
	}
	return New(constructive.Ln(c), rational.One()).Divide(Ln2()), nil
}

// Log returns the logarithm of u in the given base. It requires a positive
// argument and a positive base; a non-positive argument or base returns
// ErrNonPositive, and a base of one returns ErrInvalidBase.
func (u *Real) Log(base *Real) (*Real, error) {
	c := u.Constructive()
	if u.IsZero() || constructive.Sign(c) < 0 {
		return nil, fmt.Errorf("Log: %w", ErrNonPositive)
	}

	bc := base.Constructive()
	if base.IsZero() || constructive.Sign(bc) < 0 {
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
// so Cbrt(-8) is -2, and Cbrt(0) is 0. The cube root of a negative value is
// computed by sign extraction over Pow(|u|, 1/3), keeping the result real.
func (u *Real) Cbrt() *Real {
	if u.IsZero() {
		return Zero()
	}

	third := New(constructive.One(), rational.New64(1, 3))
	result, _ := u.Abs().Pow(third)
	if constructive.Sign(u.Constructive()) < 0 {
		return result.Negate()
	}
	return result
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
func (u *Real) Asin() (*Real, error) {
	if s := u.unitEndpoint(); s != 0 {
		if s > 0 {
			return halfPi(), nil
		}
		return halfPi().Negate(), nil
	}

	c := u.Constructive()
	s := constructive.Subtract(constructive.One(), constructive.Square(c))
	if constructive.Sign(s) < 0 {
		return nil, fmt.Errorf("Asin: %w", ErrOutsideUnitInterval)
	}
	inner := constructive.Divide(c, constructive.Sqrt(s))
	return New(constructive.Arctangent(inner), rational.One()), nil
}

// Acos returns the arccosine of u, in radians, as π/2 - asin(u). It requires an
// argument in [-1, 1] and returns ErrOutsideUnitInterval otherwise.
func (u *Real) Acos() (*Real, error) {
	asin, err := u.Asin()
	if err != nil {
		return nil, fmt.Errorf("Acos: %w", ErrOutsideUnitInterval)
	}
	return halfPi().Subtract(asin), nil
}

// Atan2 returns the angle, in radians, of the point (x, y) measured from the
// positive x-axis, where the receiver is y. The result lies in (-π, π]. Both
// arguments zero returns ErrUndefinedAtOrigin.
func (y *Real) Atan2(x *Real) (*Real, error) {
	xZero := x.IsZero()
	yZero := y.IsZero()
	if xZero && yZero {
		return nil, fmt.Errorf("Atan2: %w", ErrUndefinedAtOrigin)
	}
	if xZero {
		if y.sign() > 0 {
			return halfPi(), nil
		}
		return halfPi().Negate(), nil
	}

	base := y.Divide(x).Atan()
	if x.sign() > 0 {
		return base, nil
	}
	if y.sign() >= 0 {
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

// sign returns the sign of u: 0 when u is structurally zero, otherwise the
// constructive sign, which terminates for nonzero values.
func (u *Real) sign() int {
	if u.IsZero() {
		return 0
	}
	return constructive.Sign(u.Constructive())
}
