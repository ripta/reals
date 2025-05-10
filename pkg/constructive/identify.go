package constructive

import (
	"errors"
	"math/big"
)

var ErrNotConstructive = errors.New("not constructive")

func Identify(c Real) (*big.Int, bool, error) {
	if c == nil {
		return nil, false, ErrNotConstructive
	}

	//switch v := c.(type) {
	//case *constructiveInteger:
	//	return NewRational(v.i, 1), true, nil
	//}

	return nil, true, nil
}
