// dataset generates a reproducible corpus.
//
// Binary format (little-endian):
//   [int32 n][int32 dim][n * dim * float32]
package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
)

func main() {
	n := flag.Int("n", 1_000_000, "number of vectors")
	dim := flag.Int("dim", 128, "vector dimension")
	seed := flag.Int64("seed", 42, "RNG seed")
	out := flag.String("out", "/data/dataset.bin", "output path")
	flag.Parse()
	if err := generate(*out, *n, *dim, *seed); err != nil {
		log.Fatalf("generate: %v", err)
	}
}

func generate(path string, n, dim int, seed int64) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	hdr := make([]byte, 8)
	binary.LittleEndian.PutUint32(hdr[0:4], uint32(n))
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(dim))
	if _, err := f.Write(hdr); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	rng := rand.New(rand.NewSource(seed))
	buf := make([]byte, dim*4)
	for i := 0; i < n; i++ {
		for j := 0; j < dim; j++ {
			v := float32(rng.NormFloat64())
			binary.LittleEndian.PutUint32(buf[j*4:], math.Float32bits(v))
		}
		if _, err := f.Write(buf); err != nil {
			return fmt.Errorf("write vector %d: %w", i, err)
		}
	}
	log.Printf("done: n=%d dim=%d seed=%d", n, dim, seed)
	return nil
}
