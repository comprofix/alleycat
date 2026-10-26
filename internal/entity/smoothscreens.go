package entity

import (
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// The new look's art for the screens between the game's scenes (DESIGN
// §3.6, §3.7): the tally's cat faces, the HERE KITTY dog, Freddy's head in
// the rejection, and the dancing cats. Each is drawn in its sprite's box.

// FaceInks are a tally face's colours: fur, its shade, the eyes and the
// nose.
type FaceInks struct{ Fur, Shade, Eye, Nose int }

var (
	// LitFace is a tally face that has lit up.
	LitFace = FaceInks{resurrect64.White, resurrect64.Mist, resurrect64.Ink, resurrect64.Cerise}
	// DimFace is one still waiting.
	DimFace = FaceInks{resurrect64.Lilac, resurrect64.Dusk, resurrect64.Shadow, resurrect64.Grape}
)

// DrawTallyFace draws a cat's face, face on, in a 48×32 box, in inks c.
func DrawTallyFace(p Pen, c FaceInks) {
	p.Poly(c.Fur, 5, 13, 9, 0, 19, 7)
	p.Poly(c.Fur, 29, 7, 39, 0, 43, 13)
	p.Ellipse(25, 18, 21, 12.5, 0, c.Shade)
	p.Ellipse(23, 16.5, 20, 11.5, 0, c.Fur)
	for _, x := range []float64{14, 33} {
		p.Ellipse(x, 15, 2.6, 3.4, 0, c.Eye)
	}
	p.Poly(c.Nose, 21.5, 20, 26.5, 20, 24, 23)
	p.Stroke(c.Eye, 1.2, false, 24, 23, 21, 25.5)
	p.Stroke(c.Eye, 1.2, false, 24, 23, 27, 25.5)
}

// DrawKittyDog draws the HERE KITTY dog in a 128×272 box, in coat co: the
// bulldog stood up on one hind leg, face on and grinning, beckoning with a
// front paw, up in frame 0 and out in frame 1.
func DrawKittyDog(p Pen, frame int, co DogCoat) {
	// The hind legs: one stood on, the other bent with its foot raised.
	p.Stroke(co.Coat, 20, false, 50, 176, 50, 248)
	p.Ellipse(48, 254, 19, 10, 0, co.Shade)
	p.Stroke(co.Far, 18, false, 80, 180, 96, 214, 84, 230)
	p.Ellipse(84, 232, 14, 7, -0.3, co.Shade)
	// The body, shaded down its right side, with a patch or two.
	p.Ellipse(64, 142, 38, 52, 0, co.Far)
	p.Ellipse(59, 139, 34, 50, 0, co.Coat)
	p.Ellipse(60, 156, 14, 5, 0, co.Shade)
	p.Ellipse(60, 180, 14, 5, 0, co.Shade)
	// The hanging front leg, and the beckoning one.
	p.Stroke(co.Coat, 15, false, 30, 108, 22, 160)
	p.Ellipse(22, 166, 10, 8, 0, co.Shade)
	if frame%2 == 0 {
		p.Stroke(co.Coat, 15, false, 90, 104, 112, 92, 118, 70)
		p.Circle(118, 64, 10, co.Shade)
	} else {
		p.Stroke(co.Coat, 15, false, 90, 104, 116, 108, 122, 92)
		p.Circle(122, 86, 9, co.Shade)
	}
	// The collar and its tag.
	p.Poly(co.Collar, 34, 82, 94, 82, 92, 94, 36, 94)
	p.Stroke(co.CollarShade, 3, false, 36, 93, 92, 93)
	p.Circle(64, 98, 4, resurrect64.Honey)
	// The head: broad, with heavy jowls and folded ears.
	p.Ellipse(64, 44, 46, 36, 0, co.Coat)
	p.Ellipse(38, 60, 20, 16, 0, co.Coat)
	p.Ellipse(90, 60, 20, 16, 0, co.Coat)
	p.Ellipse(64, 72, 30, 6, 0, co.Far)
	p.Poly(co.Shade, 20, 12, 40, 3, 46, 22, 26, 32)
	p.Poly(co.Shade, 108, 12, 88, 3, 82, 22, 102, 32)
	p.Ellipse(82, 38, 9, 8, 0, co.Shade) // the patch round one eye
	for _, x := range []float64{46, 82} {
		p.Circle(x, 38, 4.5, resurrect64.Ink)
		p.Circle(x-1.5, 36.5, 1.4, resurrect64.White)
	}
	p.Ellipse(64, 50, 10, 6.5, 0, resurrect64.Ink)
	p.Ellipse(61, 48, 3, 1.5, 0, resurrect64.Dusk)
	// The grin, with its teeth.
	p.Poly(resurrect64.Ink, 40, 62, 52, 66, 76, 66, 88, 62, 80, 72, 48, 72)
	for _, x := range []float64{50, 60, 70, 80} {
		p.Poly(resurrect64.White, x-3, 65, x+3, 65, x, 70)
	}
}

// greyCat is the grey of Freddy's head in the rejection and of the dancing
// cats: fog fur, lilac shade, white light.
var greyCat = Coat{resurrect64.Fog, resurrect64.Lilac, resurrect64.White, resurrect64.Mist, resurrect64.Honey}

// DrawRejectHead draws Freddy's head as it falls in the rejection, face on
// in a 96×48 box: a light grey cat that shows on the black screen, with
// dark eyes and whiskers.
func DrawRejectHead(p Pen) {
	co := greyCat
	for _, e := range [][6]float64{{6, 14, 12, 0, 26, 8}, {54, 8, 68, 0, 74, 14}} {
		p.Poly(co.Fur, e[:]...)
	}
	p.Poly(co.Ear, 10, 11, 13, 4, 21, 8)
	p.Poly(co.Ear, 59, 8, 67, 4, 70, 11)
	p.Ellipse(41, 27, 36, 18, 0, co.Shade)
	p.Ellipse(39, 25, 34, 17, 0, co.Fur)
	p.Ellipse(24, 14, 10, 3.5, -0.3, co.Light)
	for _, x := range []float64{20, 60} {
		p.Ellipse(x, 20, 3.5, 4.5, 0, resurrect64.Ink)
		p.Circle(x-1, 18.5, 1.2, resurrect64.White)
	}
	p.Ellipse(40, 27, 3.5, 2.5, 0, resurrect64.Cerise)
	p.Stroke(resurrect64.Ink, 1.6, true, 32, 33, 36, 35, 40, 30, 44, 35, 48, 33)
	// White whiskers, the dancing cats' length and weight.
	for _, w := range [][4]float64{{14, 24, 2.4, 22.3}, {14, 29, 2.4, 30.6}, {64, 24, 75.6, 22.3}, {64, 29, 75.6, 30.6}} {
		p.Stroke(resurrect64.White, 1.3, false, w[0], w[1], w[2], w[3])
	}
}

// DrawDanceCat draws a dancing cat in a 128×140 box: a grey cat in a frilled
// skirt doing the can-can, paws in the air, kicking out its left leg in
// frame 0 and its right in frame 1.
func DrawDanceCat(p Pen, frame int) {
	co := greyCat
	// The paws held high either side.
	for _, a := range [][4]float64{{54, 56, 18, 38}, {74, 56, 110, 38}} {
		p.Stroke(co.Shade, 8, false, a[0], a[1], a[2], a[3])
		p.Stroke(co.Fur, 5.5, false, a[0], a[1], a[2], a[3])
		p.Circle(a[2], a[3], 5, co.Fur)
	}
	// The legs: the one stood on, and the one kicked out, each in a shoe.
	stand, kick := [6]float64{74, 86, 74, 106, 74, 124}, [6]float64{50, 88, 28, 96, 6, 104}
	if frame%2 == 1 {
		stand, kick = [6]float64{50, 86, 50, 106, 50, 124}, [6]float64{74, 88, 100, 96, 122, 104}
	}
	for _, l := range [][6]float64{stand, kick} {
		p.Stroke(co.Shade, 9, false, l[:]...)
		p.Stroke(co.Fur, 6, false, l[:]...)
		p.Ellipse(l[4], l[5]+3, 7, 4, 0, resurrect64.Ink)
	}
	// The body, then the skirt: a violet band, frills and a plum hem.
	p.Ellipse(64, 60, 20, 15, 0, co.Shade)
	p.Ellipse(62, 58, 18, 14, 0, co.Fur)
	p.Poly(resurrect64.Grape, 36, 72, 98, 72, 114, 84, 20, 84)
	for x := 30.0; x <= 104; x += 12 {
		p.Circle(x, 78, 3, resurrect64.White)
	}
	p.Stroke(resurrect64.Violet, 5, false, 38, 70, 96, 70)
	p.Stroke(resurrect64.Plum, 4, false, 20, 85, 114, 85)
	for x := 26.0; x <= 110; x += 12 {
		p.Circle(x, 87, 2.5, resurrect64.White)
	}
	// The face, with a nose, mouth and whiskers as the rejection head has;
	// the whiskers are white, as ink would be lost on the black screen.
	f := p.At(64, 20, 2.1)
	drawFace(f, 0, 0, co)
	f.Ellipse(0, 5.4, 1.8, 1.3, 0, resurrect64.Cerise)
	f.Stroke(resurrect64.Ink, 0.7, true, -2.6, 7.6, -1.3, 8.4, 0, 6.8, 1.3, 8.4, 2.6, 7.6)
	for _, w := range [][4]float64{{-5, 5, -10.5, 4}, {-5, 7, -10.5, 7.75}, {5, 5, 10.5, 4}, {5, 7, 10.5, 7.75}} {
		f.Stroke(resurrect64.White, 0.6, false, w[0], w[1], w[2], w[3])
	}
}
