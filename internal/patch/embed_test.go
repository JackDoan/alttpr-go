package patch

import "testing"

func TestLoadEmbedded(t *testing.T) {
	entries, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("embedded patch contains no entries")
	}
	if got := len(EmbeddedRaw()); got < 100_000 {
		t.Errorf("embedded blob smaller than expected: %d bytes", got)
	}
}
