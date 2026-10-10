package unified

import (
	"slices"
	"testing"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type linSpaceTest struct {
	name     string
	start    *Real
	stop     *Real
	n        int
	expected []*Real
	wantErr  error
}

var linSpaceTests = []linSpaceTest{
	{
		name:     "exact-rational endpoints emit n points with stop last",
		start:    rat(0, 1),
		stop:     rat(1, 1),
		n:        5,
		expected: []*Real{rat(0, 1), rat(1, 4), rat(1, 2), rat(3, 4), rat(1, 1)},
	},
	{
		name:     "irrational stop is included exactly as the last point",
		start:    rat(0, 1),
		stop:     irr(constructive.Sqrt2()),
		n:        3,
		expected: []*Real{rat(0, 1), New(constructive.Sqrt2(), rational.New64(1, 2)), irr(constructive.Sqrt2())},
	},
	{
		name:     "zero count yields an empty sequence",
		start:    rat(0, 1),
		stop:     rat(1, 1),
		n:        0,
		expected: nil,
	},
	{
		name:     "unit count yields a single point at start",
		start:    rat(2, 1),
		stop:     rat(9, 1),
		n:        1,
		expected: []*Real{rat(2, 1)},
	},
	{
		name:    "negative count errors",
		start:   rat(0, 1),
		stop:    rat(1, 1),
		n:       -1,
		wantErr: ErrNegativeCount,
	},
}

func TestLinSpace(t *testing.T) {
	for _, test := range linSpaceTests {
		t.Run(test.name, func(t *testing.T) {
			seq, err := LinSpace(test.start, test.stop, test.n)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			got := slices.Collect(seq)
			require.Len(t, got, len(test.expected))
			for i := range test.expected {
				assertEqualAtPrecision(t, test.expected[i], got[i], -100)
			}
		})
	}
}

type rangeTest struct {
	name     string
	start    *Real
	stop     *Real
	step     *Real
	expected []*Real
	wantErr  error
}

var rangeTests = []rangeTest{
	{
		name:     "even divide excludes stop under the half-open contract",
		start:    rat(0, 1),
		stop:     rat(1, 1),
		step:     rat(1, 4),
		expected: []*Real{rat(0, 1), rat(1, 4), rat(1, 2), rat(3, 4)},
	},
	{
		name:     "non-integer quotient uses Ceil count",
		start:    rat(0, 1),
		stop:     rat(1, 1),
		step:     rat(3, 10),
		expected: []*Real{rat(0, 1), rat(3, 10), rat(6, 10), rat(9, 10)},
	},
	{
		name:     "start equal to stop yields an empty sequence",
		start:    rat(1, 1),
		stop:     rat(1, 1),
		step:     rat(1, 4),
		expected: nil,
	},
	{
		name:    "irrational operand errors",
		start:   rat(0, 1),
		stop:    irr(constructive.Sqrt2()),
		step:    rat(1, 10),
		wantErr: ErrInexactOperand,
	},
	{
		name:    "zero step errors",
		start:   rat(0, 1),
		stop:    rat(1, 1),
		step:    rat(0, 1),
		wantErr: ErrZeroStep,
	},
	{
		name:    "step pointing away from stop errors",
		start:   rat(0, 1),
		stop:    rat(1, 1),
		step:    rat(-1, 4),
		wantErr: ErrDirectionMismatch,
	},
}

func TestRange(t *testing.T) {
	for _, test := range rangeTests {
		t.Run(test.name, func(t *testing.T) {
			seq, err := Range(test.start, test.stop, test.step)
			if test.wantErr != nil {
				assert.ErrorIs(t, err, test.wantErr)
				return
			}

			require.NoError(t, err)
			got := slices.Collect(seq)
			require.Len(t, got, len(test.expected))
			for i := range test.expected {
				assertEqualAtPrecision(t, test.expected[i], got[i], -100)
			}
		})
	}
}
