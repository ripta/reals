package unified

import (
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// assertLoEqHi checks the interval collapsed to a point at the given precision.
func assertLoEqHi(t *testing.T, iv Interval, precision int) {
	t.Helper()
	if mustPreciseCmp(t, iv.lo.Constructive(), iv.hi.Constructive(), precision) != 0 {
		t.Errorf("expected lo == hi, got lo != hi at precision %d", precision)
	}
}

// assertRationalEndpoint checks an endpoint kept its exact rational factor
// rather than fusing it into a constructive Min/Max node. cr must be the shared
// One() instance and rr must equal the expected rational.
func assertRationalEndpoint(t *testing.T, u *Real, rr *rational.Number) {
	t.Helper()
	if u.cr != constructive.One() {
		t.Errorf("expected endpoint constructive factor to be One(), rational visibility was fused away")
	}
	if u.rr.Cmp(rr) != 0 {
		t.Errorf("expected endpoint rational factor to be preserved")
	}
}

type unifiedIntervalNewTest struct {
	name  string
	p, q  *Real
	loNum int64
	hiNum int64
}

var unifiedIntervalNewTests = []unifiedIntervalNewTest{
	{
		name:  "already ordered rationals",
		p:     rat(3, 1),
		q:     rat(5, 1),
		loNum: 3,
		hiNum: 5,
	},
	{
		name:  "swaps unordered rationals",
		p:     rat(5, 1),
		q:     rat(3, 1),
		loNum: 3,
		hiNum: 5,
	},
}

func TestUnifiedIntervalNewPreservesRationals(t *testing.T) {
	for _, test := range unifiedIntervalNewTests {
		t.Run(test.name, func(t *testing.T) {
			iv := NewInterval(test.p, test.q)
			assertRationalEndpoint(t, iv.lo, rational.New64(test.loNum, 1))
			assertRationalEndpoint(t, iv.hi, rational.New64(test.hiNum, 1))
		})
	}
}

func TestUnifiedIntervalNewEqualEndpoints(t *testing.T) {
	iv := NewInterval(rat(3, 1), rat(3, 1))
	assertLoEqHi(t, iv, -100)
}

func TestUnifiedIntervalNewIrrationalFallback(t *testing.T) {
	iv := NewInterval(Pi(), rat(3, 1))
	if mustPreciseCmp(t, iv.lo.Constructive(), iv.hi.Constructive(), -100) > 0 {
		t.Errorf("normalization violated: lo > hi")
	}
	assertEqualAtPrecision(t, iv.lo, rat(3, 1), -100)
	assertEqualAtPrecision(t, iv.hi, Pi(), -100)
}

type unifiedIntervalUnaryTest struct {
	name     string
	iv       Interval
	expected *Real
}

var unifiedIntervalWidthTests = []unifiedIntervalUnaryTest{
	{
		name:     "integer band",
		iv:       NewInterval(rat(3, 1), rat(5, 1)),
		expected: rat(2, 1),
	},
	{
		name:     "rational band",
		iv:       NewInterval(rat(1, 2), rat(3, 2)),
		expected: rat(1, 1),
	},
}

func TestUnifiedIntervalWidth(t *testing.T) {
	for _, test := range unifiedIntervalWidthTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.iv.Width(), test.expected, -100)
		})
	}
}

var unifiedIntervalMidTests = []unifiedIntervalUnaryTest{
	{
		name:     "integer midpoint",
		iv:       NewInterval(rat(3, 1), rat(5, 1)),
		expected: rat(4, 1),
	},
	{
		name:     "exact half",
		iv:       NewInterval(rat(0, 1), rat(1, 1)),
		expected: rat(1, 2),
	},
}

func TestUnifiedIntervalMid(t *testing.T) {
	for _, test := range unifiedIntervalMidTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.iv.Mid(), test.expected, -100)
		})
	}
}

type unifiedIntervalClampTest struct {
	name     string
	iv       Interval
	x        *Real
	expected *Real
}

var unifiedIntervalClampTests = []unifiedIntervalClampTest{
	{
		name:     "inside the band",
		iv:       NewInterval(rat(2, 1), rat(4, 1)),
		x:        rat(3, 1),
		expected: rat(3, 1),
	},
	{
		name:     "below the band",
		iv:       NewInterval(rat(2, 1), rat(4, 1)),
		x:        rat(0, 1),
		expected: rat(2, 1),
	},
	{
		name:     "above the band",
		iv:       NewInterval(rat(2, 1), rat(4, 1)),
		x:        rat(9, 1),
		expected: rat(4, 1),
	},
}

func TestUnifiedIntervalClamp(t *testing.T) {
	for _, test := range unifiedIntervalClampTests {
		t.Run(test.name, func(t *testing.T) {
			assertEqualAtPrecision(t, test.iv.Clamp(test.x), test.expected, -100)
		})
	}
}

type unifiedIntervalBinaryTest struct {
	name   string
	a, b   Interval
	lo, hi *Real
}

var unifiedIntervalHullTests = []unifiedIntervalBinaryTest{
	{
		name: "disjoint inputs",
		a:    NewInterval(rat(0, 1), rat(1, 1)),
		b:    NewInterval(rat(3, 1), rat(4, 1)),
		lo:   rat(0, 1),
		hi:   rat(4, 1),
	},
}

func TestUnifiedIntervalHull(t *testing.T) {
	for _, test := range unifiedIntervalHullTests {
		t.Run(test.name, func(t *testing.T) {
			h := test.a.Hull(test.b)
			assertEqualAtPrecision(t, h.lo, test.lo, -100)
			assertEqualAtPrecision(t, h.hi, test.hi, -100)
		})
	}
}

var unifiedIntervalIntersectTests = []unifiedIntervalBinaryTest{
	{
		name: "overlapping inputs",
		a:    NewInterval(rat(0, 1), rat(3, 1)),
		b:    NewInterval(rat(2, 1), rat(5, 1)),
		lo:   rat(2, 1),
		hi:   rat(3, 1),
	},
	{
		name: "disjoint inputs return a value with lo > hi",
		a:    NewInterval(rat(0, 1), rat(1, 1)),
		b:    NewInterval(rat(3, 1), rat(4, 1)),
		lo:   rat(3, 1),
		hi:   rat(1, 1),
	},
}

func TestUnifiedIntervalIntersect(t *testing.T) {
	for _, test := range unifiedIntervalIntersectTests {
		t.Run(test.name, func(t *testing.T) {
			r := test.a.Intersect(test.b)
			assertEqualAtPrecision(t, r.lo, test.lo, -100)
			assertEqualAtPrecision(t, r.hi, test.hi, -100)
		})
	}
}
