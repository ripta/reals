package unified

import (
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

// irr wraps an irrational constructive real as a unified Real, so its
// constructive factor is not One() and the predicates take the budgeted path.
func irr(cr constructive.Real) *Real {
	return New(cr, rational.One())
}

type unifiedLocateTest struct {
	name     string
	iv       Interval
	x        *Real
	p        int
	expected Location
}

var unifiedLocateTests = []unifiedLocateTest{
	{
		name:     "rational point below rational interval",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		x:        rat(1, 1),
		p:        -100,
		expected: Below,
	},
	{
		name:     "rational point above rational interval",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		x:        rat(9, 1),
		p:        -100,
		expected: Above,
	},
	{
		name:     "rational point strictly inside rational interval",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		x:        rat(3, 1),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "exact lo endpoint resolves to Inside",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		x:        rat(2, 1),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "exact hi endpoint resolves to Inside",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		x:        rat(5, 1),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "degenerate rational interval at its point resolves to Inside",
		iv:       NewInterval(rat(3, 1), rat(3, 1)),
		x:        rat(3, 1),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "rational point at rational lo stays exact against irrational hi",
		iv:       Interval{lo: rat(1, 1), hi: irr(constructive.Sqrt2())},
		x:        rat(1, 1),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "irrational point inside rational interval is decided at budget",
		iv:       NewInterval(rat(1, 1), rat(2, 1)),
		x:        irr(constructive.Sqrt2()),
		p:        -100,
		expected: Inside,
	},
	{
		name:     "irrational point below rational interval is decided at budget",
		iv:       NewInterval(rat(2, 1), rat(3, 1)),
		x:        irr(constructive.Sqrt2()),
		p:        -100,
		expected: Below,
	},
}

func TestUnifiedIntervalLocate(t *testing.T) {
	for _, test := range unifiedLocateTests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.iv.Locate(test.x, test.p); got != test.expected {
				t.Errorf("Locate = %v, want %v", got, test.expected)
			}
		})
	}
}

type unifiedIsEmptyTest struct {
	name     string
	iv       Interval
	p        int
	expected Decision
}

var unifiedIsEmptyTests = []unifiedIsEmptyTest{
	{
		name:     "empty rational interval reports Yes",
		iv:       Interval{lo: rat(5, 1), hi: rat(2, 1)},
		p:        -100,
		expected: Yes,
	},
	{
		name:     "non-empty rational interval reports No",
		iv:       NewInterval(rat(2, 1), rat(5, 1)),
		p:        -100,
		expected: No,
	},
	{
		name:     "degenerate rational interval reports No exactly",
		iv:       NewInterval(rat(3, 1), rat(3, 1)),
		p:        -100,
		expected: No,
	},
	{
		name:     "irrational touching interval reports Undecided at budget",
		iv:       Interval{lo: irr(constructive.Sqrt2()), hi: irr(constructive.Sqrt2())},
		p:        -100,
		expected: Undecided,
	},
}

func TestUnifiedIntervalIsEmpty(t *testing.T) {
	for _, test := range unifiedIsEmptyTests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.iv.IsEmpty(test.p); got != test.expected {
				t.Errorf("IsEmpty = %v, want %v", got, test.expected)
			}
		})
	}
}

type unifiedOverlapTest struct {
	name     string
	a, b     Interval
	p        int
	expected Decision
}

var unifiedOverlapTests = []unifiedOverlapTest{
	{
		name:     "disjoint rational intervals report No",
		a:        NewInterval(rat(1, 1), rat(2, 1)),
		b:        NewInterval(rat(3, 1), rat(4, 1)),
		p:        -100,
		expected: No,
	},
	{
		name:     "overlapping rational intervals report Yes",
		a:        NewInterval(rat(1, 1), rat(3, 1)),
		b:        NewInterval(rat(2, 1), rat(4, 1)),
		p:        -100,
		expected: Yes,
	},
	{
		name:     "touching-at-a-point intervals report Undecided at budget",
		a:        NewInterval(rat(1, 1), rat(2, 1)),
		b:        NewInterval(rat(2, 1), rat(3, 1)),
		p:        -100,
		expected: Undecided,
	},
}

func TestUnifiedIntervalOverlap(t *testing.T) {
	for _, test := range unifiedOverlapTests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.a.Overlap(test.b, test.p); got != test.expected {
				t.Errorf("Overlap = %v, want %v", got, test.expected)
			}
		})
	}
}
