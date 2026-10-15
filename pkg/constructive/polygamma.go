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
	// requested precision. The series' smallest term falls like e^(−2πy), so the shift only
	// needs y ≈ bits·ln2/(2π). The margin keeps the truncation search from running to the
	// series' turning point.
	polygammaMargin = 8

	// polygammaFloatSlack is the relative slack taken off the floating-point estimate of the
	// argument, so the lower bound used for the truncation search stays below the true value
	// despite the estimate's rounding.
	polygammaFloatSlack = 0x1p-20

	// polygammaInfEstimate stands in for an estimate that overflowed. floatEstimate returns
	// +Inf only when the argument is at least about 2^964, so 2^960 is a lower bound for it.
	// Finite estimates need no cap: they are at most about 2^964.
	polygammaInfEstimate = 0x1p960
)

// Digamma computes the digamma function ψ(c) = Γ'(c)/Γ(c) for real arguments.
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
	return newDeferredReduction("Digamma", c, -2, reduceDigamma)
}

func reduceDigamma(c Real, rough *big.Int) (Real, error) {
	if rough.Cmp(big.NewInt(2)) < 0 {
		// x < 1/2: reflect into the convergent region, ψ(x) = ψ(1−x) − π·cot(πx).
		px := Multiply(Pi(), c)
		cot := Divide(Cosine(px), Sine(px))

		return Subtract(newPolygammaSeries(Subtract(One(), c)), Multiply(Pi(), cot)), nil
	}

	return newPolygammaSeries(c), nil
}

// polygammaSeries is the digamma function of an argument of at least about 1/4. A smaller
// argument is reflected by reduceDigamma before it gets here, so the shift always lands on a
// positive y.
type polygammaSeries struct {
	precisionTracker
	r Real
}

func newPolygammaSeries(c Real) Real {
	return &polygammaSeries{
		r: c,
	}
}

// approximate evaluates the asymptotic expansion at y = x + m:
//
//	ψ(x) = ln y − 1/(2y) − Σ_{k=1}^{K} B₂ₖ / (2k·y^(2k)) − Σ_{j<m} 1/(x+j) + R_K
//
// This is DLMF 5.11.2 together with the recurrence ψ(z+1) = ψ(z) + 1/z. For real positive y,
// DLMF 5.11(ii) bounds the remainder R_K by the first neglected term, |B₂ₖ₊₂| / ((2K+2)·y^(2K+2)),
// and gives it that term's sign.
//
// The shift m and the term count K are chosen per call, so the truncation error falls below
// the requested precision. The terms are summed in w = 1/y² by Horner's rule.
func (c *polygammaSeries) approximate(ctx context.Context, p int) (*big.Int, error) {
	est, err := floatEstimate(ctx, c.r)
	if err != nil {
		return nil, err
	}

	m, terms, err := polygammaPlan(ctx, est, p-polygammaGuard)
	if err != nil {
		return nil, err
	}

	y := c.r
	if m > 0 {
		y = Add(c.r, FromInt(m))
	}

	series, err := hornerBernoulli(ctx, y, terms)
	if err != nil {
		return nil, err
	}

	shifts := make([]Real, 0, m)
	for j := 0; j < m; j++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if j == 0 {
			shifts = append(shifts, Inverse(c.r))
			continue
		}

		shifts = append(shifts, Inverse(Add(c.r, FromInt(j))))
	}

	head := Subtract(Ln(y), ShiftRight(Inverse(y), 1))
	result := Subtract(Subtract(head, series), sumBalanced(shifts))

	ir, err := approx(ctx, result, p-polygammaGuard)
	if err != nil {
		return nil, err
	}

	return scale(ir, -polygammaGuard), nil
}

