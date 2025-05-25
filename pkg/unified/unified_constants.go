package unified

import (
	"sync"

	"github.com/ripta/reals/pkg/constructive"
	"github.com/ripta/reals/pkg/rational"
)

var Zero = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.Zero())
})

var One = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.One())
})

var Two = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.New64(2, 1))
})

var Ten = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.New64(10, 1))
})

var Half = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.New64(1, 2))
})

var NegativeOne = sync.OnceValue(func() *Real {
	return New(constructive.One(), rational.New64(-1, 1))
})

var E = sync.OnceValue(func() *Real {
	return New(constructive.E(), rational.One())
})

var Pi = sync.OnceValue(func() *Real {
	return New(constructive.Pi(), rational.One())
})

var Phi = sync.OnceValue(func() *Real {
	return New(constructive.Phi(), rational.One())
})

var Sqrt2 = sync.OnceValue(func() *Real {
	return New(constructive.Sqrt2(), rational.One())
})

var Ln2 = sync.OnceValue(func() *Real {
	return New(constructive.Ln2(), rational.One())
})
