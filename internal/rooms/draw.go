package rooms

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Birdcage room layout, measured from the reference for layout only (DESIGN
// §3.4, "Birdcage room layout"). The art is our own.
const (
	wallX0, wallY0, wallX1, wallY1 = 16, 56, 304, 152
	pictureX, pictureY             = 88, 88
	openingX, openingY             = 232, 88
	openingW, openingH             = 32, 16
	curtainW, curtainDY            = 16, 2
	lampX, lampY                   = 16, 112
	furnitureY                     = 128 // chair tops
	tableX, tableY                 = 128, 144
)

var chairs = []struct {
	x    int
	left bool // facing left
}{{56, false}, {208, true}, {248, true}}

// broomPoses maps the broom's 12-step swing to its 7 poses.
var broomPoses = [config.BroomFrames]int{0, 1, 2, 3, 4, 5, 6, 5, 4, 3, 2, 1}

// Renderer draws a room. It caches the backdrop; it holds no game state.
type Renderer struct {
	sheet  *gfx.Sheet // room sprites
	actors *gfx.Sheet // Freddy and the dog
	scheme gfx.Scheme
	back   *ebiten.Image
	water  *ebiten.Image // the underwater backdrop
	kiss   *ebiten.Image // the heart room backdrop in the kiss colours
}

// NewRenderer returns a renderer using the room and actor sheets.
func NewRenderer(sheet, actors *gfx.Sheet) *Renderer {
	return &Renderer{sheet: sheet, actors: actors, scheme: gfx.BirdcageScheme}
}

func fill(dst *ebiten.Image, x, y, w, h, idx int) {
	dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(gfx.Palette[idx])
}

func line(dst *ebiten.Image, x0, y0, x1, y1, idx int) {
	vector.StrokeLine(dst, float32(x0)+0.5, float32(y0)+0.5, float32(x1)+0.5, float32(y1)+0.5, 1, gfx.Palette[idx], false)
}

// buildWalls fills the backdrop and draws the room in perspective, with the
// back wall from y0 to y1 (M8f): a darker ceiling, wallpaper on the back
// wall with a picture rail and skirting, lightly shaded side walls and
// floorboards. The floor keeps the room's Ground colour, so footprints
// still show (DESIGN §3.4).
func (r *Renderer) buildWalls(y0, y1 int) {
	s, b := r.scheme, r.back
	b.Fill(gfx.Palette[s.Wall])
	const sw = gfx.ScreenWidth
	// edge is how far the side wall reaches in from the screen edge on a row
	// of the ceiling or the floor, which slope over 16 rows to the back wall.
	edge := func(d int) int { return max(wallX0-d*wallX0/16, 0) }

	for y := 0; y < y0; y++ { // the ceiling
		x := edge(y0 - y)
		fill(b, x, y, sw-2*x, 1, s.WallShade)
	}
	for y := y0; y < y1; y++ { // the side walls, lightly shaded
		for x := (y % 2); x < wallX0; x += 2 {
			if y%4 < 2 {
				fill(b, x, y, 1, 1, s.WallShade)
				fill(b, sw-1-x, y, 1, 1, s.WallShade)
			}
		}
	}
	for x := wallX0 + 6; x < wallX1; x += 8 { // wallpaper pinstripes
		fill(b, x, y0+6, 1, y1-y0-9, s.WallShade)
	}
	fill(b, wallX0, y0+3, wallX1-wallX0, 1, s.WallLight) // the picture rail, lit on top
	fill(b, wallX0, y0+4, wallX1-wallX0, 1, s.WallShade)
	fill(b, wallX0, y1-3, wallX1-wallX0, 2, s.Trim) // the skirting
	fill(b, wallX0, y1-1, wallX1-wallX0, 1, s.Outline)
	// Floorboards: rows that widen towards us, and joints that fan out.
	for y, gap := y1+3, 3; y < gfx.ScreenHeight; y, gap = y+gap, gap+1 {
		x := edge(y - y1)
		fill(b, x, y, sw-2*x, 1, s.WallShade)
		for lx := x + y%2; lx < sw-x; lx += 2 { // the next board's lit edge, dithered
			fill(b, lx, y+1, 1, 1, s.GroundLight)
		}
	}
	for i := 1; i < 8; i++ {
		bx := wallX0 + i*(wallX1-wallX0)/8
		line(b, bx, y1, sw/2+(bx-sw/2)*3, gfx.ScreenHeight*2, s.WallShade)
	}

	ink := s.Outline
	line(b, wallX0, y0, wallX1, y0, ink)
	line(b, wallX0, y1, wallX1, y1, ink)
	line(b, wallX0, y0, wallX0, y1, ink)
	line(b, wallX1, y0, wallX1, y1, ink)
	line(b, 0, y0-16, wallX0, y0, ink)
	line(b, gfx.ScreenWidth-1, y0-16, wallX1, y0, ink)
	line(b, 0, y1+16, wallX0, y1, ink)
	line(b, gfx.ScreenWidth-1, y1+16, wallX1, y1, ink)
}

// buildWindow draws the window opening at (x, y) with a curtain either
// side.
func (r *Renderer) buildWindow(x, y int, curtainName string) {
	s, b := r.scheme, r.back
	fill(b, x-1, y-1, openingW+2, openingH+2, s.Trim)
	fill(b, x, y, openingW, openingH, s.Glass)
	curtain := r.sheet.Sprite(curtainName).Frame(0)
	entity.DrawAt(b, curtain, x-curtainW, y-curtainDY, false)
	entity.DrawAt(b, curtain, x+openingW, y-curtainDY, true)
}

// buildBirdcage draws the birdcage room's backdrop: the walls in
// perspective, the picture, the window with its curtains, the lamp, the
// chairs and the table.
func (r *Renderer) buildBirdcage() {
	b := r.back
	r.buildWalls(wallY0, wallY1)
	entity.DrawAt(b, r.sheet.Sprite("picture").Frame(0), pictureX, pictureY, false)
	r.buildWindow(openingX, openingY, "curtain")
	entity.DrawAt(b, r.sheet.Sprite("lamp").Frame(0), lampX, lampY, false)
	for _, c := range chairs {
		entity.DrawAt(b, r.sheet.Sprite("chair").Frame(0), c.x, furnitureY, c.left)
	}
	entity.DrawAt(b, r.sheet.Sprite("table").Frame(0), tableX, tableY, false)
}

// Draw renders the room.
func (r *Renderer) Draw(dst *ebiten.Image, room Room) {
	if h, ok := room.(*HeartRoom); ok {
		r.drawHeart(dst, h)
		return
	}
	if fb, ok := room.(*FishbowlRoom); ok && fb.Phase != InRoom {
		r.drawUnderwater(dst, fb)
		return
	}
	if r.back == nil {
		r.back = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		switch room.Kind() {
		case Cheese:
			r.scheme = gfx.CheeseScheme
			r.buildCheese()
		case Fishbowl:
			r.scheme = gfx.FishbowlScheme
			r.buildFishbowl()
		case Dogs:
			r.scheme = gfx.DogsScheme
			r.buildDogs()
		case Bookshelf:
			r.scheme = gfx.BookshelfScheme
			r.buildBookshelf(room.(*BookshelfRoom))
		default:
			r.buildBirdcage()
		}
	}
	dst.DrawImage(r.back, nil)
	b := room.Shared()

	feet := r.sheet.Sprite("footprint")
	for col, n := range b.Broom.Marks {
		if n > 0 {
			drawTinted(dst, feet.Frame(n-1), col*8, config.FootprintY, false, r.scheme.Prints)
		}
	}
	switch rm := room.(type) {
	case *BirdcageRoom:
		r.drawBirdcage(dst, rm)
	case *CheeseRoom:
		r.drawCheese(dst, rm)
	case *FishbowlRoom:
		r.drawBowl(dst, rm)
	case *DogsRoom:
		r.drawDogs(dst, rm)
	case *BookshelfRoom:
		r.drawBookshelf(dst, rm)
	}
	entity.DrawDog(dst, r.actors, &b.Dog, "dog_run_orange")
	entity.DrawAt(dst, r.sheet.Sprite("broom").Frame(broomPoses[b.Broom.Frame]), b.Broom.X, b.Broom.Y, false)
	if dr, ok := room.(*DogsRoom); ok && dr.Eating {
		r.drawEating(dst, dr)
	} else {
		entity.DrawFreddy(dst, r.actors, b.Freddy)
	}
	if br, ok := room.(*BookshelfRoom); ok {
		r.drawBite(dst, br)
	}
}

func (r *Renderer) drawBirdcage(dst *ebiten.Image, bc *BirdcageRoom) {
	c := &bc.Cage
	if c.Broken {
		entity.DrawAt(dst, r.sheet.Sprite("cage_broken").Frame(0), c.X-4, c.Y, false)
		bd := &bc.Bird
		entity.DrawAt(dst, r.sheet.Sprite("bird").Frame(bd.Frame), bd.X, bd.Y, bd.VX < 0)
		return
	}
	entity.DrawAt(dst, r.sheet.Sprite("cage").Frame(0), c.X, c.Y, false)
}

// DrawDebug outlines the room's collision boxes for debug builds: the
// surfaces (as the zone Freddy's top-left must be in), the window, the
// broom, the cage and bird, and Freddy.
func (r *Renderer) DrawDebug(dst *ebiten.Image, room Room) {
	box := func(x, y, w, h, idx int) {
		vector.StrokeRect(dst, float32(x)+0.5, float32(y)+0.5, float32(w)-1, float32(h)-1, 1, gfx.Palette[idx], false)
	}
	if h, ok := room.(*HeartRoom); ok {
		drawHeartDebug(dst, h)
		return
	}
	b := room.Shared()
	if fb, ok := room.(*FishbowlRoom); ok && fb.Phase != InRoom {
		for i := range fb.Swimmers {
			if !fb.Swimmers[i].Off {
				sr := fb.SwimmerRect(i)
				box(sr.X, sr.Y, sr.W, sr.H, resurrect64.White)
			}
		}
		h := b.Freddy.HitBox()
		box(h.X, h.Y, h.W, h.H, resurrect64.Leaf)
		return
	}
	for _, s := range b.Layout.Surfaces {
		idx := resurrect64.Honey
		if s.Perch {
			idx = resurrect64.Seafoam
		}
		box(s.X, s.Band, s.Reach+8, 8, idx)
	}
	e := b.ExitRect()
	box(e.X, e.Y, e.W, e.H, resurrect64.Blue)
	br := b.Broom.Rect()
	box(br.X, br.Y, br.W, br.H, resurrect64.Coral)
	switch rm := room.(type) {
	case *BirdcageRoom:
		if rm.Cage.Broken {
			br := rm.birdRect()
			box(br.X, br.Y, br.W, br.H, resurrect64.White)
		} else {
			cr := rm.cageRect()
			box(cr.X, cr.Y, cr.W, cr.H, resurrect64.White)
		}
	case *FishbowlRoom:
		br := rm.BowlRect()
		box(br.X, br.Y, br.W, br.H, resurrect64.White)
	case *BookshelfRoom:
		if l := b.Layout.Ladder; l != nil {
			for y := l.TopY; y <= l.BottomY; y += l.Gap {
				box(l.MinX, y+l.Snap, l.MaxX-l.MinX+4, 8, resurrect64.Seafoam)
			}
		}
		for i := range rm.Ferns {
			if !rm.Ferns[i] {
				fr := FernRect(i)
				box(fr.X, fr.Y, fr.W, fr.H, resurrect64.Leaf)
			}
		}
		sr := rm.SpiderRect()
		box(sr.X, sr.Y, sr.W, sr.H, resurrect64.White)
	case *DogsRoom:
		for i, sd := range rm.Sleepers {
			if sd.On {
				sr := SleeperRect(i)
				box(sr.X, sr.Y, sr.W, sr.H, []int{resurrect64.White, resurrect64.Honey, resurrect64.Cerise}[sd.Alert])
			}
		}
	case *CheeseRoom:
		for i := range rm.Mice {
			if rm.Mice[i].Visible() {
				mr := rm.MouseRect(i)
				box(mr.X, mr.Y, mr.W, mr.H, resurrect64.White)
			}
		}
	}
	if f := b.Freddy; !f.Gone {
		h := f.HitBox()
		box(h.X, h.Y, h.W, h.H, resurrect64.Leaf)
	}
}
