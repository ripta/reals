package constructive

import (
	"fmt"
	"math"
	"math/big"
	"strings"
)

const IntSize = 32 << (^uint(0) >> 63) // 32 or 64

// IsPrecisionValid checks if the precision is within 4 bits of tolerance.
// That value depends on the size of the integer on our architecture, 32 or 64.
func IsPrecisionValid(p int) bool {
	return IsIntWithinBitTolerance(p, 4)
}

// IsIntWithinBitTolerance checks if the integer value is within the bit
// tolerance. A bit tolerance of 4 means that there must be at least 4 bits
// unused in the integer representation.
func IsIntWithinBitTolerance(value, tolerance int) bool {
	highBits := value >> (IntSize - 4)
	topBits := value >> (IntSize - 3)
	return (highBits ^ topBits) == 0
}

// Text converts a Real number to a string representation.
// The function takes a Real number, a non-negative decimal
// precision, and a radix (base) for the conversion.
func Text(c Real, dec, radix int) (text string) {
	defer func() {
		if err := recover(); err != nil {
			text = fmt.Sprintf("<undefined: %v>", err)
		}
	}()

	var sc Real
	if radix == 16 {
		sc = ShiftLeft(c, 4*dec)
	} else {
		sf := bigExp(big.NewInt(int64(radix)), big.NewInt(int64(dec)), nil)
		sc = Multiply(c, newInteger(sf))
	}

	si := Approximate(sc, 0)
	ss := bigAbs(si).Text(radix)

	out := ss
	if dec > 0 {
		if sl := len(ss); sl <= dec {
			ss = strings.Repeat("0", dec+1-sl) + ss
			sl = dec + 1
		}

		out = ss[:len(ss)-dec] + "." + ss[len(ss)-dec:]
	}

	if si.Sign() < 0 {
		out = "-" + out
	}
	return out
}

// Approximate computes the approximation of a Real number,
// given a precision p. When possible, the approximation is cached
// to save time on future calls.
func Approximate(c Real, p int) *big.Int {
	if !IsPrecisionValid(p) {
		return nil
	}

	t := c.tracker()
	if s, ok := t.Get(p); ok {
		return s
	}

	s := c.approximate(p)
	return t.Set(p, s)
}

// AsConstruction returns a string representing the construction of the
// Real number c, which may provide insight into how the number is constructed.
// The construction is returned as a single line string.
func AsConstruction(c Real) string {
	return AsConstructionIndent(c, "")
}

// AsConstructionIndent returns a string representing the construction of the
// Real number c, which may provide insight into how the number is constructed.
//
// When indent is empty, the construction is returned as a single line string.
//
// Otherwise, every opening parenthesis increases the indentation level by one,
// and every closing parenthesis decreases it by one. Arguments are separated by
// commas, and each argument starts on a new line with the appropriate indentation.
func AsConstructionIndent(c Real, indent string) string {
	data := c.asConstruction()
	if len(indent) == 0 {
		return data
	}

	out := strings.Builder{}
	currentIndent := 0
	sawComma := false
	for i := 0; i < len(data); i++ {
		ch := data[i]
		switch ch {
		case '(':
			out.WriteByte(ch)
			currentIndent++
			out.WriteByte('\n')
			out.WriteString(strings.Repeat(indent, currentIndent))
		case ')':
			currentIndent--
			out.WriteByte(',')
			out.WriteByte('\n')
			out.WriteString(strings.Repeat(indent, currentIndent))
			out.WriteByte(ch)
		case ',':
			out.WriteByte(ch)
			out.WriteByte('\n')
			out.WriteString(strings.Repeat(indent, currentIndent))
			sawComma = true
			continue
		case ' ':
			if !sawComma {
				out.WriteByte(ch)
			}
		default:
			out.WriteByte(ch)
		}
		sawComma = false
	}

	return out.String()
}

// Cmp compares two Real numbers a and b with higher and higher
// precision until a non-zero result is found. It returns 1 if `a > b`,
// -1 if `a < b`.
//
// This function never terminates if `a == b`; use PreciseCmp instead.
func Cmp(a, b Real) int {
	for p := -20; ; p *= 2 {
		if !IsPrecisionValid(p) {
			return 0
		}
		if v := PreciseCmp(a, b, p); v != 0 {
			return v
		}
	}
}

