package constructive

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

type signumTest struct {
	input    Real
	expected int
}

var signumTests = []signumTest{
	{FromInt64(-100), -1},
	{FromInt64(-10), -1},
	{FromInt64(-3), -1},
	{FromInt64(-2), -1},
	{FromInt64(-1), -1},
	// {FromInt64(0), 0},
	{FromInt64(1), 1},
	{FromInt64(2), 1},
	{FromInt64(3), 1},
	{FromInt64(10), 1},
	{FromInt64(100), 1},
}

func TestSignum(t *testing.T) {
	for _, test := range signumTests {
		if result := Sign(test.input); result != test.expected {
			t.Errorf("expected %d, got %d", test.expected, result)
		}
	}
}

type approximateTest struct {
	input     Real
	expecteds map[int]*big.Int
}

var approximateTests = []approximateTest{
	{
		input: FromInt64(1),
		expecteds: map[int]*big.Int{
			-3: big.NewInt(8),
			-2: big.NewInt(4),
			-1: big.NewInt(2),
			0:  big.NewInt(1),
			1:  big.NewInt(1),
		},
	},
}

func TestApproximate(t *testing.T) {
	for _, test := range approximateTests {
		for precision, expected := range test.expecteds {
			if result := Approximate(test.input, precision); result.Cmp(expected) != 0 {
				t.Errorf("precision %d, expected %v, got %v", precision, expected, result)
			}
		}
	}
}

type cmpTest struct {
	inputA   Real
	inputB   Real
	expected int
}

var cmpTests = []cmpTest{
	{
		inputA:   FromInt64(1),
		inputB:   FromInt64(2),
		expected: -1,
	},
	{
		inputA:   FromInt64(2),
		inputB:   FromInt64(1),
		expected: 1,
	},
}

func TestCmp(t *testing.T) {
	for _, test := range cmpTests {
		if result := Cmp(test.inputA, test.inputB); result != test.expected {
			t.Errorf("expected %d, got %d", test.expected, result)
		}
	}
}

type preciseCmpTest struct {
	inputA   Real
	inputB   Real
	expected int
}

var preciseCmpTests = []preciseCmpTest{
	{
		inputA:   FromInt64(1),
		inputB:   FromInt64(2),
		expected: -1,
	},
	{
		inputA:   FromInt64(2),
		inputB:   FromInt64(1),
		expected: 1,
	},
	{
		inputA:   FromInt64(5),
		inputB:   FromInt64(5),
		expected: 0,
	},
}

