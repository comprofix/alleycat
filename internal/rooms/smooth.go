package rooms

import (
	"cmp"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// k is a position on the 320×200 grid in the new look's screen pixels.
func k(v int) float64 { return float64(v * entity.K) }

// SmoothRenderer draws the rooms in the new look (DESIGN §2.1), in code
// from shapes, on the classic look's layout and in its depth order. Only
// the rooms Drawn reports are done; the rest are left to the pixel
// Renderer, enlarged. It caches the backdrop and holds no game state.
type SmoothRenderer struct {
	freddy entity.SmoothFreddy
	dog    entity.SmoothDog
	canvas entity.Canvas   // for the room's things, drawn one at a time
	text   *gfx.VectorFont // "BYTE!"
	scheme gfx.SceneScheme
	fur    furnish
	back   *ebiten.Image
}

// furnish is a room's furniture colours, as its sprites are recoloured:
// the body, its lit edge and its shade, the shade under the trim, the
// curtain rod and its ends, and the shine on a table top.
type furnish struct {
	body, light, shade, trimShade, rod, rodEnd, shine int
}

var (
	greenFurnish = furnish{resurrect64.Emerald, resurrect64.Leaf, resurrect64.Forest, resurrect64.Caramel, resurrect64.Copper, resurrect64.Coral, resurrect64.Lime}
	tealFurnish  = furnish{resurrect64.Seafoam, resurrect64.White, resurrect64.Jade, resurrect64.Jade, resurrect64.Indigo, resurrect64.Cobalt, resurrect64.White}
)

// NewSmoothRenderer returns the new look's room renderer.
func NewSmoothRenderer() *SmoothRenderer {
	return &SmoothRenderer{text: gfx.NewVectorFont(36), scheme: gfx.BirdcageScheme, fur: greenFurnish}
}

// Drawn reports whether room has its new look yet.
func (r *SmoothRenderer) Drawn(room Room) bool {
	switch room.(type) {
	case *BirdcageRoom, *BookshelfRoom, *CheeseRoom, *DogsRoom:
		return true
	}
	return false
}

// Draw renders the room, as Renderer.Draw does, onto a 1280×800 dst.
func (r *SmoothRenderer) Draw(dst *ebiten.Image, room Room) {
	if r.back == nil {
		r.back = ebiten.NewImage(gfx.ScreenWidth*entity.K, gfx.ScreenHeight*entity.K)
		switch rm := room.(type) {
		case *BookshelfRoom:
			r.scheme, r.fur = gfx.BookshelfScheme, tealFurnish
			r.buildBookshelf(rm)
		case *CheeseRoom:
			r.scheme = gfx.CheeseScheme
			r.buildCheese()
		case *DogsRoom:
			r.scheme, r.fur = gfx.DogsScheme, tealFurnish
			r.buildDogs()
		default:
			r.buildBirdcage()
		}
	}
	dst.DrawImage(r.back, nil)
	b := room.Shared()
	for col, n := range b.Broom.Marks {
		r.drawPrints(dst, col*8, config.FootprintY, n)
	}
	var floor []depthed
	switch rm := room.(type) {
	case *BirdcageRoom:
		r.drawBirdcage(dst, rm)
	case *BookshelfRoom:
		r.drawBookshelf(dst, rm)
	case *CheeseRoom:
		r.drawCheese(dst, rm)
	case *DogsRoom:
		floor = r.dogsFloor(dst, rm)
	}

	// The room dog, the broom and Freddy: on the floor, each in depth order
	// among the room's floor things by the line it stands on, the nearest
	// last, so things pass behind and in front of each other as the floor's
	// perspective says; on one line, in the classic look's order, Freddy
	// last, so he stays in view. Off the floor, up on the furniture or in
	// the air, they are drawn over it in that order.
	var up []func()
	put := func(onFloor bool, line int, draw func()) {
		if onFloor {
			floor = append(floor, depthed{line, draw})
		} else {
			up = append(up, draw)
		}
	}
	d, br, f := &b.Dog, &b.Broom, b.Freddy
	put(d.Capture == 0, config.FloorY, func() { r.dog.Draw(dst, d, entity.OrangeDog) })
	put(br.Y == config.BroomFloorY, config.FloorY, func() { r.drawBroom(dst, br) })
	if dr, ok := room.(*DogsRoom); ok && dr.Eating {
		put(true, f.Y, func() { r.drawEating(dst, dr) })
	} else {
		put(onFloor(room, f), f.Y, func() { r.freddy.Draw(dst, f) })
	}
	slices.SortStableFunc(floor, func(a, b depthed) int { return cmp.Compare(a.line, b.line) })
	for _, t := range floor {
		t.draw()
	}
	if dr, ok := room.(*DogsRoom); ok {
		r.drawSnores(dst, dr)
	}
	for _, draw := range up {
		draw()
	}
	if bs, ok := room.(*BookshelfRoom); ok {
		r.drawBite(dst, bs)
	}
}

// depthed is something drawn in depth order on the floor: the line it
// stands on, as the Y Freddy stands at there on the grid, and how to draw
// it.
type depthed struct {
	line int
	draw func()
}

// onFloor reports whether Freddy is standing on the room's floor, not up
// on the furniture or in the air: at the floor, or on one of the sleeping
// dogs room's floor rows.
func onFloor(room Room, f *entity.Freddy) bool {
	if f.Airborne() || f.Seat != entity.NotSeated {
		return false
	}
	if f.Y >= config.FloorY {
		return true
	}
	if _, ok := room.(*DogsRoom); ok {
		return slices.Contains(dogsRows[:], f.Y)
	}
	return false
}

// fadePoly fills the polygon through the x, y pairs with palette colour c
// at opacity a (0–1), for soft shading.
func fadePoly(dst *ebiten.Image, c int, a float64, pts ...float64) {
	var p vector.Path
	p.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i+1 < len(pts); i += 2 {
		p.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	p.Close()
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(gfx.RGBA(c))
	op.ColorScale.Scale(float32(a), float32(a), float32(a), float32(a))
	vector.FillPath(dst, &p, nil, op)
}

// polyline is the line through the x, y pairs, for shapes.StrokePath.
func polyline(pts []float64) *vector.Path {
	var p vector.Path
	p.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i+1 < len(pts); i += 2 {
		p.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	return &p
}

// fadeRect fills a rectangle with palette colour c at opacity a.
func fadeRect(dst *ebiten.Image, x, y, w, h float64, c int, a float64) {
	fadePoly(dst, c, a, x, y, x+w, y, x+w, y+h, x, y+h)
}

// oval is the points round the ellipse with radii rx and ry about
// (cx, cy), as x, y pairs.
func oval(cx, cy, rx, ry float64) []float64 {
	const n = 32
	pts := make([]float64, 0, 2*n)
	for i := range n {
		t := 2 * math.Pi * float64(i) / n
		pts = append(pts, cx+rx*math.Cos(t), cy+ry*math.Sin(t))
	}
	return pts
}

// buildBirdcage draws the birdcage room's backdrop, as the classic look
// lays it out (DESIGN §3.4, "Birdcage room layout"): the walls in
// perspective, the picture, the window with its curtains, the lamp, the
// chairs and the table, each casting a soft shadow on the floor.
func (r *SmoothRenderer) buildBirdcage() {
	r.buildWalls(wallY0, wallY1)
	r.buildPicture(k(pictureX), k(pictureY))
	r.buildWindow(openingX, openingY)
	r.buildLamp(lampX, lampY)
	for _, c := range chairs {
		r.buildChair(c.x, furnitureY, c.left)
	}
	r.buildTable(tableX, tableY)
}

// buildTable draws the round table at (x, y) on the grid, in the room's
// furniture colours, casting a soft shadow on the floor.
func (r *SmoothRenderer) buildTable(x, y int) {
	b, ink := r.back, r.scheme.Outline
	fadePoly(b, ink, 0.3, oval(k(x)+112, k(y+36)-4, 112, 10)...)
	r.canvas.Draw(b, entity.Picture{X: x, Y: y, W: 56, H: 36, Outline: ink}, func(p entity.Pen) { drawTable(p, r.fur) })
}

// buildWalls draws the room in perspective with the back wall from y0 to
// y1, as the classic look does: a darker ceiling, striped wallpaper on the
// back wall with a picture rail and skirting, shaded side walls, and
// floorboards that widen towards us in the room's Ground colour, so the
// footprints still show (DESIGN §3.4).
func (r *SmoothRenderer) buildWalls(y0, y1 int) {
	s, b := r.scheme, r.back
	w, h := float64(b.Bounds().Dx()), float64(b.Bounds().Dy())
	x0, x1, top, bot := k(wallX0), k(wallX1), k(y0), k(y1)
	d := k(16) // how far the ceiling and floor slope to the screen's edge
	b.Fill(gfx.Palette[s.Wall])

	// The ceiling, darkening towards us, and its shadow on the walls.
	shapes.FillPoly(b, s.WallShade, 0, 0, w, 0, w, top-d, x1, top, x0, top, 0, top-d)
	for i := range 6 {
		fadeRect(b, 0, float64(i)*(top-d)/6, w, (top-d)/6, s.Outline, 0.3-0.05*float64(i))
	}
	fadeRect(b, x0, top, x1-x0, 10, s.Outline, 0.35)
	fadeRect(b, x0, top+10, x1-x0, 8, s.Outline, 0.15)

	// The wallpaper: a stripe every 8 of the grid's pixels, as the classic
	// look's pinstripes, with a small diamond between each pair.
	for i, x := 0, x0+k(6); x < x1; i, x = i+1, x+k(8) {
		shapes.FillRect(b, shapes.R(x-1, k(y0+6), 3, k(y1-y0-9)), s.WallShade)
		for y := k(y0+12) + float64(i%2)*24; y < k(y1-6); y += 48 {
			dx := x + 16
			shapes.FillPoly(b, s.WallShade, dx, y-5, dx+4, y, dx, y+5, dx-4, y)
		}
	}
	// The picture rail, lit along its top, and the skirting.
	shapes.FillRect(b, shapes.R(x0, k(y0+3), x1-x0, 5), s.WallLight)
	shapes.FillRect(b, shapes.R(x0, k(y0+3)+5, x1-x0, 3), s.WallShade)
	shapes.FillRect(b, shapes.R(x0, k(y1-3), x1-x0, 8), s.Trim)
	shapes.FillRect(b, shapes.R(x0, k(y1-3)+8, x1-x0, 2), r.fur.trimShade)
	shapes.FillRect(b, shapes.R(x0, bot-4, x1-x0, 4), s.Outline)

	// The side walls, in shade, with the rail and skirting carried round.
	for _, side := range []float64{1, -1} {
		ex, wx := 0.0, x0
		if side < 0 {
			ex, wx = w, x1
		}
		fadePoly(b, s.Outline, 0.3, ex, top-d, wx, top, wx, bot, ex, bot+d)
		shapes.Line(b, ex, top-d+k(3)+2, wx, k(y0+3)+2, 5, s.WallLight)
		shapes.FillPoly(b, s.Trim, wx, k(y1-3), wx, bot, ex, bot+d, ex, bot+d-k(3)-6)
	}

	// The floor: boards that widen towards us, each lit along its near
	// edge, with joints fanning out from the back wall.
	shapes.FillPoly(b, s.Ground, x0, bot, x1, bot, w, bot+d, w, h, 0, h, 0, bot+d)
	for i := 1; i < 8; i++ {
		bx := x0 + float64(i)*(x1-x0)/8
		shapes.Line(b, bx, bot, w/2+(bx-w/2)*3, h*2, 2.5, s.WallShade)
	}
	for y, gap := y1+3, 3; y < gfx.ScreenHeight; y, gap = y+gap, gap+1 {
		e := max(d-(k(y)-bot), 0)
		shapes.Line(b, e, k(y), w-e, k(y), 3, s.WallShade)
		shapes.Line(b, e, k(y)+3, w-e, k(y)+3, 2, s.GroundLight)
	}
	fadeRect(b, x0, bot, x1-x0, 8, s.Outline, 0.3)

	// The room's edges.
	ink := s.Outline
	for _, l := range [][4]float64{
		{x0, top, x1, top}, {x0, bot, x1, bot}, {x0, top, x0, bot}, {x1, top, x1, bot},
		{0, top - d, x0, top}, {w, top - d, x1, top}, {0, bot + d, x0, bot}, {w, bot + d, x1, bot},
	} {
		shapes.Line(b, l[0], l[1], l[2], l[3], 3, ink)
	}
}

// buildPicture draws the framed picture at (x, y), 64 pixels square: a
// little blue bird on its nest, in a gilt frame casting a shadow.
func (r *SmoothRenderer) buildPicture(x, y float64) {
	b := r.back
	fadeRect(b, x+5, y+5, 64, 64, r.scheme.Outline, 0.35)
	shapes.FillRect(b, shapes.R(x, y, 64, 64), resurrect64.Honey)
	shapes.FillPoly(b, resurrect64.Caramel, x+64, y, x+64, y+64, x, y+64, x+8, y+56, x+56, y+56, x+56, y+8)
	shapes.FillRect(b, shapes.R(x, y, 64, 3), resurrect64.Coral)
	shapes.FillRect(b, shapes.R(x, y+61, 64, 3), resurrect64.Copper)
	shapes.FillRect(b, shapes.R(x+8, y+8, 48, 48), resurrect64.Navy)
	shapes.FillPoly(b, resurrect64.Coral, oval(x+32, y+46, 17, 6)...)
	shapes.Line(b, x+19, y+45, x+45, y+47, 1.5, resurrect64.Copper)
	shapes.FillPoly(b, resurrect64.Blue, oval(x+32, y+35, 10, 11)...)
	shapes.FillCircle(b, x+32, y+24, 8, resurrect64.Blue)
	shapes.FillPoly(b, resurrect64.Cobalt, oval(x+37, y+36, 4, 7)...)
	shapes.FillCircle(b, x+29, y+23, 1.6, resurrect64.Ink)
	shapes.FillCircle(b, x+35, y+23, 1.6, resurrect64.Ink)
	shapes.FillPoly(b, resurrect64.Honey, x+30, y+27, x+34, y+27, x+32, y+31)
}

// buildWindow draws the window opening at (x, y) on the grid: night
// through the glass, a frame and sill, and the curtains either side on
// their rod, where the classic look hangs them.
func (r *SmoothRenderer) buildWindow(x, y int) {
	s, b, f := r.scheme, r.back, r.fur
	wx, wy, ww, wh := k(x), k(y), k(openingW), k(openingH)
	shapes.FillRoundRect(b, shapes.R(wx-6, wy-6, ww+12, wh+12), 2, s.Trim)
	shapes.FillRect(b, shapes.R(wx-2, wy-2, ww+4, wh+4), f.trimShade)
	shapes.FillRect(b, shapes.R(wx, wy, ww, wh), s.Glass)
	fadeRect(b, wx, wy+wh/2, ww, wh/2, resurrect64.Indigo, 0.5)
	for _, st := range [][2]float64{{14, 10}, {40, 24}, {92, 8}, {110, 30}, {70, 16}} {
		shapes.FillCircle(b, wx+st[0], wy+st[1], 1.5, resurrect64.Fog)
	}
	shapes.Line(b, wx+20, wy+56, wx+42, wy+34, 3, s.Glint)
	shapes.Line(b, wx+84, wy+56, wx+106, wy+34, 3, s.Glint)
	shapes.FillRect(b, shapes.R(wx+ww/2-3, wy, 6, wh), s.Trim)
	sill := shapes.R(wx-12, wy+wh+6, ww+24, 8)
	shapes.FillRoundRect(b, sill, 2, s.Trim)
	shapes.FillRect(b, shapes.R(sill.X, sill.Y+6, sill.W, 2), f.trimShade)
	fadeRect(b, sill.X+4, sill.Y+sill.H, sill.W-8, 5, s.Outline, 0.4)

	// The curtains, gathered in folds and scalloped at the hem.
	cy, cw, ch := wy-k(curtainDY), k(curtainW), k(26)
	for _, cx := range []float64{wx - cw + 6, wx + ww - 6} {
		shapes.FillPoly(b, f.body, cx, cy, cx+cw, cy, cx+cw+2, cy+ch-4, cx-2, cy+ch-4)
		for i := range 5 {
			hx := cx + float64(i)*cw/4
			shapes.FillCircle(b, hx, cy+ch-4, 7, f.body)
		}
		for _, fold := range []float64{12, 32, 52} {
			shapes.Line(b, cx+fold, cy+8, cx+fold+1, cy+ch, 4, f.shade)
			shapes.Line(b, cx+fold-5, cy+8, cx+fold-5, cy+ch-6, 2, f.light)
		}
		fadeRect(b, cx, cy+ch+4, cw, 4, s.Outline, 0.25)
		for i := range 4 {
			shapes.FillCircle(b, cx+8+float64(i)*16, cy+3, 3.5, f.rod)
		}
	}
	shapes.Line(b, wx-cw-4, cy, wx+ww+cw+4, cy, 5, f.rod)
	shapes.FillCircle(b, wx-cw-6, cy, 6, f.rodEnd)
	shapes.FillCircle(b, wx+ww+cw+6, cy, 6, f.rodEnd)
}

// buildLamp draws a standard lamp at (x, y) on the grid, in the room's
// furniture colours, casting a soft shadow on the floor: a striped shade
// lit underneath, a pole with its pull cord, and a heavy base, in a
// 104×256 box.
func (r *SmoothRenderer) buildLamp(x, y int) {
	f := r.fur
	fadePoly(r.back, r.scheme.Outline, 0.3, oval(k(x)+52, k(y+64)-4, 44, 7)...)
	r.canvas.Draw(r.back, entity.Picture{X: x, Y: y, W: 26, H: 64, Outline: r.scheme.Outline}, func(p entity.Pen) {
		p.Poly(f.body, 10, 0, 94, 0, 104, 36, 0, 36)
		for i := range 7 {
			tx, bx := 14+float64(i)*12.6, 6+float64(i)*15.4
			p.Stroke(f.shade, 3, false, tx+4, 1, bx+5, 35)
			p.Stroke(f.light, 2.5, false, tx, 1, bx, 35)
		}
		p.Stroke(f.light, 3, false, 11, 1.5, 93, 1.5)
		p.Stroke(f.shade, 4, false, 1, 34, 103, 34)
		p.Ellipse(52, 38, 22, 3.5, 0, resurrect64.Honey)
		p.Stroke(resurrect64.DeepTeal, 12, false, 50, 38, 50, 232)
		p.Stroke(f.body, 3, false, 47, 40, 47, 230)
		p.Stroke(resurrect64.Honey, 2, false, 74, 36, 74, 74)
		p.Circle(74, 78, 4, resurrect64.Coral)
		p.Poly(f.body, 26, 232, 74, 232, 80, 254, 20, 254)
		p.Poly(f.shade, 60, 232, 74, 232, 80, 254, 64, 254)
		p.Ellipse(50, 232, 25, 5, 0, f.light)
	})
}

// buildChair draws a balloon-backed chair at (x, y) on the grid, facing
// left when left is set, in the room's furniture colours, casting a soft
// shadow on the floor: the padded back, the seat and four legs, the far two
// darker, in a 128×208 box.
func (r *SmoothRenderer) buildChair(x, y int, left bool) {
	f := r.fur
	fadePoly(r.back, r.scheme.Outline, 0.3, oval(k(x)+64, k(y+52)-4, 64, 9)...)
	r.canvas.Draw(r.back, entity.Picture{X: x, Y: y, W: 32, H: 52, Flip: left, Outline: r.scheme.Outline}, func(p entity.Pen) {
		for _, l := range [][3]float64{{40, 200}, {88, 200}} {
			p.Stroke(resurrect64.Pine, 7, false, l[0], 140, l[0], l[1])
		}
		p.Ellipse(32, 64, 28, 64, 0, f.shade)
		p.Ellipse(29, 63, 24, 60, 0, f.body)
		p.Ellipse(18, 60, 7, 46, 0, f.light)
		p.Stroke(f.shade, 14, false, 14, 137, 120, 137)
		p.Stroke(f.body, 11, false, 14, 132, 120, 132)
		p.Stroke(f.light, 4, false, 16, 127, 118, 127)
		p.Stroke(resurrect64.DeepTeal, 8, false, 20, 142, 19, 205)
		p.Stroke(resurrect64.DeepTeal, 8, false, 108, 142, 109, 205)
		p.Stroke(f.body, 2.5, false, 106, 144, 107, 202)
	})
}

// drawTable draws the round table in a 224×144 box, in furniture colours
// f: its top seen from a little above, the apron, and four legs, the far
// two darker.
func drawTable(p entity.Pen, f furnish) {
	for _, x := range []float64{74, 150} {
		p.Stroke(resurrect64.Pine, 8, false, x, 30, x, 126)
	}
	for _, x := range []float64{22, 202} {
		p.Stroke(f.shade, 9, false, x, 30, x, 141)
		p.Stroke(f.body, 4, false, x-1.5, 32, x-1.5, 139)
	}
	p.Stroke(resurrect64.DeepTeal, 12, false, 14, 29, 210, 29)
	p.Stroke(f.shade, 9, false, 6, 20, 218, 20)
	p.Stroke(f.body, 6, false, 6, 18, 168, 18)
	p.Ellipse(112, 10, 110, 10, 0, f.light)
	p.Ellipse(80, 8, 50, 3, 0, f.shine)
}

// drawBirdcage draws the cage on its table or, once it has fallen, its
// wreck and the bird flying free.
func (r *SmoothRenderer) drawBirdcage(dst *ebiten.Image, bc *BirdcageRoom) {
	c, ink := &bc.Cage, r.scheme.Outline
	if c.Broken {
		r.canvas.Draw(dst, entity.Picture{X: c.X - 4, Y: c.Y, W: config.CageBrokenW, H: config.CageBrokenH, Outline: ink}, drawWreck)
		bd := &bc.Bird
		r.canvas.Draw(dst, entity.Picture{X: bd.X, Y: bd.Y, W: 8, H: 5, Flip: bd.VX < 0, Outline: ink}, func(p entity.Pen) {
			drawBird(p, bd.Frame)
		})
		return
	}
	r.canvas.Draw(dst, entity.Picture{X: c.X, Y: c.Y, W: config.CageW, H: config.CageH, Outline: ink}, drawCage)
}

// drawCage draws the birdcage in a 96×64 box: a domed cage of brass bars
// on a tray, hung from a ring, with the canary on its perch inside.
func drawCage(p entity.Pen) {
	const cx, cy, rx, ry, base = 48, 28, 42, 22, 52
	dome := []float64{}
	for i := range 17 {
		t := math.Pi + math.Pi*float64(i)/16
		dome = append(dome, cx+rx*math.Cos(t), cy+ry*math.Sin(t))
	}
	p.Poly(resurrect64.Grape, append(dome, cx+rx, base, cx-rx, base)...)
	p.Stroke(resurrect64.Copper, 3, false, 28, 40, 68, 40)
	drawBird(p.At(30, 22, 1), 1)
	for x := cx - rx + 4.0; x <= cx+rx-4; x += 9.5 {
		u := (x - cx) / rx
		c := resurrect64.Honey
		if x > cx+12 {
			c = resurrect64.Caramel
		}
		p.Stroke(c, 2.5, false, x, cy-ry*math.Sqrt(1-u*u), x, base)
	}
	p.Stroke(resurrect64.Honey, 3, true, dome...)
	p.Stroke(resurrect64.Honey, 3, false, cx-rx, 26, cx+rx, 26)
	p.Stroke(resurrect64.Honey, 3, false, cx-rx, 44, cx+rx, 44)
	ring := []float64{}
	for i := range 13 {
		t := 2 * math.Pi * float64(i) / 12
		ring = append(ring, cx+4*math.Cos(t), 2+4*math.Sin(t))
	}
	p.Stroke(resurrect64.Honey, 2, false, ring...)
	p.Stroke(resurrect64.Coral, 7, false, 3, 55, 93, 55)
	p.Stroke(resurrect64.Copper, 4, false, 3, 60, 93, 60)
	p.Stroke(resurrect64.Ink, 3, false, 2, 63, 94, 63)
}

// drawWreck draws the broken cage in a 128×68 box: the tray, with the bars
// bent and snapped above it and a piece of the dome hanging off.
func drawWreck(p entity.Pen) {
	for _, bar := range [][]float64{
		{24, 54, 20, 38, 30, 22, 40, 14},
		{42, 54, 46, 34, 40, 18},
		{60, 54, 56, 40, 58, 28},
		{80, 54, 84, 36, 78, 28},
	} {
		p.Stroke(resurrect64.Honey, 3, true, bar...)
	}
	for _, bar := range [][]float64{
		{94, 54, 100, 36, 108, 16},
		{108, 54, 114, 40, 120, 34},
	} {
		p.Stroke(resurrect64.Caramel, 3, true, bar...)
	}
	p.Stroke(resurrect64.Honey, 3, true, 50, 14, 60, 8, 74, 8, 86, 12)
	p.Stroke(resurrect64.Coral, 3, true, 56, 12, 68, 12, 80, 14)
	p.Stroke(resurrect64.Coral, 7, false, 11, 56, 117, 56)
	p.Stroke(resurrect64.Copper, 4, false, 11, 61, 117, 61)
	p.Stroke(resurrect64.Ink, 3, false, 10, 65, 118, 65)
}

// birdWings is where the bird's wingtip is in each of its four frames:
// up, level, down and level again.
var birdWings = [4][2]float64{{9, 0}, {4, 8}, {9, 18}, {4, 8}}

// drawBird draws the canary facing right in a 32×20 box, its wing in
// frame's position.
func drawBird(p entity.Pen, frame int) {
	p.Poly(resurrect64.Forest, 0, 7, 8, 10, 8, 13, 0, 14)
	p.Ellipse(16, 11, 10, 5.5, -0.15, resurrect64.Honey)
	p.Ellipse(15, 14, 6, 2.5, 0, resurrect64.Caramel)
	p.Circle(25, 8, 5, resurrect64.Honey)
	p.Poly(resurrect64.Coral, 29, 6, 33, 8.5, 29, 10)
	p.Circle(26, 7, 1.3, resurrect64.Ink)
	tip := birdWings[frame%len(birdWings)]
	p.Poly(resurrect64.Leaf, 13, 11, 21, 9, tip[0], tip[1])
	p.Circle(17, 10, 3.2, resurrect64.Leaf)
	p.Stroke(resurrect64.Forest, 1.2, false, 17, 10, (17+tip[0])/2, (10+tip[1])/2)
}

// drawBroom draws the broom in its pose, as the classic look does: the
// handle swings over a fixed head, and the bristles trail a little behind.
func (r *SmoothRenderer) drawBroom(dst *ebiten.Image, b *entity.Broom) {
	pose := float64(broomPoses[b.Frame])
	pic := entity.Picture{X: b.X, Y: b.Y, W: config.BroomW, H: config.BroomH, Outline: r.scheme.Outline}
	r.canvas.Draw(dst, pic, func(p entity.Pen) {
		top := 18 + 4*pose
		p.Stroke(resurrect64.Coral, 6, false, top, 2, 30, 84)
		p.Stroke(resurrect64.Honey, 2, false, top+1.5, 3, 31.5, 82)
		sway := -(pose - 3) * 2
		p.Poly(resurrect64.Honey, 18, 88, 44, 88, 58+sway, 116, 6+sway, 116)
		for i := range 6 {
			x := 22 + float64(i)*4
			p.Stroke(resurrect64.Coral, 2.5, false, x, 90, 9+float64(i)*9.6+sway, 115)
		}
		for i := range 9 {
			x := 10 + float64(i)*5.5 + sway
			p.Stroke(resurrect64.Caramel, 2.5, false, x, 110, x, 117)
		}
		p.Stroke(resurrect64.DeepTeal, 9, false, 21, 84, 41, 84)
		p.Stroke(resurrect64.Emerald, 2, false, 21, 82, 41, 82)
	})
}

// pawSpots is where each of the four footprints a column can hold lies in
// its 8×5 cell, as the classic look's sprite has them.
var pawSpots = [config.FootprintMax][2]int{{1, 3}, {5, 1}, {2, 0}, {6, 3}}

// drawPrints draws the first n paw prints of the column whose cell is at
// (x, y) on the grid, in the room's footprint colour.
func (r *SmoothRenderer) drawPrints(dst *ebiten.Image, x, y, n int) {
	c := r.scheme.Prints
	for _, s := range pawSpots[:min(n, len(pawSpots))] {
		px, py := k(x+s[0])+4, k(y+s[1])+4
		shapes.FillPoly(dst, c, oval(px, py+2, 3.5, 2.8)...)
		for _, t := range [][2]float64{{-3.5, -2.5}, {0, -4}, {3.5, -2.5}} {
			shapes.FillCircle(dst, px+t[0], py+t[1], 1.4, c)
		}
	}
}
