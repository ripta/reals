package constructive

import "testing"

// freshPi builds a new Pi tree on every call. The shared Pi caches its
// approximations, which would turn the benchmark into a cache lookup.
func freshPi() Real {
	m1 := Multiply(FromInt(6), newIntegralArctan(FromInt(8)))
	m2 := Multiply(FromInt(2), newIntegralArctan(FromInt(57)))
	m3 := newIntegralArctan(FromInt(239))

	return Multiply(FromInt(4), Add(m1, Add(m2, m3)))
}

// deepExpLn builds a fresh nest of Exp and Ln over an irrational argument.
func deepExpLn() Real {
	x := Add(FromInt(2), Sqrt(FromInt(2)))
	for i := 0; i < 3; i++ {
		x = Ln(Exp(x))
	}

	return x
}

func benchTextOnce(b *testing.B, c Real, digits int) string {
	b.Helper()

	return mustText(b, c, digits, 10)
}

func benchText(b *testing.B, digits int, build func() Real) {
	b.Helper()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := benchTextOnce(b, build(), digits); len(got) == 0 {
			b.Fatal("empty text")
		}
	}
}

func BenchmarkTextPi1000(b *testing.B) {
	benchText(b, 1000, freshPi)
}

func BenchmarkTextDeepExpLn1000(b *testing.B) {
	benchText(b, 1000, deepExpLn)
}

// digammaBench builds a fresh digamma at a non-trivial rational argument.
func digammaBench() Real {
	return Digamma(FromRat(7, 2))
}

// polygammaBench builds a fresh polygamma of order 3 at the same argument.
func polygammaBench() Real {
	return Polygamma(3, FromRat(7, 2))
}

// reset empties the shared Bernoulli cache, so the next approximation pays for
// filling it. Only benchmarks use it: in use, the cache fills once and stays.
func (b *bernoulliEvens) reset() {
	b.values.Store(nil)
}

// benchDigamma runs digamma over a fresh tree per iteration. A cold run empties the Bernoulli
// cache first, so it measures the one-time cost of filling it. A warm run fills the cache
// before timing starts, so it measures the cost of an approximation that finds the numbers
// already there.
func benchDigamma(b *testing.B, digits int, cold bool) {
	b.Helper()

	benchSeries(b, digammaBench, digits, cold)
}

// benchPolygamma is benchDigamma for a polygamma of order 3.
func benchPolygamma(b *testing.B, digits int, cold bool) {
	b.Helper()

	benchSeries(b, polygammaBench, digits, cold)
}

func benchSeries(b *testing.B, build func() Real, digits int, cold bool) {
	b.Helper()

	if !cold {
		benchTextOnce(b, build(), digits)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if cold {
			bernoulliCache.reset()
		}

		if got := benchTextOnce(b, build(), digits); len(got) == 0 {
			b.Fatal("empty text")
		}
	}
}

func BenchmarkDigamma50(b *testing.B) {
	benchDigamma(b, 50, false)
}

func BenchmarkDigamma1000(b *testing.B) {
	benchDigamma(b, 1000, false)
}

func BenchmarkDigammaCold50(b *testing.B) {
	benchDigamma(b, 50, true)
}

func BenchmarkDigammaCold1000(b *testing.B) {
	benchDigamma(b, 1000, true)
}

func BenchmarkPolygamma3At50(b *testing.B) {
	benchPolygamma(b, 50, false)
}

func BenchmarkPolygamma3At1000(b *testing.B) {
	benchPolygamma(b, 1000, false)
}

func BenchmarkPolygamma3ColdAt50(b *testing.B) {
	benchPolygamma(b, 50, true)
}

func BenchmarkPolygamma3ColdAt1000(b *testing.B) {
	benchPolygamma(b, 1000, true)
}
