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
	shardSize := cfg.NumVectors / cfg.NumLeaves
	offset := cfg.LeafID * shardSize
	if cfg.LeafID == cfg.NumLeaves-1 {
		shardSize = cfg.NumVectors - offset
	}

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
	if dim != cfg.Dim || n != cfg.NumVectors {
		return nil, fmt.Errorf("dataset mismatch")
	}

	seekBytes := int64(offset) * int64(dim) * 4
	if _, err := f.Seek(seekBytes, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("seek: %w", err)
	}
	vecs := make([]float32, shardSize*dim)
	if err := binary.Read(f, binary.LittleEndian, vecs); err != nil {
		return nil, fmt.Errorf("read vectors: %w", err)
	}

	var idx *C.FaissIndexFlatL2
	if rc := C.faiss_IndexFlatL2_new_with(&idx, C.idx_t(dim)); rc != 0 {
		return nil, fmt.Errorf("new_with: %d", rc)
	}
	if rc := C.faiss_Index_add(
		(*C.FaissIndex)(unsafe.Pointer(idx)),
		C.idx_t(shardSize),
		(*C.float)(unsafe.Pointer(&vecs[0])),
	); rc != 0 {
		C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(idx)))
		return nil, fmt.Errorf("faiss_Index_add: %d", rc)
	}
	log.Printf("leaf %d: indexed %d vectors (offset=%d)", cfg.LeafID, shardSize, offset)
	return &faissIndex{idx: idx, dim: dim}, nil
}

func (fi *faissIndex) Search(query []float32, k int) ([]int64, []float32, error) {
	return nil, nil, fmt.Errorf("search not yet implemented")
}

func (fi *faissIndex) Close() {
	C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(fi.idx)))
}
