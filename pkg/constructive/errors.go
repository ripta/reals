package constructive

import "errors"

var (
	// ErrInvalidPrecision is returned when a precision is outside the range that
	// the integer width can represent. It differs from ErrPrecisionOverflow, which
	// reports that the caller's own limit was hit.
	ErrInvalidPrecision = errors.New("invalid precision")

	// ErrUndecided is returned when a decision loop leaves the valid precision
	// range before the operands were told apart. It is also what an inverse of an
	// exact zero ends in, because zero cannot be told from a tiny value in bounded
	// work.
	ErrUndecided = errors.New("undecided within the valid precision range")

	// ErrNilReal is returned when a nil Real is approximated.
	ErrNilReal = errors.New("nil real")

	// ErrNegative is returned when a value is negative where the operation needs a
	// non-negative one, such as the square root.
	ErrNegative = errors.New("negative value out of domain")

	// ErrNonPositive is returned when a value is not positive where the operation
	// needs a positive one, such as the natural logarithm.
	ErrNonPositive = errors.New("non-positive value out of domain")

	// ErrNegativeOrder is returned when the order of a polygamma function is negative.
	// Only non-negative orders are derivatives of the digamma function.
	ErrNegativeOrder = errors.New("negative order out of domain")

	// ErrBadApproximation is returned when an approximation callback breaks its
	// contract, such as by returning a nil approximation with a nil error, or a
	// slice of approximations of the wrong length.
	ErrBadApproximation = errors.New("approximation callback broke its contract")

	// ErrReentrantApproximation is returned when an approximation callback, while
	// it runs, asks for an approximation of its own value, or of a sibling output
	// that shares its computation, that would run it again.
	ErrReentrantApproximation = errors.New("approximation callback re-entered itself")

	errNoKnownMSD = errors.New("most significant digit is not known")
)
