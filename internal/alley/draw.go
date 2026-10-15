package alley

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/physics"
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
	canFloorY   = 185 // can bodies stand on the ground here
	canInset    = 3   // body starts this far right of the can's X
	canW        = 27
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
	scheme      gfx.Scheme
	back, front *ebiten.Image
	builtFor    []config.Can
}

// NewRenderer returns a renderer using the alley sprite sheet.
func NewRenderer(sheet *gfx.Sheet) *Renderer {
	return &Renderer{sheet: sheet, scheme: gfx.AlleyScheme}
}

// Draw renders the alley. Freddy is drawn behind the fence or in front of
// it according to his depth (DESIGN §3.8, "Fence depth").
func (r *Renderer) Draw(dst *ebiten.Image, a *Alley) {
	r.build(a.Cans)
	dst.DrawImage(r.back, nil)
	for i := range a.Lines {
		r.drawLine(dst, &a.Lines[i])
	}
	f := a.Freddy
	if f.Depth == entity.Behind {
		r.drawFreddy(dst, f)
	}
	dst.DrawImage(r.front, nil)
	if f.Depth != entity.Behind {
		r.drawFreddy(dst, f)
	}
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

	for _, c := range clutter {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(c.x), float64(c.y))
		f.DrawImage(r.sheet.Sprite(c.name).Frame(0), op)
	}
	for _, c := range cans {
		r.drawCan(f, c)
	}
}

// drawCan draws a lidded trash can whose lid lines up with Freddy's feet
// when he stands on it.
func (r *Renderer) drawCan(dst *ebiten.Image, c config.Can) {
	s := r.scheme
	x := c.X + canInset
	lid := CanTop(c) + config.FreddyH             // top row of the lid
	fill(dst, x+canW/2-3, lid-2, 7, 1, s.Outline) // handle
	fill(dst, x+canW/2-3, lid-1, 1, 1, s.Outline)
	fill(dst, x+canW/2+3, lid-1, 1, 1, s.Outline)
	fill(dst, x-1, lid, canW+2, 3, s.Outline)
	fill(dst, x, lid, canW, 2, s.Metal)
	fill(dst, x, lid+3, canW, canFloorY-lid-3, s.Outline)
	fill(dst, x+1, lid+3, canW-2, canFloorY-lid-4, s.Metal)
	for y := lid + 7; y < canFloorY-2; y += 8 {
		fill(dst, x+1, y, canW-2, 2, s.Shade) // hoops
	}
	for sx := x + 5; sx < x+canW-2; sx += 6 {
		fill(dst, sx, lid+3, 1, canFloorY-lid-4, s.Shade) // ribs
	}
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

func (r *Renderer) freddyFrame(f *entity.Freddy) *ebiten.Image {
	switch {
	case f.Seat != entity.NotSeated:
		return r.sheet.Sprite("freddy_sit").Frame(0)
	case f.Arc.Dir == physics.Rising:
		return r.sheet.Sprite("freddy_jump").Frame(0)
	case f.Arc.Dir == physics.Falling:
		return r.sheet.Sprite("freddy_jump").Frame(1)
	case f.Dir != 0:
		return r.sheet.Sprite("freddy_walk").Frame(f.WalkFrame)
	}
	return r.sheet.Sprite("freddy_stand").Frame(0)
}

func (r *Renderer) drawFreddy(dst *ebiten.Image, f *entity.Freddy) {
	if f.Hidden() {
		return
	}
	img := r.freddyFrame(f)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	x := f.X
	if px, entering := f.EnterSlice(); entering {
		if px == 0 {
			return
		}
		x = px - w
		if f.Facing < 0 {
			x = gfx.ScreenWidth - px
		}
	}
	if f.Depth == entity.Behind {
		vis := config.FenceHideY - f.Y
		if vis <= 0 {
			return
		}
		if vis < h {
			img = img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+vis)).(*ebiten.Image)
		}
	}
	op := &ebiten.DrawImageOptions{}
	if f.Facing < 0 {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(w), 0)
	}
	op.GeoM.Translate(float64(x), float64(f.Y))
	dst.DrawImage(img, op)
}
