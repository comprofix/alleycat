package entity

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/shapes"
)

// Drawing for the new look (DESIGN §2.1): the actors drawn in code from
// shapes, on the same 320×200 layout, each of its pixels K screen pixels.
// The poses, boxes and colours are the sprites', so the game reads the
// same; only the drawing is finer.

// K is how many of the new look's screen pixels each of the 320×200
// layout's pixels covers.
const K = 4

// margin is the room round a picture's box on its canvas, for anything
// that reaches past the box: whiskers, a tail as Freddy jumps, an outline.
const margin = 16

// Pen draws a picture on its canvas. Points are in the picture's box, in
// screen pixels, facing right; the pen can move them, scale them about the
// box's corner and turn them about a centre.
type Pen struct {
	dst              *ebiten.Image
	ox, oy, s        float64
	cx, cy, sin, cos float64
}

// NewPen is a pen on dst turned by angle about (cx, cy) in the box.
func NewPen(dst *ebiten.Image, cx, cy, angle float64) Pen {
	return Pen{dst: dst, s: 1, cx: cx, cy: cy, sin: math.Sin(angle), cos: math.Cos(angle)}
}

// At is the pen drawing at scale s with its points moved to (ox, oy).
func (p Pen) At(ox, oy, s float64) Pen {
	p.ox, p.oy, p.s = ox, oy, s
	return p
}

// Pt is the canvas position of box point (x, y).
func (p Pen) Pt(x, y float64) (float64, float64) {
	x, y = p.ox+x*p.s, p.oy+y*p.s
	dx, dy := x-p.cx, y-p.cy
	return margin + p.cx + dx*p.cos - dy*p.sin, margin + p.cy + dx*p.sin + dy*p.cos
}

func (p Pen) pts(xy ...float64) []float64 {
	out := make([]float64, len(xy))
	for i := 0; i+1 < len(xy); i += 2 {
		out[i], out[i+1] = p.Pt(xy[i], xy[i+1])
	}
	return out
}

// Circle fills a circle of radius r about (x, y) in palette colour c.
func (p Pen) Circle(x, y, r float64, c int) {
	px, py := p.Pt(x, y)
	shapes.FillCircle(p.dst, px, py, r*p.s, c)
}

// Poly fills the polygon through the x, y pairs in palette colour c.
func (p Pen) Poly(c int, xy ...float64) { shapes.FillPoly(p.dst, c, p.pts(xy...)...) }

// Ellipse fills an ellipse with radii rx and ry about (x, y), tilted by
// tilt as well as the pen's turn.
func (p Pen) Ellipse(x, y, rx, ry, tilt float64, c int) {
	const n = 24
	pts := make([]float64, 0, 2*n)
	s, co := math.Sin(tilt), math.Cos(tilt)
	for i := range n {
		t := 2 * math.Pi * float64(i) / n
		ex, ey := rx*math.Cos(t), ry*math.Sin(t)
		pts = append(pts, x+ex*co-ey*s, y+ex*s+ey*co)
	}
	p.Poly(c, pts...)
}

// Stroke draws a line w wide through the points, rounded at the ends and
// corners, smoothed into a curve when curve is set.
func (p Pen) Stroke(c int, w float64, curve bool, xy ...float64) {
	if curve {
		xy = smoothCurve(xy, 6)
	}
	pts := p.pts(xy...)
	var path vector.Path
	path.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i+1 < len(pts); i += 2 {
		path.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(gfx.RGBA(c))
	vector.StrokePath(p.dst, &path, &vector.StrokeOptions{
		Width: float32(w * p.s), LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound,
	}, op)
}

// Text writes s in font f with its top-left at (x, y), in palette colour
// c. It is placed by the pen but neither scaled nor turned.
func (p Pen) Text(f gfx.Font, s string, x, y float64, c int) {
	px, py := p.Pt(x, y)
	f.Draw(p.dst, s, int(math.Round(px)), int(math.Round(py)), c)
}

// smoothCurve is a Catmull-Rom curve through the points, n steps between
// each pair.
func smoothCurve(xy []float64, n int) []float64 {
	m := len(xy) / 2
	at := func(i int) (float64, float64) {
		i = min(max(i, 0), m-1)
		return xy[2*i], xy[2*i+1]
	}
	out := []float64{xy[0], xy[1]}
	for i := range m - 1 {
		x0, y0 := at(i - 1)
		x1, y1 := at(i)
		x2, y2 := at(i + 1)
		x3, y3 := at(i + 2)
		for s := 1; s <= n; s++ {
			t := float64(s) / float64(n)
			t2, t3 := t*t, t*t*t
			f := func(a, b, c, d float64) float64 {
				return 0.5 * (2*b + (c-a)*t + (2*a-5*b+4*c-d)*t2 + (3*b-a-3*c+d)*t3)
			}
			out = append(out, f(x0, x1, x2, x3), f(y0, y1, y2, y3))
		}
	}
	return out
}

// NoOutline is a Picture with no outline.
const NoOutline = -1

// Picture is where and how a picture is laid on the screen, as its sprite
// would be: its w×h box on the layout with the top-left at (X, Y),
// mirrored when Flip is set. Only its rows From to To of the box are
// shown, the first of them at Y (To 0 shows them all), as a sprite cut by
// rows is. Outline, unless NoOutline, rims it in that palette colour.
type Picture struct {
	X, Y, W, H int
	Flip       bool
	From, To   int
	Outline    int
}

// Canvas is the image a picture is drawn on before it is laid on the
// screen, a render cache.
type Canvas struct{ img *ebiten.Image }

// Draw has draw draw the picture facing right on the canvas, with a pen
// whose points are the box's screen pixels, then lays it on dst as pic
// says.
func (c *Canvas) Draw(dst *ebiten.Image, pic Picture, draw func(Pen)) {
	w, h := pic.W*K+2*margin, pic.H*K+2*margin
	if c.img == nil || c.img.Bounds().Dx() < w || c.img.Bounds().Dy() < h {
		if c.img != nil {
			c.img.Deallocate()
		}
		c.img = ebiten.NewImage(max(w, 64), max(h, 64))
	}
	c.img.Clear()
	draw(NewPen(c.img, 0, 0, 0))

	to := pic.To
	if to <= 0 || to > pic.H {
		to = pic.H
	}
	if pic.From >= to {
		return
	}
	// The box's rows From to To, with the margin above and below kept
	// where the box is not cut.
	top, bot := 0, h
	if pic.From > 0 {
		top = margin + pic.From*K
	}
	if to < pic.H {
		bot = margin + to*K
	}
	img := c.img.SubImage(image.Rect(0, top, w, bot)).(*ebiten.Image)
	x, y := pic.X*K-margin, pic.Y*K-margin
	if pic.From > 0 {
		y = pic.Y * K // the cut row, not the margin, at Y
	}
	if pic.Outline != NoOutline {
		for _, d := range [8][2]int{{-2, 0}, {2, 0}, {0, -2}, {0, 2}, {-1, -1}, {1, -1}, {-1, 1}, {1, 1}} {
			DrawFlat(dst, img, x+d[0], y+d[1], pic.Flip, pic.Outline)
		}
	}
	DrawAt(dst, img, x, y, pic.Flip)
}
