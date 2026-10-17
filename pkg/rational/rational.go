package rational

import (
	"context"
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

// ShiftLeft shifts the rational number left by n bits (multiplies by 2^n).
// Supports signed shifts: negative n shifts right.
func (r *Number) ShiftLeft(n int) *Number {
	return r.shift(n)
}

// ShiftRight shifts the rational number right by n bits (divides by 2^n).
// Supports signed shifts: negative n shifts left.
func (r *Number) ShiftRight(n int) *Number {
	return r.shift(-n)
}

// shift shifts the rational number by n bits. Positive n shifts left, negative n shifts right.
func (r *Number) shift(n int) *Number {
	if n == 0 {
		return r
	}

	num := new(big.Int).Set(r.r.Num())
	denom := new(big.Int).Set(r.r.Denom())
	if n > 0 {
		num.Lsh(num, uint(n))
	} else {
		denom.Lsh(denom, uint(-n))
	}

	return &Number{
		r: new(big.Rat).SetFrac(num, denom),
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

// Pow returns r raised to the integer power e. A negative e returns the reciprocal of the power,
// and a zero base with a negative e returns constructive.ErrNonPositive. The power of zero to zero
// is one.
//
// The numerator and denominator of a normalized fraction are coprime, so their powers are coprime
// too and the result is built without normalizing it again.
//
// The precision limit carried by ctx bounds the size of the result. When the estimated result, e
// times the larger bit length of the numerator and the denominator, exceeds the limit, Pow returns
// constructive.ErrPrecisionOverflow before computing anything. A context with no limit places no
// bound on the size. The bases zero, one, and negative one are exact for any e and are never
// bounded. Pow checks ctx before and between squaring steps and returns its error as it is.
func (r *Number) Pow(ctx context.Context, e *big.Int) (*Number, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	exp := new(big.Int).Abs(e)
	invert := e.Sign() < 0
	if r.IsZero() {
		switch {
		case e.Sign() == 0:
			return One(), nil
		case invert:
			return nil, constructive.ErrNonPositive
		default:
			return Zero(), nil
		}
	}

	num, den := new(big.Int).Abs(r.r.Num()), r.r.Denom()
	negative := r.r.Sign() < 0 && exp.Bit(0) == 1
	if num.BitLen() == 1 && den.BitLen() == 1 {
		if negative {
			return New64(-1, 1), nil
		}
		return One(), nil
	}

	if limit, ok := constructive.PrecisionLimit(ctx); ok && limit >= 0 {
		bits := max(num.BitLen(), den.BitLen())
		estimate := new(big.Int).Mul(exp, big.NewInt(int64(bits)))
		if estimate.Cmp(big.NewInt(int64(limit))) > 0 {
			return nil, constructive.ErrPrecisionOverflow
		}
	}

	powNum, powDen := big.NewInt(1), big.NewInt(1)
	base, baseDen := new(big.Int).Set(num), new(big.Int).Set(den)
	for i, n := 0, exp.BitLen(); i < n; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if exp.Bit(i) == 1 {
			powNum.Mul(powNum, base)
			powDen.Mul(powDen, baseDen)
		}
		if i < n-1 {
			base.Mul(base, base)
			baseDen.Mul(baseDen, baseDen)
		}
	}

	if invert {
		powNum, powDen = powDen, powNum
	}
	if negative {
		powNum.Neg(powNum)
	}

	// SetInt64 initializes the denominator, so Denom returns a reference to it rather than a copy.
	out := new(big.Rat).SetInt64(1)
	out.Num().Set(powNum)
	out.Denom().Set(powDen)
	return &Number{r: out}, nil
}

// Num returns a copy of the numerator of the rational number in lowest terms.
// For an integer, it is the value itself.
func (r *Number) Num() *big.Int {
	return new(big.Int).Set(r.r.Num())
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

// IsInteger returns whether the rational number is an integer.
func (r *Number) IsInteger() bool {
	return r.r.IsInt()
}

// Cmp compares two rational numbers: -1 if r < other, 0 if r == other, 1 if r > other.
func (r *Number) Cmp(other *Number) int {
	return r.r.Cmp(other.r)
}

// floorRem returns the floor of r and the remainder num - floor*den, where den
// is the positive denominator. The remainder satisfies 0 <= rem < den.
func (r *Number) floorRem() (f, rem, den *big.Int) {
	num, den := r.r.Num(), r.r.Denom()
	f = new(big.Int).Div(num, den)
	rem = new(big.Int).Sub(num, new(big.Int).Mul(f, den))
	return f, rem, den
}

// fromInt builds an integer-valued rational from a big.Int.
func fromInt(i *big.Int) *Number {
	return &Number{r: new(big.Rat).SetInt(i)}
}

// Floor returns the greatest integer less than or equal to r.
func (r *Number) Floor() *Number {
	f, _, _ := r.floorRem()
	return fromInt(f)
}

// Ceil returns the least integer greater than or equal to r.
func (r *Number) Ceil() *Number {
	f, rem, _ := r.floorRem()
	if rem.Sign() != 0 {
		f = new(big.Int).Add(f, big.NewInt(1))
	}
	return fromInt(f)
}

// Round returns the nearest integer to r, rounding half away from zero.
func (r *Number) Round() *Number {
	f, rem, den := r.floorRem()
	switch new(big.Int).Lsh(rem, 1).Cmp(den) {
	case -1:
		return fromInt(f)
	case 1:
		return fromInt(new(big.Int).Add(f, big.NewInt(1)))
	default:
		if r.r.Sign() < 0 {
			return fromInt(f)
		}
		return fromInt(new(big.Int).Add(f, big.NewInt(1)))
	}
}

// RoundToEven returns the nearest integer to r, rounding ties to even.
func (r *Number) RoundToEven() *Number {
	f, rem, den := r.floorRem()
	switch new(big.Int).Lsh(rem, 1).Cmp(den) {
	case -1:
		return fromInt(f)
	case 1:
		return fromInt(new(big.Int).Add(f, big.NewInt(1)))
	default:
		if f.Bit(0) == 0 {
			return fromInt(f)
		}
		return fromInt(new(big.Int).Add(f, big.NewInt(1)))
	}
}

// String returns the string representation of the rational number. If the
// denominator is 1, it returns just the numerator. Otherwise, it returns
// "numerator/denominator".
func (r *Number) String() string {
	return r.r.RatString()
}
