package unified

import (
	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// Floor returns the greatest integer less than or equal to u. A purely rational
// value is floored exactly; otherwise the boundary is decided at precision p.
func (u *Real) Floor(p int) *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Floor())
	}
	return New(constructive.Floor(u.Constructive(), p), rational.One())
}

// Ceil returns the least integer greater than or equal to u. A purely rational
// value is ceiled exactly; otherwise the boundary is decided at precision p.
func (u *Real) Ceil(p int) *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Ceil())
	}
	return New(constructive.Ceil(u.Constructive(), p), rational.One())
}

// Round returns the nearest integer to u, rounding half away from zero. A purely
// rational value is rounded exactly; otherwise the boundary is decided at
// precision p, where an exact halfway tie is not finitely decidable.
func (u *Real) Round(p int) *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.Round())
	}
	return New(constructive.Round(u.Constructive(), p), rational.One())
}

// RoundToEven returns the nearest integer to u, rounding ties to even. A purely
// rational value is rounded exactly; otherwise the boundary is decided at
// precision p, where an exact halfway tie is not finitely decidable.
func (u *Real) RoundToEven(p int) *Real {
	if u.cr == constructive.One() {
		return New(constructive.One(), u.rr.RoundToEven())
	}
	return New(constructive.RoundToEven(u.Constructive(), p), rational.One())
}

// Min returns the lesser of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at precision p with PreciseCmp, so
// operands equal within p may return either operand.
func (u *Real) Min(other *Real, p int) *Real {
	if u.cr == constructive.One() && other.cr == constructive.One() {
		if u.rr.Cmp(other.rr) <= 0 {
			return u
		}
		return other
	}
	if constructive.PreciseCmp(u.Constructive(), other.Constructive(), p) <= 0 {
		return u
	}
	return other
}

// Max returns the greater of u and other. When both are purely rational the
// comparison is exact; otherwise it is decided at precision p with PreciseCmp, so
// operands equal within p may return either operand.
func (u *Real) Max(other *Real, p int) *Real {
	if u.cr == constructive.One() && other.cr == constructive.One() {
		if u.rr.Cmp(other.rr) >= 0 {
			return u
		}
		return other
	}
	if constructive.PreciseCmp(u.Constructive(), other.Constructive(), p) >= 0 {
		return u
	}
	return other
}