// PreciseCmp compares two Real numbers a and b with a precision p.
func PreciseCmp(a, b Real, p int) int {
	if a == nil || b == nil {
		return 0
	}

	ia := Approximate(a, p-1)
	ib := Approximate(b, p-1)
	if ia == nil || ib == nil {
		return 0
	}

	if ia.Cmp(bigAdd(ib, big.NewInt(1))) > 0 {
		return 1
	}
	if ia.Cmp(bigSub(ib, big.NewInt(1))) < 0 {
		return -1
	}

	return 0
}

// Real represents a constructive real number.
type Real interface {
	approximate(int) *big.Int
	asConstruction() string
	tracker() *precisionTracker
}

// knownMSD computes the position of the most significant digit (MSD). When
// the MSD is n, then 2^(n-1) < |c| < 2^(n+1).
func knownMSD(c Real) int {
	t := c.tracker()
	if t.MaxApproximation.Sign() >= 0 {
		return t.MinPrecision + t.MaxApproximation.BitLen() - 1
	}

	return t.MinPrecision + bigNeg(t.MaxApproximation).BitLen() - 1
}

func msd(c Real, n int) int {
	t := c.tracker()
	if !t.IsValid || (t.MaxApproximation.Cmp(big.NewInt(1)) <= 0 && t.MaxApproximation.Cmp(big.NewInt(-1)) >= 0) {
		_ = Approximate(c, n-1) // for side effects :(
		if bigAbs(t.MaxApproximation).Cmp(big.NewInt(1)) <= 0 {
			return math.MinInt
		}
	}

	return knownMSD(c)
}

// PreciseSign computes the sign of a Real number c given precision p.
func PreciseSign(c Real, p int) int {
	if t := c.tracker(); t.IsValid {
		v := t.MaxApproximation.Sign()
		if v != 0 {
			return v
		}
	}

	ic := Approximate(c, p-1)
	if ic == nil {
		return 0
	}

	return ic.Sign()
}

// Sign computes the sign of a Real number c. It returns 1 if c > 0,
// or -1 if c < 0.
//
// This function never terminates if c == 0; use PreciseSign instead.
func Sign(c Real) int {
	for p := -20; ; p *= 2 {
		if r := PreciseSign(c, p-1); r != 0 {
			return r
		}
	}
}

// scale is a rounded multiplication by 2^n.
func scale(i *big.Int, n int) *big.Int {
	if n >= 0 {
		return bigLsh(i, uint(n))
	}

	adj := bigAdd(signedShift(i, n+1), big.NewInt(1))
	return bigRsh(adj, 1)
}

// signedShift is a signed shift function.
func signedShift(i *big.Int, n int) *big.Int {
	switch {
	case n < 0:
		return bigRsh(i, uint(-n))
	case n > 0:
		return bigLsh(i, uint(n))
	default: // n == 0
		return i
	}
}

// constructiveInteger represents a constructive integer.
type constructiveInteger struct {
	precisionTracker
	i *big.Int
}

// FromBigInt creates a Real number from a big.Int.
func FromBigInt(i *big.Int) Real {
	if i == nil {
		return nil
	}

	if i.Sign() == 0 {
		return FromInt64(0)
	}

	return newInteger(i)
}

// FromBigIntSlice creates a slice of Real numbers from a slice of big.Int.
func FromBigIntSlice(ints []*big.Int) []Real {
	reals := make([]Real, len(ints))
	for idx, val := range ints {
		reals[idx] = FromBigInt(val)
	}
	return reals
}

// FromInt64 creates a Real number from an int64.
func FromInt64(i int64) Real {
	return newInteger(big.NewInt(i))
}

// FromInt64Slice creates a slice of Real numbers from a slice of int64.
func FromInt64Slice(ints []int64) []Real {
	reals := make([]Real, len(ints))
	for idx, val := range ints {
		reals[idx] = FromInt64(val)
	}
	return reals
}

// FromInt creates a Real number from an int.
func FromInt(i int) Real {
	return FromInt64(int64(i))
}

// FromIntSlice creates a slice of Real numbers from a slice of int.
func FromIntSlice(ints []int) []Real {
	reals := make([]Real, len(ints))
	for idx, val := range ints {
		reals[idx] = FromInt(val)
	}
	return reals
}

// FromFloat32 creates a Real number from a float32.
func FromFloat32(f float32) Real {
	return FromFloat64(float64(f))
}

// FromFloat32Slice creates a slice of Real numbers from a slice of float32.
func FromFloat32Slice(floats []float32) []Real {
	reals := make([]Real, len(floats))
	for idx, val := range floats {
		reals[idx] = FromFloat32(val)
	}
	return reals
}

