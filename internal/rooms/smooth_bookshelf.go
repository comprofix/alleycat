package rooms

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// buildBookshelf draws the bookshelf room's backdrop, as the classic look
// lays it out (DESIGN §3.4, "Bookshelf room layout"): the walls, the
// window with its curtains, two lamps, two chairs and the bookcase with
// this visit's books, each casting a soft shadow on the floor.
func (r *SmoothRenderer) buildBookshelf(room *BookshelfRoom) {
	b, ink := r.back, r.scheme.Outline
	r.buildWalls(wallY0, wallY1)
	r.buildWindow(bookOpeningX, openingY)

	floor := k(furnitureY + 52)
	for _, x := range []int{bookLampLX, bookLampRX} {
		fadePoly(b, ink, 0.3, oval(k(x)+52, k(lampY+64)-4, 44, 7)...)
	}
	for _, x := range []int{bookChairLX, bookChairRX} {
		fadePoly(b, ink, 0.3, oval(k(x)+64, floor-4, 64, 9)...)
	}
	r.buildLamp(bookLampLX)
	r.buildLamp(bookLampRX)
	r.buildChair(bookChairLX, false)
	r.buildChair(bookChairRX, true)
	r.buildBookcase(room)
}

// spineShade is the shade down the side of a book of each colour: the
// next step down its ramp.
var spineShade = map[int]int{
	resurrect64.White: resurrect64.Mist,
	resurrect64.Mist:  resurrect64.Fog,
	resurrect64.Fog:   resurrect64.Lilac,
	resurrect64.Lilac: resurrect64.Dusk,
}

// spines are the books' colours, the sprites' whites, greys and lilac.
var spines = []int{resurrect64.White, resurrect64.Mist, resurrect64.Fog, resurrect64.Lilac}

// buildBookcase draws the bookcase where the classic look has it: a teal
// case with a shelf along the foot of every row of cells, and in each cell
// what this visit's mix put there (DESIGN §3.4, "Bookshelf room in
// detail").
func (r *SmoothRenderer) buildBookcase(room *BookshelfRoom) {
	b, ink, f := r.back, r.scheme.Outline, r.fur
	x0, y0 := k(config.BookcaseX), k(config.BookcaseY)
	cell := k(config.BookCell)
	w, h := cell*config.BookCols, cell*config.BookRows
	fadePoly(b, ink, 0.35, x0-8, y0+h-4, x0+w+8, y0+h-4, x0+w+16, y0+h+10, x0-16, y0+h+10)

	shapes.FillRect(b, shapes.R(x0, y0, w, h), resurrect64.DeepTeal)
	for row, cells := range room.Books {
		y := y0 + float64(row)*cell
		fadeRect(b, x0, y, w, 6, ink, 0.35) // the shadow of the shelf above
		for col, kind := range cells {
			r.drawBook(b, kind, x0+float64(col)*cell, y, row*config.BookCols+col)
		}
		shapes.FillRect(b, shapes.R(x0, y+cell-8, w, 4), f.body)
		shapes.FillRect(b, shapes.R(x0, y+cell-8, w, 1.5), f.light)
		shapes.FillRect(b, shapes.R(x0, y+cell-4, w, 4), ink)
	}
	// The sides and the top, lit on their outer edges.
	for _, sx := range []float64{x0, x0 + w - 8} {
		shapes.FillRect(b, shapes.R(sx, y0, 8, h), f.body)
		shapes.FillRect(b, shapes.R(sx, y0, 2, h), f.light)
		shapes.FillRect(b, shapes.R(sx+6, y0, 2, h), f.shade)
	}
	top := shapes.R(x0-6, y0-6, w+12, 8)
	shapes.FillRect(b, top, f.body)
	shapes.FillRect(b, shapes.R(top.X, top.Y, top.W, 2), f.light)
	shapes.FillRect(b, shapes.R(top.X, top.Y+top.H-2, top.W, 2), f.shade)
	shapes.StrokeRect(b, shapes.R(x0-1, y0-1, w+2, h+2), 2, ink)
	shapes.StrokeRect(b, top.Inset(-1), 2, ink)
}

