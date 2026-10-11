package constructive

import (
	"fmt"
	"math/big"
	"testing"
)

type multiplicationErrorTest struct {
	Name string

	// Each operand is the dyadic rational N / 2^Shift, so the product is exact.
	AN, BN         int64
	AShift, BShift int
}

// The first operand is far smaller than the second, so its msd is unknown at the
// half precision the node probes first. That is the case where the node must
// swap its operands before choosing either precision.
var multiplicationErrorTests = []multiplicationErrorTest{
	{Name: "tiny-times-one", AN: 1, AShift: 200, BN: 1, BShift: 0},
	{Name: "tiny-times-large", AN: 3, AShift: 150, BN: 1 << 40, BShift: 0},
	{Name: "tiny-times-small", AN: 7, AShift: 300, BN: 5, BShift: 20},
	{Name: "tiny-odd-times-odd", AN: 12345, AShift: 120, BN: 99991, BShift: 3},
	{Name: "negative-tiny", AN: -11, AShift: 180, BN: 13, BShift: 1},
	{Name: "both-tiny", AN: 9, AShift: 90, BN: 17, BShift: 70},
}

type multiplicationOperandTest struct {
	Name string

	// The small operand is 2^-TinyShift, and the large operand is one.
	TinyShift int
	P         int
}

var multiplicationOperandTests = []multiplicationOperandTest{
	{Name: "product-rounds-to-zero", TinyShift: 200, P: -50},
	{Name: "product-nonzero", TinyShift: 80, P: -100},
	{Name: "product-nonzero-fine", TinyShift: 300, P: -1000},
}

// TestMultiplicationLargeOperandStaysCoarse checks that a tiny factor keeps the
// large factor coarse. The large factor is needed only to about p - msd(tiny),
// and not at all once the tiny factor rounds to zero. The finest precision the
// large factor sees is either that bound or the half-precision msd probe.
func TestMultiplicationLargeOperandStaysCoarse(t *testing.T) {
	for _, test := range multiplicationOperandTests {
		for _, tinyFirst := range []bool{false, true} {
			name := fmt.Sprintf("%s/tinyFirst=%t", test.Name, tinyFirst)
			t.Run(name, func(t *testing.T) {
				tiny := ShiftRight(One(), test.TinyShift)
				large := FromInt(1)

				product := Multiply(large, tiny)
				if tinyFirst {
					product = Multiply(tiny, large)
				}

				mustApproximate(t, product, test.P)

				probe := (test.P >> 1) - 2
				bound := min(probe, test.P+test.TinyShift-3) - 1

				got := large.tracker().latest().precision
				if got < bound {
					t.Fatalf("large operand refined to %d, want no finer than %d", got, bound)
				}
			})
		}
	}
}

// TestMultiplicationError checks the approximation contract on products whose
// exact value is known: at precision p, the result is within one unit of
// x / 2^p. Each precision builds a fresh tree so no cache carries over.
func TestMultiplicationError(t *testing.T) {
	for _, test := range multiplicationErrorTests {
		for _, swapped := range []bool{false, true} {
			name := fmt.Sprintf("%s/swapped=%t", test.Name, swapped)
			t.Run(name, func(t *testing.T) {
				exact := new(big.Rat).SetFrac(
					big.NewInt(test.AN*test.BN),
					new(big.Int).Lsh(big.NewInt(1), uint(test.AShift+test.BShift)),
				)

				for p := 10; p >= -700; p-- {
					a := ShiftRight(FromInt64(test.AN), test.AShift)
					b := ShiftRight(FromInt64(test.BN), test.BShift)
					if swapped {
						a, b = b, a
					}

					got := mustApproximate(t, Multiply(a, b), p)

					want := new(big.Rat).Set(exact)
					if p >= 0 {
						want.Quo(want, new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(p))))
					} else {
						want.Mul(want, new(big.Rat).SetInt(new(big.Int).Lsh(big.NewInt(1), uint(-p))))
					}

					diff := new(big.Rat).Sub(new(big.Rat).SetInt(got), want)
					if diff.Abs(diff).Cmp(big.NewRat(1, 1)) >= 0 {
						t.Fatalf("p=%d: got %s, want within 1 of %s", p, got, want.FloatString(4))
					}
				}
			})
		}
	}
}
