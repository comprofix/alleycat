package scene

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// banner is a word in the new look's big lettering, drawn once in white so
// it can be laid in any colour, whole or in bands, as the classic look's
// Font.DrawBanded paints its logo and result words.
type banner struct {
	img  *ebiten.Image
	w, h int
	// top and bottom are the first and last rows the letters fill.
	top, bottom int
}

func newBanner(f *gfx.VectorFont, s string) banner {
	b := banner{w: f.Width(s), h: f.LineHeight()}
	b.img = ebiten.NewImage(b.w, b.h)
	f.Draw(b.img, s, 0, 0, resurrect64.White)
	b.top, b.bottom = inkRows(b.img)
	return b
}

// inkRows are the first and last rows of img with anything drawn in them.
func inkRows(img *ebiten.Image) (top, bottom int) {
	r := img.Bounds()
	px := make([]byte, 4*r.Dx()*r.Dy())
	img.ReadPixels(px)
	top = -1
	for y := range r.Dy() {
		for x := range r.Dx() {
			if px[4*(y*r.Dx()+x)+3] > 0 {
				if top < 0 {
					top = y
				}
				bottom = y
				break
			}
		}
	}
	return max(top, 0), bottom
}

// lay lays the banner's rows from to to in colour c, with the top of its
// letters at (x, y).
func (b banner) lay(dst *ebiten.Image, x, y float64, c, from, to int) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y-float64(b.top)+float64(from))
	op.ColorScale.ScaleWithColor(gfx.RGBA(c))
	dst.DrawImage(b.img.SubImage(image.Rect(0, from, b.w, to)).(*ebiten.Image), op)
}

// layAll lays the whole banner in colour c, with the top of its letters at
// (x, y).
func (b banner) layAll(dst *ebiten.Image, x, y float64, c int) { b.lay(dst, x, y, c, 0, b.h) }

// banded lays the letters with the top of them at (x, y) in three bands,
// light, middle and dark from the top, as Font.DrawBanded does.
func (b banner) banded(dst *ebiten.Image, x, y float64, light, mid, dark int) {
	h := b.bottom - b.top + 1
	b1, b2 := b.top+h*3/10, b.top+h*7/10
	b.lay(dst, x, y, light, 0, b1)
	b.lay(dst, x, y, mid, b1, b2)
	b.lay(dst, x, y, dark, b2, b.h)
}

// outline lays the banner in colour c round where it lies at (x, y), r
// pixels out.
func (b banner) outline(dst *ebiten.Image, x, y, r float64, c int) {
	for _, d := range [12][2]float64{{-1, 0}, {1, 0}, {0, -1}, {0, 1}, {-0.8, -0.6}, {0.8, -0.6}, {-0.8, 0.6}, {0.8, 0.6}, {-0.6, -0.8}, {0.6, -0.8}, {-0.6, 0.8}, {0.6, 0.8}} {
		b.layAll(dst, x+d[0]*r, y+d[1]*r, c)
	}
}
