package alley

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
)

// Layout positions, measured from the reference for layout only (DESIGN
// §3.3, "Layout positions"). The art itself is our own.
const (
	windowW, windowH = 40, 16
	windowLeft       = 20 // first column's left edge
	windowPitchX     = 80
	windowTop        = 24 // first row's top edge
	windowPitchY     = 32

	fenceTop    = 104 // highest point of the ragged board tops
	fenceBottom = 175
	groundTop   = 176
	canSpriteDX = 1 // can sprites sit this far right of the can's X
	canSpriteDY = 5 // and this far below the Y Freddy stands at
)

// clutter is the litter along the ground: sprite name and position.
var clutter = []struct {
	name string
	x, y int
}{
	{"fishbones", 124, 189},
	{"bone", 186, 186},
	{"tincan", 222, 188},
}

// Renderer draws the alley. It caches the backdrop layers, which only
// change with the can layout; it holds no game state.
type Renderer struct {
	sheet       *gfx.Sheet
	font        *gfx.Font
	scheme      gfx.Scheme
	back, front *ebiten.Image
	builtFor    []config.Can
}

// NewRenderer returns a renderer using the alley sprite sheet and the font.
func NewRenderer(sheet *gfx.Sheet, font *gfx.Font) *Renderer {
	return &Renderer{sheet: sheet, font: font, scheme: gfx.AlleyScheme}
}

// Draw renders the alley with the fence HUD showing the high score. Freddy
// is drawn behind the fence or in front of it according to his depth
// (DESIGN §3.8, "Fence depth"); thrown objects sink behind it.
func (r *Renderer) Draw(dst *ebiten.Image, a *Alley, high int) {
	r.build(a.Cans)
	dst.DrawImage(r.back, nil)
	r.drawWindow(dst, &a.Window)
	for i := range a.Lines {
		r.drawLine(dst, &a.Lines[i])
	}
	r.drawMice(dst, a)
	r.drawThrown(dst, &a.Thrown)
	f := a.Freddy
	if f.Depth == entity.Behind {
		entity.DrawFreddyRimmed(dst, r.sheet, f, r.scheme.Shade)
	}
	dst.DrawImage(r.front, nil)
	r.drawHUD(dst, a, high)
	r.drawTrashCat(dst, &a.TrashCat)
	entity.DrawDog(dst, r.sheet, &a.Dog, "dog_run")
	if f.Depth != entity.Behind {
		entity.DrawFreddyRimmed(dst, r.sheet, f, r.scheme.Shade)
	}
	r.drawImpact(dst, a.Impact)
}

func fill(dst *ebiten.Image, x, y, w, h, idx int) {
	dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(gfx.Palette[idx])
}

func (r *Renderer) build(cans []config.Can) {
	if r.back != nil && slices.Equal(r.builtFor, cans) {
		return
	}
	r.builtFor = slices.Clone(cans)
	if r.back == nil {
		r.back = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.front = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
	}
	r.buildBack()
	r.buildFront(cans)
}

// The backdrop's art (DESIGN §8, M8a): our own, drawn once into the back
// layer. None of it changes the layout of §3.3.
const (
	skyBottom = 16 // rows above this are night sky
	wallTop   = 20 // the cornice runs from skyBottom to here
)

// skyBands are the night sky's colour bands, darkest at the top.
var skyBands = []struct{ to, idx int }{{4, gfx.Night}, {8, gfx.Charcoal}, {skyBottom, gfx.Navy}}

// stars are single pixels in the upper sky.
var stars = [][2]int{{12, 2}, {37, 5}, {71, 1}, {98, 6}, {133, 3}, {166, 1}, {188, 6}, {219, 2}, {241, 5}, {301, 3}}

// skyline is the distant rooftops: left edge, width and top row. Lit windows
// are a pixel or two in each.
var skyline = []struct{ x, w, top int }{
	{0, 22, 12}, {22, 14, 10}, {36, 26, 13}, {62, 10, 11}, {72, 30, 13},
	{102, 18, 12}, {120, 24, 14}, {144, 12, 11}, {156, 34, 13}, {190, 16, 12},
	{206, 28, 13}, {234, 12, 10}, {246, 30, 13}, {276, 20, 12}, {296, 24, 13},
}

