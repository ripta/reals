package constructive

import "errors"

var (
	// ErrInvalidPrecision is returned when a precision is outside the range that
	// the integer width can represent. It differs from ErrPrecisionOverflow, which
	// reports that the caller's own limit was hit.
	ErrInvalidPrecision = errors.New("invalid precision")

	// ErrUndecided is returned when a decision loop leaves the valid precision
	// range before the operands were told apart.
	ErrUndecided = errors.New("undecided within the valid precision range")

	// ErrNilReal is returned when a nil Real is approximated.
	ErrNilReal = errors.New("nil real")

	// ErrDivisionByZero is returned when an inverse is taken of a value that
	// approximates to zero at the precision the inverse needs.
	ErrDivisionByZero = errors.New("division by zero")

	// ErrNegative is returned when a value is negative where the operation needs a
	// non-negative one, such as the square root.
	ErrNegative = errors.New("negative value out of domain")

	// ErrNonPositive is returned when a value is not positive where the operation
	// needs a positive one, such as the natural logarithm.
	ErrNonPositive = errors.New("non-positive value out of domain")

	errNoKnownMSD = errors.New("most significant digit is not known")
)
