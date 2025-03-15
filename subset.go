package main

import (
	"flag"
	"fmt"
	"math/rand/v2"
)

// Flags.
var (
	subsetPct = flag.Uint("subsetpct", 100,
		"percentage of files to process (0 = none, 100 = all)")
	randSeed = flag.Uint64("subsetseed", 0,
		"seed for the subset selection PRNG, useful for testing (0 = random)")
)

type Subset struct {
	// Percentage of files to process (0 = none, 100 = all).
	percent uint

	// Random source for subset selection.
	rand *rand.Rand
}

func NewSubset() (*Subset, error) {
	if *subsetPct > 100 {
		return nil, fmt.Errorf(
			"subset percentage %d must be in the [0, 100] range",
			*subsetPct)
	}

	// Seed the PRNG. If the user didn't specify a seed, use two random
	// numbers.
	seed1, seed2 := rand.Uint64(), rand.Uint64()
	if *randSeed != 0 {
		seed1 = 0
		seed2 = *randSeed
	}

	return &Subset{
		percent: *subsetPct,
		rand:    rand.New(rand.NewPCG(seed1, seed2)),
	}, nil
}

func (s *Subset) ShouldProcess() bool {
	// Special-case 0% and 100% to avoid picking a random number
	// unnecessarily.
	if s.percent == 100 {
		return true
	} else if s.percent == 0 {
		return false
	}

	// Note this is NOT thread-safe, but the caller is single threaded so it's
	// fine for our use case.
	return s.rand.UintN(100) < s.percent
}
