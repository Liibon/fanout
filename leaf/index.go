package main

// Index is the abstraction over FAISS (real) and synthetic (sleep-based) backends.
type Index interface {
	// Search returns the top-k nearest neighbours for query. Returns (ids, distances, err).
	Search(query []float32, k int) ([]int64, []float32, error)
	// Close releases resources.
	Close()
}
