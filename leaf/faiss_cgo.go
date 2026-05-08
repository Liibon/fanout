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
	"sort"
	"unsafe"
)

type faissIndex struct {
	idx    *C.FaissIndexFlatL2
	dim    int
	offset int64     // global base index of this shard's first vector
	vecs   []float32 // raw shard vectors retained for SearchByIDs exact scoring
}

// NewFaissIndex builds a flat-L2 FAISS index from the leaf's shard of the dataset.
func NewFaissIndex(cfg *Config) (Index, error) {
	shardSize := cfg.NumVectors / cfg.NumLeaves
	offset := cfg.LeafID * shardSize
	// Last leaf absorbs remainder.
	if cfg.LeafID == cfg.NumLeaves-1 {
		shardSize = cfg.NumVectors - offset
	}

	log.Printf("leaf %d: loading %d vectors (offset=%d) from %s",
		cfg.LeafID, shardSize, offset, cfg.DatasetPath)

	f, err := os.Open(cfg.DatasetPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// Binary format: int32 n, int32 dim, n*dim float32 row-major.
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

	// Seek to shard start.
	seekBytes := int64(offset) * int64(dim) * 4
	if _, err := f.Seek(seekBytes, io.SeekCurrent); err != nil {
		return nil, fmt.Errorf("seek to shard: %w", err)
	}

	vecs := make([]float32, shardSize*dim)
	if err := binary.Read(f, binary.LittleEndian, vecs); err != nil {
		return nil, fmt.Errorf("read vectors: %w", err)
	}

	var idx *C.FaissIndexFlatL2
	rc := C.faiss_IndexFlatL2_new_with(&idx, C.idx_t(dim))
	if rc != 0 {
		return nil, fmt.Errorf("faiss_IndexFlatL2_new_with returned %d", rc)
	}

	rc = C.faiss_Index_add(
		(*C.FaissIndex)(unsafe.Pointer(idx)),
		C.idx_t(shardSize),
		(*C.float)(unsafe.Pointer(&vecs[0])),
	)
	if rc != 0 {
		C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(idx)))
		return nil, fmt.Errorf("faiss_Index_add returned %d", rc)
	}

	log.Printf("leaf %d: FAISS index built (%d vectors, dim=%d)", cfg.LeafID, shardSize, dim)
	return &faissIndex{idx: idx, dim: dim, offset: int64(offset), vecs: vecs}, nil
}

func (fi *faissIndex) Search(query []float32, k int) ([]int64, []float32, error) {
	distances := make([]float32, k)
	labels := make([]int64, k)

	rc := C.faiss_Index_search(
		(*C.FaissIndex)(unsafe.Pointer(fi.idx)),
		1,
		(*C.float)(unsafe.Pointer(&query[0])),
		C.idx_t(k),
		(*C.float)(unsafe.Pointer(&distances[0])),
		(*C.idx_t)(unsafe.Pointer(&labels[0])),
	)
	if rc != 0 {
		return nil, nil, fmt.Errorf("faiss_Index_search returned %d", rc)
	}
	// Translate local shard labels to global dataset IDs.
	// FAISS returns -1 for unfilled slots when k > shard size.
	for i, l := range labels {
		if l >= 0 {
			labels[i] = l + fi.offset
		}
	}
	return labels, distances, nil
}

// SearchByIDs computes exact L2 distances from query to each candidate ID,
// translating global IDs to local shard positions. Candidates outside this
// shard's range are silently skipped. Returned IDs are global.
func (fi *faissIndex) SearchByIDs(query []float32, candidateIDs []int64, k int) ([]int64, []float32, error) {
	type pair struct {
		id   int64
		dist float32
	}
	pairs := make([]pair, 0, len(candidateIDs))
	shardSize := int64(len(fi.vecs)) / int64(fi.dim)
	for _, gid := range candidateIDs {
		lid := gid - fi.offset
		if lid < 0 || lid >= shardSize {
			continue
		}
		base := lid * int64(fi.dim)
		vec := fi.vecs[base : base+int64(fi.dim)]
		pairs = append(pairs, pair{gid, l2sq(query, vec)})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].dist < pairs[j].dist })
	if k > len(pairs) {
		k = len(pairs)
	}
	ids := make([]int64, k)
	dists := make([]float32, k)
	for i := 0; i < k; i++ {
		ids[i] = pairs[i].id
		dists[i] = pairs[i].dist
	}
	return ids, dists, nil
}

func l2sq(a, b []float32) float32 {
	var s float32
	for i := range a {
		d := a[i] - b[i]
		s += d * d
	}
	return s
}

func (fi *faissIndex) Close() {
	C.faiss_Index_free((*C.FaissIndex)(unsafe.Pointer(fi.idx)))
}
