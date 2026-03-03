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
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"unsafe"
)

type faissIndex struct {
	idx *C.FaissIndexFlatL2
	dim int
}

func NewFaissIndex(cfg *Config) (Index, error) {
	f, err := os.Open(cfg.DatasetPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var hdr [8]byte
	if _, err := io.ReadFull(f, hdr[:]); err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	n := int(binary.LittleEndian.Uint32(hdr[:4]))
	dim := int(binary.LittleEndian.Uint32(hdr[4:]))
	if dim != cfg.Dim {
		return nil, fmt.Errorf("dataset dim=%d, expected %d", dim, cfg.Dim)
	}
	if n != cfg.NumVectors {
		return nil, fmt.Errorf("dataset n=%d, expected %d", n, cfg.NumVectors)
	}
	log.Printf("leaf %d: dataset ok (n=%d dim=%d)", cfg.LeafID, n, dim)

	var idx *C.FaissIndexFlatL2
	rc := C.faiss_IndexFlatL2_new_with(&idx, C.idx_t(dim))
	if rc != 0 {
		return nil, fmt.Errorf("faiss_IndexFlatL2_new_with returned %d", rc)
	}
	_ = unsafe.Pointer(idx)
	return &faissIndex{idx: idx, dim: dim}, nil
}

func (fi *faissIndex) Search(query []float32, k int) ([]int64, []float32, error) {
	return nil, nil, fmt.Errorf("search not yet implemented")
}

func (fi *faissIndex) Close() {
	C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(fi.idx)))
}
