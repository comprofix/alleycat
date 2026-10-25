package rooms

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// buildFishbowl draws the fishbowl room's backdrop, as the classic look
// lays it out (DESIGN §3.4, "Fishbowl room layout"): the walls, the window
// with its curtains, the chairs, the table, the bowl's pedestal table and
// the lamp, each casting a soft shadow on the floor. The bowl itself is
// drawn each frame.
func (r *SmoothRenderer) buildFishbowl() {
	r.buildWalls(wallY0, wallY1)
	r.buildWindow(fishOpeningX, openingY)
	r.buildChair(fishChairLX, furnitureY, false)
	r.buildTable(fishTableX, tableY)
	r.buildChair(fishChairMX, furnitureY, true)
	fadePoly(r.back, r.scheme.Outline, 0.3, oval(k(bowlTableX)+76, k(pedestalY+34)-2, 48, 7)...)
	r.buildPedestal(bowlTableX, pedestalY)
	r.buildLamp(fishLampX, lampY)
}

// drawFishbowl draws the bowl on its pedestal table, its goldfish swimming
// round.
func (r *SmoothRenderer) drawFishbowl(dst *ebiten.Image, fb *FishbowlRoom) {
	r.canvas.Draw(dst, entity.Picture{X: bowlDrawX, Y: bowlDrawY, W: 20, H: 12, Outline: r.scheme.Outline}, func(p entity.Pen) {
		drawGlassBowl(p, fb.BowlFrame)
	})
}

// bowlFish is where the goldfish is in the bowl in each of its four
// frames, and whether it faces right: round the bowl and back.
var bowlFish = [config.BowlFrames]struct {
	x, y  float64
	right bool
}{{24, 34, true}, {40, 30, true}, {56, 33, false}, {40, 38, false}}

// drawGlassBowl draws the fishbowl in an 80×48 box: a round glass bowl with
// a lip round its opening, the water filling it two thirds up, darker at
// the bottom over a little gravel, its surface lit, and the goldfish in
// frame's place, with glints on the glass.
func drawGlassBowl(p entity.Pen, frame int) {
	const cx, cy, rx, ry, water, deep = 40, 26, 36, 21, 14, 31
	// arc is the bowl's side from where it is at height y on the left,
	// round the bottom, to the same height on the right.
	arc := func(y, in float64) []float64 {
		t0 := math.Asin((y - cy) / (ry - in))
		pts := []float64{}
		for i := range 25 {
			t := math.Pi - t0 - (math.Pi-2*t0)*float64(i)/24
			pts = append(pts, cx+(rx-in)*math.Cos(t), cy+(ry-in)*math.Sin(t))
		}
		return pts
	}
	p.Poly(resurrect64.Blue, arc(water, 0)...)
	p.Poly(resurrect64.Cobalt, arc(deep, 0)...)
	for i, x := range []float64{22, 29, 36, 44, 51, 58} {
		c := resurrect64.Honey
		if i%2 == 1 {
			c = resurrect64.Coral
		}
		p.Ellipse(x, 43-float64(i%3), 3, 2, 0, c)
	}
	surface := math.Sqrt(1-math.Pow(float64(water-cy)/ry, 2)) * rx
	p.Ellipse(cx, water, surface, 3, 0, resurrect64.Sky)
	p.Ellipse(cx, water+1, surface-3, 1.6, 0, resurrect64.Blue)

	f := bowlFish[frame%len(bowlFish)]
	s := 1.0
	if !f.right {
		s = -1
	}
	p.Poly(resurrect64.Vermilion, f.x-s*5, f.y, f.x-s*11, f.y-4, f.x-s*10, f.y, f.x-s*11, f.y+4)
	p.Ellipse(f.x, f.y, 6.5, 4, 0, resurrect64.Orange)
	p.Ellipse(f.x+s, f.y+1.5, 4, 1.6, 0, resurrect64.Amber)
	p.Circle(f.x+s*3.5, f.y-1, 1.1, resurrect64.Ink)

	// The glass: its side up to the lip, the lip round the opening, and
	// glints on the upper left.
	p.Stroke(resurrect64.Seafoam, 2.5, false, arc(7.5, 0)...)
	p.Ellipse(cx, 6, 17, 3.2, 0, resurrect64.Seafoam)
	p.Ellipse(cx, 6, 14, 1.8, 0, resurrect64.Jade)
	p.Stroke(resurrect64.White, 2.5, true, 12, 22, 14, 15, 19, 11)
	p.Circle(15.5, 27, 1.5, resurrect64.White)
	p.Stroke(resurrect64.White, 1.5, true, 64, 34, 66, 30)
}

