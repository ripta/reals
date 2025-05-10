package constructive

import (
	"math"
	"math/big"
)

//var (
//	bigZero  = big.NewInt(0)
//	bigOne   = big.NewInt(1)
//	bigTwo   = big.NewInt(2)
//	bigThree = big.NewInt(3)
//	bigFour  = big.NewInt(4)
//	bigFive  = big.NewInt(5)
//	bigSix   = big.NewInt(6)
//	bigSeven = big.NewInt(7)
//	bigEight = big.NewInt(8)
//	bigNine  = big.NewInt(9)
//	bigTen   = big.NewInt(10)
//
//	bigNegOne = big.NewInt(-1)
//)

// bigAdd adds two big integers.
func bigAdd(a, b *big.Int) *big.Int {
	return new(big.Int).Add(a, b)
}

// bigSub subtracts two big integers.
func bigSub(a, b *big.Int) *big.Int {
	return new(big.Int).Sub(a, b)
}

// bigLsh left shifts a big integer by n bits.
func bigLsh(a *big.Int, n uint) *big.Int {
	return new(big.Int).Lsh(a, n)
}

// bigRsh right shifts a big integer by n bits.
func bigRsh(a *big.Int, n uint) *big.Int {
	return new(big.Int).Rsh(a, n)
}

// bigAbs returns the absolute value of a big integer.
func bigAbs(a *big.Int) *big.Int {
	return new(big.Int).Abs(a)
}

func bigBitAnd(a, b *big.Int) *big.Int {
	return new(big.Int).And(a, b)
}

// bigNeg returns the negation of a big integer.
func bigNeg(a *big.Int) *big.Int {
	return new(big.Int).Neg(a)
}

// bigMul multiplies two big integers.
func bigMul(a, b *big.Int) *big.Int {
	return new(big.Int).Mul(a, b)
}

// bigDiv divides two big integers.
func bigDiv(a, b *big.Int) *big.Int {
	return new(big.Int).Div(a, b)
}

// bigExp raises a big integer to the power of another big integer modulo m.
func bigExp(a, b, m *big.Int) *big.Int {
	return new(big.Int).Exp(a, b, m)
}

// boundLog2 calculates the base-2 logarithm of a number, rounded up.
func boundLog2(n int) int {
	return int(math.Ceil(math.Log2(math.Abs(float64(n)) + 1)))
}
