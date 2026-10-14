package gfx

import (
	"os"
	"strings"
	"testing"
)

func TestParseFontGlyph(t *testing.T) {
	src := "; test\n@A\n#.......\n.#......\n........\n........\n........\n........\n........\n#######.\n"
	glyphs, err := ParseFont(strings.NewReader(src))
	if err != nil {
		t.Fatal(err)
	}
	g := glyphs['A']
	if g[0] != 0x80 || g[1] != 0x40 || g[7] != 0xfe {
		t.Errorf("glyph A = %08b", g)
	}
}

func TestParseFontErrors(t *testing.T) {
	bad := []string{
		"A\n",            // missing @
		"@A\n#.......\n", // incomplete
		"@A\n#......\n",  // short row
		"@A\nx.......\n", // bad pixel
	}
	for _, src := range bad {
		if _, err := ParseFont(strings.NewReader(src)); err == nil {
			t.Errorf("ParseFont(%q) succeeded, want error", src)
		}
	}
}

func TestShippedFontCoversPrintableASCII(t *testing.T) {
	f, err := os.Open("../../assets/font/font8x8.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	glyphs, err := ParseFont(f)
	if err != nil {
		t.Fatal(err)
	}
	for r := rune(' '); r <= '~'; r++ {
		if _, ok := glyphs[r]; !ok {
			t.Errorf("font is missing %q", r)
		}
	}
	if g := glyphs[' ']; g != (Glyph{}) {
		t.Error("space should be blank")
	}
}

func TestWidth(t *testing.T) {
	if Width("PAWS") != 32 {
		t.Errorf("Width(PAWS) = %d", Width("PAWS"))
	}
}