// FromFloat64 creates a Real number from a float64.
func FromFloat64(f float64) Real {
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}

	bits := math.Float64bits(f) &^ (1 << 63)
	mantissa := bits & ((1 << 52) - 1)
	// 11-bit biased form https://en.wikipedia.org/wiki/IEEE_754-1985#Double_precision
	exponent := int((bits>>52)&((1<<11)-1)) - 1075
	if exponent != 0 {
		mantissa += (1 << 52)
	} else {
		mantissa <<= 1
	}

	r := ShiftLeft(newInteger(big.NewInt(int64(mantissa))), exponent)
	if f < 0 {
		r = newNegation(r)
	}
	return r
}

// FromFloat64Slice creates a slice of Real numbers from a slice of float64.
func FromFloat64Slice(floats []float64) []Real {
	reals := make([]Real, len(floats))
	for idx, val := range floats {
		reals[idx] = FromFloat64(val)
	}
	return reals
}

// FromRat creates a Real number from a rational number a/b, where b != 0.
func FromRat(a, b int) Real {
	return Divide(FromInt(a), FromInt(b))
}

func newInteger(i *big.Int) Real {
	return &constructiveInteger{
		i: i,
	}
}

func (c *constructiveInteger) approximate(p int) *big.Int {
	return scale(c.i, -p)
}

func (c *constructiveInteger) asConstruction() string {
	return fmt.Sprintf("Int(%s)", c.i.Text(10))
}

// Add computes the addition `a + b`.
func Add(a, b Real) Real {
	return newAddition(a, b)
}

// Subtract computes the subtraction `a + (-b)`.
func Subtract(a, b Real) Real {
	return newAddition(a, Negate(b))
}

type constructiveAddition struct {
	precisionTracker
	a Real
	b Real
}

func newAddition(a, b Real) Real {
	return &constructiveAddition{
		a: a,
		b: b,
	}
}

func (c *constructiveAddition) approximate(p int) *big.Int {
	sum := bigAdd(Approximate(c.a, p-2), Approximate(c.b, p-2))
	return scale(sum, -2)
}

func (c *constructiveAddition) asConstruction() string {
	return fmt.Sprintf("Add(%s, %s)", c.a.asConstruction(), c.b.asConstruction())
}

type constructiveMultiplication struct {
	precisionTracker
	a Real
	b Real
}

// Square computes the square `c * c`.
func Square(c Real) Real {
	return newMultiplication(c, c)
}

// Multiply computes the multiplication `a * b`.
func Multiply(a, b Real) Real {
	return newMultiplication(a, b)
}

func newMultiplication(a, b Real) Real {
	return &constructiveMultiplication{
		a: a,
		b: b,
	}
}

func (c *constructiveMultiplication) approximate(p int) *big.Int {
	hp := (p >> 1) - 1
	ma := msd(c.a, hp)
	if ma == math.MinInt {
		mb := msd(c.b, hp)
		if mb == math.MinInt {
			return big.NewInt(0)
		}

		ma, mb = mb, ma
	}

	p2 := p - ma - 3
	ib := Approximate(c.b, p2)
	if ib.Sign() == 0 {
		return big.NewInt(0)
	}

	mb := knownMSD(c.b)
	p1 := p - mb - 3
	ia := Approximate(c.a, p1)

	return scale(bigMul(ia, ib), p1+p2-p)
}

func (c *constructiveMultiplication) asConstruction() string {
	return fmt.Sprintf("Multiply(%s, %s)", c.a.asConstruction(), c.b.asConstruction())
}

// Inverse computes the multiplicative inverse, which is 1/c.
func Inverse(c Real) Real {
	return newMultiplicativeInverse(c)
}

// Divide computes the division `a * (1/b)`, where `1/b` is the multiplicative
// inverse of b.
func Divide(a, b Real) Real {
	return Multiply(a, Inverse(b))
}

type constructiveMultiplicativeInverse struct {
	precisionTracker
	r Real
}

func newMultiplicativeInverse(r Real) Real {
	return &constructiveMultiplicativeInverse{
		r: r,
	}
}

func (c *constructiveMultiplicativeInverse) approximate(p int) *big.Int {
	mr := msd(c.r, p)
	ir := 1 - mr

	digits := ir - p + 3
	pn := mr - digits

	lsf := -p - pn
	if lsf < 0 {
		return big.NewInt(0)
	}

	dividend := bigLsh(big.NewInt(1), uint(lsf))
	divisor := Approximate(c.r, pn)
	absolute := bigAbs(divisor)
	adj := bigAdd(dividend, bigRsh(absolute, 1))

	res := bigDiv(adj, divisor)
	if res.Sign() < 0 {
		return bigNeg(res)
	}
	return res
}

