package rational

import "sync"

var Zero = sync.OnceValue(func() *Number {
	return New64(0, 1)
})

var One = sync.OnceValue(func() *Number {
	return New64(1, 1)
})
