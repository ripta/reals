package constructive

import (
	"fmt"
	"math/big"
)

// arithmeticGeometricMean computes the Arithmetic-Geometric Mean (AGM) of two
// Real numbers using the iterative algorithm:
//
//	a₀ = a, b₀ = b
//	aₙ₊₁ = (aₙ + bₙ) / 2
//	bₙ₊₁ = √(aₙ × bₙ)
//
// The sequences converge quadratically to the same value, which is AGM(a, b).
type arithmeticGeometricMean struct {
	precisionTracker
	a Real
	b Real
}

// newArithmeticGeometricMean creates a new AGM computation for the given
// Real numbers a and b.
func newArithmeticGeometricMean(a, b Real) Real {
	return &arithmeticGeometricMean{
		a: a,
		b: b,
	}
}

func (c *arithmeticGeometricMean) approximate(p int) *big.Int {
	if p >= 1 {
		return big.NewInt(0)
	}

	// quadratic convergence, so we need roughly log2(-p) iterations, plus some
	// safety buffer
	maxIters := boundLog2(-p) + 15
	calcPrec := p - 20

	aReal := c.a
	bReal := c.b
	for i := 0; i < maxIters; i++ {
		aApprox := Approximate(aReal, calcPrec)
		bApprox := Approximate(bReal, calcPrec)

		// Check convergence: |aₙ - bₙ| < threshold
		// at calcPrec, a difference of 1 represents 2^calcPrec,
		diff := bigAbs(bigSub(aApprox, bApprox))
		threshold := big.NewInt(4)
		if diff.Cmp(threshold) < 0 {
			break
		}

		// aₙ₊₁ = (aₙ + bₙ) / 2
		// bₙ₊₁ = √(aₙ × bₙ)
		aNext := Divide(Add(aReal, bReal), FromInt(2))
		bNext := Sqrt(Multiply(aReal, bReal))

		aReal, bReal = aNext, bNext
	}

	return Approximate(aReal, p)
}

func (c *arithmeticGeometricMean) asConstruction() string {
	return fmt.Sprintf("AGM(%s, %s)", c.a.asConstruction(), c.b.asConstruction())
}
