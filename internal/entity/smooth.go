package entity

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/physics"
	"github.com/comprofix/alleycat/internal/shapes"
)

// Drawing for the new look (DESIGN §2.1): the actors drawn in code from
// shapes, on the same 320×200 layout, each of its pixels K screen pixels.
// The poses, boxes and colours are the sprites', so the game reads the
// same; only the drawing is finer.

// K is how many of the new look's screen pixels each of the 320×200
// layout's pixels covers.
const K = 4

// freddyW is the width of Freddy's box on the layout, in every pose, and
// freddySitH the height of his sitting pose (standing is config.FreddyH).
const (
	freddyW    = 24
	freddySitH = 14
)

// margin is the room round Freddy's box on his canvas, for anything that
// reaches past the box: his whiskers, his tail as he jumps.
const margin = 16

// SmoothFreddy draws Freddy in the new look. It holds only the canvas he
// is drawn on, a render cache.
type SmoothFreddy struct {
	canvas *ebiten.Image
}

// pen draws a pose on Freddy's canvas: points are in the box's own screen
// pixels, facing right, turned by an angle about a centre for the jumps.
type pen struct {
	dst              *ebiten.Image
	cx, cy, sin, cos float64
}

func newPen(dst *ebiten.Image, cx, cy, angle float64) pen {
	return pen{dst: dst, cx: cx, cy: cy, sin: math.Sin(angle), cos: math.Cos(angle)}
}

// pt is the canvas position of box point (x, y).
func (p pen) pt(x, y float64) (float64, float64) {
	dx, dy := x-p.cx, y-p.cy
	return margin + p.cx + dx*p.cos - dy*p.sin, margin + p.cy + dx*p.sin + dy*p.cos
}

func (p pen) pts(xy ...float64) []float64 {
	out := make([]float64, len(xy))
	for i := 0; i+1 < len(xy); i += 2 {
		out[i], out[i+1] = p.pt(xy[i], xy[i+1])
	}
	return out
}

func (p pen) circle(x, y, r float64, c int) {
	px, py := p.pt(x, y)
	shapes.FillCircle(p.dst, px, py, r, c)
}

func (p pen) poly(c int, xy ...float64) { shapes.FillPoly(p.dst, c, p.pts(xy...)...) }

// ellipse fills an ellipse with radii rx and ry about (x, y), tilted by
// tilt as well as the pen's turn.
func (p pen) ellipse(x, y, rx, ry, tilt float64, c int) {
	const n = 24
	pts := make([]float64, 0, 2*n)
	s, co := math.Sin(tilt), math.Cos(tilt)
	for i := range n {
		t := 2 * math.Pi * float64(i) / n
		ex, ey := rx*math.Cos(t), ry*math.Sin(t)
		pts = append(pts, x+ex*co-ey*s, y+ex*s+ey*co)
	}
	p.poly(c, pts...)
}

// stroke draws a line w wide through the points, rounded at the ends and
// corners, smoothed into a curve when curve is set.
func (p pen) stroke(c int, w float64, curve bool, xy ...float64) {
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
		Width: float32(w), LineJoin: vector.LineJoinRound, LineCap: vector.LineCapRound,
	}, op)
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

// leg is one of Freddy's legs: where it joins the body and where its paw
// is, both in the box's screen pixels.
type leg struct {
	jx, jy, fx, fy float64
	hind, far      bool
}

// draw draws the leg in two parts, bending at the knee: forwards for a
// front leg, back at the hock for a hind one. A far leg is a little
// thinner; it stays black, as shadow would vanish into the cobbles.
func (l leg) draw(p pen) {
	c, w := resurrect64.Ink, 6.0
	if l.far {
		w = 5
	}
	kx, ky := (l.jx+l.fx)/2+1.5, (l.jy+l.fy)/2
	if l.hind {
		kx -= 5
	}
	p.stroke(c, w, false, l.jx, l.jy, kx, ky, l.fx, l.fy)
	p.ellipse(l.fx+1.5, l.fy-1, 3.6, 2.4, 0, c)
}

// pose is everything that changes between Freddy's poses: his turn, his
// legs and the line of his tail.
type pose struct {
	angle float64
	legs  [4]leg
	tail  []float64
}

// joints are where the legs join the body: hind near and far, front near
// and far.
var joints = [4][2]float64{{30, 27}, {40, 27}, {68, 27}, {78, 27}}

// legsAt are the legs with each paw moved by the given offsets from under
// its joint, standing on the ground at the foot of the box.
func legsAt(off [4][2]float64) [4]leg {
	var ls [4]leg
	for i, j := range joints {
		ls[i] = leg{jx: j[0], jy: j[1], fx: j[0] + off[i][0], fy: float64(config.FreddyH*K) - 2 + off[i][1], hind: i < 2, far: i%2 == 1}
	}
	return ls
}

