package gfx

import (
	"os"
	"testing"
)

func TestShippedFontCoversPrintableASCII(t *testing.T) {
	f, err := LoadFont(os.DirFS("../../assets"), "font/font8x8.txt")
	if err != nil {
		t.Fatal(err)
	}
	if f.Size != GlyphSize {
		t.Fatalf("font size %d, want %d", f.Size, GlyphSize)
	}
	for r := rune(' '); r <= '~'; r++ {
		if !f.Has(r) {
			t.Errorf("font is missing %q", r)
		}
	}
	for y := range GlyphSize {
		for x := range GlyphSize {
			if f.Ink(' ', x, y) {
				t.Fatal("space should be blank")
			}
		}
	}
}

func TestWidth(t *testing.T) {
	if Width("PAWS") != 32 {
		t.Errorf("Width(PAWS) = %d", Width("PAWS"))
	}
}
