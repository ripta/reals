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
	{name: "undecided", loc: Undecided, expected: "Undecided"},
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
		name:     "exact lower endpoint is undecided at the budget",
		iv:       New(FromInt(2), FromInt(4)),
		x:        FromInt(2),
		p:        -100,
		expected: Undecided,
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
		name:     "degenerate exact center is undecided at the budget",
		iv:       New(FromInt(4), FromInt(4)),
		x:        FromInt(4),
		p:        -100,
		expected: Undecided,
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
