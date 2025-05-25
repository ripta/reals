package unified

import (
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
	"github.com/stretchr/testify/assert"
)

type newTest struct {
	cr constructive.Real
	rr *rational.Number
}

var newTests = []newTest{
	{},
	{
		cr: constructive.One(),
	},
	{
		cr: constructive.Negate(constructive.One()),
	},
	{
		rr: rational.One(),
	},
	{
		rr: rational.New64(1, 2),
	},
	{
		cr: constructive.E(),
		rr: rational.New64(1, 2),
	},
}

func TestNew(t *testing.T) {
	for _, test := range newTests {
		u := New(test.cr, test.rr)
		assert.NotNil(t, u)
	}
}

func assertEqualAtPrecision(t *testing.T, expected, actual *Real, precision int) {
	t.Helper()
	a := expected.Constructive()
	b := actual.Constructive()
	if result := constructive.PreciseCmp(a, b, precision); result != 0 {
		t.Errorf("expected [1] to be equal to [2] at precision %d\n[1]: %s\n     %#v\n[2]: %s\n     %#v",
			precision,
			constructive.Text(a, -precision, 10),
			a,
			constructive.Text(b, -precision, 10),
			b)
	}
}

type constructiveTest struct {
	name     string
	input    *Real
	expected constructive.Real
}

var constructiveTests = []constructiveTest{
	{
		name:     "Half",
		input:    Half(),
		expected: constructive.Divide(constructive.One(), constructive.Two()),
	},
	{
		name:     "E times half",
		input:    New(constructive.E(), rational.New64(1, 2)),
		expected: constructive.Divide(constructive.E(), constructive.Two()),
	},
	{
		name:     "Pi times three quarters",
		input:    New(constructive.Pi(), rational.New64(3, 4)),
		expected: constructive.Divide(constructive.Multiply(constructive.FromInt(3), constructive.Pi()), constructive.FromInt(4)),
	},
	{
		name:     "NegativeOne",
		input:    NegativeOne(),
		expected: constructive.FromInt(-1),
	},
	{
		name:     "Zero",
		input:    Zero(),
		expected: constructive.FromInt(0),
	},
	{
		name:     "One",
		input:    One(),
		expected: constructive.One(),
	},
	{
		name:     "E",
		input:    E(),
		expected: constructive.E(),
	},
	{
		name:     "Pi",
		input:    Pi(),
		expected: constructive.Pi(),
	},
}

func TestConstructive(t *testing.T) {
	for _, test := range constructiveTests {
		t.Run(test.name, func(t *testing.T) {
			actual := test.input.Constructive()
			if result := constructive.PreciseCmp(test.expected, actual, -100); result != 0 {
				t.Errorf("expected %s, got %s",
					constructive.Text(test.expected, 100, 10),
					constructive.Text(actual, 100, 10))
			}
		})
	}
}

type addTest struct {
	name     string
	a        *Real
	b        *Real
	expected *Real
}

var addTests = []addTest{
	{
		name:     "same constructive: E/2 + E/4 = 3E/4",
		a:        New(constructive.E(), rational.New64(1, 2)),
		b:        New(constructive.E(), rational.New64(1, 4)),
		expected: New(constructive.E(), rational.New64(3, 4)),
	},
	{
		name:     "zero identity: 0 + x = x",
		a:        Zero(),
		b:        Half(),
		expected: Half(),
	},
	{
		name:     "zero identity: x + 0 = x",
		a:        Half(),
		b:        Zero(),
		expected: Half(),
	},
	{
		name:     "commutativity: 1/2 + 1/4 = 1/4 + 1/2",
		a:        Half(),
		b:        New(constructive.One(), rational.New64(1, 4)),
		expected: New(constructive.One(), rational.New64(3, 4)),
	},
	{
		name:     "negative values: 1/2 + (-1) = -1/2",
		a:        Half(),
		b:        NegativeOne(),
		expected: New(constructive.One(), rational.New64(-1, 2)),
	},
	{
		name:     "different constructive: One + Half = 3/2",
		a:        One(),
		b:        Half(),
		expected: New(constructive.One(), rational.New64(3, 2)),
	},
	{
		name:     "transcendentals: E + Pi",
		a:        E(),
		b:        Pi(),
		expected: New(constructive.Add(constructive.E(), constructive.Pi()), rational.New64(2, 1)),
	},
}