func (c *constructiveMultiplicativeInverse) asConstruction() string {
	return fmt.Sprintf("Inverse(%s)", c.r.asConstruction())
}

type constructiveShift struct {
	precisionTracker
	r Real
	n int
}

// ShiftLeft computes the left shift `c * 2^n`.
func ShiftLeft(c Real, n int) Real {
	return newShift(c, n)
}

// ShiftRight computes the right shift `c * 2^-n`.
func ShiftRight(c Real, n int) Real {
	return newShift(c, -n)
}

func newShift(r Real, n int) Real {
	return &constructiveShift{
		r: r,
		n: n,
	}
}

func (c *constructiveShift) approximate(p int) *big.Int {
	return Approximate(c.r, p-c.n)
}

func (c *constructiveShift) asConstruction() string {
	dir := "Left"
	if c.n < 0 {
		dir = "Right"
	}

	amt := c.n
	if amt < 0 {
		amt = -amt
	}

	return fmt.Sprintf("Shift%s(%s, %d)", dir, c.r.asConstruction(), amt)
}

// Negate computes the negation `-c`. The approximation is actually `-Approximate(c, p)`.
func Negate(c Real) Real {
	return newNegation(c)
}

type constructiveNegation struct {
	precisionTracker
	r Real
}

func newNegation(r Real) Real {
	return &constructiveNegation{
		r: r,
	}
}

func (c *constructiveNegation) approximate(p int) *big.Int {
	return bigNeg(Approximate(c.r, p))
}

func (c *constructiveNegation) asConstruction() string {
	return fmt.Sprintf("Negate(%s)", c.r.asConstruction())
}

// Abs computes the absolute value of c.
func Abs(c Real) Real {
	return newCondsign(c, Negate(c), c)
}

// Max computes the maximum of a and b.
func Max(a, b Real) Real {
	return newCondsign(Subtract(a, b), b, a)
}

// Min computes the minimum of a and b.
func Min(a, b Real) Real {
	return newCondsign(Subtract(a, b), a, b)
}

type constructiveCondsign struct {
	precisionTracker
	a Real
	b Real
	r Real
}

func newCondsign(r, a, b Real) Real {
	return &constructiveCondsign{
		a: a,
		b: b,
		r: r,
	}
}

func (c *constructiveCondsign) approximate(p int) *big.Int {
	switch sign := Approximate(c.r, -20).Sign(); {
	case sign < 0:
		return Approximate(c.a, p)
	case sign > 0:
		return Approximate(c.b, p)
	default:
	}

	ia := Approximate(c.a, p-1)
	ib := Approximate(c.b, p-1)
	delta := bigAbs(bigSub(ia, ib))
	if delta.Cmp(big.NewInt(1)) <= 0 {
		return scale(ia, -1)
	}

	if Sign(c.r) < 0 {
		return scale(ia, -1)
	}

	return scale(ib, -1)
}

func (c *constructiveCondsign) asConstruction() string {
	return fmt.Sprintf("CondSign(%s, %s, %s)", c.r.asConstruction(), c.a.asConstruction(), c.b.asConstruction())
}

// Exp computes the e^c.
func Exp(c Real) Real {
	rough := Approximate(c, -3)
	// e^-c = 1/e^c
	if rough.Sign() < 0 {
		return Inverse(Exp(Negate(c)))
	}

	if rough.Cmp(big.NewInt(2)) > 0 {
		return Square(Exp(ShiftRight(c, 1)))
	}

	return newPrescaledExponential(c)
}

type prescaledExponential struct {
	precisionTracker
	r Real
}

// newPrescaledExponential computes the exponential using a Taylor series
// expansion.
func newPrescaledExponential(c Real) Real {
	return &prescaledExponential{
		r: c,
	}
}

func (c *prescaledExponential) approximate(p int) *big.Int {
	if p >= 1 {
		return big.NewInt(0)
	}

	iters := -p/2 + 2
	calcPrec := p - boundLog2(2*iters) - 4
	opPrec := p - 3
	opAppr := Approximate(c.r, opPrec)

	term := bigLsh(big.NewInt(1), uint(-calcPrec))
	sum := bigLsh(big.NewInt(1), uint(-calcPrec))
	n := int64(0)
	maxTruncError := bigLsh(big.NewInt(1), uint(p-4-calcPrec))
	for bigAbs(term).Cmp(maxTruncError) >= 0 {
		n++
		term = scale(bigMul(term, opAppr), opPrec)
		term = bigDiv(term, big.NewInt(n))
		sum = bigAdd(sum, term)
	}
	return scale(sum, calcPrec-p)
}

