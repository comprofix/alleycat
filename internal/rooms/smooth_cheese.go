package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// buildCheese draws the cheese room's backdrop, as the classic look lays
// it out (DESIGN §3.4, "Cheese room layout"): the walls, the window with
// its curtains, the chair, the pedestal table, and the wedge with its pits
// and holes.
func (r *SmoothRenderer) buildCheese() {
	b, ink := r.back, r.scheme.Outline
	r.buildWalls(wallY0, wallY1)
	r.buildWindow(cheeseOpeningX, openingY)
	fadePoly(b, ink, 0.3, oval(k(pedestalX)+76, k(pedestalY+34)-2, 48, 7)...)
	r.buildChair(cheeseChairX, furnitureY, false)
	r.buildPedestal(pedestalX, pedestalY)
	r.buildWedge()
}

// buildPedestal draws the small round table at (x, y) on the grid, in a
// 160×144 box: its top seen from a little above, one column and a
// spreading foot.
func (r *SmoothRenderer) buildPedestal(x, y int) {
	f := r.fur
	r.canvas.Draw(r.back, entity.Picture{X: x, Y: y, W: 40, H: 36, Outline: r.scheme.Outline}, func(p entity.Pen) {
		p.Poly(resurrect64.DeepTeal, 68, 114, 84, 114, 112, 134, 40, 134)
		p.Poly(f.body, 68, 114, 72, 114, 46, 134, 40, 134)
		p.Stroke(f.shade, 3, false, 40, 135, 112, 135)
		p.Stroke(resurrect64.DeepTeal, 12, false, 76, 20, 76, 116)
		p.Stroke(f.body, 3, false, 72, 22, 72, 114)
		p.Stroke(resurrect64.DeepTeal, 9, false, 6, 15, 150, 15)
		p.Ellipse(78, 8, 76, 8, 0, f.body)
		p.Ellipse(130, 9, 26, 6, 0, f.shade)
		p.Ellipse(60, 6, 40, 2.5, 0, f.light)
	})
}

// buildWedge draws the wedge of cheese where the classic look has it: its
// top lit, its far side in shade, a scatter of pits in a fixed pattern
// (drawing never uses the game's RNG), the 16 holes and its shadow on the
// floor.
func (r *SmoothRenderer) buildWedge() {
	s, b := r.scheme, r.back
	outline := []float64{k(wedgeX0), k(wedgeY0)}
	for _, e := range wedgeEdge {
		outline = append(outline, k(e[1]), k(e[0]))
	}
	outline = append(outline, k(wedgeX0), k(wedgeY1))
	bot := k(wedgeY1)
	fadePoly(b, s.Outline, 0.35, k(wedgeX0)-4, bot-4, k(216)+8, bot-4, k(216)+24, bot+12, k(wedgeX0)-16, bot+12)
	shapes.FillPoly(b, s.Cheese, outline...)

	// The far side in shade: a band along the right edge, solid by the
	// edge and fading further in.
	edge := func(in float64) []float64 {
		pts := []float64{}
		for _, e := range wedgeEdge {
			pts = append(pts, k(e[1])-in, k(e[0]))
		}
		return pts
	}
	band := func(in float64) []float64 {
		pts := edge(0)
		inner := edge(in)
		for i := len(inner) - 2; i >= 0; i -= 2 {
			pts = append(pts, inner[i], inner[i+1])
		}
		return pts
	}
	fadePoly(b, s.CheeseShade, 0.55, band(32)...)
	shapes.FillPoly(b, s.CheeseShade, band(14)...)
	shapes.StrokePath(b, polyline(edge(2)), 4, s.Hole)
	// The rind's line round the top and down the left.
	shapes.Line(b, k(wedgeX0), k(wedgeY0)+1, k(wedgeEdge[0][1]), k(wedgeY0)+1, 3, s.Hole)
	shapes.Line(b, k(wedgeX0)+1, k(wedgeY0), k(wedgeX0)+1, bot, 3, s.CheeseShade)

	for y := wedgeY0 + 4; y < wedgeY1-4; y += pitCell {
		for x := wedgeX0 + 4; x < wedgeRight(y)-6; x += pitCell {
			if h := (x*7 + y*13) % 11; h < 3 {
				cx, cy := k(x+h)+4, k(y+h)+2
				rx, ry := 5+float64(h)*1.5, 3.5+float64(h)
				shapes.FillPoly(b, s.Hole, oval(cx, cy, rx, ry)...)
				shapes.FillPoly(b, s.CheeseShade, oval(cx, cy+1.5, rx*0.9, ry*0.75)...)
			}
		}
	}
	for h := range config.CheeseHoleCount {
		x, y := holeDrawAt(h)
		r.drawHole(b, k(x), k(y), h < config.CheeseTopHoles)
	}
}

