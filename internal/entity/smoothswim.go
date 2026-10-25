package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// swimH is the height of Freddy's swimming box on the layout, as his swim
// sprite's.
const swimH = 14

// DrawSwimming draws Freddy underwater in the new look, in his swimming
// pose (DESIGN §3.4, "Underwater"), in coat co, which carries the air
// warning: paddling along, tilted up or down, turned to face us, or sinking
// on his back as he drowns.
func (s *SmoothFreddy) DrawSwimming(dst *ebiten.Image, f *Freddy, co Coat) {
	sw := &f.Swim
	pic := Picture{X: f.X, Y: f.Y, W: freddyW, H: swimH, Flip: f.Facing < 0, Outline: NoOutline}
	angle, phi, kick := 0.0, 0.0, 1.0
	switch sw.Pose {
	case SwimSide:
		phi = float64(sw.Frame) * math.Pi / 2
	case SwimUp:
		angle, phi = -0.45, float64(sw.Frame)*math.Pi
	case SwimDown:
		angle, phi = 0.45, float64(sw.Frame)*math.Pi
	case SwimSink:
		// On his back, legs limp: turned half round, and mirrored back so
		// he still faces his way.
		angle, kick, pic.Flip = math.Pi, 0.3, !pic.Flip
	}
	s.canvas.Draw(dst, pic, func(p Pen) {
		q := NewPen(p.dst, freddyW*K/2, swimH*K/2, angle)
		if sw.Pose == SwimTurn {
			drawSwimFace(q, co)
			return
		}
		drawSwimmer(q, phi, kick, co)
	})
}

// drawSwimmer draws Freddy paddling, facing right, in a 96×56 box, in coat
// co: body level, tail trailing behind with a ripple, head held up, and his
// legs paddling in diagonal pairs at phase phi, kick strong, the front
// paws reaching forward and the hind ones kicking back.
func drawSwimmer(p Pen, phi, kick float64, co Coat) {
	paddle := func(jx, jy, ph, w float64, hind bool) {
		reach := 7.0
		if hind {
			reach = -7
		}
		fx := jx + reach + 11*kick*math.Sin(phi+ph)
		fy := jy + 8 + 4*kick*math.Cos(phi+ph)
		kx, ky := (jx+fx)/2, (jy+fy)/2+2
		p.Stroke(co.Fur, w, false, jx, jy, kx, ky, fx, fy)
		p.Ellipse(fx, fy, 3.6, 2.2, 0, co.Fur)
	}
	// The far legs, then the tail and body, then the near legs.
	paddle(40, 34, math.Pi, 4.5, true)
	paddle(68, 34, 0, 4.5, false)
	wave := 2.5 * math.Sin(phi)
	p.Stroke(co.Fur, 5, true, 32, 30, 22, 31+wave/2, 12, 29-wave, 3, 26+wave)
	p.Circle(3, 26+wave, 1.8, co.Light)
	p.Poly(co.Fur, 34, 22, 56, 22, 70, 20, 72, 38, 36, 38)
	p.Circle(37, 30, 8.5, co.Fur)
	p.Circle(68, 28, 9.5, co.Fur)
	p.Ellipse(36, 31, 7, 6, 0, co.Shade)
	p.Ellipse(37, 32, 5.5, 4.5, 0, co.Fur)
	p.Stroke(co.Light, 2.5, true, 34, 22.5, 46, 21.8, 58, 22, 66, 20.5)
	paddle(34, 34, 0, 5.5, true)
	paddle(62, 34, math.Pi, 5.5, false)
	drawHead(p, 78, 14, 0, co)
}

// drawSwimFace draws Freddy turning, face on to us, in a 96×56 box, in coat
// co: his legs splayed out either side as he treads water.
func drawSwimFace(p Pen, co Coat) {
	for _, side := range []float64{-1, 1} {
		p.Stroke(co.Fur, 5, false, 48+side*8, 32, 48+side*20, 28, 48+side*29, 33)
		p.Stroke(co.Fur, 5, false, 48+side*8, 42, 48+side*18, 46, 48+side*24, 52)
		p.Ellipse(48+side*30, 34, 3.4, 2.2, 0, co.Fur)
		p.Ellipse(48+side*25, 52, 3.4, 2.2, 0, co.Fur)
	}
	p.Ellipse(48, 37, 12, 11, 0, co.Fur)
	p.Stroke(co.Light, 2.2, true, 40, 31, 44, 27.5, 48, 27)
	drawFace(p, 48, 16, co)
}
