package rooms

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// buildDogs draws the sleeping dogs room's backdrop, as the classic look
// lays it out (DESIGN §3.4, "Sleeping dogs room layout"): the taller
// walls, the window with its curtains, the chairs, the picture and the
// lamp.
func (r *SmoothRenderer) buildDogs() {
	r.buildWalls(dogsWallY0, dogsWallY1)
	r.buildWindow(dogsOpeningX, dogsOpeningY)
	r.buildChair(dogsChairLX, dogsChairY, false)
	r.buildChair(dogsChairRX, dogsChairY, true)
	r.buildPicture(k(dogsPictureX), k(dogsPictureY))
	r.buildLamp(dogsLampX, dogsLampY)
}

// dogsRows are the sleeping dogs room's floor rows, which Freddy walks.
var dogsRows = [...]int{136, 152, 168}

// dogsFloor is what stands on the sleeping dogs room's floor, for drawing
// in depth order with Freddy: the bowls, a little in front of their row,
// and the sleeping dogs, on theirs. An empty bowl is gone, and the bowl
// Freddy eats from is drawn with him.
func (r *SmoothRenderer) dogsFloor(dst *ebiten.Image, d *DogsRoom) []depthed {
	var floor []depthed
	for i, b := range config.FoodBowls {
		if !eatenBowl(d, i) {
			floor = append(floor, depthed{b.Y + config.FoodBowlDY, func() { r.drawBowl(dst, d, i) }})
		}
	}
	for i, s := range d.Sleepers {
		if sl := config.Sleepers[i]; s.On {
			sr := SleeperRect(i)
			floor = append(floor, depthed{sl.Y, func() {
				r.canvas.Draw(dst, entity.Picture{X: sr.X, Y: sr.Y, W: sr.W, H: sr.H, Flip: sl.Left, Outline: r.scheme.Outline},
					func(p entity.Pen) { drawSleeper(p, s.Alert, entity.WhiteDog) })
			}})
		}
	}
	return floor
}

// eatPoseH is the height of the eating pose's box.
const eatPoseH = 10

// drawBowl draws bowl i where it stands, unless it is empty.
func (r *SmoothRenderer) drawBowl(dst *ebiten.Image, d *DogsRoom, i int) {
	b := config.FoodBowls[i]
	if level := d.MilkLevel(i); level > 0 {
		r.canvas.Draw(dst, entity.Picture{X: b.X, Y: b.Y + config.FoodBowlDY, W: config.FoodBowlW, H: config.FoodBowlH, Outline: r.scheme.Outline},
			func(p entity.Pen) { drawBowlArt(p, level) })
	}
}

// eatenBowl reports whether Freddy is eating from bowl i.
func eatenBowl(d *DogsRoom, i int) bool { return d.Eating && i == d.EatBowl }

// drawBowlArt draws a dog's bowl in a 64×32 box, holding milk to level 1–4
// (full): a teal dish on a dark foot, the milk filling it.
func drawBowlArt(p entity.Pen, level int) {
	p.Ellipse(32, 26, 18, 3.5, 0, resurrect64.Pine)
	p.Poly(resurrect64.Seafoam, 4, 12, 60, 12, 50, 26, 14, 26)
	p.Poly(resurrect64.Jade, 44, 12, 60, 12, 50, 26, 42, 26)
	p.Ellipse(32, 11, 29, 6, 0, resurrect64.Seafoam)
	p.Ellipse(32, 11, 25, 4.5, 0, resurrect64.DeepTeal)
	f := float64(level) / config.MilkLevels
	p.Ellipse(32, 11+(1-f)*2, 25*f, 4.5*f, 0, resurrect64.White)
	p.Ellipse(32+25*f*0.55, 11+(1-f)*2, 25*f*0.35, 3*f, 0, resurrect64.Mist)
}

