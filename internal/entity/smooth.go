package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/physics"
)

// freddyW is the width of Freddy's box on the layout, in every pose, and
// freddySitH the height of his sitting pose (standing is config.FreddyH).
const (
	freddyW    = 24
	freddySitH = 14
)

// Coat is a cat's colours: fur, its shading and highlights, the inside of
// the ears, and the eyes. Freddy's is the sprite's; other cats drawn from
// his pieces wear their own.
type Coat struct {
	Fur, Shade, Light, Ear, Eye int
}

// FreddyCoat is Freddy's: a black cat with honey eyes.
var FreddyCoat = Coat{resurrect64.Ink, resurrect64.Shadow, resurrect64.Lilac, resurrect64.Mauve, resurrect64.Honey}

// SmoothFreddy draws Freddy in the new look. It holds only the canvas he
// is drawn on, a render cache.
type SmoothFreddy struct {
	canvas Canvas
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
func (l leg) draw(p Pen) {
	c, w := resurrect64.Ink, 6.0
	if l.far {
		w = 5
	}
	kx, ky := (l.jx+l.fx)/2+1.5, (l.jy+l.fy)/2
	if l.hind {
		kx -= 5
	}
	p.Stroke(c, w, false, l.jx, l.jy, kx, ky, l.fx, l.fy)
	p.Ellipse(l.fx+1.5, l.fy-1, 3.6, 2.4, 0, c)
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
	p := NewPen(dst, 52, 30, ps.angle)
	for _, l := range ps.legs {
		if l.far {
			l.draw(p)
		}
	}
	// The tail, then the body: haunch and chest joined along the back.
	p.Stroke(resurrect64.Ink, 5.5, true, ps.tail...)
	tx, ty := ps.tail[len(ps.tail)-4], ps.tail[len(ps.tail)-3]
	p.Circle(tx, ty, 1.8, resurrect64.Lilac)
	p.Poly(resurrect64.Ink, 32, 13, 56, 14, 72, 12, 74, 31, 34, 32)
	p.Circle(33, 23, 9.5, resurrect64.Ink)
	p.Circle(71, 22, 10, resurrect64.Ink)
	p.Ellipse(32, 24, 8, 7, 0, resurrect64.Shadow)
	p.Ellipse(33, 25, 6.5, 5.5, 0, resurrect64.Ink)
	p.Stroke(resurrect64.Lilac, 2.5, true, 30, 14.5, 44, 13.8, 58, 14, 66, 12.8)
	for _, l := range ps.legs {
		if !l.far {
			l.draw(p)
		}
	}
	drawHead(p, 78, 10)
}

// drawHead draws Freddy's head in profile, facing right, round (x, y): his
// ears, a honey eye and his muzzle, with whiskers.
func drawHead(p Pen, x, y float64) {
	p.Poly(resurrect64.Ink, x-11, y-1, x-10, y-10, x-3, y-6)
	p.Poly(resurrect64.Ink, x+2, y-7, x+6, y-10, x+8, y-1)
	p.Poly(resurrect64.Mauve, x-9.5, y-3, x-9, y-7.5, x-5, y-5.5)
	p.Poly(resurrect64.Mauve, x+3, y-5.5, x+5.5, y-7.5, x+6.5, y-3)
	p.Circle(x-1, y+1, 9, resurrect64.Ink)
	p.Ellipse(x+7, y+4, 5, 4, 0.1, resurrect64.Ink)
	p.Stroke(resurrect64.Lilac, 1.8, true, x-9, y-5, x-6, y-8, x-1, y-8.5)
	p.Circle(x+11, y+2.5, 1.4, resurrect64.Lilac) // the nose
	p.Ellipse(x+2, y, 2.6, 2.9, 0, resurrect64.Honey)
	p.Stroke(resurrect64.Ink, 1.3, false, x+2.4, y-2, x+2.4, y+2)
	p.Circle(x+1.3, y-1, 0.8, resurrect64.White)
	p.Stroke(resurrect64.Lilac, 0.8, false, x+8, y+5, x+17, y+3)
	p.Stroke(resurrect64.Lilac, 0.8, false, x+8, y+6, x+17, y+8)
}

// DrawSitting draws a cat in Freddy's sitting pose, facing out of the
// screen, in a 96×56 box (Freddy on a clothesline), in coat co.
func DrawSitting(p Pen, co Coat) {
	p.Stroke(co.Fur, 5.5, true, 60, 46, 72, 44, 79, 36, 81, 29, 78, 26)
	p.Circle(78, 27, 1.8, co.Light)
	p.Ellipse(50, 36, 16, 13, 0, co.Fur)
	p.Poly(co.Fur, 40, 24, 60, 24, 64, 36, 36, 36)
	p.Ellipse(41, 40, 8, 7.5, 0, co.Shade)
	p.Ellipse(42, 41, 6.5, 6, 0, co.Fur)
	p.Stroke(co.Light, 2.2, true, 38, 30, 36, 25, 40, 21)
	p.Stroke(co.Light, 2.2, true, 62, 30, 64, 25, 60, 21)
	for _, x := range []float64{38, 62} {
		p.Ellipse(x, 50, 5, 3, 0, co.Fur)
	}
	// The head, face on: both ears and both eyes.
	x, y := 50.0, 12.0
	p.Poly(co.Fur, x-10, y-1, x-9, y-12, x-2, y-7)
	p.Poly(co.Fur, x+2, y-7, x+9, y-12, x+10, y-1)
	p.Poly(co.Ear, x-8.5, y-3, x-8, y-9, x-4, y-6)
	p.Poly(co.Ear, x+4, y-6, x+8, y-9, x+8.5, y-3)
	p.Circle(x, y+1, 9.5, co.Fur)
	p.Stroke(co.Light, 1.8, true, x-8, y-6, x-5, y-9.5, x-2, y-8)
	for _, ex := range []float64{x - 4, x + 4} {
		p.Ellipse(ex, y+1, 2.4, 2.8, 0, co.Eye)
		p.Stroke(resurrect64.Ink, 1.2, false, ex, y-1, ex, y+3)
	}
	p.Circle(x, y+5.5, 1.3, co.Light)
}

// Draw draws Freddy as DrawFreddy does, onto the new look's screen. He has
// no rim in the new look: a black cat, as the owner chose.
func (s *SmoothFreddy) Draw(dst *ebiten.Image, f *Freddy) {
	if f.Hidden() || f.Gone {
		return
	}
	pic := Picture{X: f.X, Y: f.Y, W: freddyW, H: config.FreddyH, Flip: f.Facing < 0, Outline: NoOutline}
	if f.Seat != NotSeated {
		pic.H = freddySitH
	}
	if px, entering := f.EnterSlice(); entering {
		if px == 0 {
			return
		}
		pic.X = EdgeX(f.Facing > 0, px, freddyW)
	}
	if f.Depth == Behind {
		pic.To = config.FenceHideY - f.Y
		if pic.To <= 0 {
			return
		}
	}
	s.canvas.Draw(dst, pic, func(p Pen) {
		if f.Seat != NotSeated {
			DrawSitting(p, FreddyCoat)
		} else {
			drawCat(p.dst, freddyPose(f))
		}
	})
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
