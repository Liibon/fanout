// dataset generates a reproducible 1M × 128-dim float32 corpus.
//
// Binary format (little-endian):
//
//	[int32 n][int32 dim][n * dim * float32]
//
// A SHA-256 digest is written to <output>.sha256 so downstream consumers can
// verify dataset integrity before use.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
)

func main() {
	n := flag.Int("n", 1_000_000, "number of vectors")
	dim := flag.Int("dim", 128, "vector dimension")
	seed := flag.Int64("seed", 42, "RNG seed (change breaks reproducibility)")
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

	h := sha256.New()
	w := io.MultiWriter(f, h)

	// Write 8-byte header.
	hdr := make([]byte, 8)
	binary.LittleEndian.PutUint32(hdr[0:4], uint32(n))
	binary.LittleEndian.PutUint32(hdr[4:8], uint32(dim))
	if _, err := w.Write(hdr); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	rng := rand.New(rand.NewSource(seed))
	buf := make([]byte, dim*4)

	for i := 0; i < n; i++ {
		for j := 0; j < dim; j++ {
			v := float32(rng.NormFloat64())
			binary.LittleEndian.PutUint32(buf[j*4:], math.Float32bits(v))
		}
		if _, err := w.Write(buf); err != nil {
			return fmt.Errorf("write vector %d: %w", i, err)
		}
		if i > 0 && i%100_000 == 0 {
			log.Printf("generated %d / %d vectors (%.0f%%)", i, n, float64(i)/float64(n)*100)
		}
	}

	digest := hex.EncodeToString(h.Sum(nil))
	log.Printf("done: n=%d dim=%d seed=%d sha256=%s", n, dim, seed, digest)

	return os.WriteFile(path+".sha256",
		[]byte(fmt.Sprintf("%s  %s\n", digest, path)), 0644)
}