func TestAdd(t *testing.T) {
	for _, test := range addTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.a.Add(test.b)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}

	t.Run("commutativity property", func(t *testing.T) {
		a := New(constructive.Pi(), rational.New64(2, 3))
		b := New(constructive.E(), rational.New64(3, 5))
		ab := a.Add(b)
		ba := b.Add(a)
		assertEqualAtPrecision(t, ab, ba, -100)
	})
}

type subtractTest struct {
	name     string
	a        *Real
	b        *Real
	expected *Real
}

var subtractTests = []subtractTest{
	{
		name:     "self subtraction: x - x = 0",
		a:        Half(),
		b:        Half(),
		expected: Zero(),
	},
	{
		name:     "zero identity: x - 0 = x",
		a:        Half(),
		b:        Zero(),
		expected: Half(),
	},
	{
		name:     "basic subtraction: 3/4 - 1/4 = 1/2",
		a:        New(constructive.One(), rational.New64(3, 4)),
		b:        New(constructive.One(), rational.New64(1, 4)),
		expected: Half(),
	},
	{
		name:     "negative result: 1/4 - 3/4 = -1/2",
		a:        New(constructive.One(), rational.New64(1, 4)),
		b:        New(constructive.One(), rational.New64(3, 4)),
		expected: New(constructive.One(), rational.New64(-1, 2)),
	},
	{
		name:     "subtracting negative: 1/2 - (-1) = 3/2",
		a:        Half(),
		b:        NegativeOne(),
		expected: New(constructive.One(), rational.New64(3, 2)),
	},
	{
		name:     "One - Half = Half",
		a:        One(),
		b:        Half(),
		expected: Half(),
	},
}

func TestSubtract(t *testing.T) {
	for _, test := range subtractTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.a.Subtract(test.b)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}

	t.Run("non-commutativity: a - b != b - a", func(t *testing.T) {
		a := One()
		b := Half()
		ab := a.Subtract(b)
		ba := b.Subtract(a)

		abConst := ab.Constructive()
		baConst := ba.Constructive()

		if constructive.PreciseCmp(abConst, baConst, -100) == 0 {
			t.Errorf("expected a-b != b-a, but got equal values")
		}

		assertEqualAtPrecision(t, ab, ba.Negate(), -100)
	})
}

type multiplyTest struct {
	name     string
	a        *Real
	b        *Real
	expected *Real
}

var multiplyTests = []multiplyTest{
	{
		name:     "identity: One * x = x",
		a:        One(),
		b:        Half(),
		expected: Half(),
	},
	{
		name:     "identity: x * One = x",
		a:        Half(),
		b:        One(),
		expected: Half(),
	},
	{
		name:     "zero: Zero * x = Zero",
		a:        Zero(),
		b:        Half(),
		expected: Zero(),
	},
	{
		name:     "zero: x * Zero = Zero",
		a:        Half(),
		b:        Zero(),
		expected: Zero(),
	},
	{
		name:     "Half * Half = 1/4",
		a:        Half(),
		b:        Half(),
		expected: New(constructive.One(), rational.New64(1, 4)),
	},
	{
		name:     "negative: Half * NegativeOne = -1/2",
		a:        Half(),
		b:        NegativeOne(),
		expected: New(constructive.One(), rational.New64(-1, 2)),
	},
	{
		name:     "negative * negative = positive: (-1) * (-1) = 1",
		a:        NegativeOne(),
		b:        NegativeOne(),
		expected: One(),
	},
	{
		name:     "transcendental: Pi * 1/2 = Pi/2",
		a:        Pi(),
		b:        Half(),
		expected: New(constructive.Pi(), rational.New64(1, 2)),
	},
	{
		name:     "Two * Half = One",
		a:        Two(),
		b:        Half(),
		expected: One(),
	},
}

func TestMultiply(t *testing.T) {
	for _, test := range multiplyTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.a.Multiply(test.b)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}

	t.Run("commutativity property", func(t *testing.T) {
		a := New(constructive.Pi(), rational.New64(2, 3))
		b := New(constructive.E(), rational.New64(3, 5))
		ab := a.Multiply(b)
		ba := b.Multiply(a)
		assertEqualAtPrecision(t, ab, ba, -100)
	})
}

type divideTest struct {
	name     string
	a        *Real
	b        *Real
	expected *Real
}

