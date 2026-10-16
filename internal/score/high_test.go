package score

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHighRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "scores.json")
	if err := SaveHigh(path, 12345); err != nil {
		t.Fatal(err)
	}
	if got := LoadHigh(path); got != 12345 {
		t.Errorf("loaded %d, want 12345", got)
	}
}

// A missing or corrupt file gives 0; out of range is clamped (DESIGN §9).
func TestHighFallBack(t *testing.T) {
	dir := t.TempDir()
	if got := LoadHigh(filepath.Join(dir, "none.json")); got != 0 {
		t.Errorf("missing file: %d", got)
	}
	for bad, want := range map[string]int{"{oops": 0, `{"high":"lots"}`: 0, `{"high":-5}`: 0, `{"high":123456789}`: MaxHigh} {
		path := filepath.Join(dir, "bad.json")
		os.WriteFile(path, []byte(bad), 0o644)
		if got := LoadHigh(path); got != want {
			t.Errorf("%q: %d, want %d", bad, got, want)
		}
	}
}