// drawUnderwater draws the underwater scene in the new look (DESIGN §3.4,
// "Underwater layout"): the water, the weed swaying in it, the surface's
// waves, the fish and eels, Freddy swimming, his bubbles and "ZAP!". Freddy
// and the eels take the air warning's colours, as the classic look does,
// flashing as an eel zaps him.
func (r *SmoothRenderer) drawUnderwater(dst *ebiten.Image, fb *FishbowlRoom) {
	if r.water == nil {
		r.water = ebiten.NewImage(gfx.ScreenWidth*entity.K, gfx.ScreenHeight*entity.K)
		r.buildWater()
	}
	dst.DrawImage(r.water, nil)
	drawWeed(dst, fb.Ticks)
	drawSurface(dst, fb)

	ink, lit := gfx.AirColours[fb.AirStage()], gfx.AirLights[fb.AirStage()]
	if fb.Phase == Zapped {
		ink, lit = resurrect64.White, resurrect64.White
		if fb.Zap%2 == 0 {
			ink, lit = resurrect64.Cobalt, resurrect64.Blue
		}
	}
	for i := range fb.Swimmers {
		sw := &fb.Swimmers[i]
		if sw.Off {
			continue
		}
		if IsEel(i) {
			frame := (fb.EelFrame + i) % 4
			r.canvas.Draw(dst, entity.Picture{X: sw.X, Y: sw.Y, W: config.EelW, H: config.EelH, Outline: resurrect64.Ink},
				func(p entity.Pen) { drawEel(p, frame, ink, lit) })
			continue
		}
		frame := fb.FishFrame ^ (i & 1)
		r.canvas.Draw(dst, entity.Picture{X: sw.X, Y: sw.Y, W: config.FishW, H: config.FishH, Flip: sw.HDir < 0, Outline: resurrect64.Ink},
			func(p entity.Pen) { drawFish(p, frame) })
	}
	for _, b := range fb.Bubbles {
		r.canvas.Draw(dst, entity.Picture{X: b.X + config.FreddyW/2, Y: b.Y, W: 4, H: 4, Outline: entity.NoOutline}, drawBubble)
	}
	co := entity.Coat{Fur: ink, Shade: lit, Light: lit, Ear: resurrect64.Mauve, Eye: resurrect64.Honey}
	r.freddy.DrawSwimming(dst, fb.Freddy, co)
	if fb.Phase == Zapped {
		r.canvas.Draw(dst, entity.Picture{X: fb.ZapX, Y: fb.ZapY, W: 40, H: 18, Outline: entity.NoOutline}, func(p entity.Pen) {
			entity.DrawSplat(p, r.text, "ZAP!", 40, 18)
		})
	}
}

// sandY is the top of the sandy bottom on the grid, below Freddy's deepest
// dive.
const sandY = 194

// weedX are where the weed grows from the sand, on the grid.
var weedX = [...]int{18, 74, 150, 233, 297}

// buildWater draws the underwater backdrop: the water, lit near the surface
// and darkening with depth, faint rays of light slanting down, the classic
// look's fixed scatter of ripple marks, and the sandy bottom with its
// pebbles.
func (r *SmoothRenderer) buildWater() {
	s, b := gfx.UnderwaterScheme, r.water
	w, h := float64(b.Bounds().Dx()), float64(b.Bounds().Dy())
	b.Fill(gfx.Palette[s.Water])
	for i := range 8 {
		fadeRect(b, 0, float64(i)*30, w, 30, resurrect64.Teal, 0.32-0.04*float64(i))
	}
	for i := range 12 {
		fadeRect(b, 0, 420+float64(i)*30, w, 30, resurrect64.Navy, 0.05*float64(i+1))
	}
	// The rays, slanting as the classic look's, widening and fading as
	// they go down.
	const raySteps = 10
	for _, rx := range []int{40, 112, 196, 268} {
		x := k(rx)
		for i := range raySteps {
			y0, y1 := 24+float64(i)*40, 24+float64(i+1)*40
			w0, w1 := 10+float64(i)*4, 14+float64(i)*4
			fadePoly(b, resurrect64.Seafoam, 0.1*float64(raySteps-i)/raySteps,
				x+y0/3, y0, x+y0/3+w0, y0, x+y1/3+w1, y1, x+y1/3, y1)
		}
	}
	for y := 16; y < gfx.ScreenHeight-4; y += rippleCell {
		for x := 4; x < gfx.ScreenWidth-8; x += rippleCell {
			if hh := (x*5 + y*11) % 13; hh < 2 {
				rx, ry := k(x+hh*3), k(y+hh)+2
				shapes.StrokePath(b, polyline([]float64{rx, ry, rx + 6, ry - 3, rx + 12, ry, rx + 18, ry - 3, rx + 24, ry}), 2, s.WallShade)
			}
		}
	}

	// The sand, its top rippled, darker below, with pebbles.
	top := k(sandY)
	sand := []float64{0, h}
	for x := 0.0; x <= w; x += 16 {
		sand = append(sand, x, top+2*math.Sin(x/37)+1.5*math.Sin(x/11))
	}
	sand = append(sand, w, h)
	shapes.FillPoly(b, resurrect64.Honey, sand...)
	fadeRect(b, 0, top+10, w, h-top-10, resurrect64.Caramel, 0.6)
	for x := 3; x < gfx.ScreenWidth; x += 11 {
		c := resurrect64.Coral
		if x%3 == 0 {
			c = resurrect64.Lilac
		}
		shapes.FillPoly(b, c, oval(k(x)+2, top+8+float64(x%4)*3, 4, 2.5)...)
	}
}

