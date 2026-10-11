package constructive

import (
	"context"
	"fmt"
	"math"
	"math/big"
)

// log2TwoPi is log2(2π), the per-term decay rate of Spouge's relative error.
var log2TwoPi = math.Log2(2 * math.Pi)

// Gamma computes the gamma function Γ(c) for real arguments.
//
// The positive-argument engine is Spouge's approximation with a
// precision-dependent parameter, so each approximation refines to within 2^p
// like the rest of the package. Arguments below 1/2 are folded into the
// well-conditioned region with the reflection formula
// Γ(x)·Γ(1−x) = π / sin(πx). Γ is composed entirely from existing primitives
// rather than bespoke series machinery; see (*spougeGamma).approximate.
//
// Γ has poles at the non-positive integers. This layer does not guard them, and
// such an argument drives a non-terminating refinement; callers reject exact
// poles upstream.
func Gamma(c Real) Real {
	rough := roughApproximation(c, -2)
	if rough == nil {
		return nil
	}

	if rough.Cmp(big.NewInt(2)) < 0 {
		// x < 1/2: reflect into the convergent region Γ(1−x).
		reflected := Multiply(Sine(Multiply(Pi(), c)), Gamma(Subtract(One(), c)))
		return Divide(Pi(), reflected)
	}

	return newSpougeGamma(c)
}

type spougeGamma struct {
	precisionTracker
	r Real
}

func newSpougeGamma(c Real) Real {
	return &spougeGamma{
		r: c,
	}
}

// approximate evaluates Spouge's approximation, with z = x − 1:
//
//	Γ(x) = (x+N−1)^(x−1/2) · e^(−(x+N−1)) · S
//	S    = c_0 + Σ_{k=1}^{N−1} c_k / (x + k − 1)
//	c_0  = √(2π)
//	c_k  = (−1)^(k−1) / (k−1)! · (N−k)^(k−1/2) · e^(N−k)
//
// The integer parameter N doubles as Spouge's free parameter a, chosen per call
// so the relative truncation error (2π)^−N drops below the requested precision.
func (c *spougeGamma) approximate(ctx context.Context, p int) (*big.Int, error) {
	n, err := c.spougeParameter(ctx, p)
	if err != nil {
		return nil, err
	}

	sum := Sqrt(Multiply(Two(), Pi())) // c_0 = √(2π)
	factorial := big.NewInt(1)         // (k−1)! carried across iterations
	for k := 1; k < n; k++ {
		// Building each term runs the eager Pow and Exp constructors, which do real
		// work before any approximation reaches the context check.
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if k > 1 {
			factorial = bigMul(factorial, big.NewInt(int64(k-1)))
		}

		coef := Multiply(
			Pow(FromInt(n-k), FromRat(2*k-1, 2)),
			Exp(FromInt(n-k)),
		)
		coef = Divide(coef, FromBigInt(factorial))
		if k%2 == 0 {
			coef = Negate(coef)
		}

		sum = Add(sum, Divide(coef, Add(c.r, FromInt(k-1))))
	}

	w := Add(c.r, FromInt(n-1)) // x + N − 1
	result := Multiply(
		Multiply(Pow(w, Subtract(c.r, FromRat(1, 2))), Exp(Negate(w))),
		sum,
	)

	// The series carries absolute precision through its additions, so the
	// alternating-sign cancellation in S needs no special handling; a handful of
	// guard bits covers rounding in the final product.
	const guard = 4
	ir, err := approx(ctx, result, p-guard)
	if err != nil {
		return nil, err
	}

	return scale(ir, -guard), nil
}

// spougeParameter sizes Spouge's free parameter for precision p. Spouge's
// relative error decays as (2π)^−N, so an absolute error below 2^p needs
// N > (log2|Γ(x)| − p) / log2(2π). The magnitude term is estimated cheaply in
// floating point, in the spirit of Sqrt seeding itself with math.Sqrt;
// overestimating only enlarges N and sharpens the result.
func (c *spougeGamma) spougeParameter(ctx context.Context, p int) (int, error) {
	est, err := floatEstimate(ctx, c.r)
	if err != nil {
		return 0, err
	}

	lg, _ := math.Lgamma(est)
	mag := lg/math.Ln2 + 4

	n := int(math.Ceil((mag-float64(p)+1)/log2TwoPi)) + 5
	if n < 10 {
		n = 10
	}
	return n, nil
}

func (c *spougeGamma) asConstruction() string {
	return fmt.Sprintf("Gamma(%s)", c.r.asConstruction())
}

// floatEstimate returns a float64 approximation of c, used only to seed
// precision heuristics.
func floatEstimate(ctx context.Context, c Real) (float64, error) {
	const fp = -60
	i, err := approx(ctx, c, fp)
	if err != nil {
		return 0, err
	}

	v, _ := i.Float64()
	return math.Ldexp(v, fp), nil
}
