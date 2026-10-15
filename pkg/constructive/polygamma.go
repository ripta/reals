package constructive

import (
	"context"
	"fmt"
	"math"
	"math/big"
)

const (
	// polygammaGuard is the number of extra bits the series is evaluated with. The truncation
	// error stays below 2^(p−guard−1), since the term search keeps one bit of slack, and the
	// node tree's own rounding stays within 2^(p−guard). Together they leave the sum within 2^p
	// after scaling.
	polygammaGuard = 8

	// polygammaMargin is the number of extra bits of decay the upward shift aims for beyond the
	// requested precision. The series' smallest term falls like (2π)ⁿ·e^(−2πy), so the shift
	// only needs 2πy ≈ bits·ln2 + n·ln(2π). The margin keeps the truncation search from running
	// to the series' turning point.
	polygammaMargin = 8

	// polygammaFloatSlack is the relative slack taken off the floating-point estimate of the
	// argument, so the lower bound used for the truncation search stays below the true value
	// despite the estimate's rounding.
	polygammaFloatSlack = 0x1p-20

	// polygammaInfEstimate stands in for an estimate that overflowed. floatEstimate returns
	// +Inf only when the argument is at least about 2^964, so 2^960 is a lower bound for it.
	// Finite estimates need no cap: they are at most about 2^964.
	polygammaInfEstimate = 0x1p960

	// polygammaMaxShift is the largest upward shift the plan will take. A shift needs one term
	// per unit, so a larger one cannot be built, and it is also where converting the target to
	// an int stops being safe. No order or precision that reaches it can be requested within
	// the valid precision range, so the plan fails with ErrInvalidPrecision, the error for a
	// request beyond what the integer width represents.
	polygammaMaxShift = 1 << 62

	// polygammaMaxExpected clamps the cache-sizing estimate so converting it to an int is
	// always in range.
	polygammaMaxExpected = 1 << 30

	// polygammaCheckMask spaces the context checks of a cheap loop whose length is the order: a
	// check every polygammaCheckMask+1 steps.
	polygammaCheckMask = 1<<12 - 1
)

// Digamma computes the digamma function ψ(c) = Γ'(c)/Γ(c) for real arguments. It is Polygamma
// of order zero.
//
// The engine shifts the argument upward by a precision-dependent amount, using
// ψ(x) = ψ(x+m) − Σ_{j<m} 1/(x+j), and then sums the Bernoulli asymptotic expansion of ψ at
// x+m, truncated by its error bound. Arguments below 1/2 are folded into the convergent region
// with the reflection formula ψ(1−x) − ψ(x) = π·cot(πx).
//
// ψ has poles at the non-positive integers. This layer does not guard them. At a pole the
// reflection divides by a zero sine, and the inverse refines until it ends in ErrUndecided or
// the context's error. Callers reject exact poles upstream.
//
// The reflection is chosen on the first approximation, not here.
func Digamma(c Real) Real {
	reduce := func(ctx context.Context, c Real, rough *big.Int) (Real, error) {
		return reducePolygamma(ctx, 0, c, rough)
	}

	return newDeferredReduction("Digamma", c, -2, reduce)
}

