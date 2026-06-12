package unified

import (
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

var floorTests = []unaryTest{
	{name: "Floor(5/2)=2", input: rat(5, 2), expected: rat(2, 1)},
	{name: "Floor(-5/2)=-3", input: rat(-5, 2), expected: rat(-3, 1)},
	{name: "Floor(3)=3", input: rat(3, 1), expected: rat(3, 1)},
	{name: "Floor(Pi)=3", input: Pi(), expected: rat(3, 1)},
}

func TestFloor(t *testing.T) {
	for _, test := range floorTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Floor(), -100)
		})
	}
}

var ceilTests = []unaryTest{
	{name: "Ceil(5/2)=3", input: rat(5, 2), expected: rat(3, 1)},
	{name: "Ceil(-5/2)=-2", input: rat(-5, 2), expected: rat(-2, 1)},
	{name: "Ceil(3)=3", input: rat(3, 1), expected: rat(3, 1)},
	{name: "Ceil(Pi)=4", input: Pi(), expected: rat(4, 1)},
}

func TestCeil(t *testing.T) {
	for _, test := range ceilTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Ceil(), -100)
		})
	}
}

var roundTests = []unaryTest{
	{name: "Round(5/2)=3", input: rat(5, 2), expected: rat(3, 1)},
	{name: "Round(-5/2)=-3", input: rat(-5, 2), expected: rat(-3, 1)},
	{name: "Round(7/3)=2", input: rat(7, 3), expected: rat(2, 1)},
	{name: "Round(Pi)=3", input: Pi(), expected: rat(3, 1)},
}

func TestRound(t *testing.T) {
	for _, test := range roundTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.Round(), -100)
		})
	}
}

var roundToEvenTests = []unaryTest{
	{name: "RoundToEven(5/2)=2", input: rat(5, 2), expected: rat(2, 1)},
	{name: "RoundToEven(7/2)=4", input: rat(7, 2), expected: rat(4, 1)},
	{name: "RoundToEven(-5/2)=-2", input: rat(-5, 2), expected: rat(-2, 1)},
	{name: "RoundToEven(Pi)=3", input: Pi(), expected: rat(3, 1)},
}

func TestRoundToEven(t *testing.T) {
	for _, test := range roundToEvenTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.input.RoundToEven(), -100)
		})
	}
}

type binaryTest struct {
	name     string
	a        *Real
	b        *Real
	expected *Real
}

var minTests = []binaryTest{
	{name: "Min(2,3)=2", a: rat(2, 1), b: rat(3, 1), expected: rat(2, 1)},
	{name: "Min(3,2)=2", a: rat(3, 1), b: rat(2, 1), expected: rat(2, 1)},
	{name: "Min(2,Pi)=2", a: Two(), b: Pi(), expected: rat(2, 1)},
	{name: "Min(Pi,2)=2", a: Pi(), b: Two(), expected: rat(2, 1)},
}

func TestMin(t *testing.T) {
	for _, test := range minTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.a.Min(test.b), -100)
		})
	}
}

var maxTests = []binaryTest{
	{name: "Max(2,3)=3", a: rat(2, 1), b: rat(3, 1), expected: rat(3, 1)},
	{name: "Max(3,2)=3", a: rat(3, 1), b: rat(2, 1), expected: rat(3, 1)},
	{name: "Max(2,Pi)=Pi", a: Two(), b: Pi(), expected: New(constructive.Pi(), rational.One())},
	{name: "Max(Pi,2)=Pi", a: Pi(), b: Two(), expected: New(constructive.Pi(), rational.One())},
}

func TestMax(t *testing.T) {
	for _, test := range maxTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.expected, test.a.Max(test.b), -100)
		})
	}
}
