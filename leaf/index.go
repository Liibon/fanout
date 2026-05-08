package main

// Index is the abstraction over FAISS (real) and synthetic (sleep-based) backends.
type Index interface {
	// Search returns the global top-k nearest neighbours for query by searching the full shard.
	Search(query []float32, k int) ([]int64, []float32, error)
	// SearchByIDs computes exact L2 distances from query to each global ID in candidateIDs
	// and returns the top-k closest. This is the candidate-scoring path used when the root
	// has already performed approximate candidate selection (analogous to MicroSuite's bucket
	// receiving a candidate list from the mid-tier's FLANN index).
	SearchByIDs(query []float32, candidateIDs []int64, k int) ([]int64, []float32, error)
	// Close releases resources.
	Close()
}
