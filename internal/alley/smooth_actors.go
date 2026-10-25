package alley

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// The alley's actors in the new look, each in its sprite's box and
// colours, drawn by entity's pen.

// thrownH are the thrown objects' heights on the layout, by kind: boot,
// phone, bottle, pot (all 16 wide).
var thrownH = [...]int{9, 6, 12, 12}

// FeliciaCoat is Felicia's: a white cat with blue eyes.
var FeliciaCoat = entity.Coat{Fur: resurrect64.White, Shade: resurrect64.Mist, Light: resurrect64.Mist, Ear: resurrect64.Pink, Eye: resurrect64.Cobalt}

// drawLineS draws a clothesline and its laundry: a rope, and each piece
// pegged on by its top.
func (r *SmoothRenderer) drawLineS(dst *ebiten.Image, l *Line) {
	y := k(config.LineY[l.Row]) + 2
	w := float64(dst.Bounds().Dx())
	shapes.Line(dst, 0, y, w, y, 3, r.scheme.Trim)
	for x := 0.0; x < w; x += 6 {
		shapes.Line(dst, x, y+1, x+3, y-1, 1, r.scheme.Shade)
	}
	for i, p := range l.Pieces {
		px := l.Offset + i*config.LinePiecePx
		for h, kind := range p {
			x := px + h*config.LinePiecePx/2
			pic := entity.Picture{X: x, Y: config.LineY[l.Row], Outline: entity.NoOutline}
			var draw func(entity.Pen)
			switch {
			case kind == BigLeft:
				pic.W, pic.H, draw = 16, 16, drawShirt
			case kind == MediumLeft:
				pic.W, pic.H, draw = 16, 8, drawTowel
			case kind >= Small:
				pic.W, pic.H = 8, 8
				switch kind - Small {
				case 0:
					draw = func(p entity.Pen) { drawSock(p, resurrect64.Cerise, resurrect64.Flamingo, resurrect64.Wine, false) }
				case 1:
					draw = drawShorts
				default:
					draw = func(p entity.Pen) { drawSock(p, resurrect64.Fog, resurrect64.Mist, resurrect64.Lilac, true) }
				}
			default:
				continue
			}
			r.canvas.Draw(dst, pic, draw)
		}
	}
}

// peg draws a clothes peg over the line at x.
func peg(p entity.Pen, x float64) {
	p.Poly(resurrect64.Lilac, x-2, -2, x+2, -2, x+1.5, 8, x-1.5, 8)
	p.Circle(x, 1, 1, resurrect64.Shadow)
}

// drawShirt draws a white shirt hung by its shoulders, 64×64.
func drawShirt(p entity.Pen) {
	p.Poly(resurrect64.Mist, 0, 3, 23, 3, 32, 9, 41, 3, 64, 3, 63, 18, 51, 26, 51, 58, 13, 58, 13, 26, 1, 18)
	p.Poly(resurrect64.White, 0, 3, 23, 3, 32, 9, 41, 3, 60, 3, 59, 16, 47, 24, 47, 55, 13, 55, 13, 26, 1, 18)
	p.Stroke(resurrect64.Mist, 2, false, 23, 4, 32, 13, 41, 4)
	p.Stroke(resurrect64.Mist, 1.5, false, 13, 26, 15, 22)
	p.Stroke(resurrect64.Mist, 1.5, false, 51, 26, 49, 22)
	for _, y := range []float64{24, 34, 44} {
		p.Circle(31, y, 1.8, resurrect64.Fog)
	}
	p.Poly(resurrect64.Fog, 13, 55, 51, 55, 51, 59, 13, 59)
	peg(p, 10)
	peg(p, 54)
}

// drawTowel draws a striped towel, 64×32, its fringe hanging below.
func drawTowel(p entity.Pen) {
	p.Poly(resurrect64.Blue, 0, 3, 64, 3, 64, 27, 0, 27)
	p.Poly(resurrect64.Sky, 0, 3, 5, 3, 5, 27, 0, 27)
	p.Poly(resurrect64.Cobalt, 56, 3, 64, 3, 64, 27, 56, 27)
	for _, y := range []float64{11, 21} {
		p.Poly(resurrect64.White, 0, y, 64, y, 64, y+3.5, 0, y+3.5)
		p.Poly(resurrect64.Mist, 56, y, 64, y, 64, y+3.5, 56, y+3.5)
	}
	for x := 3.0; x < 62; x += 6 {
		p.Stroke(resurrect64.Blue, 2, false, x, 27, x, 31)
	}
	peg(p, 6)
	peg(p, 58)
}

