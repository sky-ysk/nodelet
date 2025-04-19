package utils

import (
	"context"
	"math"
)

const DefaultParallelism int = 16

// Parallelizer holds the parallelism for scheduler.
type Parallelizer struct {
	parallelism int
}

// NewParallelizer returns an object holding the parallelism.
func NewParallelizer(p int) Parallelizer {
	return Parallelizer{parallelism: p}
}

// chunkSizeFor returns a chunk size for the given number of items to use for
// parallel work. The size aims to produce good CPU utilization.
// returns max(1, min(sqrt(n), n/Parallelism))
func chunkSizeFor(n, parallelism int) int {
	s := int(math.Sqrt(float64(n)))

	if r := n/parallelism + 1; s > r {
		s = r
	} else if s < 1 {
		s = 1
	}
	return s
}

// Until is a wrapper around workqueue.ParallelizeUntil to use in scheduling algorithms.
// A given operation will be a label that is recorded in the goroutine metric.
func (p Parallelizer) Until(ctx context.Context, pieces int, doWorkPiece DoWorkPieceFunc) {
	ParallelizeUntil(ctx, p.parallelism, pieces, doWorkPiece, WithChunkSize(chunkSizeFor(pieces, p.parallelism)))
}
