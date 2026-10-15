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
		entity.DrawFreddy(dst, r.sheet, f)
	}
	dst.DrawImage(r.front, nil)
	r.drawHUD(dst, a, high)
	r.drawTrashCat(dst, &a.TrashCat)
	entity.DrawDog(dst, r.sheet, &a.Dog, "dog_run")
	if f.Depth != entity.Behind {
		entity.DrawFreddy(dst, r.sheet, f)
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

// buildBack draws the building: brickwork and the 3×4 closed windows.
func (r *Renderer) buildBack() {
	s, b := r.scheme, r.back
	b.Fill(gfx.Palette[s.Wall])
	for y := 4; y < fenceTop; y += 8 {
		for x := (y / 8 % 2) * 6; x < gfx.ScreenWidth; x += 12 {
			fill(b, x, y, 4, 1, s.WallShade)
		}
	}
	for row := range 3 {
		for col := range 4 {
			x, y := windowLeft+col*windowPitchX, windowTop+row*windowPitchY
			fill(b, x, y, windowW, windowH, s.Trim)
			fill(b, x+2, y+1, windowW-4, windowH-2, s.Outline)
			fill(b, x+3, y+2, windowW-6, windowH-4, s.Glass)
			fill(b, x+windowW/2-1, y+1, 2, windowH-2, s.Trim)
			for _, gx := range []int{x + 6, x + 25} {
				for i := range 4 {
					fill(b, gx+i, y+8-i, 1, 1, s.Glint)
				}
			}
		}
	}
}

// buildFront draws the fence, the ground, the litter and the trash cans.
func (r *Renderer) buildFront(cans []config.Can) {
	s, f := r.scheme, r.front
	f.Clear()
	for i, x := 0, 0; x < gfx.ScreenWidth; i, x = i+1, x+8 {
		top := fenceTop + (i*37+i/3)%4
		fill(f, x, top, 7, fenceBottom-top+1, s.Fence)
		fill(f, x+7, top+2, 1, fenceBottom-top-1, s.Gap)
		fill(f, x+3, top+(i*11)%40+12, 1, 1, s.Gap) // a knot
	}
	fill(f, 0, 128, gfx.ScreenWidth, 1, s.Gap)
	fill(f, 0, 160, gfx.ScreenWidth, 1, s.Gap)
	fill(f, 0, groundTop, gfx.ScreenWidth, gfx.ScreenHeight-groundTop, s.Ground)
	fill(f, 0, groundTop, gfx.ScreenWidth, 1, s.Shade)

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
