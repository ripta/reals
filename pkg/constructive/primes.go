package constructive

// generatePrimes returns all prime numbers up to maxPrime using the Sieve of
// Eratosthenes. This is used for computing the Prime Constant ρ = Σ(1/2^p) for
// all primes p.
func generatePrimes(maxPrime int) []int {
	// fast path
	if maxPrime < 2 {
		return []int{}
	}

	// init: consider all numbers prime
	sieve := make([]bool, maxPrime+1)
	for i := 2; i <= maxPrime; i++ {
		sieve[i] = true
	}

	// sieve: mark each multiple as composite
	for i := 2; i*i <= maxPrime; i++ {
		if sieve[i] {
			for j := i * i; j <= maxPrime; j += i {
				sieve[j] = false
			}
		}
	}

	// collect primes
	primes := []int{}
	for i := 2; i <= maxPrime; i++ {
		if sieve[i] {
			primes = append(primes, i)
		}
	}

	return primes
}
