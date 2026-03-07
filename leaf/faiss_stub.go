//go:build !cgo

package main

import "errors"

// NewFaissIndex is unavailable without CGo + FAISS.
// Build with CGO_ENABLED=1 and FAISS installed (see leaf/Dockerfile).
func NewFaissIndex(_ *Config) (Index, error) {
	return nil, errors.New("FAISS requires CGO_ENABLED=1; use Docker to build the leaf")
}