func (c *prescaledExponential) asConstruction() string {
	return fmt.Sprintf("Pow(E, %s)", c.r.asConstruction())
}

func Ln(c Real) Real {
	rough := Approximate(c, -4)
	if rough.Sign() < 0 {
		return nil
	}
	if rough.Cmp(big.NewInt(8)) < 0 {
		return Negate(Ln(Inverse(c)))
	}
	if rough.Cmp(big.NewInt(24)) > 0 {
		return ShiftLeft(Ln(Sqrt(Sqrt(c))), 2)
	}
	return SimpleLn(c)
}

// SimpleLn computes the natural logarithm of `c`, for `1 < |c| < 2`.
func SimpleLn(c Real) Real {
	return newPrescaledNaturalLog(Subtract(c, One()))
}

type prescaledNaturalLog struct {
	precisionTracker
	r Real
}

func newPrescaledNaturalLog(c Real) Real {
	return &prescaledNaturalLog{
		r: c,
	}
}

func (c *prescaledNaturalLog) approximate(p int) *big.Int {
	if p >= 0 {
		return big.NewInt(0)
	}

	iters := -p - 1
	calcPrec := p - boundLog2(2*iters) - 4
	opPrec := p - 3
	opAppr := Approximate(c.r, opPrec)

	xToTheN := scale(opAppr, opPrec-calcPrec)
	term := xToTheN
	sum := term
	n := int64(1)
	sign := int64(1)
	maxTruncError := bigLsh(big.NewInt(1), uint(p-4-calcPrec))
	for bigAbs(term).Cmp(maxTruncError) >= 0 {
		n++
		sign = -sign
		xToTheN = scale(bigMul(xToTheN, opAppr), opPrec)
		term = bigDiv(xToTheN, big.NewInt(sign*n))
		sum = bigAdd(sum, term)
	}
	return scale(sum, calcPrec-p)
}

func (c *prescaledNaturalLog) asConstruction() string {
	return fmt.Sprintf("Ln(%s)", c.r.asConstruction())
}

type integralArctan struct {
	precisionTracker
	a Real
}

func newIntegralArctan(c Real) Real {
	return &integralArctan{
		a: c,
	}
}

func (c *integralArctan) approximate(p int) *big.Int {
	if p >= 1 {
		return big.NewInt(0)
	}

	iters := -p/2 + 2
	calcPrec := p - boundLog2(2*iters) - 4

	ia := Approximate(c.a, 0)
	isq := bigMul(ia, ia)

	power := bigDiv(bigLsh(big.NewInt(1), uint(-calcPrec)), ia)
	term := power
	sum := power
	sign := int64(1)

	n := int64(1)
	maxTruncError := bigLsh(big.NewInt(1), uint(p-4-calcPrec))
	for bigAbs(term).Cmp(maxTruncError) >= 0 {
		n += 2
		power = bigDiv(power, isq)
		sign = -sign

		term = bigDiv(power, bigMul(big.NewInt(sign), big.NewInt(n)))
		sum = bigAdd(sum, term)
	}
	return scale(sum, calcPrec-p)
}

func (c *integralArctan) asConstruction() string {
	return fmt.Sprintf("IntegralArctan(%s)", c.a.asConstruction())
}

// Sqrt computes the square root of c.
func Sqrt(c Real) Real {
	return newPrescaledSqrt(c)
}

type prescaledSqrt struct {
	precisionTracker
	r Real
}

func newPrescaledSqrt(c Real) Real {
	return &prescaledSqrt{
		r: c,
	}
}

func (c *prescaledSqrt) approximate(p int) *big.Int {
	pn := 2*p - 1
	mr := msd(c.r, pn)
	if mr <= pn {
		return big.NewInt(0)
	}

	digits := mr/2 - p
	if digits > 40 {
		pa := mr/2 - (digits/2 + 6)
		ic := Approximate(c, pa)
		ir := Approximate(c.r, 2*pa)

		numerator := scale(bigAdd(bigMul(ic, ic), ir), pa-p)
		return bigRsh(bigAdd(bigDiv(numerator, ic), big.NewInt(1)), 1)
	}

	pa := (mr - 60) &^ 1
	ir := bigLsh(Approximate(c.r, pa), 60)
	if ir.Sign() < 0 {
		return nil
	}

	fp, _ := ir.Float64()
	return signedShift(big.NewInt(int64(math.Sqrt(fp))), (pa-60)/2-p)
}

