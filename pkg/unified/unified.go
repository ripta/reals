package unified

import (
	"fmt"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// Real represents a real number as a unification of a constructive real and
// a rational number.
type Real struct {
	cr constructive.Real
	rr *rational.Number
}

// New creates a new Real number from the given constructive real and
// rational number. The actual value being represented is `cr * rr`;
// if either argument is nil, it defaults to one.
func New(cr constructive.Real, rr *rational.Number) *Real {
	if cr == nil {
		cr = constructive.One()
	}
	if rr == nil {
		rr = rational.One()
	}

	return &Real{
		cr: cr,
		rr: rr,
	}
}

// Constructive returns the constructive real representation of the unified
// real number.
func (u *Real) Constructive() constructive.Real {
	return constructive.Multiply(u.cr, u.rr.Constructive())
}

// Add adds the current number and another number together, returning a new
// Real number.
func (u *Real) Add(other *Real) *Real {
	if u.cr == other.cr {
		return New(u.cr, u.rr.Add(other.rr))
	}
	if other.IsZero() {
		return u
	}
	if u.IsZero() {
		return other
	}
	return New(constructive.Add(u.Constructive(), other.Constructive()), rational.One())
}

// Subtract `other` from the current number, returning a new Real number.
func (u *Real) Subtract(other *Real) *Real {
	return u.Add(other.Negate())
}

// Multiply multiplies the current number by another number, returning a new
// Real number.
func (u *Real) Multiply(other *Real) *Real {
	if u.cr == constructive.One() {
		return New(other.cr, u.rr.Multiply(other.rr))
	}
	if other.cr == constructive.One() {
		return New(u.cr, u.rr.Multiply(other.rr))
	}

	if u.IsZero() || other.IsZero() {
		return New(constructive.One(), rational.Zero())
	}

	return New(constructive.Multiply(u.cr, other.cr), u.rr.Multiply(other.rr))
}

// Divide divides the current number by another number, returning a new
// Real number.
func (u *Real) Divide(other *Real) *Real {
	return u.Multiply(other.Inverse())
}

// ShiftLeft shifts the number to the left by the specified number of bits,
// which is equivalent to multiplying the number by 2^n.
func (u *Real) ShiftLeft(n int) *Real {
	return New(u.cr, u.rr.ShiftLeft(n))
}

// ShiftRight shifts the number to the right by the specified number of bits,
// which is equivalent to dividing the number by 2^n.
func (u *Real) ShiftRight(n int) *Real {
	return New(u.cr, u.rr.ShiftRight(n))
}

// Negate returns the negation of the current number as a new Real number.
func (u *Real) Negate() *Real {
	return New(u.cr, u.rr.Negate())
}

// Inverse returns the multiplicative inverse of the current number as a new
// Real number.
func (u *Real) Inverse() *Real {
	return New(constructive.Inverse(u.cr), u.rr.Inverse())
}

// IsZero returns true if the current number is zero. In order for the number
// to be zero, the rational component must be zero. The constructive component
// cannot be used to determine if the number is zero, since constructive reals
// can only approximate zero at a specific precision (unless it's the zero object).
func (u *Real) IsZero() bool {
	return u.rr.IsZero()
}

// FormattedString returns a string representation of the unified real number
// with the specified number of decimal digits and radix.
func (u *Real) FormattedString(decimalDigits, radix int) string {
	return constructive.Text(u.Constructive(), decimalDigits, radix)
}

var _ fmt.Formatter = (*Real)(nil)

// Format implements the fmt.Formatter interface for custom formatting.
func (u *Real) Format(f fmt.State, c rune) {
	switch c {
	case 'f':
		precision, ok := f.Precision()
		if ok {
			fmt.Fprint(f, u.FormattedString(precision, 10))
			return
		}

	case 's', 'q':
		if u.cr == constructive.One() {
			fmt.Fprint(f, u.rr.String())
			return
		}

	default:
	}

	fmt.Fprint(f, u.FormattedString(30, 10))
}