// drawSock draws a sock hung by its top, 32×32: body, a lit side and a
// darker heel and toe; mirror points its toe to the left.
func drawSock(p entity.Pen, body, light, dark int, mirror bool) {
	x := func(v float64) float64 {
		if mirror {
			return 32 - v
		}
		return v
	}
	p.Poly(body, x(5), 2, x(15), 2, x(15), 19, x(27), 23, x(27), 30, x(10), 30, x(5), 25)
	p.Poly(light, x(5), 2, x(8), 2, x(8), 27, x(5), 25)
	p.Ellipse(x(24), 26.5, 4, 3.5, 0, dark)
	p.Ellipse(x(9), 26, 4, 4, 0, dark)
	p.Poly(dark, x(5), 2, x(15), 2, x(15), 5, x(5), 5)
	peg(p, x(10))
}

// drawShorts draws a pair of yellow shorts hung by the waistband, 32×32.
func drawShorts(p entity.Pen) {
	p.Poly(resurrect64.Honey, 1, 3, 31, 3, 31, 26, 20, 26, 17, 14, 15, 14, 12, 26, 1, 26)
	p.Poly(resurrect64.Caramel, 27, 3, 31, 3, 31, 26, 27, 26)
	p.Poly(resurrect64.Caramel, 1, 3, 31, 3, 31, 7, 1, 7)
	p.Stroke(resurrect64.Caramel, 1.5, false, 16, 7, 16, 14)
	peg(p, 5)
	peg(p, 27)
}

// drawMiceS draws the mice on their lines, each walking or still, or its
// points where it was caught.
func (r *SmoothRenderer) drawMiceS(dst *ebiten.Image, a *Alley) {
	for i, m := range a.Mice {
		y := config.MouseY[i]
		if m.Caught {
			r.drawText(dst, fmt.Sprint(config.MousePoints[i]), m.X, y, resurrect64.White, r.scheme.Outline, 2)
			continue
		}
		frame := -1
		if m.Dir != 0 {
			frame = m.Frame
		}
		r.canvas.Draw(dst, entity.Picture{X: m.X, Y: y, W: 16, H: 8, Flip: m.Dir < 0, Outline: entity.NoOutline},
			func(p entity.Pen) { drawMouse(p, frame) })
	}
}

// mouseFeet are where a mouse's feet are, by walk frame and standing
// (last), as its sprite's.
var mouseFeet = [3][2]float64{{14, 42}, {22, 34}, {18, 38}}

// drawMouse draws a grey mouse facing right, 64×32; frame is its walk
// frame, or -1 standing.
func drawMouse(p entity.Pen, frame int) {
	feet := mouseFeet[2]
	if frame >= 0 {
		feet = mouseFeet[frame%2]
	}
	p.Stroke(resurrect64.Lilac, 2, true, 10, 21, 4, 22, 1, 25, 4, 29, 8, 30)
	for _, x := range feet {
		p.Stroke(resurrect64.Lilac, 2.5, false, x, 24, x+1, 29)
		p.Ellipse(x+2.5, 29.5, 3, 1.6, 0, resurrect64.Lilac)
	}
	p.Ellipse(30, 18, 21, 9, 0, resurrect64.Fog)
	p.Stroke(resurrect64.Mist, 3, true, 14, 13, 22, 10.5, 32, 10)
	p.Ellipse(37, 24, 10, 3, 0, resurrect64.Lilac)
	p.Ellipse(50, 16, 11, 7.5, 0.15, resurrect64.Fog)
	p.Poly(resurrect64.Fog, 54, 11, 63, 18, 54, 22)
	p.Circle(62, 18, 2.2, resurrect64.Cerise)
	p.Circle(44, 7, 5.5, resurrect64.Fog)
	p.Circle(44.5, 7.5, 3, resurrect64.Pink)
	p.Circle(52, 13.5, 2, resurrect64.Ink)
	p.Circle(51.4, 12.8, 0.7, resurrect64.White)
	p.Stroke(resurrect64.Lilac, 0.8, false, 58, 19, 64, 16)
	p.Stroke(resurrect64.Lilac, 0.8, false, 58, 20, 64, 22)
}

// drawThrownS draws what is thrown from a window, showing only as much as
// has come out of it.
func (r *SmoothRenderer) drawThrownS(dst *ebiten.Image, t *Thrown) {
	if !t.Active {
		return
	}
	draw := [...]func(entity.Pen){drawBoot, drawPhone, drawBottle, drawPot}[t.Kind]
	r.canvas.Draw(dst, entity.Picture{X: t.X, Y: t.Y, W: 16, H: thrownH[t.Kind], Flip: t.Dir < 0, To: t.Visible(), Outline: entity.NoOutline}, draw)
}