// Polygamma computes the polygamma function ψ⁽ⁿ⁾(c), the nth derivative of the digamma
// function, for an order n ≥ 0 and a real argument. Order zero is Digamma.
//
// The engine is the same as Digamma's. It shifts the argument upward with
// ψ⁽ⁿ⁾(x) = ψ⁽ⁿ⁾(x+m) − (−1)ⁿ·n!·Σ_{j<m} 1/(x+j)ⁿ⁺¹ and sums the Bernoulli asymptotic
// expansion at x+m. Arguments below 1/2 are folded into the convergent region with the
// reflection formula ψ⁽ⁿ⁾(x) = (−1)ⁿ·ψ⁽ⁿ⁾(1−x) − πⁿ⁺¹·Pₙ(cot πx), where Pₙ is the integer
// polynomial of cotangentPolynomial.
//
// A negative order has no meaning. Polygamma returns a node that fails with ErrNegativeOrder
// when it is approximated, without doing any work here.
//
// ψ⁽ⁿ⁾ has poles at the non-positive integers, which this layer does not guard, as in Digamma.
// The reflection is chosen on the first approximation, not here.
//
// The order is not capped. The cost grows with n, since the series needs n!-sized coefficients,
// a shift of about n/(2π) units, and, below 1/2, a polynomial of degree n+1, so callers are
// expected to keep the order small. A very large order is slow and does not fail, but every
// loop that scales with it checks the context, so a cancelled context ends the work promptly.
func Polygamma(n int, c Real) Real {
	switch {
	case n < 0:
		return &negativeOrderPolygamma{n: n, r: c}
	case n == 0:
		return Digamma(c)
	}

	reduce := func(ctx context.Context, c Real, rough *big.Int) (Real, error) {
		return reducePolygamma(ctx, n, c, rough)
	}

	return newDeferredReductionLead("Polygamma", fmt.Sprintf("%d, ", n), c, -2, reduce)
}

// negativeOrderPolygamma is the polygamma function of a negative order. It fails every
// approximation, so the failure is never cached.
type negativeOrderPolygamma struct {
	precisionTracker
	n int
	r Real
}

func (c *negativeOrderPolygamma) approximate(ctx context.Context, _ int) (*big.Int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return nil, ErrNegativeOrder
}

func (c *negativeOrderPolygamma) asConstruction() string {
	return fmt.Sprintf("Polygamma(%d, %s)", c.n, c.r.asConstruction())
}

// reducePolygamma picks the range reduction of ψ⁽ⁿ⁾(c) from its rough approximation. Building
// the reflection takes work that grows with n, so it follows the context.
func reducePolygamma(ctx context.Context, n int, c Real, rough *big.Int) (Real, error) {
	if rough.Cmp(big.NewInt(2)) >= 0 {
		return newPolygammaSeries(n, c), nil
	}

	// x < 1/2: reflect into the convergent region, ψ⁽ⁿ⁾(x) = (−1)ⁿ·ψ⁽ⁿ⁾(1−x) − πⁿ⁺¹·Pₙ(cot πx).
	px := Multiply(Pi(), c)
	cot := Divide(Cosine(px), Sine(px))

	poly, err := cotangentDerivative(ctx, n, cot)
	if err != nil {
		return nil, err
	}

	reflected := newPolygammaSeries(n, Subtract(One(), c))
	if n%2 == 1 {
		reflected = Negate(reflected)
	}

	return Subtract(reflected, Multiply(powInt(Pi(), uint(n)+1), poly)), nil
}

// powInt builds c^k for k ≥ 1 by repeated squaring. The exponent is unsigned so that an
// exponent of n+1 for the largest int order does not wrap, and the recursion is at most 64
// deep.
func powInt(c Real, k uint) Real {
	switch {
	case k <= 1:
		return c
	case k%2 == 0:
		return Square(powInt(c, k/2))
	}

	return Multiply(c, Square(powInt(c, k/2)))
}

// cotangentPolynomial returns the coefficients of Pₙ, lowest degree first, or the context's
// error. Pₙ is the chain-rule form of the derivative in the reflection formula, DLMF 5.15.6:
// dⁿ/dzⁿ cot(πz) = πⁿ·Pₙ(cot πz), found from d/dz cot(πz) = −π(1 + cot²(πz)) by
//
//	P₀(t) = t,  Pₖ₊₁(t) = −(1 + t²)·Pₖ'(t).
//
// Pₙ has integer coefficients, degree n+1, and the parity of n+1. It is computed on demand and
// not cached, which keeps it free of shared state. The work is O(n²) big-integer operations on
// numbers that grow with n, so it checks the context at every step.
func cotangentPolynomial(ctx context.Context, n int) ([]*big.Int, error) {
	p := []*big.Int{big.NewInt(0), big.NewInt(1)}
	for k := 0; k < n; k++ {
		d := make([]*big.Int, len(p)-1)
		for i := range d {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			d[i] = new(big.Int).Mul(p[i+1], big.NewInt(int64(i+1)))
		}

		next := make([]*big.Int, len(d)+2)
		for i := range next {
			next[i] = new(big.Int)
		}

		for i, v := range d {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			next[i].Sub(next[i], v)
			next[i+2].Sub(next[i+2], v)
		}

		p = next
	}

	return p, nil
}