// drawBook draws what one 32-pixel cell of the bookcase holds, standing on
// the shelf at its foot: the case's sides, bare shelf, or a kind of book
// in a colour picked by its place n.
func (r *SmoothRenderer) drawBook(b *ebiten.Image, kind Book, x, y float64, n int) {
	ink := r.scheme.Outline
	h := uint32(n) * 0x9e3779b1 >> 24
	c := spines[h%uint32(len(spines))]
	c2 := spines[(h/4+1)%uint32(len(spines))]
	foot := y + 24 // the shelf's top
	book := func(bx, top, bw float64, c int) {
		shapes.FillRect(b, shapes.R(bx-1.5, top-1.5, bw+3, foot-top+1.5), ink)
		shapes.FillRect(b, shapes.R(bx, top, bw, foot-top), c)
		shapes.FillRect(b, shapes.R(bx+bw-3, top, 3, foot-top), spineShade[c])
		shapes.FillRect(b, shapes.R(bx, top+4, bw-3, 2), spineShade[c])
		shapes.FillRect(b, shapes.R(bx, foot-6, bw-3, 2), spineShade[c])
	}
	switch kind {
	case BookUpright:
		book(x+5, y+2+float64(h%3)*2, 18, c)
	case BookPair:
		book(x+3, y+5, 9, c)
		book(x+13, y+2, 11, c2)
	case BookLeaning:
		// Leaning on the next book, its foot against this cell's left.
		shapes.FillPoly(b, ink, x+1, foot+1, x+12, foot+1, x+33, y+4, x+24, y-2)
		shapes.FillPoly(b, c, x+3, foot, x+11, foot, x+30, y+4, x+24, y)
		shapes.FillPoly(b, spineShade[c], x+9, foot, x+11, foot, x+30, y+4, x+28, y+3)
	case BookStack:
		for i, s := range [][3]float64{{3, 26, 18}, {6, 22, 12}, {4, 24, 6}} {
			sc := spines[(h+uint32(i))%uint32(len(spines))]
			bk := shapes.R(x+s[0], y+s[2], s[1], 6)
			shapes.FillRoundRect(b, bk.Inset(-1.5), 2, ink)
			shapes.FillRect(b, bk, sc)
			shapes.FillRect(b, shapes.R(bk.X, bk.Y+bk.H-2, bk.W, 2), spineShade[sc])
			shapes.FillRect(b, shapes.R(bk.X+bk.W-4, bk.Y+1, 2, bk.H-3), resurrect64.White)
		}
	}
}

// drawBookshelf draws the ferns still standing and the spider on its
// thread.
func (r *SmoothRenderer) drawBookshelf(dst *ebiten.Image, room *BookshelfRoom) {
	for i, taken := range room.Ferns {
		if !taken {
			r.canvas.Draw(dst, entity.Picture{X: config.FernX[i], Y: config.FernY, W: config.FernW, H: config.FernH, Outline: entity.NoOutline}, drawFern)
		}
	}
	s := &room.Spider
	if s.Y > config.SpiderTopY {
		x := k(s.X) + 48
		shapes.Line(dst, x, k(config.SpiderTopY)+16, x, k(s.Y)+12, 1.5, resurrect64.Mist)
	}
	r.canvas.Draw(dst, entity.Picture{X: s.X, Y: s.Y, W: config.SpiderW, H: config.SpiderH, Outline: entity.NoOutline}, func(p entity.Pen) {
		drawSpider(p, s.Frame)
	})
}