// drawHole draws a hole in the wedge in its 64-pixel box at (x, y): in
// the face, a cut rim round a deep inside, darkest at the back under the
// top, its floor lit, with a lit lip below; in the top, seen from above,
// flatter.
func (r *SmoothRenderer) drawHole(b *ebiten.Image, x, y float64, top bool) {
	s := r.scheme
	if top {
		shapes.FillPoly(b, s.Hole, oval(x+32, y+28, 30, 10)...)
		shapes.FillPoly(b, resurrect64.Rust, oval(x+32, y+29, 27, 8)...)
		shapes.FillPoly(b, resurrect64.Maroon, oval(x+32, y+27, 24, 5)...)
		shapes.FillPoly(b, s.Hole, oval(x+32, y+35, 20, 2.5)...)
		shapes.Line(b, x+12, y+40, x+52, y+40, 3, s.CheeseShade)
		return
	}
	shapes.FillPoly(b, s.Hole, oval(x+32, y+29, 30, 27)...)
	shapes.FillPoly(b, resurrect64.Rust, oval(x+32, y+30, 27, 24)...)
	shapes.FillPoly(b, resurrect64.Maroon, oval(x+32, y+23, 24, 15)...)
	shapes.FillPoly(b, s.Hole, oval(x+32, y+44, 21, 9)...)
	shapes.StrokePath(b, polyline([]float64{x + 12, y + 56, x + 22, y + 60, x + 32, y + 61, x + 42, y + 60, x + 52, y + 56}), 4, s.CheeseShade)
}

// drawCheese draws the mice in their holes, Freddy going through a hole,
// and the caught mice along the top of the screen.
func (r *SmoothRenderer) drawCheese(dst *ebiten.Image, c *CheeseRoom) {
	for i := range c.Mice {
		m := &c.Mice[i]
		if !m.Visible() {
			continue
		}
		// Peeking shows the top of its head, half out its face, out the
		// whole mouse, its whiskers twitching, as its sprite's frames.
		pic := entity.Picture{W: config.CheeseMouseW, H: config.CheeseMouseH, Outline: entity.NoOutline}
		frame := 2 + m.Frame
		switch m.Count {
		case config.MouseHidden - 1, 0:
			frame, pic.To = 0, 5
		case config.MouseEmerge, 1:
			frame, pic.To = 1, 8
		}
		mr := c.MouseRect(i)
		pic.X, pic.Y = mr.X, mr.Y
		r.canvas.Draw(dst, pic, func(p entity.Pen) { drawHoleMouse(p, frame) })
	}
	if t := &c.Transfer; t.Active {
		h := t.From
		if t.Frame >= config.TransferSwapAfter {
			h = t.To
		}
		x, y := holeDrawAt(h)
		r.canvas.Draw(dst, entity.Picture{X: x, Y: y, W: 16, H: 16, Outline: entity.NoOutline}, func(p entity.Pen) {
			drawFreddyHole(p, t.Frame)
		})
	}
	for i := range c.Caught {
		r.canvas.Draw(dst, entity.Picture{X: config.MouseIconX + i*config.MouseIconStep, Y: config.MouseIconY, W: 14, H: 8, Outline: entity.NoOutline},
			func(p entity.Pen) { entity.DrawMouse(p.At(0, 0, 0.875), -1) })
	}
}

