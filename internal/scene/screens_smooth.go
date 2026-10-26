package scene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// The screens between the game's scenes in the new look (DESIGN §2.1,
// §3.6, §3.7): each as the classic look lays it out, drawn in code at four
// times the size. Their pictures are drawn once and kept, render caches.

func (*Result) native()    {}
func (*Tally) native()     {}
func (*HereKitty) native() {}
func (*Reject) native()    {}
func (*Dance) native()     {}

// The new look's lettering on these screens, as big as the classic look's
// font at three, two and one times.
const (
	bigTextSize    = 160
	middleTextSize = 105
	smallTextSize  = 40
)

// smoothScreen reports whether dst is the new look's screen.
func smoothScreen(dst *ebiten.Image) bool { return dst.Bounds().Dx() != gfx.ScreenWidth }

// drawSmooth draws the result word in the new look: big, banded, with a
// dark shadow and paint running down from some letters.
func (r *Result) drawSmooth(dst *ebiten.Image) {
	dst.Fill(gfx.Palette[resurrect64.Ink])
	if r.ink < 0 {
		return
	}
	if r.font == nil {
		r.font = gfx.NewVectorFont(bigTextSize)
		r.big = newBanner(r.font, r.text)
	}
	b := r.big
	x, y := float64(dst.Bounds().Dx()-b.w)/2, tk(resultY)
	band := resultBands[r.ink]
	b.layAll(dst, x+12, y+12, resurrect64.Shadow)
	b.banded(dst, x, y, band[0], r.ink, band[1])
	cell, foot := float64(r.font.CellWidth()), y+float64(b.bottom-b.top)
	for i := range len(r.text) {
		if i%2 == 0 {
			continue
		}
		dx := x + float64(i)*cell + cell/2
		end := foot + 12 + float64(i%3)*12
		shapes.Line(dst, dx, foot-8, dx, end, 7, band[1])
		shapes.FillCircle(dst, dx, end+2, 5, band[1])
	}
}

// tallyArt is the tally's pictures in the new look.
type tallyArt struct {
	faces [2]*ebiten.Image // lit, dim
	text  *gfx.VectorFont
}

// drawSmooth draws the tally in the new look: the columns of cat faces
// lighting up from the bottom, then the bonus.
func (t *Tally) drawSmooth(dst *ebiten.Image) {
	if t.art == nil {
		t.art = &tallyArt{text: gfx.NewVectorFont(middleTextSize)}
		for i, c := range []entity.FaceInks{entity.LitFace, entity.DimFace} {
			t.art.faces[i] = entity.Bake(12, 8, entity.NoOutline, func(p entity.Pen) { entity.DrawTallyFace(p, c) })
		}
	}
	bg := resurrect64.Ink
	if t.lit >= t.rows() && t.held/4%2 == 1 {
		bg = resurrect64.Navy // the flashing at the end
	}
	dst.Fill(gfx.Palette[bg])
	for row, y := 0, tallyBottomY; y >= 0; row, y = row+1, y-tallyRowH {
		face := t.art.faces[1]
		if row < t.lit {
			face = t.art.faces[0]
		}
		for _, x := range tallyColumns {
			entity.DrawBaked(dst, face, x, y, false)
		}
	}
	if t.lit >= t.rows() {
		f := t.art.text
		text := fmt.Sprintf("%04d", t.bonus%10000)
		f.DrawShadowed(dst, text, f.CentreX(dst, text), int(tk(t.stopY)), resurrect64.Honey, resurrect64.Ink)
	}
}

// kittyArt is HERE KITTY's pictures in the new look.
type kittyArt struct {
	dog  [2]*ebiten.Image
	text *gfx.VectorFont
}

// drawSmooth draws HERE KITTY in the new look: the dog beckoning, and the
// words appearing, each letter a little higher than the last.
func (h *HereKitty) drawSmooth(dst *ebiten.Image) {
	if h.art == nil {
		h.art = &kittyArt{text: gfx.NewVectorFont(smallTextSize)}
		for i := range h.art.dog {
			h.art.dog[i] = entity.Bake(32, 68, resurrect64.Ink, func(p entity.Pen) { entity.DrawKittyDog(p, i, entity.WhiteDog) })
		}
	}
	dst.Fill(gfx.Palette[resurrect64.Ink])
	frame := h.ticks / 4
	entity.DrawBaked(dst, h.art.dog[frame%2], kittyDogX, kittyDogY, false)
	f := h.art.text
	for _, w := range kittyWords {
		if frame < w.frame {
			continue
		}
		for i, r := range w.text {
			x := int(tk(w.x+i*8)) + (32-f.CellWidth())/2
			y := int(tk(w.y-i*3)) - 6
			f.DrawShadowed(dst, string(r), x, y, resurrect64.Seafoam, resurrect64.Ink)
		}
	}
}

// rejectArt is the rejection's pictures in the new look.
type rejectArt struct {
	head, face *ebiten.Image
}

// grawlix is what Freddy says in his bubble, in lettering small enough
// for its five symbols to sit inside it.
const (
	grawlix        = "#@$%!"
	bubbleTextSize = 20
)

// drawSmooth draws the rejection in the new look: Freddy's grey head
// falling, then landed with a speech bubble of cartoon symbols.
func (r *Reject) drawSmooth(dst *ebiten.Image) {
	if r.art == nil {
		f := gfx.NewVectorFont(bubbleTextSize)
		r.art = &rejectArt{
			head: entity.Bake(config.RejectHeadW, config.RejectHeadH, resurrect64.Ink, entity.DrawRejectHead),
			face: entity.Bake(config.RejectFaceW, config.RejectFaceH, resurrect64.Ink, func(p entity.Pen) {
				entity.DrawRejectHead(p)
				drawBubble(p, f)
			}),
		}
	}
	dst.Fill(gfx.Palette[resurrect64.Ink])
	if r.landed {
		entity.DrawBaked(dst, r.art.face, config.RejectX, config.RejectLandY, false)
		return
	}
	entity.DrawBaked(dst, r.art.head, config.RejectX, r.y, false)
}

// drawBubble draws the speech bubble beside Freddy's head in the 192×68
// box the face fills: cyan, shaded underneath, its tail towards his mouth,
// holding the cartoon symbols in font f.
func drawBubble(p entity.Pen, f *gfx.VectorFont) {
	p.Poly(resurrect64.Jade, 116, 30, 130, 34, 100, 54)
	p.Ellipse(146, 22, 46, 20, 0, resurrect64.Jade)
	p.Poly(resurrect64.Seafoam, 116, 28, 128, 31, 104, 50)
	p.Ellipse(144, 19, 45, 18, 0, resurrect64.Seafoam)
	p.Ellipse(126, 8, 12, 3.5, -0.2, resurrect64.White)
	p.Text(f, grawlix, 144-float64(f.Width(grawlix))/2, 19-float64(f.LineHeight())/2, resurrect64.Ink)
}

// drawSmooth draws the dancing cats in the new look.
func (d *Dance) drawSmooth(dst *ebiten.Image) {
	if d.art[0] == nil {
		for i := range d.art {
			d.art[i] = entity.Bake(32, 35, resurrect64.Ink, func(p entity.Pen) { entity.DrawDanceCat(p, i) })
		}
	}
	dst.Fill(gfx.Palette[resurrect64.Ink])
	step := (d.ticks / config.DanceFrameTicks) % 2
	for _, c := range d.dancers {
		frame := step
		if c.Off {
			frame ^= 1
		}
		entity.DrawBaked(dst, d.art[frame], c.X, c.Y, false)
	}
}