// drawFern draws a potted fern in a 64×64 box: a bush of fronds arching
// out of a black pot, thick with leaflets, the inner ones darker.
func drawFern(p entity.Pen) {
	const steps = 8
	p.Ellipse(32, 36, 13, 7, 0, resurrect64.Forest)
	for _, a := range []float64{-2.85, -2.4, -2.0, -1.57, -1.14, -0.74, -0.29} {
		out := math.Abs(a + math.Pi/2)
		l := 30 - out*4
		c, shade := resurrect64.Emerald, resurrect64.Forest
		if out > 0.6 {
			c, shade = resurrect64.Leaf, resurrect64.Emerald
		}
		pts := make([]float64, 0, 2*(steps+1))
		for s := range steps + 1 {
			t := float64(s) / steps
			pts = append(pts, 32+math.Cos(a)*l*t, 38+math.Sin(a)*l*t+14*t*t*out)
		}
		for s := 1; s <= steps; s++ {
			x, y := pts[2*s], pts[2*s+1]
			dx, dy := x-pts[2*s-2], y-pts[2*s-1]
			dir := math.Atan2(dy, dx)
			r := 7.5 - 0.7*float64(s)
			for _, side := range []float64{-1, 1} {
				lx, ly := x+side*r*0.8*math.Cos(dir+math.Pi/2), y+side*r*0.8*math.Sin(dir+math.Pi/2)
				p.Ellipse(lx, ly, r, 2.4, dir+side*1.1, c)
			}
		}
		p.Stroke(shade, 2, true, pts...)
	}
	p.Poly(resurrect64.Shadow, 20, 42, 44, 42, 40, 59, 24, 59)
	p.Poly(resurrect64.Ink, 34, 42, 44, 42, 40, 59, 32, 59)
	p.Stroke(resurrect64.Ink, 6, false, 18, 42, 46, 42)
	p.Stroke(resurrect64.Shadow, 1.5, false, 20, 40.5, 40, 40.5)
}

// spiderLegs is where each of a spider's left legs bends and ends, from
// its hip, in each of its two frames: reaching out and up, then bent in
// and down. The right legs mirror them.
var spiderLegs = [2][4][4]float64{
	{{-14, -12, -30, -4}, {-16, -8, -34, 2}, {-16, -2, -32, 12}, {-14, 4, -26, 18}},
	{{-10, -8, -22, 0}, {-12, -4, -26, 8}, {-12, 2, -24, 16}, {-10, 8, -18, 22}},
}

// drawSpider draws the spider, seen from above, in a 96×56 box: eight
// jointed legs in frame's pose, a round head with two white eyes, and a
// pale body with a red mark.
func drawSpider(p entity.Pen, frame int) {
	const cx = 48
	for i, l := range spiderLegs[frame%2] {
		hy := 18 + float64(i)*5
		for _, side := range []float64{-1, 1} {
			hx := cx + side*6
			p.Stroke(resurrect64.Ink, 2.5, false, hx, hy, hx+side*-l[0], hy+l[1], hx+side*-l[2], hy+l[3])
		}
	}
	p.Ellipse(cx, 32, 11, 14, 0, resurrect64.Seafoam)
	p.Ellipse(cx+5, 33, 5, 12, 0, resurrect64.Jade)
	p.Circle(cx, 15, 7, resurrect64.Seafoam)
	p.Poly(resurrect64.Cerise, cx-4, 34, cx+4, 34, cx, 39, cx+4, 44, cx-4, 44, cx, 39)
	for _, ex := range []float64{cx - 3, cx + 3} {
		p.Circle(ex, 13, 2.2, resurrect64.White)
		p.Circle(ex, 13.5, 1, resurrect64.Ink)
	}
}

// drawBite draws "BYTE!" by Freddy while the bite plays.
func (r *SmoothRenderer) drawBite(dst *ebiten.Image, room *BookshelfRoom) {
	if room.Bite > 0 {
		r.canvas.Draw(dst, entity.Picture{X: room.ByteX, Y: room.ByteY, W: 48, H: 21, Outline: entity.NoOutline}, func(p entity.Pen) {
			entity.DrawSplat(p, r.text, "BYTE!", 48, 21)
		})
	}
}
