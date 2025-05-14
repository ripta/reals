package rational

import (
	"math/big"

	"github.com/ripta/reals/pkg/constructive"
)

// Number represents a rational number.
type Number struct {
	r *big.Rat
}

// New creates a new rational number from int numerator and denominator.
func New(a, b *big.Int) *Number {
	return &Number{
		r: new(big.Rat).SetFrac(a, b),
	}
}

// New64 creates a new rational number from int64 numerator and denominator.
func New64(a, b int64) *Number {
	return &Number{
		r: new(big.Rat).SetFrac64(a, b),
	}
}

// FromRational creates a new rational number from a big.Rat.
func FromRational(r *big.Rat) *Number {
	return &Number{
		r: new(big.Rat).Set(r),
	}
}

// Constructive converts the rational number to a constructive real.
func (r *Number) Constructive() constructive.Real {
	return constructive.Divide(constructive.FromBigInt(r.r.Num()), constructive.FromBigInt(r.r.Denom()))
}

// Add adds two rational numbers.
func (r *Number) Add(other *Number) *Number {
	return &Number{
		r: new(big.Rat).Add(r.r, other.r),
	}
}

// Subtract subtracts two rational numbers.
func (r *Number) Subtract(other *Number) *Number {
	return &Number{
		r: new(big.Rat).Sub(r.r, other.r),
	}
}

// Multiply multiplies two rational numbers.
func (r *Number) Multiply(other *Number) *Number {
	return &Number{
		r: new(big.Rat).Mul(r.r, other.r),
	}
}

// Divide divides two rational numbers.
func (r *Number) Divide(other *Number) *Number {
	return &Number{
		r: new(big.Rat).Quo(r.r, other.r),
	}
}

// Negate negates the rational number.
func (r *Number) Negate() *Number {
	return &Number{
		r: new(big.Rat).Neg(r.r),
	}
}

// Inverse returns the multiplicative inverse of the rational number.
func (r *Number) Inverse() *Number {
	if r.r.Num().Sign() == 0 {
		return nil
	}
	return &Number{
		r: new(big.Rat).Inv(r.r),
	}
}

// Sign returns the sign of the rational number: -1 for negative, 0 for zero,
// 1 for positive.
func (r *Number) Sign() int {
	return r.r.Sign()
}

// IsZero checks if the rational number is zero.
func (r *Number) IsZero() bool {
	return r.r.Sign() == 0
}

// Cmp compares two rational numbers: -1 if r < other, 0 if r == other, 1 if r > other.
func (r *Number) Cmp(other *Number) int {
	return r.r.Cmp(other.r)
}
