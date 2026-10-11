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

func benchTextOnce(b *testing.B, c Real) string {
	b.Helper()

	return mustText(b, c, 1000, 10)
}

func benchText(b *testing.B, build func() Real) {
	b.Helper()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if got := benchTextOnce(b, build()); len(got) == 0 {
			b.Fatal("empty text")
		}
	}
}

func BenchmarkTextPi1000(b *testing.B) {
	benchText(b, freshPi)
}

func BenchmarkTextDeepExpLn1000(b *testing.B) {
	benchText(b, deepExpLn)
}