func TestPreciseCmp(t *testing.T) {
	for _, test := range preciseCmpTests {
		if result := PreciseCmp(test.inputA, test.inputB, -50); result != test.expected {
			t.Errorf("expected %d, got %d", test.expected, result)
		}
	}

	// 1 << 10 = 1024, 1 >> 10 = 1/1024
	assertEqualAtPrecision(t, FromInt(1024), ShiftLeft(FromInt(1), 10), -100)
	assertEqualAtPrecision(t, Inverse(FromInt(1024)), ShiftRight(FromInt(1), 10), -100)

	// 1/φ = φ - 1
	phi := Phi()
	assertEqualAtPrecision(t, Inverse(phi), Subtract(phi, FromInt(1)), -100)

	// e^1 = e, e^0 = 1, e^-1 = 1/e
	assertEqualAtPrecision(t, E(), Exp(FromInt(1)), -100)
	assertEqualAtPrecision(t, FromInt(1), Exp(FromInt(0)), -100)
	assertEqualAtPrecision(t, Inverse(E()), Exp(Negate(FromInt(1))), -100)

	// ln(2) = log_e(2)
	assertEqualAtPrecision(t, Ln2(), Ln(FromInt(2)), -70)

	// cos(0) = 1, cos(π/4) = √2/2, cos(π/3) = 1/2, cos(π/2) = 0, cos(π) = -1, cos(2π) = 1
	assertEqualAtPrecision(t, FromInt(1), Cosine(FromInt(0)), -100)
	assertEqualAtPrecision(t, Divide(Sqrt2(), FromInt(2)), Cosine(Divide(Pi(), FromInt(4))), -100)
	assertEqualAtPrecision(t, FromRat(1, 2), Cosine(Divide(Pi(), FromInt(3))), -100)
	assertEqualAtPrecision(t, Zero(), Cosine(Divide(Pi(), FromInt(2))), -100)
	assertEqualAtPrecision(t, FromInt(-1), Cosine(Pi()), -100)
	assertEqualAtPrecision(t, FromInt(1), Cosine(Multiply(FromInt(2), Pi())), -100)

	// sin(0) = 0, sin(π/4) = √2/2, sin(π/3) = √3/2, sin(π/2) = 1, sin(π) = 0, sin(2π) = 0
	assertEqualAtPrecision(t, Zero(), Sine(FromInt(0)), -100)
	assertEqualAtPrecision(t, Divide(Sqrt2(), FromInt(2)), Sine(Divide(Pi(), FromInt(4))), -100)
	assertEqualAtPrecision(t, Divide(Sqrt(FromInt(3)), FromInt(2)), Sine(Divide(Pi(), FromInt(3))), -100)
	assertEqualAtPrecision(t, FromInt(1), Sine(Divide(Pi(), FromInt(2))), -100)
	assertEqualAtPrecision(t, Zero(), Sine(Pi()), -100)
	assertEqualAtPrecision(t, Zero(), Sine(Multiply(FromInt(2), Pi())), -100)

	// tan(0) = 0, tan(π/4) = 1, tan(π/3) = √3, tan(π/2) = undefined, tan(π) = 0, tan(2π) = 0
	assertEqualAtPrecision(t, Zero(), Tangent(FromInt(0)), -100)
	assertEqualAtPrecision(t, FromInt(1), Tangent(Divide(Pi(), FromInt(4))), -100)
	assertEqualAtPrecision(t, Sqrt(FromInt(3)), Tangent(Divide(Pi(), FromInt(3))), -100)
	assertEqualAtPrecision(t, Zero(), Tangent(Pi()), -100)
	assertEqualAtPrecision(t, Zero(), Tangent(Multiply(FromInt(2), Pi())), -100)

	// TODO(ripta): never terminates
	// atan(0) = 0, atan(1) = π/4, atan(√3) = π/3, atan(∞) = π/2
	// assertEqualAtPrecision(t, FromInt(0), Arctangent(FromInt(0)), -100)
	// assertEqualAtPrecision(t, Divide(Pi(), FromInt(4)), Arctangent(FromInt(1)), -100)
	// assertEqualAtPrecision(t, Divide(Pi(), FromInt(3)), Arctangent(Sqrt(FromInt(3))), -100)
	// assertEqualAtPrecision(t, Divide(Pi(), FromInt(2)), Arctangent(FromInt(1<<1000)), -100)

	// 47/17 = [2; 1, 3, 4]
	assertEqualAtPrecision(t, Divide(FromInt(47), FromInt(17)), ContinuedFraction64([]int64{2, 1, 3, 4}), -100)
	assertEqualAtPrecision(t, Divide(FromInt(47), FromInt(17)), ContinuedFraction(FromInt64Slice([]int64{2, 1, 3, 4})), -200)
	assertEqualAtPrecision(t, Divide(FromInt(47), FromInt(17)), ContinuedFraction(FromIntSlice([]int{2, 1, 3, 4})), -200)
	assertEqualAtPrecision(t, Divide(FromInt(47), FromInt(17)), ContinuedFraction(FromFloat32Slice([]float32{2, 1, 3, 4})), -200)
	assertEqualAtPrecision(t, Divide(FromInt(47), FromInt(17)), ContinuedFraction(FromFloat64Slice([]float64{2, 1, 3, 4})), -200)

	// 81047/107501 = [0; 1, 3, 15, 1, 2, 3, 33, 2, 2]
	assertEqualAtPrecision(t, Divide(FromInt(81047), FromInt(107501)), ContinuedFraction64([]int64{0, 1, 3, 15, 1, 2, 3, 33, 2, 2}), -100)
}

