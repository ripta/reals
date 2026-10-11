package constructive

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// deferredBuildTest builds one lazily reduced function over the given argument.
type deferredBuildTest struct {
	Name  string
	Build func(arg Real) Real
}

var deferredBuildTests = []deferredBuildTest{
	{Name: "Exp", Build: Exp},
	{Name: "Ln", Build: Ln},
	{Name: "Cosine", Build: Cosine},
	{Name: "Sine", Build: Sine},
	{Name: "Tangent", Build: Tangent},
	{Name: "Arctangent", Build: Arctangent},
	{Name: "Gamma", Build: Gamma},
	{Name: "Pow, base", Build: func(arg Real) Real { return Pow(arg, FromInt(2)) }},
	{Name: "Pow, exponent", Build: func(arg Real) Real { return Pow(FromInt(2), arg) }},
	{Name: "Pow10", Build: Pow10},
}

func TestDeferredConstructionDoesNotApproximate(t *testing.T) {
	for _, test := range deferredBuildTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &failingReal{err: errLeaf}

			node := test.Build(leaf)
			assert.NotNil(t, node)
			assert.Zero(t, leaf.calls)
			assert.Nil(t, leaf.tracker().latest())
			assert.Nil(t, node.tracker().latest())
			assert.NotEmpty(t, AsConstruction(node))
			assert.Zero(t, leaf.calls)
		})
	}
}

func TestDeferredFirstApproximationSurfacesLeafError(t *testing.T) {
	for _, test := range deferredBuildTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &failingReal{err: errLeaf}
			node := test.Build(leaf)

			_, err := Approximate(testCtx, node, -20)
			assert.Equal(t, errLeaf, err)
			assert.Positive(t, leaf.calls)
			assert.Nil(t, node.tracker().latest())
		})
	}
}

func TestDeferredCancelledContextRunsNothing(t *testing.T) {
	for _, test := range deferredBuildTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &failingReal{err: errLeaf}
			node := test.Build(leaf)

			_, err := Approximate(cancelledContext(), node, -20)
			assert.ErrorIs(t, err, context.Canceled)
			assert.Zero(t, leaf.calls)
		})
	}
}

func TestDeferredFailedFirstApproximationIsRetried(t *testing.T) {
	for _, test := range deferredBuildTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &flakyReal{failures: 1}
			node := test.Build(leaf)

			_, err := Approximate(testCtx, node, -20)
			assert.ErrorIs(t, err, errLeaf)
			assert.Nil(t, node.tracker().latest())

			got, err := Approximate(testCtx, node, -20)
			assert.NoError(t, err)

			want := mustApproximate(t, test.Build(FromInt(7)), -20)
			assert.Zero(t, want.Cmp(got))
		})
	}
}

type reductionMemoTest struct {
	Name  string
	Build func(arg Real) Real
}

var reductionMemoTests = []reductionMemoTest{
	{Name: "Exp", Build: Exp},
	{Name: "Ln", Build: Ln},
	{Name: "Cosine", Build: Cosine},
	{Name: "Arctangent", Build: Arctangent},
	{Name: "Gamma", Build: Gamma},
}

func TestDeferredReductionIsMemoizedOnlyOnSuccess(t *testing.T) {
	for _, test := range reductionMemoTests {
		t.Run(test.Name, func(t *testing.T) {
			leaf := &flakyReal{failures: 1}
			node, ok := test.Build(leaf).(*deferredReduction)
			assert.True(t, ok)
			assert.Nil(t, node.reduced.Load())

			_, err := Approximate(testCtx, node, -20)
			assert.ErrorIs(t, err, errLeaf)
			assert.Nil(t, node.reduced.Load())

			_, err = Approximate(testCtx, node, -20)
			assert.NoError(t, err)
			assert.NotNil(t, node.reduced.Load())

			reduced := node.reduced.Load()

			_, err = Approximate(testCtx, node, -40)
			assert.NoError(t, err)
			assert.Same(t, reduced, node.reduced.Load())
		})
	}
}

type constructionTest struct {
	Name     string
	Build    func() Real
	Before   string
	AfterHas string
}

var constructionTests = []constructionTest{
	{
		Name:     "Exp of a small argument",
		Build:    func() Real { return Exp(FromInt(0)) },
		Before:   "Exp(Int(0))",
		AfterHas: "Pow(E, Int(0))",
	},
	{
		Name:     "Ln of a mid-range argument",
		Build:    func() Real { return Ln(FromInt(10)) },
		Before:   "Ln(Int(10))",
		AfterHas: "Ln(",
	},
}