var divideTests = []divideTest{
	{
		name:     "identity: x / One = x",
		a:        Half(),
		b:        One(),
		expected: Half(),
	},
	{
		name:     "self division: x / x = One",
		a:        Half(),
		b:        Half(),
		expected: One(),
	},
	{
		name:     "self division: phi / phi = One",
		a:        Phi(),
		b:        Phi(),
		expected: One(),
	},
	{
		name:     "reciprocal: One / Half = Two",
		a:        One(),
		b:        Half(),
		expected: Two(),
	},
	{
		name:     "Half / Two = 1/4",
		a:        Half(),
		b:        Two(),
		expected: New(constructive.One(), rational.New64(1, 4)),
	},
	{
		name:     "negative: Half / NegativeOne = -1/2",
		a:        Half(),
		b:        NegativeOne(),
		expected: New(constructive.One(), rational.New64(-1, 2)),
	},
	{
		name:     "transcendental: Pi / Two",
		a:        Pi(),
		b:        Two(),
		expected: New(constructive.Pi(), rational.New64(1, 2)),
	},
}

func TestDivide(t *testing.T) {
	for _, test := range divideTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.a.Divide(test.b)
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}
}

type negateTest struct {
	name     string
	input    *Real
	expected *Real
}

var negateTests = []negateTest{
	{
		name:     "negate One = NegativeOne",
		input:    One(),
		expected: NegativeOne(),
	},
	{
		name:     "negate NegativeOne = One",
		input:    NegativeOne(),
		expected: One(),
	},
	{
		name:     "negate Zero = Zero",
		input:    Zero(),
		expected: Zero(),
	},
	{
		name:     "negate Half = -1/2",
		input:    Half(),
		expected: New(constructive.One(), rational.New64(-1, 2)),
	},
	{
		name:     "negate Pi = -Pi",
		input:    Pi(),
		expected: New(constructive.Negate(constructive.Pi()), rational.One()),
	},
}

func TestNegate(t *testing.T) {
	for _, test := range negateTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.Negate()
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}

	t.Run("double negation: -(-x) = x", func(t *testing.T) {
		x := New(constructive.E(), rational.New64(3, 7))
		negNegX := x.Negate().Negate()
		assertEqualAtPrecision(t, x, negNegX, -100)
	})
}

type inverseTest struct {
	name     string
	input    *Real
	expected *Real
}

var inverseTests = []inverseTest{
	{
		name:     "inverse One = One",
		input:    One(),
		expected: One(),
	},
	{
		name:     "inverse Half = Two",
		input:    Half(),
		expected: Two(),
	},
	{
		name:     "inverse Two = Half",
		input:    Two(),
		expected: Half(),
	},
	{
		name:     "inverse NegativeOne = NegativeOne",
		input:    NegativeOne(),
		expected: NegativeOne(),
	},
	{
		name:     "inverse of 1/4 = 4",
		input:    New(constructive.One(), rational.New64(1, 4)),
		expected: New(constructive.One(), rational.New64(4, 1)),
	},
}

func TestInverse(t *testing.T) {
	for _, test := range inverseTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.Inverse()
			assertEqualAtPrecision(t, test.expected, result, -100)
		})
	}

	t.Run("double inverse: 1/(1/x) = x", func(t *testing.T) {
		x := New(constructive.E(), rational.New64(3, 7))
		invInvX := x.Inverse().Inverse()
		assertEqualAtPrecision(t, x, invInvX, -100)
	})
}

type isZeroTest struct {
	name     string
	input    *Real
	expected bool
}

var isZeroTests = []isZeroTest{
	{
		name:     "Zero() is zero",
		input:    Zero(),
		expected: true,
	},
	{
		name:     "E with zero rational is zero",
		input:    New(constructive.E(), rational.Zero()),
		expected: true,
	},
	{
		name:     "Pi with zero rational is zero",
		input:    New(constructive.Pi(), rational.Zero()),
		expected: true,
	},
	{
		name:     "One is not zero",
		input:    One(),
		expected: false,
	},
	{
		name:     "Half is not zero",
		input:    Half(),
		expected: false,
	},
	{
		name:     "NegativeOne is not zero",
		input:    NegativeOne(),
		expected: false,
	},
	{
		name:     "E is not zero",
		input:    E(),
		expected: false,
	},
	{
		name:     "Pi is not zero",
		input:    Pi(),
		expected: false,
	},
	{
		name:     "constructive.Zero with rational.One is not zero",
		input:    New(constructive.Zero(), rational.One()),
		expected: false,
	},
}

func TestIsZero(t *testing.T) {
	for _, test := range isZeroTests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.IsZero()
			if result != test.expected {
				t.Errorf("expected IsZero() = %v, got %v", test.expected, result)
			}
		})
	}
}