func TestText(t *testing.T) {
	assert.True(t, true)

	ten := FromInt(10)
	assert.Equal(t, "10.00000", Text(ten, 5, 10))
	assert.Equal(t, "-10.00000", Text(Negate(ten), 5, 10))
	assert.Equal(t, "a.00000", Text(ten, 5, 16))
	assert.Equal(t, "-a.00000", Text(Negate(ten), 5, 16))

	assert.Equal(t, "5.00000", Text(Add(FromInt(3), FromInt(2)), 5, 10))
	assert.Equal(t, "1.00000", Text(Subtract(FromInt(3), FromInt(2)), 5, 10))

	assert.Equal(t, "6.00000", Text(Multiply(FromInt(3), FromInt(2)), 5, 10))
	assert.Equal(t, "6.75000", Text(Multiply(FromInt(3), FromFloat32(2.25)), 5, 10))

	assert.Equal(t, "0.50000", Text(Inverse(FromInt(2)), 5, 10))
	assert.Equal(t, "0.33333", Text(Inverse(FromFloat32(3)), 5, 10))

	assert.Equal(t, "3.00000", Text(Divide(FromInt(6), FromInt(2)), 5, 10))

	assert.Equal(t, "0.30000000447034835815", Text(Add(FromFloat32(0.1), FromFloat32(0.2)), 20, 10))
	assert.Equal(t, "0.30000000000000001665", Text(Add(FromFloat64(0.1), FromFloat64(0.2)), 20, 10))
	assert.Equal(t, "0.30000000000000000000", Text(Add(Inverse(FromInt(10)), Inverse(FromInt(5))), 20, 10))

	assert.Equal(t, "2.71828182845904509080", Text(FromFloat64(math.E), 20, 10))
	e := Exp(FromInt(1))
	assert.Equal(t, "2.7182818284590452353602874713526624977572470936999595749669676277240766", Text(e, 70, 10))

	nine := FromInt(9)
	assert.Equal(t, "1.00000", Text(Multiply(Inverse(nine), nine), 5, 10))

	ninth := Inverse(nine)
	assert.Equal(t, "0.11111111111111111111", Text(ninth, 20, 10))
	assert.Equal(t, "0.00011100011100011101", Text(ninth, 20, 2))
	assert.Equal(t, "0.01301301301301301302", Text(ninth, 20, 4))
	assert.Equal(t, "0.07070707070707070707", Text(ninth, 20, 8))
	assert.Equal(t, "0.14000000000000000000", Text(ninth, 20, 12))
	assert.Equal(t, "0.1c71c71c71c71c71c71c", Text(ninth, 20, 16))
	negNinth := Negate(ninth)
	assert.Equal(t, "-0.11111111111111111111", Text(negNinth, 20, 10))
	absNinth := Abs(negNinth)
	assert.Equal(t, "0.11111111111111111111", Text(absNinth, 20, 10))
	assert.Equal(t, "0.11111111111111111111", Text(Abs(absNinth), 20, 10))

	sqrt2 := Sqrt2()
	assert.Equal(t, "1.4142135623730950488016887242096980785696718753769480731766797379907325", Text(sqrt2, 70, 10))
	assertEqualAtPrecision(t, FromInt(4), Multiply(sqrt2, Sqrt(FromInt(8))), -100)

	sqrt11i := Sqrt(FromInt(11))
	assert.Equal(t, "3.31662", Text(sqrt11i, 5, 10))
	assert.Equal(t, "11.00000", Text(Square(sqrt11i), 5, 10))
	sqrt11f := Sqrt(FromFloat64(11))
	assert.Equal(t, Text(Square(sqrt11f), 70, 10), Text(Square(sqrt11i), 70, 10))

	phi := Phi()
	assert.Equal(t, "1.6180339887498948482045868343656381177203091798057628621354486227052605", Text(phi, 70, 10))

	pi := Pi()
	assert.Equal(t, "3.1415926535897932384626433832795028841971693993751058209749445923078164", Text(pi, 70, 10))

	ln2 := Ln2()
	assert.Equal(t, "0.6931471805599453094172321214581765680755001343602552541206800094933936", Text(ln2, 70, 10))

	assert.Equal(t, "1.0471975511965977461542144610931676280657231331250352736583148641026055", Text(Divide(Pi(), FromInt(3)), 70, 10))
	assert.Equal(t, "0.5000000000000000000000000000000000000000000000000000000000000000000000", Text(Cosine(Divide(Pi(), FromInt(3))), 70, 10))

	assert.Equal(t, "<undefined: division by zero>", Text(Tangent(Divide(Pi(), FromInt(2))), 70, 10))

	// 2 ^ 3
	assert.Equal(t,
		"8.0000000000000000000000000000000000000000000000000000000000000000000000",
		Text(Pow(FromInt(2), FromInt(3)), 70, 10),
	)

	// 2 ^ -3
	assert.Equal(t,
		"0.1250000000000000000000000000000000000000000000000000000000000000000000",
		Text(Pow(FromInt(2), FromInt(-3)), 70, 10),
	)

	// (√π - √3) ^ 8
	assert.Equal(t,
		"0.0000000000071008875411429851278570030225300893747800769074951130688105",
		Text(Pow(Subtract(Sqrt(Pi()), Sqrt(FromInt(3))), FromInt(8)), 70, 10),
	)

	// 3 ^ (9/7)
	assert.Equal(t,
		"4.1062143199266050245271033659920889591493609394572477980497607290832348",
		Text(Pow(FromInt(3), FromRat(9, 7)), 70, 10),
	)

	// π^e
	assert.Equal(t,
		"22.4591577183610454734271522045437350275893151339966922492030025540669260",
		Text(Pow(Pi(), E()), 70, 10),
	)
}

func assertEqualAtPrecision(t *testing.T, a, b Real, precision int) {
	t.Helper()
	if result := PreciseCmp(a, b, precision); result != 0 {
		t.Errorf("expected [1] to be equal to [2] at precision %d\n[1]: %s\n     %#v\n[2]: %s\n     %#v", precision, Text(a, -precision, 10), a, Text(b, -precision, 10), b)
	}
}