// cotangentDerivative builds Pₙ(t) as a tree, by Horner's rule in t². Pₙ has only the powers
// of t with the parity of n+1, so it is t^parity·Q(t²).
func cotangentDerivative(ctx context.Context, n int, t Real) (Real, error) {
	if n == 0 {
		return t, nil
	}

	p, err := cotangentPolynomial(ctx, n)
	if err != nil {
		return nil, err
	}

	parity := (n + 1) % 2

	last := (len(p) - 1 - parity) / 2
	t2 := Square(t)
	acc := FromBigInt(p[parity+2*last])

	for i := last - 1; i >= 0; i-- {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		acc = Add(FromBigInt(p[parity+2*i]), Multiply(t2, acc))
	}

	if parity == 1 {
		return Multiply(t, acc), nil
	}

	return acc, nil
}

// polygammaSeries is the polygamma function of order n of an argument of at least about 1/4.
// A smaller argument is reflected by reducePolygamma before it gets here, so the shift always
// lands on a positive y.
type polygammaSeries struct {
	precisionTracker
	order int
	r     Real
}

func newPolygammaSeries(order int, c Real) Real {
	return &polygammaSeries{
		order: order,
		r:     c,
	}
}

// approximate evaluates the asymptotic expansion at y = x + m. With the series
//
//	S = Σ_{k=1}^{K} c_k · y^(−2k),   c_k = B₂ₖ·(2k+n−1)!/(2k)!,
//
// the digamma function is
//
//	ψ(x) = ln y − 1/(2y) − S − Σ_{j<m} 1/(x+j) + R_K
//
// by DLMF 5.11.2 together with ψ(z+1) = ψ(z) + 1/z, and for n ≥ 1
//
//	ψ⁽ⁿ⁾(x) = (−1)ⁿ⁺¹ · [ y⁻ⁿ·((n−1)! + n!/(2y) + S) + n!·Σ_{j<m} 1/(x+j)ⁿ⁺¹ ] + R_K
//
// by DLMF 5.15.9 together with DLMF 5.15.5. The coefficient c_k is exact for n = 0 as well,
// where it is B₂ₖ/(2k).
//
// The remainder R_K is bounded in magnitude by the first neglected term,
// |c_{K+1}|·y^(−2K−2−n). For n = 0 this is DLMF 5.11(ii). For n ≥ 1 DLMF 5.15 states no bound.
// It follows from the Hurwitz zeta function: ψ⁽ⁿ⁾(y) = (−1)ⁿ⁺¹·n!·ζ(n+1, y) (DLMF 25.11.12),
// and DLMF 25.11.28, in §25.11(vii), writes ζ(s, y) as its Euler–Maclaurin sum through B₂ₖ plus
// the remainder
//
//	(1/Γ(s)) ∫₀^∞ ρ_K(x)·x^(s−1)·e^(−yx) dx,
//	ρ_K(x) = 1/(eˣ−1) − 1/x + 1/2 − Σ_{k=1}^{K} B₂ₖ·x^(2k−1)/(2k)!.
//
// The partial fractions 1/(eˣ−1) − 1/x + 1/2 = Σ_{m≥1} 2x/(x² + 4π²m²), expanded in x²/(4π²m²)
// with the geometric series' remainder, show that for x > 0, ρ_K has the sign of B₂ₖ₊₂ and
// is at most |B₂ₖ₊₂|·x^(2K+1)/(2K+2)! in magnitude. The weight x^(s−1)·e^(−yx)/Γ(s) is
// positive, so the integral is at most the integral of that bound, which is exactly
// |c_{K+1}|·y^(−2K−2−n) after the factor n!. The bound holds at every K, so it does not need
// the terms to shrink.
//
// The shift m and the term count K are chosen per call, so the truncation error falls below
// the requested precision. The coefficients are exact rationals, formed from exact Bernoulli
// numbers and exact factorial ratios. The terms are summed in w = 1/y² by Horner's rule.
func (c *polygammaSeries) approximate(ctx context.Context, p int) (*big.Int, error) {
	est, err := floatEstimate(ctx, c.r)
	if err != nil {
		return nil, err
	}

	m, terms, err := polygammaPlan(ctx, c.order, est, p-polygammaGuard)
	if err != nil {
		return nil, err
	}

	y := c.r
	if m > 0 {
		y = Add(c.r, FromInt(m))
	}

	invY := Inverse(y)

	series, err := hornerPolygamma(ctx, c.order, invY, terms)
	if err != nil {
		return nil, err
	}

	// The shift can be enormous for a huge order, and the loop below follows the context, so
	// the slice grows as the loop goes and does not reserve m entries up front.
	shifts := make([]Real, 0, min(m, 1<<12))
	for j := 0; j < m; j++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		z := c.r
		if j > 0 {
			z = Add(c.r, FromInt(j))
		}

		shifts = append(shifts, powInt(Inverse(z), uint(c.order)+1))
	}

	var result Real
	if c.order == 0 {
		head := Subtract(Ln(y), ShiftRight(invY, 1))
		result = Subtract(Subtract(head, series), sumBalanced(shifts))
	} else {
		result, err = higherOrderPolygamma(ctx, c.order, invY, series, sumBalanced(shifts))
		if err != nil {
			return nil, err
		}
	}

	ir, err := approx(ctx, result, p-polygammaGuard)
	if err != nil {
		return nil, err
	}

	return scale(ir, -polygammaGuard), nil
}

