package constructive

import "testing"

type locationStringTest struct {
	name     string
	loc      Location
	expected string
}

var locationStringTests = []locationStringTest{
	{name: "below", loc: Below, expected: "Below"},
	{name: "inside", loc: Inside, expected: "Inside"},
	{name: "above", loc: Above, expected: "Above"},
	{name: "indeterminate", loc: Indeterminate, expected: "Indeterminate"},
}

func TestLocationString(t *testing.T) {
	for _, test := range locationStringTests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.loc.String(); got != test.expected {
				t.Errorf("String() = %q, want %q", got, test.expected)
			}
		})
	}
}

type locateTest struct {
	name     string
	iv       Interval
	x        Real
	p        int
	expected Location
}

var locateTests = []locateTest{
	{
		name:     "below the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(1),
		p:        -100,
		expected: Below,
	},
	{
		name:     "inside the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(3),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "above the band",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(5),
		p:        -100,
		expected: Above,
	},
	{
		name:     "exact lower endpoint is indeterminate at the budget",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(2),
		p:        -100,
		expected: Indeterminate,
	},
	{
		name:     "degenerate point above center mirrors PreciseSign",
		iv:       New(FromInt(4), FromInt(4)),
		x:        FromInt(5),
		p:        -100,
		expected: Above,
	},
	{
		name:     "degenerate point below center mirrors PreciseSign",
		iv:       New(FromInt(4), FromInt(4)),
		x:        FromInt(3),
		p:        -100,
		expected: Below,
	},
	{
		name:     "degenerate exact center is indeterminate at the budget",
		iv:       New(FromInt(4), FromInt(4)),
		x:        FromInt(4),
		p:        -100,
		expected: Indeterminate,
	},
}

func TestLocate(t *testing.T) {
	for _, test := range locateTests {
		t.Run(test.name, func(t *testing.T) {
			if got := Locate(test.iv, test.x, test.p); got != test.expected {
				t.Errorf("Locate() = %v, want %v", got, test.expected)
			}
		})
	}
}

type decisionStringTest struct {
	name     string
	dec      Decision
	expected string
}

var decisionStringTests = []decisionStringTest{
	{name: "undecided", dec: Undecided, expected: "Undecided"},
	{name: "yes", dec: Yes, expected: "Yes"},
	{name: "no", dec: No, expected: "No"},
}

func TestDecisionString(t *testing.T) {
	for _, test := range decisionStringTests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.dec.String(); got != test.expected {
				t.Errorf("String() = %q, want %q", got, test.expected)
			}
		})
	}
}

type isEmptyTest struct {
	name     string
	iv       Interval
	p        int
	expected Decision
}

var isEmptyTests = []isEmptyTest{
	{
		name:     "disjoint intersection is empty",
		iv:       Intersect(New(FromInt(0), FromInt(1)), New(FromInt(2), FromInt(3))),
		p:        -100,
		expected: Yes,
	},
	{
		name:     "proper interval is not empty",
		iv:       New(FromInt(2), FromInt(4)),
		p:        -100,
		expected: No,
	},
	{
		name:     "degenerate interval is undecided at the budget",
		iv:       New(FromInt(4), FromInt(4)),
		p:        -100,
		expected: Undecided,
	},
}

func TestIsEmpty(t *testing.T) {
	for _, test := range isEmptyTests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsEmpty(test.iv, test.p); got != test.expected {
				t.Errorf("IsEmpty() = %v, want %v", got, test.expected)
			}
		})
	}
}

type overlapTest struct {
	name     string
	a, b     Interval
	p        int
	expected Decision
}

var overlapTests = []overlapTest{
	{
		name:     "disjoint intervals do not overlap",
		a:        New(FromInt(0), FromInt(1)),
		b:        New(FromInt(2), FromInt(3)),
		p:        -100,
		expected: No,
	},
	{
		name:     "overlapping intervals overlap",
		a:        New(FromInt(0), FromInt(2)),
		b:        New(FromInt(1), FromInt(3)),
		p:        -100,
		expected: Yes,
	},
	{
		name:     "touching at a point is undecided at the budget",
		a:        New(FromInt(0), FromInt(1)),
		b:        New(FromInt(1), FromInt(2)),
		p:        -100,
		expected: Undecided,
	},
}

func TestOverlap(t *testing.T) {
	for _, test := range overlapTests {
		t.Run(test.name, func(t *testing.T) {
			if got := Overlap(test.a, test.b, test.p); got != test.expected {
				t.Errorf("Overlap() = %v, want %v", got, test.expected)
			}
		})
	}
}