// drawSleeper draws a bulldog lying asleep, facing right, in a 160×52 box,
// in coat co: its chin on its paws, its tail out behind, and its eye, with
// lashes and a brow, shut, half open under a heavy lid or wide open with
// alert 0, 1 or 2.
func drawSleeper(p entity.Pen, alert int, co entity.DogCoat) {
	p.Stroke(co.Coat, 5, true, 30, 34, 16, 32, 6, 34)
	p.Ellipse(70, 31, 46, 15, -0.04, co.Coat)
	p.Ellipse(34, 34, 12, 11, 0, co.Shade)
	p.Ellipse(38, 40, 16, 7, 0, co.Coat)
	p.Ellipse(72, 43, 42, 4, 0, co.Far)
	p.Stroke(co.Coat, 8, false, 102, 44, 140, 44)
	p.Ellipse(142, 44, 6, 3.5, 0, co.Coat)
	p.Stroke(co.Collar, 6, false, 106, 16, 105, 38)
	p.Stroke(co.CollarShade, 2, false, 108, 17, 107, 37)
	p.Circle(106, 40, 2.5, resurrect64.Honey)
	p.Circle(128, 24, 15, co.Coat)
	p.Ellipse(138, 32, 14, 8, 0.05, co.Coat)
	p.Ellipse(140, 36, 9, 3, 0, co.Far)
	p.Ellipse(152, 25, 4.5, 3.5, 0.2, resurrect64.Ink)
	p.Stroke(resurrect64.Ink, 1.8, true, 142, 34, 148, 35, 154, 32)
	p.Poly(resurrect64.Ink, 112, 14, 116, 8, 126, 10, 122, 18)
	p.Ellipse(132, 19, 7, 5.5, 0, co.Shade)
	ink := resurrect64.Ink
	switch alert {
	case 0:
		// Shut: a lid curved down in sleep, its lashes hanging from it, and
		// a brow at rest.
		p.Stroke(ink, 2.2, true, 125.5, 14.8, 131, 14, 137.5, 15.2)
		p.Stroke(ink, 1.8, true, 127, 19, 132, 21.5, 137, 19)
		for _, l := range [][4]float64{{128.5, 20.3, 127.4, 22.6}, {131.5, 21.4, 131.2, 23.8}, {134.5, 20.8, 135.2, 23.1}} {
			p.Stroke(ink, 1.1, false, l[0], l[1], l[2], l[3])
		}
		return
	}
	// Open, or half open: the white of the eye, a honey iris and its pupil
	// looking ahead, a glint, and a heavy lid down over the top half.
	p.Ellipse(132, 19, 5, 3.8, 0, resurrect64.White)
	p.Circle(133.5, 19.5, 2.8, resurrect64.Honey)
	p.Circle(134, 19.5, 1.5, ink)
	p.Circle(133, 18.4, 0.7, resurrect64.White)
	if alert == 1 {
		// Half open: the lid low and the brow down, wary.
		lid := []float64{}
		for i := range 13 {
			t := math.Pi + math.Pi*float64(i)/12
			lid = append(lid, 132+5.6*math.Cos(t), 19.6+4.4*math.Sin(t))
		}
		p.Poly(co.Shade, lid...)
		p.Stroke(ink, 1.6, false, 126.6, 19.6, 137.4, 19.6)
		for _, l := range [][4]float64{{127.4, 19.6, 125.8, 18.4}, {128.8, 19.6, 127.6, 18}} {
			p.Stroke(ink, 1.1, false, l[0], l[1], l[2], l[3])
		}
		p.Stroke(ink, 2.4, true, 125.5, 15, 131, 15.2, 137.5, 16.4)
		return
	}
	// Wide open: the upper lid lined, lashes flicked up at the outer corner,
	// and the brow raised high in alarm.
	p.Stroke(ink, 1.6, true, 127, 18, 132, 15.2, 137, 17.2)
	for _, l := range [][4]float64{{127.6, 17.3, 125.6, 15.6}, {129.2, 16.1, 127.8, 14}} {
		p.Stroke(ink, 1.1, false, l[0], l[1], l[2], l[3])
	}
	p.Stroke(ink, 2.4, true, 125.5, 13.6, 131, 12.2, 137.5, 13)
}

// zPeriod is how many master ticks a Z takes to rise and fade, about three
// seconds; a dog sound asleep has three on the way up at once.
const zPeriod = 54

// drawSnores draws the Z's rising from each dog sound asleep, as cartoons
// show sleep: small at its head, growing as each drifts up and back and
// fading away, one after another. They are presentation only, timed by the
// room's ticks, so they never touch the game.
func (r *SmoothRenderer) drawSnores(dst *ebiten.Image, d *DogsRoom) {
	for i, s := range d.Sleepers {
		if !s.On || s.Alert != 0 {
			continue
		}
		sl, sr := config.Sleepers[i], SleeperRect(i)
		for z := range 3 {
			t := float64((d.Ticks+i*17+z*zPeriod/3)%zPeriod) / zPeriod
			bx := 134 - 18*t + 4*math.Sin(t*2*math.Pi*1.5) // in the box, facing right
			if sl.Left {
				bx = float64(sr.W*entity.K) - bx
			}
			x, y := k(sr.X)+bx, k(sr.Y)-2-40*t
			a := min(t*6, 1, (1-t)*3)
			drawZ(dst, x, y, 6+8*t, a)
		}
	}
}