// drawHoleMouse draws a mouse facing out of its hole in a 64×48 box: big
// ears, a pointed face with a pink nose, and, when right out, its body and
// paws on the hole's floor with whiskers twitching by frame.
func drawHoleMouse(p entity.Pen, frame int) {
	if frame >= 2 {
		p.Ellipse(32, 30, 13, 9, 0, resurrect64.Fog)
		p.Ellipse(32, 33, 7, 5, 0, resurrect64.Lilac)
		for _, x := range []float64{22, 28, 36, 42} {
			p.Ellipse(x, 38, 3.5, 2.5, 0, resurrect64.Fog)
		}
	}
	for _, x := range []float64{9, 55} {
		p.Circle(x, 8, 8, resurrect64.Fog)
		p.Circle(x, 9, 4.5, resurrect64.Pink)
	}
	p.Ellipse(32, 15, 22, 9, 0, resurrect64.Fog)
	p.Poly(resurrect64.Fog, 12, 16, 52, 16, 37, 30, 27, 30)
	p.Stroke(resurrect64.Mist, 3, true, 20, 9, 32, 7, 44, 9)
	for _, x := range []float64{21, 43} {
		p.Circle(x, 16, 2.6, resurrect64.Ink)
		p.Circle(x-0.7, 15.3, 0.9, resurrect64.White)
	}
	p.Circle(32, 28, 3, resurrect64.Cerise)
	if frame >= 2 {
		w := float64(frame-2) * 3
		for _, side := range []float64{-1, 1} {
			p.Stroke(resurrect64.Lilac, 0.9, false, 32+side*4, 28, 32+side*16, 23+w)
			p.Stroke(resurrect64.Lilac, 0.9, false, 32+side*4, 29, 32+side*16, 31-w)
		}
	}
}

// drawFreddyHole draws Freddy going through a hole in a 64×64 box over
// it, as its sprite's frames: diving in rump first, his tail last (0–2),
// then out of the far hole, ears, then head, then head and paws (3–5).
func drawFreddyHole(p entity.Pen, frame int) {
	ink, ear, eye := resurrect64.Ink, resurrect64.Mauve, resurrect64.Honey
	head := func(y float64) {
		p.Poly(ink, 14, y+2, 18, y-14, 28, y-2)
		p.Poly(ink, 50, y+2, 46, y-14, 36, y-2)
		p.Poly(ear, 17, y-1, 19, y-9, 24, y-2)
		p.Poly(ear, 47, y-1, 45, y-9, 40, y-2)
		p.Ellipse(32, y+8, 18, 13, 0, ink)
		for _, x := range []float64{24, 40} {
			p.Ellipse(x, y+6, 3.5, 3, 0, eye)
			p.Stroke(ink, 1.6, false, x, y+3.5, x, y+8.5)
		}
	}
	switch frame {
	case 0:
		// Head first, his rump and tail up and his hind legs on the rim.
		p.Stroke(ink, 5, true, 44, 14, 50, 6, 56, 2)
		p.Ellipse(30, 24, 20, 10, -0.55, ink)
		p.Stroke(ink, 6, false, 40, 20, 46, 38)
		p.Stroke(ink, 6, false, 36, 24, 38, 38)
	case 1:
		p.Stroke(ink, 5, true, 38, 18, 44, 8, 42, 0)
		p.Ellipse(32, 20, 14, 7, 0, ink)
	case 2:
		p.Stroke(ink, 5, true, 30, 2, 34, 12, 31, 22)
	case 3:
		p.Poly(ink, 14, 26, 18, 16, 26, 26)
		p.Poly(ink, 50, 26, 46, 16, 38, 26)
	case 4:
		head(18)
	default:
		head(10)
		p.Stroke(ink, 6, true, 8, 36, 16, 30, 26, 32)
		p.Stroke(ink, 6, true, 56, 36, 48, 30, 38, 32)
	}
}