func TestDeferredConstructionText(t *testing.T) {
	for _, test := range constructionTests {
		t.Run(test.Name, func(t *testing.T) {
			node := test.Build()
			assert.Equal(t, test.Before, AsConstruction(node))

			mustApproximate(t, node, -10)
			assert.Contains(t, AsConstruction(node), test.AfterHas)
		})
	}
}

type lnDomainTest struct {
	Name  string
	Value Real
}

var lnDomainTests = []lnDomainTest{
	{Name: "negative integer", Value: FromInt(-1)},
	{Name: "negative fraction", Value: FromRat(-1, 100)},
	{Name: "negative irrational", Value: Negate(Sqrt2())},
	{Name: "large negative", Value: FromInt(-1000)},
	{Name: "tiny negative", Value: FromRat(-1, 1000000)},
}

func TestLnNonPositive(t *testing.T) {
	for _, test := range lnDomainTests {
		t.Run(test.Name, func(t *testing.T) {
			node := Ln(test.Value)
			assert.NotNil(t, node)

			_, err := Approximate(testCtx, node, -20)
			assert.ErrorIs(t, err, ErrNonPositive)
			assert.Nil(t, node.tracker().latest())
		})
	}
}

type reductionValueTest struct {
	Name     string
	Build    func() Real
	Expected map[int]string
}

var reductionValueTests = []reductionValueTest{
	{Name: "exp(1)", Build: func() Real { return Exp(FromInt(1)) }, Expected: map[int]string{
		10: "2.7182818285",
		60: "2.718281828459045235360287471352662497757247093699959574966968",
	}},
	{Name: "exp(-3)", Build: func() Real { return Exp(FromInt(-3)) }, Expected: map[int]string{
		10: "0.0497870684",
		60: "0.049787068367863942979342415650061776631699592188423215567628",
	}},
	{Name: "exp(10)", Build: func() Real { return Exp(FromInt(10)) }, Expected: map[int]string{
		10: "22026.4657948067",
		60: "22026.465794806716516957900645284244366353512618556781074235426355",
	}},
	{Name: "ln(2)", Build: func() Real { return Ln(FromInt(2)) }, Expected: map[int]string{
		10: "0.6931471806",
		60: "0.693147180559945309417232121458176568075500134360255254120680",
	}},
	{Name: "ln(10)", Build: func() Real { return Ln(FromInt(10)) }, Expected: map[int]string{
		10: "2.3025850930",
		60: "2.302585092994045684017991454684364207601101488628772976033328",
	}},
	{Name: "ln(1/10)", Build: func() Real { return Ln(FromRat(1, 10)) }, Expected: map[int]string{
		10: "-2.3025850930",
		60: "-2.302585092994045684017991454684364207601101488628772976033328",
	}},
	{Name: "ln(1000)", Build: func() Real { return Ln(FromInt(1000)) }, Expected: map[int]string{
		10: "6.9077552790",
		60: "6.907755278982137052053974364053092622803304465886318928099984",
	}},
	{Name: "cos(1)", Build: func() Real { return Cosine(FromInt(1)) }, Expected: map[int]string{
		10: "0.5403023059",
		60: "0.540302305868139717400936607442976603732310420617922227670097",
	}},
	{Name: "cos(5)", Build: func() Real { return Cosine(FromInt(5)) }, Expected: map[int]string{
		10: "0.2836621855",
		60: "0.283662185463226264466639171513557308334422592252215944930359",
	}},
	{Name: "cos(100)", Build: func() Real { return Cosine(FromInt(100)) }, Expected: map[int]string{
		10: "0.8623188723",
		60: "0.862318872287683934101938513950842535510084008535510829280162",
	}},
	{Name: "sin(2)", Build: func() Real { return Sine(FromInt(2)) }, Expected: map[int]string{
		10: "0.9092974268",
		60: "0.909297426825681695396019865911744842702254971447890268378973",
	}},
	{Name: "tan(1)", Build: func() Real { return Tangent(FromInt(1)) }, Expected: map[int]string{
		10: "1.5574077247",
		60: "1.557407724654902230506974807458360173087250772381520038383947",
	}},
	{Name: "atan(1/2)", Build: func() Real { return Arctangent(FromRat(1, 2)) }, Expected: map[int]string{
		10: "0.4636476090",
		60: "0.463647609000806116214256231461214402028537054286120263810933",
	}},
	{Name: "atan(5)", Build: func() Real { return Arctangent(FromInt(5)) }, Expected: map[int]string{
		10: "1.3734007669",
		60: "1.373400766945015860861271926444961148650999595899700808969783",
	}},
	{Name: "atan(-5)", Build: func() Real { return Arctangent(FromInt(-5)) }, Expected: map[int]string{
		10: "-1.3734007669",
		60: "-1.373400766945015860861271926444961148650999595899700808969783",
	}},
	{Name: "atan(100)", Build: func() Real { return Arctangent(FromInt(100)) }, Expected: map[int]string{
		10: "1.5607966601",
		60: "1.560796660108231381024981575430471893537215347143176270859533",
	}},
	{Name: "gamma(7/2)", Build: func() Real { return Gamma(FromRat(7, 2)) }, Expected: map[int]string{
		10: "3.3233509704",
		60: "3.323350970447842551184064031264647217745405230229475865400890",
	}},
	{Name: "gamma(-1/2)", Build: func() Real { return Gamma(FromRat(-1, 2)) }, Expected: map[int]string{
		10: "-3.5449077018",
		60: "-3.544907701811032054596334966682290365595098912244774256427616",
	}},
	{Name: "pow(2, 1/2)", Build: func() Real { return Pow(FromInt(2), FromRat(1, 2)) }, Expected: map[int]string{
		10: "1.4142135624",
		60: "1.414213562373095048801688724209698078569671875376948073176680",
	}},
	{Name: "pow10(-3)", Build: func() Real { return Pow10(FromInt(-3)) }, Expected: map[int]string{
		10: "0.0010000000",
		60: "0.001000000000000000000000000000000000000000000000000000000000",
	}},
}