// drawWeed draws the weed growing from the sand, swaying with the water
// over ticks: a stem with leaves either side, the stretches leaning back in
// shade.
func drawWeed(dst *ebiten.Image, ticks int) {
	for _, wx := range weedX {
		h := k(14 + wx%9)
		x0, y0 := k(wx)+2, k(sandY)+4
		sway := 7 * math.Sin(float64(ticks)*0.09+float64(wx))
		at := func(t float64) (float64, float64) {
			return x0 + 3*math.Sin(t*5) + sway*t*t, y0 - h*t
		}
		pts := []float64{}
		for i := range 13 {
			x, y := at(float64(i) / 12)
			pts = append(pts, x, y)
		}
		shapes.StrokePath(dst, polyline(pts), 4, resurrect64.Forest)
		for i := 1; i < 12; i++ {
			t := float64(i) / 12
			x, y := at(t)
			side := float64(1 - 2*(i%2))
			l := 20 - 8*t // the leaves shorten towards the top
			c := resurrect64.Emerald
			if i%3 == 0 {
				c = resurrect64.Forest
			}
			lean := sway * t / 3
			shapes.FillPoly(dst, c, x, y, x+side*l*0.4+lean, y-l*0.5, x+side*l+lean, y-l*0.6, x+side*l*0.5+lean, y-l*0.15)
		}
	}
}

// drawSurface draws the surface of the water along the top: the band of
// the room above, its waves rising and falling with the classic look's
// wave columns, lit along their edge.
func drawSurface(dst *ebiten.Image, fb *FishbowlRoom) {
	pts := []float64{}
	for c, f := range fb.Surface {
		for half := range 2 {
			d := surfaceDepth + (c*3+f+half*2)%3
			pts = append(pts, k(c*8+half*4)+8, k(d))
		}
	}
	var line, band vector.Path
	band.MoveTo(0, 0)
	band.LineTo(0, float32(pts[1]))
	line.MoveTo(0, float32(pts[1]))
	for i := 0; i+3 < len(pts); i += 2 {
		mx, my := float32(pts[i]+pts[i+2])/2, float32(pts[i+1]+pts[i+3])/2
		band.QuadTo(float32(pts[i]), float32(pts[i+1]), mx, my)
		line.QuadTo(float32(pts[i]), float32(pts[i+1]), mx, my)
	}
	w, last := float32(dst.Bounds().Dx()), float32(pts[len(pts)-1])
	band.LineTo(w, last)
	band.LineTo(w, 0)
	band.Close()
	line.LineTo(w, last)
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(gfx.RGBA(gfx.UnderwaterScheme.Trim))
	vector.FillPath(dst, &band, nil, op)
	shapes.StrokePath(dst, &line, 3, resurrect64.Seafoam)
}

// drawFish draws a goldfish facing right in a 32×24 box, its tail swishing
// with frame 0 or 1.
func drawFish(p entity.Pen, frame int) {
	if frame == 0 {
		p.Poly(resurrect64.Vermilion, 9, 12, 1, 4, 4, 12, 1, 20)
	} else {
		p.Poly(resurrect64.Vermilion, 9, 12, 3, 7, 5, 12, 3, 17)
	}
	p.Poly(resurrect64.Vermilion, 12, 7, 17, 2, 21, 7)
	p.Poly(resurrect64.Vermilion, 15, 16, 18, 21, 21, 16)
	p.Ellipse(17, 12, 10, 6.5, 0, resurrect64.Orange)
	p.Ellipse(18, 15, 7, 2.6, 0, resurrect64.Amber)
	p.Stroke(resurrect64.Vermilion, 1.2, true, 21.5, 8, 20.5, 12, 21.5, 16)
	p.Circle(24.5, 10, 2.2, resurrect64.White)
	p.Circle(25, 10, 1.2, resurrect64.Ink)
}

// drawEel draws an eel in a 64×8 box, its head to the right and its body
// wriggling with frame 0–3, in ink with its crests lit.
func drawEel(p entity.Pen, frame, ink, lit int) {
	body, crest := []float64{}, []float64{}
	for i := range 15 {
		x := float64(i) * 3.6
		a := 2.6 * (1 - x/60)
		y := 4 + a*math.Sin(x/8-float64(frame)*math.Pi/2)
		body = append(body, x, y)
		crest = append(crest, x, y-1.6)
	}
	p.Stroke(ink, 5, true, body...)
	p.Stroke(lit, 1.4, true, crest[:24]...)
	p.Ellipse(55, 4, 8, 3.8, 0, ink)
	p.Circle(59, 3, 1.2, lit)
}

// drawBubble draws a bubble in a 16×16 box: a thin pale ring with a glint.
func drawBubble(p entity.Pen) {
	ring := []float64{}
	for i := range 17 {
		t := 2 * math.Pi * float64(i) / 16
		ring = append(ring, 8+6*math.Cos(t), 8+6*math.Sin(t))
	}
	p.Stroke(resurrect64.Mist, 1.6, false, ring...)
	p.Circle(5.5, 5.5, 1.4, resurrect64.White)
}