// drawZ draws a Z of height h centred on (x, y) at opacity a, in blue with
// an ink rim so it reads on the wallpaper.
func drawZ(dst *ebiten.Image, x, y, h, a float64) {
	w := h * 0.8
	var p vector.Path
	p.MoveTo(float32(x-w/2), float32(y-h/2))
	p.LineTo(float32(x+w/2), float32(y-h/2))
	p.LineTo(float32(x-w/2), float32(y+h/2))
	p.LineTo(float32(x+w/2), float32(y+h/2))
	for _, l := range []struct {
		c int
		w float64
	}{{resurrect64.Ink, h/5 + 3}, {resurrect64.Blue, h / 5}} {
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(gfx.RGBA(l.c))
		op.ColorScale.Scale(float32(a), float32(a), float32(a), float32(a))
		vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: float32(l.w), LineJoin: vector.LineJoinMiter, LineCap: vector.LineCapSquare}, op)
	}
}

// drawEating draws Freddy's eating pose in place of Freddy, to one side of
// him as the bowl dictates, as the classic look does: crouched over the
// bowl, lapping, his tongue out on the second frame.
func (r *SmoothRenderer) drawEating(dst *ebiten.Image, d *DogsRoom) {
	f, b := d.Freddy, config.FoodBowls[d.EatBowl]
	x := f.X &^ 3
	if b.Left {
		x = max(x-config.EatPoseDX, 0)
	} else {
		x = min(x+config.EatPoseDX, config.EatPoseMaxX)
	}
	// Where Freddy stands by the bowl moves the pose against it, so his head
	// reaches for the middle of the milk, as near as his neck allows.
	milk := k(b.X) + k(config.FoodBowlW)/2 - k(x)
	if !b.Left {
		milk = k(24) - milk
	}
	reach := min(max(lapX, milk-lapSlack), milk+lapSlack) - lapX
	lap := d.EatAcc >= eatFrameFull
	// He stands behind the bowl with his head over it: his body, then the
	// bowl, then his head.
	pic := entity.Picture{X: x, Y: f.Y + config.EatPoseDY, W: 24, H: eatPoseH, Flip: !b.Left, Outline: entity.NoOutline}
	r.canvas.Draw(dst, pic, drawEatingBody)
	r.drawBowl(dst, d, d.EatBowl)
	r.canvas.Draw(dst, pic, func(p entity.Pen) { drawEatingHead(p.At(reach, 0, 1), lap) })
}

// lapX is where Freddy's tongue meets the milk in the eating pose's box,
// with his head stretched out, and lapSlack how far from the middle of the
// milk it may land, so it is in the milk however low the milk is.
const lapX, lapSlack = 89, 6

// drawEatingBody draws Freddy crouched to eat, facing right, in a 96×40
// box: haunches down, tail up behind, front legs braced.
func drawEatingBody(p entity.Pen) {
	ink := resurrect64.Ink
	p.Stroke(ink, 5, true, 18, 22, 8, 14, 6, 6, 10, 2)
	p.Ellipse(44, 22, 30, 8, 0.08, ink)
	p.Ellipse(24, 26, 13, 11, 0, ink)
	p.Stroke(ink, 6, false, 20, 30, 16, 38)
	p.Stroke(ink, 6, false, 58, 26, 62, 38)
	p.Stroke(ink, 6, false, 66, 26, 72, 38)
}

// drawEatingHead draws Freddy's head lapping, in the eating pose's box
// stretched right out: low over the bowl with his chin above the milk's
// near edge, his tongue dipping into the milk, further on the second
// frame.
func drawEatingHead(h entity.Pen, lap bool) {
	ink := resurrect64.Ink
	h.Poly(ink, 70, 16, 72, 4, 80, 12)
	h.Poly(ink, 84, 16, 88, 6, 92, 16)
	h.Poly(resurrect64.Mauve, 73, 13, 74, 8, 78, 12)
	h.Ellipse(82, 22, 11, 8, 0.3, ink)
	h.Ellipse(84, 20, 2.2, 1.6, 0.3, resurrect64.Honey)
	tongue := 2.0
	if lap {
		tongue = 5
	}
	h.Ellipse(lapX, 28+tongue/2, 2.5, tongue/2+1, 0, resurrect64.Pink)
}