func (c *prescaledSqrt) asConstruction() string {
	return fmt.Sprintf("Sqrt(%s)", c.r.asConstruction())
}

// Cosine computes the cosine of c.
func Cosine(c Real) Real {
	rough := Approximate(c, -1)
	if rough.CmpAbs(big.NewInt(6)) >= 0 {
		mult := bigDiv(rough, big.NewInt(6))
		adj := Multiply(Pi(), FromBigInt(mult))
		if bigBitAnd(mult, big.NewInt(1)).Sign() != 0 {
			return Negate(Cosine(Subtract(c, adj)))
		}

		return Cosine(Subtract(c, adj))
	}

	if rough.CmpAbs(big.NewInt(2)) >= 0 {
		return Subtract(ShiftLeft(Square(Cosine(ShiftRight(c, 1))), 1), One())
	}

	return newPrescaledCosine(c)
}

// Sine computes the sine of c, using the identity `sin(c) = cos(π/2 - c)`.
func Sine(c Real) Real {
	return Cosine(Subtract(Divide(Pi(), Two()), c))
}

// Tangent computes the tangent of c, using the identity `tan(c) = sin(c) / cos(c)`.
func Tangent(c Real) Real {
	return Divide(Sine(c), Cosine(c))
}

// Arctangent computes the arctangent of c, using the integral formula.
//
// TODO(ripta): never terminates
// func Arctangent(c Real) Real {
//	return newIntegralArctan(Inverse(c))
// }

type prescaledCosine struct {
	precisionTracker
	r Real
}

func newPrescaledCosine(c Real) Real {
	return &prescaledCosine{
		r: c,
	}
}

func (c *prescaledCosine) approximate(p int) *big.Int {
	if p >= 1 {
		return big.NewInt(0)
	}

	iters := -p/2 - 2
	calcPrec := p - boundLog2(2*iters) - 4
	opPrec := p - 3
	opAppr := Approximate(c.r, opPrec)

	term := bigLsh(big.NewInt(1), uint(-calcPrec))
	sum := term
	n := int64(0)
	maxTruncError := bigLsh(big.NewInt(1), uint(p-4-calcPrec))
	for bigAbs(term).Cmp(maxTruncError) >= 0 {
		n += 2

		term = scale(bigMul(term, opAppr), opPrec)
		term = scale(bigMul(term, opAppr), opPrec) // [sic]
		term = bigDiv(term, big.NewInt(-n*(n-1)))
		sum = bigAdd(sum, term)
	}

	return scale(sum, calcPrec-p)
}

func (c *prescaledCosine) asConstruction() string {
	return fmt.Sprintf("Cosine(%s)", c.r.asConstruction())
}

// Pow computes the power c^n.
func Pow(c, n Real) Real {
	return Exp(Multiply(Ln(c), n))
}

// Pow10 computes the power 10^n.
func Pow10(n Real) Real {
	return Pow(Ten(), n)
}

type named struct {
	Real
	Name string
}

func newNamed(name string, c Real) Real {
	return &named{
		Real: c,
		Name: name,
	}
}

func (c *named) asConstruction() string {
	return fmt.Sprintf("Named(%q, %s)", c.Name, c.Real.asConstruction())
}

// ConstructiveName returns the name of the constructive Real number c,
// if it has one. The second return value indicates whether a name was found.
func ConstructiveName(c Real) (string, bool) {
	if n, ok := c.(*named); ok {
		return n.Name, true
	}
	return "", false
}

// ContinuedFraction64 computes the continued fraction from the given
// slice of int64 values.
func ContinuedFraction64(fracs []int64) Real {
	if len(fracs) == 0 {
		return Zero()
	}

	c := FromInt64(fracs[len(fracs)-1])
	for i := len(fracs) - 2; i >= 0; i-- {
		c = Add(FromInt64(fracs[i]), Inverse(c))
	}

	return c
}

// ContinuedFraction computes the continued fraction from the given
// slice of constructive Real values.
func ContinuedFraction(fracs []Real) Real {
	if len(fracs) == 0 {
		return Zero()
	}

	c := fracs[len(fracs)-1]
	for i := len(fracs) - 2; i >= 0; i-- {
		c = Add(fracs[i], Inverse(c))
	}

	return c
}
