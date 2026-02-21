//go:build !cgo

package main

import "errors"

func NewFaissIndex(_ *Config) (Index, error) {
	return nil, errors.New("FAISS requires CGO_ENABLED=1; use Docker to build the leaf")
}
