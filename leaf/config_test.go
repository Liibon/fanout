package main

import (
	"os"
	"testing"
)

func TestConfigDefaults(t *testing.T) {
	cfg, err := configFromEnv()
	if err != nil { t.Fatalf("err: %v", err) }
	if cfg.NumVectors != 1_000_000 { t.Errorf("NumVectors: %d", cfg.NumVectors) }
	if cfg.Dim != 128 { t.Errorf("Dim: %d", cfg.Dim) }
	if cfg.NumLeaves != 16 { t.Errorf("NumLeaves: %d", cfg.NumLeaves) }
}

func TestConfigLeafID(t *testing.T) {
	os.Setenv("LEAF_ID", "7")
	defer os.Unsetenv("LEAF_ID")
	cfg, _ := configFromEnv()
	if cfg.LeafID != 7 { t.Errorf("LeafID: %d", cfg.LeafID) }
}

func TestShardOffset(t *testing.T) {
	cases := []struct{ leafID, numLeaves, numVectors, wantOff, wantSz int }{
		{0, 4, 100, 0, 25}, {1, 4, 100, 25, 25},
		{3, 4, 100, 75, 25}, {3, 4, 101, 75, 26},
	}
	for _, tc := range cases {
		sz := tc.numVectors / tc.numLeaves
		off := tc.leafID * sz
		if tc.leafID == tc.numLeaves-1 { sz = tc.numVectors - off }
		if off != tc.wantOff || sz != tc.wantSz {
			t.Errorf("case %v: off=%d sz=%d", tc, off, sz)
		}
	}
}
