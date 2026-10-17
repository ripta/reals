package unified

import (
	"context"
	"testing"
	"time"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
	"github.com/stretchr/testify/assert"
)

// constructiveZero is zero that no finite approximation can tell from a tiny
// value, and that the structural zero check does not recognize.
func constructiveZero() *Real {
	return irr(constructive.Subtract(constructive.Sqrt2(), constructive.Sqrt2()))
}

func cancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

type partialContextTest struct {
	name string
	run  func(ctx context.Context) (*Real, error)
}

var deadlinePartialTests = []partialContextTest{
	{
		name: "Sqrt of a constructive zero",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Sqrt(ctx) },
	},
	{
		name: "Ln of a constructive zero",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Ln(ctx) },
	},
}

func TestPartialFunctionsStopAtDeadline(t *testing.T) {
	for _, test := range deadlinePartialTests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
			defer cancel()

			result, err := test.run(ctx)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, context.DeadlineExceeded)
		})
	}
}

var cancelledPartialTests = []partialContextTest{
	{
		name: "Log10 of a constructive zero",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Log10(ctx) },
	},
	{
		name: "Log2 of a constructive zero",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Log2(ctx) },
	},
	{
		name: "Log of a constructive zero argument",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Log(ctx, Two()) },
	},
	{
		name: "Log with a constructive zero base",
		run:  func(ctx context.Context) (*Real, error) { return Two().Log(ctx, constructiveZero()) },
	},
	{
		name: "Pow of a constructive zero base to a fractional power",
		run:  func(ctx context.Context) (*Real, error) { return constructiveZero().Pow(ctx, Half()) },
	},
	{
		name: "Asin of a constructive one",
		run:  func(ctx context.Context) (*Real, error) { return oneButNotStructurally().Asin(ctx) },
	},
	{
		name: "Acos of a constructive one",
		run:  func(ctx context.Context) (*Real, error) { return oneButNotStructurally().Acos(ctx) },
	},
	{
		name: "Atan2 with a constructive zero x",
		run:  func(ctx context.Context) (*Real, error) { return One().Atan2(ctx, constructiveZero()) },
	},
}

func TestPartialFunctionsSurfaceCancelledContext(t *testing.T) {
	for _, test := range cancelledPartialTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.run(cancelledContext())
			assert.Nil(t, result)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

type crossLayerSentinelTest struct {
	name    string
	run     func() error
	wantErr error
	altErr  error
}

var crossLayerSentinelTests = []crossLayerSentinelTest{
	{
		name: "unified Sqrt of a negative rational",
		run: func() error {
			_, err := NegativeOne().Sqrt(testCtx)
			return err
		},
		wantErr: ErrNegative,
		altErr:  constructive.ErrNegative,
	},
	{
		name: "unified Ln of a negative rational",
		run: func() error {
			_, err := NegativeOne().Ln(testCtx)
			return err
		},
		wantErr: ErrNonPositive,
		altErr:  constructive.ErrNonPositive,
	},
	{
		name: "constructive Sqrt of a negative, approximated through FormattedString",
		run: func() error {
			_, err := irr(constructive.Sqrt(constructive.FromInt(-2))).FormattedString(testCtx, 10, 10)
			return err
		},
		wantErr: ErrNegative,
		altErr:  constructive.ErrNegative,
	},
	{
		name: "constructive Ln of a negative, approximated through FormattedString",
		run: func() error {
			_, err := irr(constructive.Ln(constructive.FromInt(-2))).FormattedString(testCtx, 10, 10)
			return err
		},
		wantErr: ErrNonPositive,
		altErr:  constructive.ErrNonPositive,
	},
	{
		name: "unified Polygamma of a negative order",
		run: func() error {
			_, err := One().Polygamma(-1)
			return err
		},
		wantErr: ErrNegativeOrder,
		altErr:  constructive.ErrNegativeOrder,
	},
	{
		name: "constructive Polygamma of a negative order, approximated through FormattedString",
		run: func() error {
			_, err := irr(constructive.Polygamma(-1, constructive.Two())).FormattedString(testCtx, 10, 10)
			return err
		},
		wantErr: ErrNegativeOrder,
		altErr:  constructive.ErrNegativeOrder,
	},
}

func TestDomainSentinelsAreSharedAcrossLayers(t *testing.T) {
	assert.Same(t, constructive.ErrNegative, ErrNegative)
	assert.Same(t, constructive.ErrNonPositive, ErrNonPositive)
	assert.Same(t, constructive.ErrNegativeOrder, ErrNegativeOrder)

	for _, test := range crossLayerSentinelTests {
		t.Run(test.name, func(t *testing.T) {
			err := test.run()
			assert.ErrorIs(t, err, test.wantErr)
			assert.ErrorIs(t, err, test.altErr)
		})
	}
}

func TestAcosPropagatesNonDomainErrors(t *testing.T) {
	result, err := oneButNotStructurally().Acos(cancelledContext())
	assert.Nil(t, result)
	assert.ErrorIs(t, err, context.Canceled)
	assert.NotErrorIs(t, err, ErrOutsideUnitInterval)

	result, err = rat(2, 1).Acos(testCtx)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrOutsideUnitInterval)
}