// higherOrderPolygamma assembles ψ⁽ⁿ⁾ for n ≥ 1 from the series sum S, the shift sum
// Σ 1/(x+j)ⁿ⁺¹, and invY = 1/y:
//
//	(−1)ⁿ⁺¹ · [ y⁻ⁿ·((n−1)! + n!/(2y) + S) + n!·Σ ].
func higherOrderPolygamma(ctx context.Context, n int, invY, series, shifts Real) (Real, error) {
	prev := big.NewInt(1) // (n−1)!
	for i := 2; i < n; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		prev.Mul(prev, big.NewInt(int64(i)))
	}

	fact := new(big.Int).Mul(prev, big.NewInt(int64(n))) // n!

	lead := Add(FromBigInt(prev), Add(ShiftRight(Multiply(FromBigInt(fact), invY), 1), series))
	sum := Add(Multiply(powInt(invY, uint(n)), lead), Multiply(FromBigInt(fact), shifts))

	if n%2 == 0 {
		return Negate(sum), nil
	}

	return sum, nil
}

// polygammaPlan picks the upward shift m and the number of series terms K so the truncation
// error falls below 2^tolExp. est is a floating-point estimate of the argument, which must be
// positive up to estimate error. A +Inf estimate, which floatEstimate returns for an argument
// beyond float64 range, is replaced by a lower bound.
//
// The bound on the neglected term is checked against a lower bound of y, so the bound holds for
// the true y. The terms are bounded in floating-point log₂, since the Bernoulli numbers
// overflow float64 long before the series ends.
func polygammaPlan(
	ctx context.Context, order int, est float64, tolExp int,
) (m, terms int, err error) {
	if math.IsInf(est, 1) {
		est = polygammaInfEstimate
	}

	lower := est - polygammaSlack(est)

	// Aim for (2π)ⁿ·e^(−2πy) below the tolerance, and for y ≥ 1 so y stays well away from the
	// singularity at zero.
	bits := float64(-tolExp+polygammaMargin)*math.Ln2 + float64(order)*math.Log(2*math.Pi)

	return polygammaShift(ctx, order, lower, tolExp, math.Max(1, bits/(2*math.Pi)))
}

