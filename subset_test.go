package main

import (
	"fmt"
	"testing"
)

// Create a Subset instance with a variety of percents, and then confirm that
// after 100k runs, the distribution of ShouldProcess() calls is as expected.
func TestSubset(t *testing.T) {
	pcts := []uint{0, 1, 10, 50, 70, 90, 99, 100}
	for _, pct := range pcts {
		t.Run(fmt.Sprintf("percent=%d", pct), func(t *testing.T) {
			testSubset(t, pct)
		})
	}
}

func testSubset(t *testing.T, pct uint) {
	*subsetPct = pct
	subset, err := NewSubset()
	if err != nil {
		t.Fatal(err)
	}

	count := uint64(1_000_000)
	selected := uint64(0)
	for range count {
		if subset.ShouldProcess() {
			selected++
		}
	}

	// Confirm that the number of selected items is within 3% of the expected.
	// With these many samples, we don't expect false positives.
	expected_min := count * uint64(max(int(pct)-3, 0)) / 100
	expected_max := count * uint64(min(int(pct)+3, 100)) / 100
	if selected < expected_min || selected > expected_max {
		t.Errorf("selected %d items, expected [%d, %d]",
			selected, expected_min, expected_max)
	}
}

// Benchmark the performance of the ShouldProcess() method, for 0%, 50%, and
// 100%. 0% and 100% should be faster since they don't pick a random number,
// and the 50% case should give us a sense of the overhead of the random
// decision.
func BenchmarkSubset(b *testing.B) {
	for _, pct := range []uint{0, 50, 100} {
		b.Run(fmt.Sprintf("percent=%d", pct), func(b *testing.B) {
			benchmarkSubset(b, pct)
		})
	}
}

func benchmarkSubset(b *testing.B, pct uint) {
	*subsetPct = pct
	subset, err := NewSubset()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		subset.ShouldProcess()
	}
}
