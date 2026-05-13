package world

import (
	"bytes"
	"os"
	"testing"
)

// TestCredits_DefaultMatchesPHP confirms that our Credits encoder produces
// the exact same bytes as PHP for the default scene table (no randomization).
// The reference dumps are at /tmp/php_credits_{data,ptrs}.bin.
func TestCredits_DefaultMatchesPHP(t *testing.T) {
	wantData, err := os.ReadFile("/tmp/php_credits_data.bin")
	if err != nil {
		t.Skipf("no PHP reference: %v", err)
	}
	wantPtrs, err := os.ReadFile("/tmp/php_credits_ptrs.bin")
	if err != nil {
		t.Skipf("no PHP reference: %v", err)
	}

	c := NewCredits()
	gotData, gotPtrs := c.BinaryData()
	gotPtrsBytes := make([]byte, len(gotPtrs)*2)
	for i, p := range gotPtrs {
		gotPtrsBytes[i*2] = byte(p)
		gotPtrsBytes[i*2+1] = byte(p >> 8)
	}

	if !bytes.Equal(gotData, wantData) {
		// Find first diff index.
		for i := 0; i < len(gotData) && i < len(wantData); i++ {
			if gotData[i] != wantData[i] {
				t.Errorf("credits data diverge at byte %d: got 0x%02X want 0x%02X", i, gotData[i], wantData[i])
				break
			}
		}
		t.Fatalf("credits data mismatch: go=%d php=%d", len(gotData), len(wantData))
	}
	if !bytes.Equal(gotPtrsBytes, wantPtrs) {
		t.Fatalf("credits pointers mismatch")
	}
	t.Logf("credits byte-identical: %d data + %d pointer bytes", len(gotData), len(gotPtrsBytes))
}
