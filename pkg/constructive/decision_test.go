package constructive

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cmpErrorTest struct {
	Name    string
	A       Real
	B       Real
	Timeout time.Duration
	Err     error
}

var cmpErrorTests = []cmpErrorTest{
	{
		Name:    "equal constructive values until the deadline",
		A:       Multiply(Sqrt2(), Sqrt2()),
		B:       FromInt(2),
		Timeout: 100 * time.Millisecond,
		Err:     context.DeadlineExceeded,
	},
	{
		Name: "nil left operand",
		A:    nil,
		B:    FromInt(2),
		Err:  ErrNilReal,
	},
	{
		Name: "nil right operand",
		A:    FromInt(2),
		B:    nil,
		Err:  ErrNilReal,
	},
	{
		Name: "leaf error",
		A:    &failingReal{err: errLeaf},
		B:    FromInt(2),
		Err:  errLeaf,
	},
}

func TestCmpErrors(t *testing.T) {
	for _, test := range cmpErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			ctx := testCtx
			if test.Timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(testCtx, test.Timeout)
				defer cancel()
			}

			got, err := Cmp(ctx, test.A, test.B)
			assert.Equal(t, test.Err, err)
			assert.Equal(t, 0, got)
		})
	}
}

func TestCmpCancelledContext(t *testing.T) {
	got, err := Cmp(cancelledContext(), FromInt(1), FromInt(2))
	assert.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 0, got)
}

type signErrorTest struct {
	Name  string
	Value Real
	Err   error
}

var signErrorTests = []signErrorTest{
	{
		Name:  "integer zero is undecided",
		Value: FromInt(0),
		Err:   ErrUndecided,
	},
	{
		Name:  "nil value",
		Value: nil,
		Err:   ErrNilReal,
	},
	{
		Name:  "leaf error",
		Value: &failingReal{err: errLeaf},
		Err:   errLeaf,
	},
}

func TestSignErrors(t *testing.T) {
	for _, test := range signErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := Sign(testCtx, test.Value)
			assert.Equal(t, test.Err, err)
			assert.Equal(t, 0, got)
		})
	}
}

type preciseNilTest struct {
	Name string
	Run  func() (int, error)
}

var preciseNilTests = []preciseNilTest{
	{
		Name: "PreciseCmp nil left",
		Run:  func() (int, error) { return PreciseCmp(testCtx, nil, FromInt(1), -10) },
	},
	{
		Name: "PreciseCmp nil right",
		Run:  func() (int, error) { return PreciseCmp(testCtx, FromInt(1), nil, -10) },
	},
	{
		Name: "PreciseSign nil",
		Run:  func() (int, error) { return PreciseSign(testCtx, nil, -10) },
	},
}

func TestPreciseNilOperands(t *testing.T) {
	for _, test := range preciseNilTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := test.Run()
			assert.Equal(t, ErrNilReal, err)
			assert.Equal(t, 0, got)
		})
	}
}

type preciseSignTest struct {
	Name     string
	Value    Real
	P        int
	Expected int
}

var preciseSignTests = []preciseSignTest{
	{Name: "positive", Value: FromInt(3), P: -50, Expected: 1},
	{Name: "negative", Value: FromInt(-3), P: -50, Expected: -1},
	{Name: "zero", Value: FromInt(0), P: -50, Expected: 0},
	{Name: "irrational positive", Value: Sqrt2(), P: -50, Expected: 1},
	{Name: "tiny value is zero at coarse precision", Value: ShiftRight(One(), 100), P: -10, Expected: 0},
}

func TestPreciseSign(t *testing.T) {
	for _, test := range preciseSignTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := PreciseSign(testCtx, test.Value, test.P)
			assert.NoError(t, err)
			assert.Equal(t, test.Expected, got)
		})
	}
}

type roundingErrorTest struct {
	Name  string
	Run   func(ctx context.Context, c Real) (Real, error)
	Value Real
	Ctx   context.Context
	Err   error
}