type tinyOperandTest struct {
	Name     string
	Build    func() Real
	Digits   int
	Expected string
}

var tinyOperandTests = []tinyOperandTest{
	{
		Name:     "ln of a tiny rational",
		Build:    func() Real { return Ln(FromRat(1, 1000000000)) },
		Digits:   30,
		Expected: "-20.723265836946411156161923092159",
	},
	{
		Name:     "ln of a tiny irrational",
		Build:    func() Real { return Ln(Multiply(Sqrt2(), FromRat(1, 1000000000000))) },
		Digits:   30,
		Expected: "-27.284447525648575553507281395483",
	},
	{
		Name:     "inverse of a tiny rational",
		Build:    func() Real { return Inverse(FromRat(1, 1000000000)) },
		Digits:   5,
		Expected: "1000000000.00000",
	},
	{
		Name:     "inverse of a tiny negative rational",
		Build:    func() Real { return Inverse(FromRat(-1, 1000000000)) },
		Digits:   5,
		Expected: "-1000000000.00000",
	},
}

func TestTinyOperands(t *testing.T) {
	for _, test := range tinyOperandTests {
		t.Run(test.Name, func(t *testing.T) {
			assert.Equal(t, test.Expected, mustText(t, test.Build(), test.Digits, 10))
		})
	}
}

type undecidedZeroTest struct {
	Name  string
	Build func() Real
}

var undecidedZeroTests = []undecidedZeroTest{
	{Name: "inverse of zero", Build: func() Real { return Inverse(FromInt(0)) }},
	{Name: "ln of zero", Build: func() Real { return Ln(FromInt(0)) }},
}

func TestExactZeroIsUndecided(t *testing.T) {
	for _, test := range undecidedZeroTests {
		t.Run(test.Name, func(t *testing.T) {
			_, err := Approximate(testCtx, test.Build(), -20)
			assert.ErrorIs(t, err, ErrUndecided)
		})
	}
}

func TestConstructiveZeroInverseHonorsDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(testCtx, 100*time.Millisecond)
	defer cancel()

	zero := Subtract(Sqrt2(), Sqrt2())

	_, err := Approximate(ctx, Inverse(zero), -20)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestDeferredValuesMatchEagerReduction(t *testing.T) {
	for _, test := range reductionValueTests {
		t.Run(test.Name, func(t *testing.T) {
			node := test.Build()

			for _, digits := range []int{10, 60} {
				assert.Equal(t, test.Expected[digits], mustText(t, node, digits, 10))
			}

			assert.Equal(t, test.Expected[10], mustText(t, node, 10, 10))
		})
	}
}