// tailUp is his tail held up and curled forward at the tip, as when he
// stands and walks; sway moves the tip.
func tailUp(sway float64) []float64 {
	return []float64{26, 20, 16, 19, 10, 13, 9 + sway/2, 6, 13 + sway, 2, 19 + sway, 4}
}

// freddyPose is the pose for Freddy's state, as FreddyImageIn picks his
// sprite.
func freddyPose(f *Freddy) pose {
	switch {
	case f.Arc.Dir == physics.Rising:
		return pose{
			angle: -0.3,
			legs:  legsAt([4][2]float64{{-13, -5}, {-8, -3}, {11, -11}, {13, -13}}),
			tail:  []float64{26, 21, 16, 23, 8, 22, 3, 18},
		}
	case f.Arc.Dir == physics.Falling:
		return pose{
			legs: legsAt([4][2]float64{{-10, 0}, {-4, -1}, {8, -1}, {12, 0}}),
			tail: tailUp(0),
		}
	case f.Dir != 0:
		// The walk: each leg swings in turn, hind then front, near then
		// far, a quarter of a stride apart, its paw lifted as it swings
		// forward.
		phi := 2 * math.Pi * float64(f.WalkFrame) / 6 // six walk frames
		var off [4][2]float64
		for i, ph := range [4]float64{0, math.Pi, math.Pi / 2, 3 * math.Pi / 2} {
			off[i] = [2]float64{7 * math.Sin(phi+ph), -4 * max(0, math.Cos(phi+ph))}
		}
		return pose{legs: legsAt(off), tail: tailUp(2 * math.Sin(phi))}
	}
	return pose{legs: legsAt([4][2]float64{{-2, 0}, {0, 0}, {0, 0}, {2, 0}}), tail: tailUp(0)}
}

// drawCat draws Freddy standing, walking or in the air, facing right in a
// 96×44 box.
func drawCat(dst *ebiten.Image, ps pose) {
	p := newPen(dst, 52, 30, ps.angle)
	for _, l := range ps.legs {
		if l.far {
			l.draw(p)
		}
	}
	// The tail, then the body: haunch and chest joined along the back.
	p.stroke(resurrect64.Ink, 5.5, true, ps.tail...)
	tx, ty := ps.tail[len(ps.tail)-4], ps.tail[len(ps.tail)-3]
	p.circle(tx, ty, 1.8, resurrect64.Lilac)
	p.poly(resurrect64.Ink, 32, 13, 56, 14, 72, 12, 74, 31, 34, 32)
	p.circle(33, 23, 9.5, resurrect64.Ink)
	p.circle(71, 22, 10, resurrect64.Ink)
	p.ellipse(32, 24, 8, 7, 0, resurrect64.Shadow)
	p.ellipse(33, 25, 6.5, 5.5, 0, resurrect64.Ink)
	p.stroke(resurrect64.Lilac, 2.5, true, 30, 14.5, 44, 13.8, 58, 14, 66, 12.8)
	for _, l := range ps.legs {
		if !l.far {
			l.draw(p)
		}
	}
	drawHead(p, 78, 10)
}

// drawHead draws Freddy's head in profile, facing right, round (x, y): his
// ears, a honey eye and his muzzle, with whiskers.
func drawHead(p pen, x, y float64) {
	p.poly(resurrect64.Ink, x-11, y-1, x-10, y-10, x-3, y-6)
	p.poly(resurrect64.Ink, x+2, y-7, x+6, y-10, x+8, y-1)
	p.poly(resurrect64.Mauve, x-9.5, y-3, x-9, y-7.5, x-5, y-5.5)
	p.poly(resurrect64.Mauve, x+3, y-5.5, x+5.5, y-7.5, x+6.5, y-3)
	p.circle(x-1, y+1, 9, resurrect64.Ink)
	p.ellipse(x+7, y+4, 5, 4, 0.1, resurrect64.Ink)
	p.stroke(resurrect64.Lilac, 1.8, true, x-9, y-5, x-6, y-8, x-1, y-8.5)
	p.circle(x+11, y+2.5, 1.4, resurrect64.Lilac) // the nose
	p.ellipse(x+2, y, 2.6, 2.9, 0, resurrect64.Honey)
	p.stroke(resurrect64.Ink, 1.3, false, x+2.4, y-2, x+2.4, y+2)
	p.circle(x+1.3, y-1, 0.8, resurrect64.White)
	p.stroke(resurrect64.Lilac, 0.8, false, x+8, y+5, x+17, y+3)
	p.stroke(resurrect64.Lilac, 0.8, false, x+8, y+6, x+17, y+8)
}