// polygammaShift is the search loop of polygammaPlan. It shifts a lower bound of the argument
// up to at least target, and then further by a quarter and a unit each time the terms turn
// around before the tolerance. The starting target only affects how many rounds that takes.
func polygammaShift(
	ctx context.Context, order int, lower float64, tolExp int, target float64,
) (m, terms int, err error) {
	for {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}

		need := math.Ceil(target - lower)
		if need > polygammaMaxShift {
			return 0, 0, ErrInvalidPrecision
		}

		m = max(0, int(need))

		var ok bool
		terms, ok, err = polygammaTerms(ctx, order, lower+float64(m), tolExp)
		if err != nil {
			return 0, 0, err
		}

		if ok {
			return m, terms, nil
		}

		// The series turned around before reaching the tolerance: shift further.
		target = target*1.25 + 1
	}
}

func polygammaSlack(est float64) float64 {
	return polygammaFloatSlack * (1 + math.Abs(est))
}

// polygammaTerms returns the smallest K such that the first neglected term of the series at y,
// |c_{K+1}|·y^(−2K−2−n) with c_k = B₂ₖ·(2k+n−1)!/(2k)!, is below 2^tolExp. The boolean is
// false when the terms stop shrinking before they reach the tolerance, which means y is too
// small.
func polygammaTerms(ctx context.Context, order int, y float64, tolExp int) (int, bool, error) {
	logY := math.Log2(y)
	limit := float64(tolExp) - 1 // one bit of slack for floating-point rounding

	// Grow the Bernoulli cache once to the expected size, instead of by small steps.
	expected, err := expectedPolygammaTerms(ctx, order, logY, limit)
	if err != nil {
		return 0, false, err
	}

	if _, err := bernoulliEven(ctx, expected); err != nil {
		return 0, false, err
	}

	prev := math.Inf(1)
	for k := 1; ; k++ {
		if err := ctx.Err(); err != nil {
			return 0, false, err
		}

		b, err := bernoulliEven(ctx, k)
		if err != nil {
			return 0, false, err
		}

		ratio, err := log2FactorialRatio(ctx, order, k)
		if err != nil {
			return 0, false, err
		}

		// The sum is in floating point, since 2k+n can overflow an int.
		term := log2Abs(b.Num()) - log2Abs(b.Denom()) + ratio - (float64(2*k)+float64(order))*logY
		if term < limit {
			return k - 1, true, nil
		}

		if term > prev {
			return 0, false, nil
		}

		prev = term
	}
}

// log2FactorialRatio returns log₂((2k+n−1)!/(2k)!), the factorial ratio of the series
// coefficient. The ratio is 1/(2k) for n = 0 and the product (2k+1)⋯(2k+n−1) for n ≥ 1. The
// loop is as long as the order, so it checks the context every few thousand steps.
func log2FactorialRatio(ctx context.Context, order, k int) (float64, error) {
	if order == 0 {
		return -math.Log2(float64(2 * k)), nil
	}

	sum := 0.0
	for j := 1; j < order; j++ {
		if j&polygammaCheckMask == 0 {
			if err := ctx.Err(); err != nil {
				return 0, err
			}
		}

		sum += math.Log2(float64(2*k) + float64(j))
	}

	return sum, nil
}

