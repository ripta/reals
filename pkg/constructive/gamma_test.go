package constructive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type gammaTest struct {
	name     string
	input    Real
	expected Real
}

// sqrtPi is √π, the value of Γ(1/2).
func sqrtPi() Real {
	return Sqrt(Pi())
}

var gammaTests = []gammaTest{
	{name: "Gamma(1)=1", input: FromInt(1), expected: FromInt(1)},
	{name: "Gamma(2)=1", input: FromInt(2), expected: FromInt(1)},
	{name: "Gamma(3)=2", input: FromInt(3), expected: FromInt(2)},
	{name: "Gamma(4)=6", input: FromInt(4), expected: FromInt(6)},
	{name: "Gamma(5)=24", input: FromInt(5), expected: FromInt(24)},
	{name: "Gamma(1/2)=sqrt(pi)", input: FromRat(1, 2), expected: sqrtPi()},
	{name: "Gamma(3/2)=sqrt(pi)/2", input: FromRat(3, 2), expected: Divide(sqrtPi(), Two())},
	{name: "Gamma(-1/2)=-2sqrt(pi)", input: FromRat(-1, 2), expected: Negate(Multiply(Two(), sqrtPi()))},
	{name: "Gamma(-3/2)=4sqrt(pi)/3", input: FromRat(-3, 2), expected: Multiply(FromRat(4, 3), sqrtPi())},
}

func TestGamma(t *testing.T) {
	for _, test := range gammaTests {
		t.Run(test.name, func(t *testing.T) {
			result := Gamma(test.input)
			assertEqualAtPrecision(t, test.expected, result, -60)
			assertEqualAtPrecision(t, test.expected, result, -120)
		})
	}
}

// TestGammaRecurrence verifies Γ(x+1) = x·Γ(x) at an irrational argument, so the
// check rests on the refinement contract rather than a hand-picked constant.
func TestGammaRecurrence(t *testing.T) {
	x := Sqrt2()
	lhs := Gamma(Add(x, One()))
	rhs := Multiply(x, Gamma(x))
	assertEqualAtPrecision(t, lhs, rhs, -120)
}

func TestGammaAsConstruction(t *testing.T) {
	assert.Equal(t, "Gamma(Int(3))", Gamma(FromInt(3)).asConstruction())
}
