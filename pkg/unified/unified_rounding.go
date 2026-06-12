package unified

import (
	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// Floor returns the greatest integer less than or equal to u. A purely rational
// value is floored exactly; otherwise the boundary is decided at a fixed
// internal precision.
func (u *Real) Floor() *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Floor())
	}
	return New(constructive.Floor(u.Constructive()), rational.One())
}

// Ceil returns the least integer greater than or equal to u. A purely rational
// value is ceiled exactly; otherwise the boundary is decided at a fixed
// internal precision.
func (u *Real) Ceil() *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Ceil())
	}
	return New(constructive.Ceil(u.Constructive()), rational.One())
}

// Round returns the nearest integer to u, rounding half away from zero. A purely
// rational value is rounded exactly; otherwise the boundary is decided at a
// fixed internal precision, where an exact halfway tie is not finitely decidable.
func (u *Real) Round() *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Round())
	}
	return New(constructive.Round(u.Constructive()), rational.One())
}

// RoundToEven returns the nearest integer to u, rounding ties to even. A purely
// rational value is rounded exactly; otherwise the boundary is decided at a
// fixed internal precision, where an exact halfway tie is not finitely decidable.
func (u *Real) RoundToEven() *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.RoundToEven())
	}
	return New(constructive.RoundToEven(u.Constructive()), rational.One())
}

// Min returns the lesser of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at a bounded precision, so
// operands equal within that precision may return either operand.
func (u *Real) Min(other *Real) *Real {
	if u.cr == constructive.One() && other.cr == constructive.One() {
		if u.rr.Cmp(other.rr) <= 0 {
			return u
		}
		return other
	}
	return New(constructive.Min(u.Constructive(), other.Constructive()), rational.One())
}

// Max returns the greater of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at a bounded precision, so
// operands equal within that precision may return either operand.
func (u *Real) Max(other *Real) *Real {
	if u.cr == constructive.One() && other.cr == constructive.One() {
		if u.rr.Cmp(other.rr) >= 0 {
			return u
		}
		return other
	}
	return New(constructive.Max(u.Constructive(), other.Constructive()), rational.One())
}