type roundingFunc func(ctx context.Context, c Real, p int) (Real, error)

var roundingFuncs = map[string]roundingFunc{
	"Floor":       Floor,
	"Ceil":        Ceil,
	"Round":       Round,
	"RoundToEven": RoundToEven,
}

var roundingErrorTests = func() []roundingErrorTest {
	var tests []roundingErrorTest

	for name, fn := range roundingFuncs {
		run := func(ctx context.Context, c Real) (Real, error) { return fn(ctx, c, -20) }

		tests = append(tests,
			roundingErrorTest{
				Name:  name + " leaf error",
				Run:   run,
				Value: &failingReal{err: errLeaf},
				Ctx:   testCtx,
				Err:   errLeaf,
			},
			roundingErrorTest{
				Name:  name + " cancelled context",
				Run:   run,
				Value: Sqrt2(),
				Ctx:   cancelledContext(),
				Err:   context.Canceled,
			},
		)
	}

	return tests
}()

func TestRoundingErrors(t *testing.T) {
	for _, test := range roundingErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := test.Run(test.Ctx, test.Value)
			assert.ErrorIs(t, err, test.Err)
			assert.Nil(t, got)
		})
	}
}

type membershipErrorTest struct {
	Name     string
	Ctx      context.Context
	Run      func(ctx context.Context) (int, error)
	Expected int
	Err      error
}

var membershipErrorTests = []membershipErrorTest{
	{
		Name: "Locate leaf error",
		Ctx:  testCtx,
		Run: func(ctx context.Context) (int, error) {
			got, err := Locate(ctx, New(FromInt(0), FromInt(1)), &failingReal{err: errLeaf}, -20)
			return int(got), err
		},
		Expected: int(Indeterminate),
		Err:      errLeaf,
	},
	{
		Name: "Locate cancelled context",
		Ctx:  cancelledContext(),
		Run: func(ctx context.Context) (int, error) {
			got, err := Locate(ctx, New(FromInt(0), FromInt(1)), Sqrt2(), -20)
			return int(got), err
		},
		Expected: int(Indeterminate),
		Err:      context.Canceled,
	},
	{
		Name: "IsEmpty leaf error",
		Ctx:  testCtx,
		Run: func(ctx context.Context) (int, error) {
			got, err := IsEmpty(ctx, New(FromInt(0), &failingReal{err: errLeaf}), -20)
			return int(got), err
		},
		Expected: int(Undecided),
		Err:      errLeaf,
	},
	{
		Name: "IsEmpty cancelled context",
		Ctx:  cancelledContext(),
		Run: func(ctx context.Context) (int, error) {
			got, err := IsEmpty(ctx, New(FromInt(0), Sqrt2()), -20)
			return int(got), err
		},
		Expected: int(Undecided),
		Err:      context.Canceled,
	},
	{
		Name: "Overlap leaf error",
		Ctx:  testCtx,
		Run: func(ctx context.Context) (int, error) {
			got, err := Overlap(ctx, New(FromInt(0), &failingReal{err: errLeaf}), New(FromInt(0), FromInt(1)), -20)
			return int(got), err
		},
		Expected: int(Undecided),
		Err:      errLeaf,
	},
	{
		Name: "Overlap cancelled context",
		Ctx:  cancelledContext(),
		Run: func(ctx context.Context) (int, error) {
			got, err := Overlap(ctx, New(FromInt(0), Sqrt2()), New(FromInt(0), FromInt(1)), -20)
			return int(got), err
		},
		Expected: int(Undecided),
		Err:      context.Canceled,
	},
}

func TestMembershipErrors(t *testing.T) {
	for _, test := range membershipErrorTests {
		t.Run(test.Name, func(t *testing.T) {
			got, err := test.Run(test.Ctx)
			assert.ErrorIs(t, err, test.Err)
			assert.Equal(t, test.Expected, got)
		})
	}
}