// oneButNotStructurally is a value equal to one that the unit-endpoint check
// does not recognize, so the domain check approximates it.
func oneButNotStructurally() *Real {
	sqrt2 := constructive.Sqrt2()

	return irr(constructive.Add(sqrt2, constructive.Subtract(constructive.One(), sqrt2)))
}

type signTest struct {
	name     string
	input    *Real
	expected int
}

var signTests = []signTest{
	{name: "structural zero", input: Zero(), expected: 0},
	{name: "positive rational", input: rat(3, 2), expected: 1},
	{name: "negative rational", input: rat(-3, 2), expected: -1},
	{name: "positive irrational", input: irr(constructive.Sqrt2()), expected: 1},
	{name: "negative irrational", input: irr(constructive.Negate(constructive.Pi())), expected: -1},
	{name: "irrational scaled by a negative rational", input: New(constructive.Pi(), rational.New64(-1, 3)), expected: -1},
}

func TestSign(t *testing.T) {
	for _, test := range signTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.input.Sign(testCtx)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

var signRationalFastPathTests = []signTest{
	{name: "structural zero", input: Zero(), expected: 0},
	{name: "positive rational", input: rat(3, 2), expected: 1},
	{name: "negative rational", input: rat(-3, 2), expected: -1},
}

func TestSignRationalFastPathIgnoresCancelledContext(t *testing.T) {
	for _, test := range signRationalFastPathTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.input.Sign(cancelledContext())
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestSignOfConstructiveZeroStopsAtDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := constructiveZero().Sign(ctx)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestSignOfIrrationalSurfacesCancelledContext(t *testing.T) {
	_, err := irr(constructive.Sqrt2()).Sign(cancelledContext())
	assert.ErrorIs(t, err, context.Canceled)
}

type cmpTest struct {
	name     string
	a, b     *Real
	expected int
}

var cmpTests = []cmpTest{
	{name: "equal rationals", a: rat(1, 2), b: rat(2, 4), expected: 0},
	{name: "lesser rational", a: rat(1, 3), b: rat(1, 2), expected: -1},
	{name: "greater rational", a: rat(2, 3), b: rat(1, 2), expected: 1},
	{name: "both structural zero", a: Zero(), b: Zero(), expected: 0},
	{name: "structural zero below irrational", a: Zero(), b: irr(constructive.Sqrt2()), expected: -1},
	{name: "irrational above structural zero", a: irr(constructive.Sqrt2()), b: Zero(), expected: 1},
	{name: "irrational below rational", a: irr(constructive.Sqrt2()), b: rat(3, 2), expected: -1},
	{name: "rational below irrational", a: rat(3, 2), b: irr(constructive.Pi()), expected: -1},
	{name: "irrational above irrational", a: irr(constructive.Pi()), b: irr(constructive.Sqrt2()), expected: 1},
}

func TestCmp(t *testing.T) {
	for _, test := range cmpTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.a.Cmp(testCtx, test.b)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

var cmpRationalFastPathTests = []cmpTest{
	{name: "equal rationals", a: rat(1, 2), b: rat(2, 4), expected: 0},
	{name: "lesser rational", a: rat(1, 3), b: rat(1, 2), expected: -1},
	{name: "both structural zero", a: Zero(), b: Zero(), expected: 0},
	{name: "rational against structural zero", a: rat(-1, 2), b: Zero(), expected: -1},
}

func TestCmpRationalFastPathIgnoresCancelledContext(t *testing.T) {
	for _, test := range cmpRationalFastPathTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.a.Cmp(cancelledContext(), test.b)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestCmpOfEqualConstructiveValuesStopsAtDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := irr(constructive.Pi()).Cmp(ctx, irr(constructive.Pi()))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestCmpOfIrrationalsSurfacesCancelledContext(t *testing.T) {
	_, err := irr(constructive.Pi()).Cmp(cancelledContext(), irr(constructive.Sqrt2()))
	assert.ErrorIs(t, err, context.Canceled)
}

type cancelledRoundingTest struct {
	name string
	run  func(ctx context.Context) (*Real, error)
}

var cancelledRoundingTests = []cancelledRoundingTest{
	{name: "Floor", run: func(ctx context.Context) (*Real, error) { return Pi().Floor(ctx, -100) }},
	{name: "Ceil", run: func(ctx context.Context) (*Real, error) { return Pi().Ceil(ctx, -100) }},
	{name: "Round", run: func(ctx context.Context) (*Real, error) { return Pi().Round(ctx, -100) }},
	{name: "RoundToEven", run: func(ctx context.Context) (*Real, error) { return Pi().RoundToEven(ctx, -100) }},
	{name: "Min", run: func(ctx context.Context) (*Real, error) { return Pi().Min(ctx, Two(), -100) }},
	{name: "Max", run: func(ctx context.Context) (*Real, error) { return Pi().Max(ctx, Two(), -100) }},
}

func TestRoundingSurfacesCancelledContext(t *testing.T) {
	for _, test := range cancelledRoundingTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.run(cancelledContext())
			assert.Nil(t, result)
			assert.ErrorIs(t, err, context.Canceled)
		})
	}
}

var cancelledRationalRoundingTests = []cancelledRoundingTest{
	{name: "Floor", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).Floor(ctx, -100) }},
	{name: "Ceil", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).Ceil(ctx, -100) }},
	{name: "Round", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).Round(ctx, -100) }},
	{name: "RoundToEven", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).RoundToEven(ctx, -100) }},
	{name: "Min", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).Min(ctx, Two(), -100) }},
	{name: "Max", run: func(ctx context.Context) (*Real, error) { return rat(5, 2).Max(ctx, Two(), -100) }},
}

func TestRationalRoundingIgnoresCancelledContext(t *testing.T) {
	for _, test := range cancelledRationalRoundingTests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.run(cancelledContext())
			assert.NoError(t, err)
			assert.NotNil(t, result)
		})
	}
}

type formattedStringTest struct {
	name     string
	input    *Real
	digits   int
	expected string
}

var formattedStringTests = []formattedStringTest{
	{name: "rational", input: rat(1, 4), digits: 4, expected: "0.2500"},
	{name: "integer", input: rat(3, 1), digits: 2, expected: "3.00"},
}

func TestFormattedString(t *testing.T) {
	for _, test := range formattedStringTests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.input.FormattedString(testCtx, test.digits, 10)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestFormattedStringSurfacesCancelledContext(t *testing.T) {
	got, err := irr(constructive.Pi()).FormattedString(cancelledContext(), 20, 10)
	assert.Empty(t, got)
	assert.ErrorIs(t, err, context.Canceled)
}