// litWindows are the windows with a lamp on, by index (row × 4 + column):
// warm light against the cool wall.
var litWindows = map[int]bool{1: true, 4: true, 7: true, 10: true}

// buildBack draws the night sky, the distant skyline, the building's
// roofline and brick wall, its drainpipes and the 3×4 windows.
func (r *Renderer) buildBack() {
	s, b := r.scheme, r.back
	b.Fill(gfx.Palette[s.Wall])

	from := 0
	for _, band := range skyBands {
		fill(b, 0, from, gfx.ScreenWidth, band.to-from, band.idx)
		from = band.to
	}
	for i, st := range stars {
		idx := gfx.Silver
		if i%3 == 0 {
			idx = gfx.White
		}
		fill(b, st[0], st[1], 1, 1, idx)
	}
	entity.DrawAt(b, r.sheet.Sprite("moon").Frame(0), 268, 1, false)
	for i, sk := range skyline {
		fill(b, sk.x, sk.top, sk.w, skyBottom-sk.top, gfx.Night)
		fill(b, sk.x+sk.w/3, sk.top-2, 2, 2, gfx.Night) // a chimney
		if i%2 == 0 {
			fill(b, sk.x+sk.w/2, sk.top+2, 1, 1, gfx.Yellow)
		}
		if i%3 == 1 {
			fill(b, sk.x+3, sk.top+3, 1, 1, gfx.Orange)
		}
	}
	// A water tower on the far roofs.
	fill(b, 150, 3, 9, 4, gfx.Night)
	fill(b, 151, 2, 7, 1, gfx.Night)
	fill(b, 151, 7, 1, 2, gfx.Night)
	fill(b, 157, 7, 1, 2, gfx.Night)

	// The cornice along our building's roof.
	fill(b, 0, skyBottom, gfx.ScreenWidth, wallTop-skyBottom, s.WallShade)
	fill(b, 0, skyBottom, gfx.ScreenWidth, 1, s.Shade)
	for x := 2; x < gfx.ScreenWidth; x += 8 {
		fill(b, x, skyBottom+2, 4, 1, s.Outline)
	}

	// Brick courses 6 rows high, the joints staggered.
	for y, course := wallTop+5, 0; y < fenceTop; y, course = y+6, course+1 {
		for x := (course % 2) * 8; x < gfx.ScreenWidth; x += 16 {
			fill(b, x, y, 12, 1, s.WallShade)
			fill(b, x+12, y-5, 1, 5, s.WallShade)
		}
	}

	// Drainpipes down both edges, with brackets.
	for _, px := range []int{3, gfx.ScreenWidth - 7} {
		fill(b, px, wallTop, 4, fenceTop-wallTop, s.WallShade)
		fill(b, px+1, wallTop, 1, fenceTop-wallTop, s.Shade)
		for y := wallTop + 10; y < fenceTop; y += 24 {
			fill(b, px-1, y, 6, 1, s.Shade)
		}
	}

	for row := range 3 {
		for col := range 4 {
			r.buildWindow(b, windowLeft+col*windowPitchX, windowTop+row*windowPitchY, litWindows[row*4+col])
		}
	}
}

// buildWindow draws one closed window: a lintel above, a sill below, the
// frame and the glass, dark or lit by a lamp inside.
func (r *Renderer) buildWindow(b *ebiten.Image, x, y int, lit bool) {
	s := r.scheme
	fill(b, x-2, y-3, windowW+4, 2, s.WallShade)         // lintel
	fill(b, x-3, y+windowH, windowW+6, 2, s.Shade)       // sill
	fill(b, x-3, y+windowH+2, windowW+6, 1, s.WallShade) // its shadow
	fill(b, x, y, windowW, windowH, s.Trim)
	fill(b, x+2, y+1, windowW-4, windowH-2, s.Outline)
	if !lit {
		fill(b, x+3, y+2, windowW-6, windowH-4, s.Glass)
		for _, gx := range []int{x + 6, x + 25} {
			for i := range 4 {
				fill(b, gx+i, y+8-i, 1, 1, s.Glint)
			}
		}
	} else {
		fill(b, x+3, y+2, windowW-6, windowH-4, gfx.Yellow)
		fill(b, x+3, y+windowH-6, windowW-6, 4, gfx.Orange) // the lamp's glow
		for _, cx := range []int{x + 3, x + windowW - 6} {
			fill(b, cx, y+2, 3, windowH-4, gfx.Red) // curtains
		}
	}
	fill(b, x+windowW/2-1, y+1, 2, windowH-2, s.Trim)
}

