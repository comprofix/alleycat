package spritegrid

import (
	"strings"
	"testing"

	"github.com/comprofix/alleycat/internal/gfx"
)

const sample = `; two sprites
sprite dot 2 2
C.
.C

0.
.0
sprite bar 3 1
39F
`

func TestParse(t *testing.T) {
	sprites, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	if len(sprites) != 2 {
		t.Fatalf("got %d sprites, want 2", len(sprites))
	}
	dot := sprites[0]
	if dot.Name != "dot" || len(dot.Frames) != 2 {
		t.Fatalf("dot = %+v", dot)
	}
	if got := dot.Frames[0]; got[0] != gfx.White || got[1] != Transparent || got[3] != gfx.White {
		t.Errorf("dot frame 0 = %v", got)
	}
	if got := sprites[1].Frames[0]; got[0] != 3 || got[1] != 9 || got[2] != 15 {
		t.Errorf("bar = %v, want [3 9 15]", got)
	}
}

func TestParseErrors(t *testing.T) {
	bad := []string{
		"sprite a 2 2\nCC\nC\n", // short row
		"sprite a 2 2\nCC\n",    // too few rows
		"sprite a 1 1\nZ\n",     // bad pixel
		"CC\n",                  // pixels before header
		"sprite a 1 1\n",        // no frames
	}
	for _, src := range bad {
		if _, err := Parse(strings.NewReader(src)); err == nil {
			t.Errorf("Parse(%q) succeeded, want error", src)
		}
	}
}

func TestPack(t *testing.T) {
	sprites, err := Parse(strings.NewReader(sample))
	if err != nil {
		t.Fatal(err)
	}
	img, index, err := Pack(sprites)
	if err != nil {
		t.Fatal(err)
	}
	dot, bar := index["dot"], index["bar"]
	if dot != (Entry{0, 0, 2, 2, 2}) || bar != (Entry{0, 2, 3, 1, 1}) {
		t.Fatalf("index = %+v", index)
	}
	if img.RGBAAt(0, 0) != gfx.Palette[gfx.White] {
		t.Error("dot frame 0 pixel missing")
	}
	if img.RGBAAt(2, 0) != gfx.Palette[gfx.Night] {
		t.Error("dot frame 1 should start at x=2")
	}
	if img.RGBAAt(1, 0).A != 0 {
		t.Error("transparent pixel should stay clear")
	}
	if img.RGBAAt(2, 2) != gfx.Palette[15] {
		t.Error("bar should start at y=2")
	}
}

func TestPackRejectsDuplicates(t *testing.T) {
	s := []Sprite{{Name: "x", W: 1, H: 1, Frames: [][]int{{0}}}, {Name: "x", W: 1, H: 1, Frames: [][]int{{0}}}}
	if _, _, err := Pack(s); err == nil {
		t.Error("duplicate names should fail")
	}
}
