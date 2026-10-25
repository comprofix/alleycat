package entity

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// dogW and dogH are the dog's box on the layout, running or fighting.
const dogW, dogH = 32, 15

// DogCoat is a bulldog's colours: coat, its shading, the far side and
// belly, and the collar with its shade.
type DogCoat struct {
	Coat, Shade, Far, Collar, CollarShade int
}

// WhiteDog is the alley dog's coat, as its sprite.
var WhiteDog = DogCoat{resurrect64.White, resurrect64.Fog, resurrect64.Mist, resurrect64.Cerise, resurrect64.Wine}

// SmoothDog draws the dog in the new look. It holds only its canvas, a
// render cache.
type SmoothDog struct {
	canvas Canvas
}

// Draw draws the dog as DrawDog does, onto the new look's screen: running
// in coat co, or the fight cloud after a catch.
func (s *SmoothDog) Draw(dst *ebiten.Image, d *Dog, co DogCoat) {
	if !d.Out() {
		return
	}
	pic := Picture{W: dogW, H: dogH, Flip: d.Dir < 0, Outline: resurrect64.Ink}
	if d.Capture != 0 {
		pic.Flip, pic.Outline = d.Frame%2 == 1, NoOutline
	}
	pic.X = d.X
	switch {
	case d.Entering > 0:
		pic.X = EdgeX(d.Dir > 0, (config.DogEdgeSteps+1-d.Entering)*8, dogW)
	case d.Exiting > 0:
		pic.X = EdgeX(d.X < gfx.ScreenWidth/2, (d.Exiting-1)*8, dogW)
	}
	pic.Y = config.DogY
	if d.Y != 0 {
		pic.Y = d.Y
	}
	s.canvas.Draw(dst, pic, func(p Pen) {
		if d.Capture != 0 {
			drawFight(p, d.Frame)
		} else {
			drawDog(p, d.Frame, co)
		}
	})
}

// dogLegs are the running frames' legs, hind near and far, front near and
// far: hip and paw, in the 128×60 box. Frame 0 is the stretch, frame 1
// the gather, as the sprite's two.
var dogLegs = [2][4][4]float64{
	{{38, 40, 22, 55}, {50, 40, 60, 55}, {80, 40, 82, 55}, {92, 40, 112, 54}},
	{{38, 40, 34, 55}, {48, 40, 52, 55}, {82, 40, 86, 55}, {92, 40, 102, 55}},
}

// drawDog draws a white bulldog running, facing right, in a 128×60 box.
func drawDog(p Pen, frame int, co DogCoat) {
	legs := dogLegs[frame%2]
	leg := func(l [4]float64, c int) {
		p.Stroke(c, 9, false, l[0], l[1], (l[0]+l[2])/2, (l[1]+l[3])/2+1, l[2], l[3])
		p.Ellipse(l[2]+2, l[3]+1, 6, 3.5, 0, c)
	}
	leg(legs[1], co.Far)
	leg(legs[3], co.Far)
	// A stub of a tail, then the barrel of a body, deep in the chest.
	p.Stroke(co.Coat, 6, false, 34, 22, 40, 12)
	p.Ellipse(58, 30, 30, 13, -0.06, co.Coat)
	p.Circle(40, 29, 13, co.Coat)
	p.Circle(80, 28, 16, co.Coat)
	p.Ellipse(34, 33, 6, 8, 0, co.Shade)
	p.Ellipse(60, 41, 26, 4, 0, co.Far)
	leg(legs[0], co.Coat)
	leg(legs[2], co.Coat)
	// The collar, then the broad head: jowls, a flat black nose, an eye in
	// a dark patch, and a folded ear.
	p.Poly(co.Collar, 76, 12, 84, 13, 82, 40, 73, 38)
	p.Stroke(co.CollarShade, 2.5, false, 82.5, 14, 80.5, 39)
	p.Circle(81, 41, 3, resurrect64.Honey)
	p.Circle(100, 21, 15, co.Coat)
	p.Ellipse(108, 30, 15, 9, 0.05, co.Coat)
	p.Ellipse(110, 34, 10, 4, 0, co.Far)
	p.Ellipse(105, 15, 6, 5, 0, co.Shade)
	p.Circle(105, 15, 2.6, resurrect64.Ink)
	p.Circle(104, 14, 0.9, resurrect64.White)
	p.Ellipse(121, 19, 5, 4, 0.2, resurrect64.Ink)
	p.Stroke(resurrect64.Ink, 2, true, 112, 31, 118, 32, 124, 28)
	p.Poly(resurrect64.Ink, 86, 12, 90, 2, 99, 6, 95, 13)
}

// drawFight draws the cloud the dog and Freddy make fighting: a tumble of
// dust with stars over it and the odd paw, tail and collar flying out. The
// frame shakes the cloud, as the sprite is mirrored on alternate steps.
func drawFight(p Pen, frame int) {
	j := float64(frame%2) * 2
	puffs := [][3]float64{
		{18, 38, 13}, {36, 30, 15}, {58, 26, 17}, {80, 28, 16}, {100, 32, 14}, {114, 40, 10},
		{28, 44, 11}, {50, 42, 13}, {74, 42, 13}, {96, 44, 11},
	}
	for _, c := range puffs {
		p.Circle(c[0], c[1]+3, c[2], resurrect64.Lilac)
	}
	for _, c := range puffs {
		p.Circle(c[0], c[1], c[2], resurrect64.Mist)
	}
	for _, c := range puffs[:6] {
		p.Circle(c[0]-2, c[1]-3, c[2]*0.8, resurrect64.White)
	}
	// What flies out of it: a black paw, a white paw, Freddy's tail, the
	// collar.
	p.Stroke(resurrect64.Ink, 5, false, 22, 30+j, 8, 22+j)
	p.Stroke(resurrect64.Ink, 5, true, 88, 22, 96, 12-j, 104, 10, 108, 14)
	p.Stroke(resurrect64.Fog, 7, false, 70, 40, 82, 54+j)
	p.Ellipse(83, 55+j, 5, 3, 0, resurrect64.Fog)
	p.Stroke(resurrect64.Cerise, 4, false, 52, 30, 62, 34)
	p.Stroke(resurrect64.Ink, 4, false, 92, 48, 104, 56)
	p.Ellipse(16, 54, 5, 3, 0, resurrect64.Ink)
	for _, s := range [][2]float64{{30, 6}, {74, 6}} {
		star(p, s[0], s[1]+j, 6, resurrect64.Honey)
	}
}

// star draws a four-pointed star of radius r about (x, y).
func star(p Pen, x, y, r float64, c int) {
	pts := make([]float64, 0, 16)
	for i := range 8 {
		a := float64(i) * math.Pi / 4
		d := r
		if i%2 == 1 {
			d = r * 0.35
		}
		pts = append(pts, x+d*math.Cos(a), y+d*math.Sin(a))
	}
	p.Poly(c, pts...)
}