// buildFront draws the fence, the ground, the litter and the trash cans.
func (r *Renderer) buildFront(cans []config.Can) {
	s, f := r.scheme, r.front
	f.Clear()
	// Weathered boards: a lit left edge, a shadowed right, grain and a knot.
	for i, x := 0, 0; x < gfx.ScreenWidth; i, x = i+1, x+8 {
		top := fenceTop + (i*37+i/3)%4
		fill(f, x, top, 7, fenceBottom-top+1, s.Fence)
		fill(f, x, top+1, 1, fenceBottom-top, s.Shade)
		fill(f, x+7, top+2, 1, fenceBottom-top-1, s.Gap)
		fill(f, x+4, top+3+(i*7)%9, 1, 6+(i*5)%8, s.Gap) // grain
		fill(f, x+2, top+(i*11)%40+14, 2, 1, s.Gap)      // a knot
		fill(f, x+1, top, 5, 1, s.Gap)                   // the cut top
	}
	// The rails behind the boards show as shadow lines, with a nail in each
	// board.
	for _, ry := range []int{128, 160} {
		fill(f, 0, ry, gfx.ScreenWidth, 1, s.Gap)
		for x := 3; x < gfx.ScreenWidth; x += 8 {
			fill(f, x, ry-2, 1, 1, s.Outline)
		}
	}
	// The ground: cobbles in rows, with a lit edge along the fence.
	fill(f, 0, groundTop, gfx.ScreenWidth, gfx.ScreenHeight-groundTop, s.Ground)
	fill(f, 0, groundTop, gfx.ScreenWidth, 1, s.Shade)
	for y, row := groundTop+6, 0; y < gfx.ScreenHeight; y, row = y+6, row+1 {
		fill(f, 0, y, gfx.ScreenWidth, 1, s.Outline)
		for x := (row % 2) * 7; x < gfx.ScreenWidth; x += 14 {
			fill(f, x, y-5, 1, 5, s.Outline)
			fill(f, x+4, y-3, 1, 1, s.Shade) // a speck
		}
	}

	r.drawGraffiti(f)
	for _, c := range clutter {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(c.x), float64(c.y))
		f.DrawImage(r.sheet.Sprite(c.name).Frame(0), op)
	}
	for _, c := range cans {
		r.drawCan(f, c)
	}
}

// drawCan draws an open trash can. Its rim lines up with Freddy's feet when
// he stands on it, and the front of the rim with the bottom of the trash
// cat rising out of it (DESIGN §3.8, "Alley systems").
func (r *Renderer) drawCan(dst *ebiten.Image, c config.Can) {
	name := "can_short"
	if c.Tall {
		name = "can_tall"
	}
	entity.DrawAt(dst, r.sheet.Sprite(name).Frame(0), c.X+canSpriteDX, CanTop(c)+canSpriteDY, false)
}

// drawLine draws a clothesline and its laundry.
func (r *Renderer) drawLine(dst *ebiten.Image, l *Line) {
	y := config.LineY[l.Row]
	for x := 0; x < gfx.ScreenWidth; x += 2 {
		fill(dst, x, y, 1, 1, r.scheme.Trim)
		fill(dst, x+1, y, 1, 1, r.scheme.Shade)
	}
	for i, p := range l.Pieces {
		px := l.Offset + i*config.LinePiecePx
		for h, k := range p {
			x := px + h*config.LinePiecePx/2
			var img *ebiten.Image
			switch {
			case k == BigLeft:
				img = r.sheet.Sprite("laundry_big").Frame(0)
			case k == MediumLeft:
				img = r.sheet.Sprite("laundry_medium").Frame(0)
			case k >= Small:
				img = r.sheet.Sprite("laundry_small").Frame(int(k - Small))
			default:
				continue
			}
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(img, op)
		}
	}
}
