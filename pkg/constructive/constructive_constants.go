package constructive

import "sync"

var E = sync.OnceValue(func() Real {
	return newNamed("e", newPrescaledExponential(FromInt(1)))
})

var Ln2 = sync.OnceValue(func() Real {
	t1 := Multiply(FromInt(7), SimpleLn(Divide(FromInt(10), FromInt(9))))
	t2 := Multiply(FromInt(2), SimpleLn(Divide(FromInt(25), FromInt(24))))
	t3 := Multiply(FromInt(3), SimpleLn(Divide(FromInt(81), FromInt(80))))
	return newNamed("ln2", Add(Subtract(t1, t2), t3))
})

var Pi = sync.OnceValue(func() Real {
	// π = 4 * (6 * arctan(1/8) + 2 * arctan(1/57) + arctan(1/239))
	m1 := Multiply(FromInt(6), newIntegralArctan(FromInt(8)))
	m2 := Multiply(FromInt(2), newIntegralArctan(FromInt(57)))
	m3 := newIntegralArctan(FromInt(239))
	return newNamed("π", Multiply(FromInt(4), Add(m1, Add(m2, m3))))
})

var Phi = sync.OnceValue(func() Real {
	// φ = (1 + √5) / 2
	return newNamed("φ", Divide(Add(FromInt(1), Sqrt(FromInt(5))), FromInt(2)))
})

var Sqrt2 = sync.OnceValue(func() Real {
	return newNamed("√2", Sqrt(FromInt(2)))
})

var Zero = sync.OnceValue(func() Real {
	return newNamed("0", FromInt(0))
})

var One = sync.OnceValue(func() Real {
	return newNamed("1", FromInt(1))
})

var Two = sync.OnceValue(func() Real {
	return newNamed("2", FromInt(2))
})

var Ten = sync.OnceValue(func() Real {
	return newNamed("10", FromInt(10))
})
