package constructive

import "testing"

type intervalNewTest struct {
	name   string
	p, q   Real
	lo, hi Real
}

var intervalNewTests = []intervalNewTest{
	{
		name: "already ordered",
		p:    FromInt(3),
		q:    FromInt(5),
		lo:   FromInt(3),
		hi:   FromInt(5),
	},
	{
		name: "swaps unordered endpoints",
		p:    FromInt(5),
		q:    FromInt(3),
		lo:   FromInt(3),
		hi:   FromInt(5),
	},
	{
		name: "equal endpoints yield Lo == Hi",
		p:    FromInt(4),
		q:    FromInt(4),
		lo:   FromInt(4),
		hi:   FromInt(4),
	},
}

func TestIntervalNew(t *testing.T) {
	for _, test := range intervalNewTests {
		t.Run(test.name, func(t *testing.T) {
			iv := New(test.p, test.q)
			if mustPreciseCmp(t, iv.Lo, iv.Hi, -100) > 0 {
				t.Errorf("normalization violated: Lo > Hi")
			}
			assertEqualAtPrecision(t, iv.Lo, test.lo, -100)
			assertEqualAtPrecision(t, iv.Hi, test.hi, -100)
		})
	}
}

type intervalUnaryTest struct {
	name     string
	iv       Interval
	expected Real
}

var intervalWidthTests = []intervalUnaryTest{
	{
		name:     "integer band",
		iv:       New(FromInt(3), FromInt(5)),
		expected: FromInt(2),
	},
	{
		name:     "rational band",
		iv:       New(FromRat(1, 2), FromRat(3, 2)),
		expected: FromInt(1),
	},
}

func TestIntervalWidth(t *testing.T) {
	for _, test := range intervalWidthTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, Width(test.iv), test.expected, -100)
		})
	}
}

var intervalMidTests = []intervalUnaryTest{
	{
		name:     "integer midpoint",
		iv:       New(FromInt(3), FromInt(5)),
		expected: FromInt(4),
	},
	{
		name:     "exact half",
		iv:       New(FromInt(0), FromInt(1)),
		expected: FromRat(1, 2),
	},
}

func TestIntervalMid(t *testing.T) {
	for _, test := range intervalMidTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, Mid(test.iv), test.expected, -100)
		})
	}
}

type intervalClampTest struct {
	name     string
	iv       Interval
	x        Real
	expected Real
}

var intervalClampTests = []intervalClampTest{
	{
		name:     "inside the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(3),
		expected: FromInt(3),
	},
	{
		name:     "below the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(0),
		expected: FromInt(2),
	},
	{
		name:     "above the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(9),
		expected: FromInt(4),
	},
}

func TestIntervalClamp(t *testing.T) {
	for _, test := range intervalClampTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, Clamp(test.iv, test.x), test.expected, -100)
		})
	}
}

type intervalBinaryTest struct {
	name   string
	a, b   Interval
	lo, hi Real
}

var intervalHullTests = []intervalBinaryTest{
	{
		name: "disjoint inputs",
		a:    New(FromInt(0), FromInt(1)),
		b:    New(FromInt(3), FromInt(4)),
		lo:   FromInt(0),
		hi:   FromInt(4),
	},
}

func TestIntervalHull(t *testing.T) {
	for _, test := range intervalHullTests {
		t.Run(test.name, func(t *testing.T) {
			h := Hull(test.a, test.b)
			assertEqualAtPrecision(t, h.Lo, test.lo, -100)
			assertEqualAtPrecision(t, h.Hi, test.hi, -100)
		})
	}
}

var intervalIntersectTests = []intervalBinaryTest{
	{
		name: "overlapping inputs",
		a:    New(FromInt(0), FromInt(3)),
		b:    New(FromInt(2), FromInt(5)),
		lo:   FromInt(2),
		hi:   FromInt(3),
	},
	{
		name: "disjoint inputs return a value with Lo > Hi",
		a:    New(FromInt(0), FromInt(1)),
		b:    New(FromInt(3), FromInt(4)),
		lo:   FromInt(3),
		hi:   FromInt(1),
	},
}

func TestIntervalIntersect(t *testing.T) {
	for _, test := range intervalIntersectTests {
		t.Run(test.name, func(t *testing.T) {
			r := Intersect(test.a, test.b)
			assertEqualAtPrecision(t, r.Lo, test.lo, -100)
			assertEqualAtPrecision(t, r.Hi, test.hi, -100)
		})
	}
}