// expectedPolygammaTerms estimates the Bernoulli index at which polygammaTerms stops, without
// looking at any Bernoulli number. It uses |B₂ₖ| ≈ 2·(2k)!/(2π)^(2k), which is accurate for all
// but the first few k, so the term is about 2·(2k+n−1)!/((2π)^(2k)·y^(2k+n)). The estimate only
// sizes the cache, so being off costs a second growth and never correctness.
func expectedPolygammaTerms(ctx context.Context, order int, logY, limit float64) (int, error) {
	n := float64(order)
	turn := math.Pi*math.Exp2(logY) - n/2 + 2 // the terms stop shrinking about here
	for k := 1; float64(k) < turn; k++ {
		if err := ctx.Err(); err != nil {
			return 0, err
		}

		lg, _ := math.Lgamma(float64(2*k) + n)
		term := 1 + lg/math.Ln2 - float64(2*k)*math.Log2(2*math.Pi) - (float64(2*k)+n)*logY
		if term < limit {
			return k + 1, nil
		}
	}

	// The clamp keeps the conversion in range. A turning point this far out is never reached
	// by a real request, since the search stops long before it.
	return max(1, int(math.Min(turn, polygammaMaxExpected))), nil
}

// log2Abs returns log₂|i| for a nonzero integer, without overflowing for large i.
func log2Abs(i *big.Int) float64 {
	abs := new(big.Int).Abs(i)

	const top = 62
	n := abs.BitLen()
	if n <= top {
		return math.Log2(float64(abs.Int64()))
	}

	return math.Log2(float64(abs.Rsh(abs, uint(n-top)).Int64())) + float64(n-top)
}

// hornerPolygamma builds S = Σ_{k=1}^{K} c_k · y^(−2k) as a tree, summed in w = 1/y² by
// Horner's rule: w·(c₁ + w·(c₂ + … + w·c_K)). invY is 1/y.
func hornerPolygamma(ctx context.Context, order int, invY Real, terms int) (Real, error) {
	if terms == 0 {
		return FromInt(0), nil
	}

	w := Square(invY)

	acc, err := polygammaCoefficient(ctx, order, terms)
	if err != nil {
		return nil, err
	}

	for k := terms - 1; k >= 1; k-- {
		coef, err := polygammaCoefficient(ctx, order, k)
		if err != nil {
			return nil, err
		}

		acc = Add(coef, Multiply(w, acc))
	}

	return Multiply(w, acc), nil
}

// polygammaCoefficient is the series coefficient c_k = B₂ₖ·(2k+n−1)!/(2k)!, as an exact Real.
// The factorial ratio is 1/(2k) for n = 0, which makes c_k = B₂ₖ/(2k), and the product
// (2k+1)⋯(2k+n−1) for n ≥ 1.
func polygammaCoefficient(ctx context.Context, order, k int) (Real, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	b, err := bernoulliEven(ctx, k)
	if err != nil {
		return nil, err
	}

	coef := new(big.Rat).Set(b)
	if order == 0 {
		coef.Quo(coef, new(big.Rat).SetInt64(int64(2*k)))
	} else {
		// The factors 2k+j are formed unsigned, so that the sum cannot wrap for any order.
		ratio := big.NewInt(1)
		factor := new(big.Int)
		for j := 1; j < order; j++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}

			ratio.Mul(ratio, factor.SetUint64(uint64(2*k)+uint64(j)))
		}

		coef.Mul(coef, new(big.Rat).SetInt(ratio))
	}

	num := FromBigInt(new(big.Int).Set(coef.Num()))
	if coef.IsInt() {
		return num, nil
	}

	return Divide(num, FromBigInt(new(big.Int).Set(coef.Denom()))), nil
}

// sumBalanced adds the terms pairwise. A left fold would nest one addition per term, and each
// addition asks its operands for two more bits.
func sumBalanced(terms []Real) Real {
	switch len(terms) {
	case 0:
		return FromInt(0)
	case 1:
		return terms[0]
	}

	mid := len(terms) / 2

	return Add(sumBalanced(terms[:mid]), sumBalanced(terms[mid:]))
}

func (c *polygammaSeries) asConstruction() string {
	if c.order == 0 {
		return fmt.Sprintf("Digamma(%s)", c.r.asConstruction())
	}

	return fmt.Sprintf("Polygamma(%d, %s)", c.order, c.r.asConstruction())
}