// drawBoot draws an old boot, 64×36.
func drawBoot(p entity.Pen) {
	p.Poly(resurrect64.Coral, 20, 0, 34, 0, 35, 17, 50, 24, 53, 32, 13, 32, 13, 24, 19, 17)
	p.Poly(resurrect64.Caramel, 20, 0, 24, 0, 23, 18, 16, 25, 16, 32, 13, 32, 13, 24, 19, 17)
	p.Poly(resurrect64.Copper, 32, 0, 34, 0, 35, 17, 50, 24, 53, 32, 48, 32, 46, 27, 32, 19)
	p.Poly(resurrect64.Cerise, 12, 31, 54, 31, 54, 36, 12, 36)
	p.Poly(resurrect64.Wine, 44, 31, 54, 31, 54, 36, 44, 36)
	for _, y := range []float64{6, 11, 16} {
		p.Stroke(resurrect64.Copper, 1.5, false, 24, y, 31, y+1)
	}
}

// drawPhone draws an old telephone, 64×24.
func drawPhone(p entity.Pen) {
	p.Poly(resurrect64.Caramel, 8, 20, 56, 20, 56, 24, 8, 24)
	roundRect(p, 4, 2, 56, 19, 6, resurrect64.Honey)
	p.Circle(20, 6, 3, resurrect64.Ink)
	p.Circle(44, 6, 3, resurrect64.Ink)
	for x := 12.0; x <= 52; x += 8 {
		p.Circle(x, 14, 1.8, resurrect64.Ink)
	}
	p.Stroke(resurrect64.Caramel, 2, false, 56, 6, 58, 18)
}

// roundRect fills a rounded rectangle with the pen.
func roundRect(p entity.Pen, x, y, w, h, r float64, c int) {
	pts := []float64{}
	corner := func(cx, cy, a0 float64) {
		for i := range 5 {
			a := a0 + float64(i)*math.Pi/8
			pts = append(pts, cx+r*math.Cos(a), cy+r*math.Sin(a))
		}
	}
	corner(x+w-r, y+r, -math.Pi/2)
	corner(x+w-r, y+h-r, 0)
	corner(x+r, y+h-r, math.Pi/2)
	corner(x+r, y+r, math.Pi)
	p.Poly(c, pts...)
}

// drawBottle draws a green bottle, 64×48.
func drawBottle(p entity.Pen) {
	p.Poly(resurrect64.Emerald, 28, 0, 36, 0, 36, 13, 43, 20, 43, 45, 21, 45, 21, 20, 28, 13)
	p.Poly(resurrect64.Forest, 38, 16, 43, 20, 43, 45, 38, 45)
	p.Stroke(resurrect64.Leaf, 3, false, 25, 22, 25, 40)
	p.Poly(resurrect64.Forest, 27, 0, 37, 0, 37, 3, 27, 3)
}

// drawPot draws a flower pot with its plant, 64×48.
func drawPot(p entity.Pen) {
	for _, s := range [][4]float64{{32, 14, 18, 2}, {32, 14, 32, 0}, {32, 14, 46, 2}} {
		p.Stroke(resurrect64.Emerald, 2, false, s[0], s[1], s[2], s[3])
		p.Ellipse(s[2], s[3]+1, 4, 2.5, math.Atan2(s[3]-s[1], s[2]-s[0]), resurrect64.Leaf)
	}
	p.Poly(resurrect64.Coral, 16, 16, 48, 16, 42, 43, 22, 43)
	p.Poly(resurrect64.Caramel, 16, 16, 22, 16, 26, 43, 22, 43)
	p.Poly(resurrect64.Copper, 42, 16, 48, 16, 42, 43, 38, 43)
	p.Poly(resurrect64.Copper, 14, 16, 50, 16, 50, 22, 14, 22)
}

// drawTrashCatS draws the cat rising out of a trash can, as much of it as
// has risen.
func (r *SmoothRenderer) drawTrashCatS(dst *ebiten.Image, t *TrashCat) {
	if t.Phase == 0 || t.H <= 0 {
		return
	}
	pic := entity.Picture{X: t.X, Y: t.Top, W: 16, H: 13, To: t.H, Outline: resurrect64.Ink}
	r.canvas.Draw(dst, pic, drawTrashCat)
	// It is in a can's mouth: the can's front goes over it. Fully up, its
	// paws are hooked over the rim, whole, never cut as it rises.
	for _, c := range r.builtFor {
		if t.X >= c.X && t.X < c.X+30 {
			r.drawCanFront(dst, c)
			if t.H >= pic.H {
				pic.Outline, pic.To = entity.NoOutline, 0
				r.canvas.Draw(dst, pic, drawTrashPaws)
			}
			break
		}
	}
}