// polygammaPlan picks the upward shift m and the number of series terms K so the truncation
// error falls below 2^tolExp. est is a floating-point estimate of the argument, which must be
// positive up to estimate error. A +Inf estimate, which floatEstimate returns for an argument
// beyond float64 range, is replaced by a lower bound.
//
// The bound on the neglected term is checked against a lower bound of y, so the bound holds for
// the true y. The terms are bounded in floating-point log₂, since the Bernoulli numbers
// overflow float64 long before the series ends.
func polygammaPlan(ctx context.Context, est float64, tolExp int) (m, terms int, err error) {
	if math.IsInf(est, 1) {
		est = polygammaInfEstimate
	}

	lower := est - polygammaSlack(est)

	// Aim for e^(−2πy) below the tolerance, and for y ≥ 1 so y stays well away from the
	// singularity at zero.
	target := math.Max(1, float64(-tolExp+polygammaMargin)*math.Ln2/(2*math.Pi))
	for {
		if err := ctx.Err(); err != nil {
			return 0, 0, err
		}

		m = max(0, int(math.Ceil(target-lower)))

		var ok bool
		terms, ok, err = polygammaTerms(ctx, lower+float64(m), tolExp)
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
// |B₂ₖ₊₂| / ((2K+2)·y^(2K+2)), is below 2^tolExp. The boolean is false when the terms stop
// shrinking before they reach the tolerance, which means y is too small.
func polygammaTerms(ctx context.Context, y float64, tolExp int) (int, bool, error) {
	logY := math.Log2(y)
	limit := float64(tolExp) - 1 // one bit of slack for floating-point rounding

	// Grow the Bernoulli cache once to the expected size, instead of by small steps.
	if _, err := bernoulliEven(ctx, expectedPolygammaTerms(logY, limit)); err != nil {
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

		term := log2Abs(b.Num()) - log2Abs(b.Denom()) - math.Log2(float64(2*k)) - float64(2*k)*logY
		if term < limit {
			return k - 1, true, nil
		}

		if term > prev {
			return 0, false, nil
		}

		prev = term
	}
}

// expectedPolygammaTerms estimates the Bernoulli index at which polygammaTerms stops, without
// looking at any Bernoulli number. It uses |B₂ₖ| ≈ 2·(2k)!/(2π)^(2k), which is accurate for all
// but the first few k. The estimate only sizes the cache, so being off costs a second growth
// and never correctness.
func expectedPolygammaTerms(logY, limit float64) int {
	turn := math.Pi*math.Exp2(logY) + 2 // the terms stop shrinking about here
	for k := 1; float64(k) < turn; k++ {
		lg, _ := math.Lgamma(float64(2*k) + 1)
		term := 1 + lg/math.Ln2 - float64(2*k)*math.Log2(2*math.Pi) - math.Log2(float64(2*k)) -
			float64(2*k)*logY
		if term < limit {
			return k + 1
		}
	}

	return int(turn)
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

// hornerBernoulli builds Σ_{k=1}^{K} B₂ₖ/(2k) · y^(−2k) as a tree, summed in w = 1/y² by
// Horner's rule: w·(c₁ + w·(c₂ + … + w·c_K)).
func hornerBernoulli(ctx context.Context, y Real, terms int) (Real, error) {
	if terms == 0 {
		return FromInt(0), nil
	}

	w := Square(Inverse(y))

	acc, err := bernoulliTerm(ctx, terms)
	if err != nil {
		return nil, err
	}

	for k := terms - 1; k >= 1; k-- {
		coef, err := bernoulliTerm(ctx, k)
		if err != nil {
			return nil, err
		}

		acc = Add(coef, Multiply(w, acc))
	}

	return Multiply(w, acc), nil
}

// bernoulliTerm is the coefficient B₂ₖ/(2k) of the series, as an exact Real.
func bernoulliTerm(ctx context.Context, k int) (Real, error) {
	b, err := bernoulliEven(ctx, k)
	if err != nil {
		return nil, err
	}

	coef := new(big.Rat).Quo(b, new(big.Rat).SetInt64(int64(2*k)))

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
	return fmt.Sprintf("Digamma(%s)", c.r.asConstruction())
}
