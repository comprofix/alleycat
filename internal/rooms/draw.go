package rooms

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
)

// Birdcage room layout, measured from the reference for layout only (DESIGN
// §3.4, "Birdcage room layout"). The art is our own.
const (
	wallX0, wallY0, wallX1, wallY1 = 16, 56, 304, 152
	pictureX, pictureY             = 88, 88
	openingX, openingY             = 232, 88
	openingW, openingH             = 32, 16
	curtainLX, curtainRX, curtainY = 216, 264, 86
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

// buildBirdcage draws the birdcage room's backdrop: the walls in
// perspective, the picture, the window with its curtains, the lamp, the
// chairs and the table.
func (r *Renderer) buildBirdcage() {
	s, b := r.scheme, r.back
	b.Fill(gfx.Palette[s.Wall])
	ink := s.Outline
	line(b, wallX0, wallY0, wallX1, wallY0, ink)
	line(b, wallX0, wallY1, wallX1, wallY1, ink)
	line(b, wallX0, wallY0, wallX0, wallY1, ink)
	line(b, wallX1, wallY0, wallX1, wallY1, ink)
	line(b, 0, wallY0-16, wallX0, wallY0, ink)
	line(b, gfx.ScreenWidth-1, wallY0-16, wallX1, wallY0, ink)
	line(b, 0, wallY1+16, wallX0, wallY1, ink)
	line(b, gfx.ScreenWidth-1, wallY1+16, wallX1, wallY1, ink)

	entity.DrawAt(b, r.sheet.Sprite("picture").Frame(0), pictureX, pictureY, false)
	fill(b, openingX-1, openingY-1, openingW+2, openingH+2, s.Trim)
	fill(b, openingX, openingY, openingW, openingH, s.Glass)
	curtain := r.sheet.Sprite("curtain").Frame(0)
	entity.DrawAt(b, curtain, curtainLX, curtainY, false)
	entity.DrawAt(b, curtain, curtainRX, curtainY, true)
	entity.DrawAt(b, r.sheet.Sprite("lamp").Frame(0), lampX, lampY, false)
	for _, c := range chairs {
		entity.DrawAt(b, r.sheet.Sprite("chair").Frame(0), c.x, furnitureY, c.left)
	}
	entity.DrawAt(b, r.sheet.Sprite("table").Frame(0), tableX, tableY, false)
}

// Draw renders the room.
func (r *Renderer) Draw(dst *ebiten.Image, room Room) {
	if r.back == nil {
		r.back = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.buildBirdcage()
	}
	dst.DrawImage(r.back, nil)
	b := room.Shared()

	feet := r.sheet.Sprite("footprint")
	for col, n := range b.Broom.Marks {
		if n > 0 {
			entity.DrawAt(dst, feet.Frame(n-1), col*8, config.FootprintY, false)
		}
	}
	if bc, ok := room.(*BirdcageRoom); ok {
		r.drawBirdcage(dst, bc)
	}
	entity.DrawDog(dst, r.actors, &b.Dog, "dog_run_orange")
	entity.DrawAt(dst, r.sheet.Sprite("broom").Frame(broomPoses[b.Broom.Frame]), b.Broom.X, b.Broom.Y, false)
	entity.DrawFreddy(dst, r.actors, b.Freddy)
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
	b := room.Shared()
	for _, s := range b.Layout.Surfaces {
		idx := gfx.Yellow
		if s.Perch {
			idx = gfx.Cyan
		}
		box(s.X, s.Band, s.Reach+8, 8, idx)
	}
	e := b.ExitRect()
	box(e.X, e.Y, e.W, e.H, gfx.Sky)
	br := b.Broom.Rect()
	box(br.X, br.Y, br.W, br.H, gfx.Orange)
	if bc, ok := room.(*BirdcageRoom); ok {
		if bc.Cage.Broken {
			br := bc.birdRect()
			box(br.X, br.Y, br.W, br.H, gfx.White)
		} else {
			cr := bc.cageRect()
			box(cr.X, cr.Y, cr.W, cr.H, gfx.White)
		}
	}
	if f := b.Freddy; !f.Gone {
		h := f.HitBox()
		box(h.X, h.Y, h.W, h.H, gfx.Lime)
	}
}
