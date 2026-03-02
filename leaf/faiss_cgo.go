//go:build cgo

package main

/*
#cgo CFLAGS: -I/usr/local/include
#cgo LDFLAGS: -L/usr/local/lib -lfaiss_c -lfaiss -lopenblas -lstdc++ -lm
#include "faiss/c_api/IndexFlat_c.h"
#include "faiss/c_api/Index_c.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"log"
	"unsafe"
)

type faissIndex struct {
	idx *C.FaissIndexFlatL2
	dim int
}

func NewFaissIndex(cfg *Config) (Index, error) {
	log.Printf("leaf %d: FAISS mode, dim=%d", cfg.LeafID, cfg.Dim)
	var idx *C.FaissIndexFlatL2
	rc := C.faiss_IndexFlatL2_new_with(&idx, C.idx_t(cfg.Dim))
	if rc != 0 {
		return nil, fmt.Errorf("faiss_IndexFlatL2_new_with returned %d", rc)
	}
	_ = unsafe.Pointer(idx)
	return &faissIndex{idx: idx, dim: cfg.Dim}, nil
}

func (fi *faissIndex) Search(query []float32, k int) ([]int64, []float32, error) {
	return nil, nil, fmt.Errorf("search not yet implemented")
}

func (fi *faissIndex) Close() {
	C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(fi.idx)))
}