// drawTrashCat draws an orange tabby looking out, its front paws hooked
// over the rim, 64×52.
func drawTrashCat(p entity.Pen) {
	p.Poly(resurrect64.Copper, 8, 36, 56, 36, 60, 52, 4, 52)
	p.Ellipse(32, 44, 9, 7, 0, resurrect64.Honey)
	p.Poly(resurrect64.Coral, 4, 18, 8, 0, 20, 10)
	p.Poly(resurrect64.Copper, 44, 10, 56, 0, 60, 18)
	p.Poly(resurrect64.Cerise, 8, 12, 9, 4, 16, 10)
	p.Poly(resurrect64.Cerise, 48, 10, 55, 4, 56, 12)
	p.Ellipse(32, 22, 27, 16, 0, resurrect64.Coral)
	p.Ellipse(46, 23, 12, 14, 0, resurrect64.Copper)
	p.Ellipse(16, 18, 9, 7, 0, resurrect64.Caramel)
	for _, x := range []float64{26, 32, 38} {
		p.Stroke(resurrect64.Copper, 2, false, x, 8, x, 14)
	}
	for _, x := range []float64{20, 43} {
		p.Ellipse(x, 22, 4, 3.5, 0, resurrect64.Leaf)
		p.Stroke(resurrect64.Ink, 1.5, false, x, 19, x, 25)
	}
	p.Ellipse(32, 31, 9, 6, 0, resurrect64.Honey)
	p.Poly(resurrect64.Cerise, 29, 27, 35, 27, 32, 30)
	p.Stroke(resurrect64.Ink, 1.2, false, 32, 30, 29, 33)
	p.Stroke(resurrect64.Ink, 1.2, false, 32, 30, 35, 33)
	drawTrashPaws(p)
}

// drawTrashPaws draws the trash cat's front paws, at the foot of its box.
func drawTrashPaws(p entity.Pen) {
	for _, x := range []float64{12, 52} {
		p.Ellipse(x, 44, 9.5, 5.5, 0, resurrect64.Ink)
		p.Ellipse(x, 44, 8, 4.5, 0, resurrect64.Coral)
		for _, d := range []float64{-4, 0, 4} {
			p.Stroke(resurrect64.Ink, 1, false, x+d, 44, x+d, 48)
		}
	}
}

// drawWindowS shows the open window: darkness rising from the bottom of
// the glass as it opens, or Felicia in it.
func (r *SmoothRenderer) drawWindowS(dst *ebiten.Image, w *Window) {
	n := w.Openness(w.Index)
	if n == 0 {
		return
	}
	o := WindowRect(w.Index)
	y := o.Y + 1 + config.WindowPhaseHold - 1 - n
	shapes.FillRect(dst, shapes.R(k(o.X-1), k(y), k(o.W+2), k(n)), r.scheme.Outline)
	if w.Felicia {
		const h = 14
		r.canvas.Draw(dst, entity.Picture{X: o.X, Y: y, W: 32, H: h, From: max(h-n, 0), Outline: entity.NoOutline}, r.drawFelicia)
	}
}

// drawFelicia draws Felicia sitting in her window, 128×56, in the middle
// of "ME OW!", its letters staggered.
func (r *SmoothRenderer) drawFelicia(p entity.Pen) {
	entity.DrawSitting(p.At(18, 9, 0.8), FeliciaCoat)
	for _, l := range []struct {
		s    string
		x, y float64
	}{{"M", 0, 16}, {"E", 22, 8}, {"O", 82, 20}, {"W", 100, 8}, {"!", 116, 16}} {
		p.Text(r.small, l.s, l.x, l.y, resurrect64.Seafoam)
	}
}

// drawImpactS shows the impact word while the game is frozen: a ragged
// splat with slanting lettering.
func (r *SmoothRenderer) drawImpactS(dst *ebiten.Image, im Impact) {
	var word string
	var w, h int
	switch im.Kind {
	case Bonk:
		word, w, h = "BONK!", 40, 18
	case Squeak:
		word, w, h = "SQUEEK!", 51, 20
	default:
		return
	}
	x := min(im.X, gfx.ScreenWidth-w)
	r.canvas.Draw(dst, entity.Picture{X: x, Y: im.Y, W: w, H: h, Outline: entity.NoOutline},
		func(p entity.Pen) { entity.DrawSplat(p, r.text, word, w, h) })
}