// drawSitting draws Freddy sitting on a clothesline, facing out of the
// screen, in a 96×56 box.
func drawSitting(dst *ebiten.Image) {
	p := newPen(dst, 0, 0, 0)
	p.stroke(resurrect64.Ink, 5.5, true, 60, 46, 72, 44, 79, 36, 81, 29, 78, 26)
	p.circle(78, 27, 1.8, resurrect64.Lilac)
	p.ellipse(50, 36, 16, 13, 0, resurrect64.Ink)
	p.poly(resurrect64.Ink, 40, 24, 60, 24, 64, 36, 36, 36)
	p.ellipse(41, 40, 8, 7.5, 0, resurrect64.Shadow)
	p.ellipse(42, 41, 6.5, 6, 0, resurrect64.Ink)
	p.stroke(resurrect64.Lilac, 2.2, true, 38, 30, 36, 25, 40, 21)
	p.stroke(resurrect64.Lilac, 2.2, true, 62, 30, 64, 25, 60, 21)
	for _, x := range []float64{38, 62} {
		p.ellipse(x, 50, 5, 3, 0, resurrect64.Ink)
	}
	// The head, face on: both ears and both eyes.
	x, y := 50.0, 12.0
	p.poly(resurrect64.Ink, x-10, y-1, x-9, y-12, x-2, y-7)
	p.poly(resurrect64.Ink, x+2, y-7, x+9, y-12, x+10, y-1)
	p.poly(resurrect64.Mauve, x-8.5, y-3, x-8, y-9, x-4, y-6)
	p.poly(resurrect64.Mauve, x+4, y-6, x+8, y-9, x+8.5, y-3)
	p.circle(x, y+1, 9.5, resurrect64.Ink)
	p.stroke(resurrect64.Lilac, 1.8, true, x-8, y-6, x-5, y-9.5, x-2, y-8)
	for _, ex := range []float64{x - 4, x + 4} {
		p.ellipse(ex, y+1, 2.4, 2.8, 0, resurrect64.Honey)
		p.stroke(resurrect64.Ink, 1.2, false, ex, y-1, ex, y+3)
	}
	p.circle(x, y+5.5, 1.3, resurrect64.Lilac)
}

// Draw draws Freddy as DrawFreddy does, onto the new look's screen. He has
// no rim in the new look: a black cat, as the owner chose.
func (s *SmoothFreddy) Draw(dst *ebiten.Image, f *Freddy) {
	if f.Hidden() || f.Gone {
		return
	}
	if s.canvas == nil {
		s.canvas = ebiten.NewImage(freddyW*K+2*margin, freddySitH*K+2*margin)
	}
	s.canvas.Clear()
	h := config.FreddyH
	if f.Seat != NotSeated {
		h = freddySitH
		drawSitting(s.canvas)
	} else {
		drawCat(s.canvas, freddyPose(f))
	}
	x := f.X
	if px, entering := f.EnterSlice(); entering {
		if px == 0 {
			return
		}
		x = EdgeX(f.Facing > 0, px, freddyW)
	}
	img := s.canvas
	if f.Depth == Behind {
		vis := config.FenceHideY - f.Y
		if vis <= 0 {
			return
		}
		if vis < h {
			img = img.SubImage(image.Rect(0, 0, img.Bounds().Dx(), margin+vis*K)).(*ebiten.Image)
		}
	}
	DrawAt(dst, img, x*K-margin, f.Y*K-margin, f.Facing < 0)
	DrawSmoothDust(dst, f)
}

// DrawSmoothDust draws the puff from Freddy's last landing in the new
// look: two little clouds rolling out to either side and thinning away.
func DrawSmoothDust(dst *ebiten.Image, f *Freddy) {
	if f.Dust <= 0 {
		return
	}
	t := float64(config.DustUpdates-f.Dust) / config.DustUpdates
	cx := float64((f.DustX+config.DustOffsetX)*K) + 64
	y := float64((f.DustY+config.FreddyH)*K) - 4
	a := 0.85 * (1 - t)
	for _, side := range []float64{-1, 1} {
		for i, d := range []float64{0, 9, 17} {
			x := cx + side*(10+d+44*t)
			r := 6 - float64(i)*1.4 + 3*t
			c := resurrect64.Fog
			if i == 0 {
				c = resurrect64.Mist
			}
			col := gfx.RGBA(c)
			col.R, col.G, col.B, col.A = uint8(float64(col.R)*a), uint8(float64(col.G)*a), uint8(float64(col.B)*a), uint8(255*a)
			vector.FillCircle(dst, float32(x), float32(y-r*0.6-6*t), float32(r), col, true)
		}
	}
}
