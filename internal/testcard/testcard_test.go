package testcard

import (
	"testing"

	"github.com/comprofix/alleycat/internal/gfx"
)

func TestBuildIsScreenSized(t *testing.T) {
	b := Build().Bounds()
	if b.Dx() != gfx.ScreenWidth || b.Dy() != gfx.ScreenHeight {
		t.Fatalf("test card is %dx%d, want %dx%d", b.Dx(), b.Dy(), gfx.ScreenWidth, gfx.ScreenHeight)
	}
}

func TestBuildUsesOnlyPaletteColours(t *testing.T) {
	img := Build()
	allowed := map[[4]uint8]bool{}
	for _, c := range gfx.Palette {
		allowed[[4]uint8{c.R, c.G, c.B, c.A}] = true
	}
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := img.RGBAAt(x, y)
			if !allowed[[4]uint8{c.R, c.G, c.B, c.A}] {
				t.Fatalf("pixel (%d,%d) = %v is not a Sweetie 16 colour", x, y, c)
			}
		}
	}
}

func TestBuildShowsEverySwatch(t *testing.T) {
	img := Build()
	for i, want := range gfx.Palette {
		x := 16 + (i%8)*36 + 16
		y := 16 + (i/8)*40 + 16
		if got := img.RGBAAt(x, y); got != want {
			t.Errorf("swatch %d centre = %v, want %v", i, got, want)
		}
	}
}
