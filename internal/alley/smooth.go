package alley

import (
	"fmt"
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/session"
	"github.com/comprofix/alleycat/internal/shapes"
)

// K is how many of the new look's screen pixels each of the alley's
// 320×200 pixels covers. The layout and the game stay on the 320×200 grid
// (DESIGN §2.1); only the drawing is finer.
const K = 4

// k is a position on the 320×200 grid in the new look's screen pixels.
func k(v int) float64 { return float64(v * K) }

// letterPitch is the lettering's character spacing on the fence: the 8×8
// font's, so the HUD lines up as in the classic look.
const letterPitch = gfx.GlyphSize * K

// SmoothRenderer draws the alley in the new look (DESIGN §2.1): the
// backdrop, the fence and its lettering drawn in code from shapes, and
// whatever has no new art yet drawn from the pixel sprites, enlarged, in
// its place among them. It caches the backdrop as the pixel Renderer does
// and holds no game state.
type SmoothRenderer struct {
	pixel       *Renderer       // the pixel art still in use
	text        *gfx.VectorFont // the lettering on the fence
	scheme      gfx.SceneScheme // the alley's colours, as the classic look
	back, front *ebiten.Image   // the backdrop, behind and in front of Freddy
	mid, top    *ebiten.Image   // 320×200 layers for the pixel art
	builtFor    []config.Can
}

// NewSmoothRenderer returns the new look's renderer, using the alley
// sprite sheet and the 8×8 font for the art not yet redrawn.
func NewSmoothRenderer(sheet *gfx.Sheet, font *gfx.PixelFont) *SmoothRenderer {
	return &SmoothRenderer{
		pixel:  NewRenderer(sheet, font),
		text:   gfx.NewVectorFont(40),
		scheme: gfx.AlleyScheme,
	}
}

// Draw renders the alley, as Renderer.Draw does, onto a 1280×800 dst.
func (r *SmoothRenderer) Draw(dst *ebiten.Image, a *Alley, high int) {
	r.build(a.Cans)
	p := r.pixel
	dst.DrawImage(r.back, nil)
	r.drawFlicker(dst, a.Tick)
	f := a.Freddy
	r.layer(dst, r.mid, func(l *ebiten.Image) {
		p.drawWindow(l, &a.Window)
		for i := range a.Lines {
			p.drawLine(l, &a.Lines[i])
		}
		p.drawMice(l, a)
		p.drawThrown(l, &a.Thrown)
		if f.Depth == entity.Behind {
			entity.DrawFreddyRimmed(l, p.sheet, f, p.scheme.Shade)
		}
	})
	dst.DrawImage(r.front, nil)
	r.drawHUD(dst, a, high)
	r.layer(dst, r.top, func(l *ebiten.Image) {
		p.drawTrashCat(l, &a.TrashCat)
		entity.DrawDog(l, p.sheet, &a.Dog, "dog_run")
		if f.Depth != entity.Behind {
			entity.DrawFreddyRimmed(l, p.sheet, f, p.scheme.Shade)
		}
		p.drawImpact(l, a.Impact)
	})
}

// layer draws pixel art on the 320×200 layer l and lays it over dst,
// enlarged with its pixels kept sharp.
func (r *SmoothRenderer) layer(dst, l *ebiten.Image, draw func(*ebiten.Image)) {
	l.Clear()
	draw(l)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(K, K)
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(l, op)
}

func (r *SmoothRenderer) build(cans []config.Can) {
	if r.back != nil && slices.Equal(r.builtFor, cans) {
		return
	}
	r.builtFor = slices.Clone(cans)
	if r.back == nil {
		w, h := gfx.ScreenWidth*K, gfx.ScreenHeight*K
		r.back, r.front = ebiten.NewImage(w, h), ebiten.NewImage(w, h)
		r.mid = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.top = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
	}
	r.buildBack()
	r.buildFront(cans)
}

// faded is palette colour c at opacity a (0–1), for soft shading.
func faded(c int, a float64) color.RGBA {
	col := gfx.RGBA(c)
	return color.RGBA{
		R: uint8(float64(col.R) * a), G: uint8(float64(col.G) * a),
		B: uint8(float64(col.B) * a), A: uint8(255 * a),
	}
}

// fadeRect fills a rectangle with palette colour c at opacity a.
func fadeRect(dst *ebiten.Image, x, y, w, h float64, c int, a float64) {
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), faded(c, a), true)
}

// blend softens the edge at row y between two bands of colour: each
// colour fades into the other over d rows either side.
func blend(dst *ebiten.Image, y, w, d float64, above, below int) {
	for i := range int(d) {
		a := 0.5 * (1 - float64(i)/d)
		fadeRect(dst, 0, y-1-float64(i), w, 1, below, a)
		fadeRect(dst, 0, y+float64(i), w, 1, above, a)
	}
}

// ellipse is n+1 points round the ellipse with radii rx and ry about
// (cx, cy), from angle a0 to a1, as x, y pairs for shapes.FillPoly.
func ellipse(cx, cy, rx, ry, a0, a1 float64, n int) []float64 {
	pts := make([]float64, 0, 2*(n+1))
	for i := range n + 1 {
		t := a0 + (a1-a0)*float64(i)/float64(n)
		pts = append(pts, cx+rx*math.Cos(t), cy+ry*math.Sin(t))
	}
	return pts
}

// path is a polyline through the x, y pairs, for shapes.StrokePath.
func path(pts []float64) *vector.Path {
	var p vector.Path
	p.MoveTo(float32(pts[0]), float32(pts[1]))
	for i := 2; i+1 < len(pts); i += 2 {
		p.LineTo(float32(pts[i]), float32(pts[i+1]))
	}
	return &p
}

// buildBack draws the night sky, the moon and the far rooftops, the
// building's cornice and brick wall, its drainpipes and the 3×4 windows,
// all where the classic look has them.
func (r *SmoothRenderer) buildBack() {
	s, b := r.scheme, r.back
	w := float64(b.Bounds().Dx())
	b.Fill(gfx.Palette[s.WallShade])

	// The sky, darkest at the top, its bands blended into each other.
	from := 0.0
	for _, band := range skyBands {
		shapes.FillRect(b, shapes.R(0, from, w, k(band.to)-from), band.idx)
		from = k(band.to)
	}
	for i := 1; i < len(skyBands); i++ {
		blend(b, k(skyBands[i-1].to), w, 6, skyBands[i-1].idx, skyBands[i].idx)
	}
	for i, st := range stars {
		x, y := k(st[0])+2, k(st[1])+2
		if i%3 == 0 {
			shapes.Line(b, x-4, y, x+4, y, 1, resurrect64.Fog)
			shapes.Line(b, x, y-4, x, y+4, 1, resurrect64.Fog)
			shapes.FillCircle(b, x, y, 2, resurrect64.White)
		} else {
			shapes.FillCircle(b, x, y, 1.5, resurrect64.Fog)
		}
	}
	r.drawMoon(b, k(268)+20, k(1)+20)

	// The far rooftops: chimneys, a ledge on some, an aerial on others,
	// and a lamp or two in their windows.
	for i, sk := range skyline {
		x, sw, top := k(sk.x), k(sk.w), k(sk.top)
		shapes.FillRect(b, shapes.R(x, top, sw, k(skyBottom)-top+1), resurrect64.Ink)
		cx := x + sw/3
		shapes.FillRect(b, shapes.R(cx, top-7, 7, 8), resurrect64.Ink)
		shapes.FillRect(b, shapes.R(cx-1.5, top-8, 10, 2.5), resurrect64.Ink)
		switch i % 3 {
		case 0:
			shapes.FillRect(b, shapes.R(x-2, top-2, sw+4, 3), resurrect64.Ink)
		case 2:
			ax := x + sw*3/4
			shapes.Line(b, ax, top, ax, top-14, 1.5, resurrect64.Ink)
			shapes.Line(b, ax-5, top-10, ax+5, top-10, 1.5, resurrect64.Ink)
		}
		if i%2 == 0 {
			shapes.FillRoundRect(b, shapes.R(x+sw/2, top+7, 4, 5), 1, resurrect64.Honey)
		}
		if i%3 == 1 {
			shapes.FillRoundRect(b, shapes.R(x+11, top+11, 4, 5), 1, resurrect64.Coral)
		}
	}
	// A water tower on the far roofs: a tank with a conical top on legs.
	shapes.FillRoundRect(b, shapes.R(k(150), k(3), k(9), k(4)), 3, resurrect64.Ink)
	shapes.FillPoly(b, resurrect64.Ink, k(150)-2, k(3)+1, k(150)+k(9)/2, k(1)-1, k(159)+2, k(3)+1)
	for _, lx := range []float64{k(151), k(155) - 2, k(158)} {
		shapes.Line(b, lx, k(7), lx+(k(155)-lx)*0.15, k(skyBottom), 2, resurrect64.Ink)
	}
	shapes.Line(b, k(151), k(9), k(158), k(12), 1.5, resurrect64.Ink)
	shapes.Line(b, k(158), k(9), k(151), k(12), 1.5, resurrect64.Ink)

	// The cornice along our building's roof: grey stone lit along the
	// top, its blocks picked out, casting a soft shadow on the wall.
	top, bot := k(skyBottom), k(wallTop)
	shapes.FillRect(b, shapes.R(0, top, w, bot-top), resurrect64.Dusk)
	shapes.FillRect(b, shapes.R(0, top, w, 3), s.Shade)
	for x := 8.0; x < w; x += 32 {
		shapes.FillRoundRect(b, shapes.R(x, top+7, 16, 5), 1.5, resurrect64.Shadow)
	}

	// Brick courses with mortar between, the joints staggered; each brick
	// lit along its top, a few weathered or chipped.
	courseH := k(6)
	for y, course := bot, 0; y < k(fenceTop); y, course = y+courseH, course+1 {
		for x := -k(8) * float64(course%2); x < w; x += k(16) {
			br := shapes.R(x+2, y+2, k(16)-4, courseH-4)
			h := uint32(course*977+int(x)*131) * 0x9e3779b1 >> 24
			c := s.Wall
			if h%7 == 0 {
				c = resurrect64.Mauve
			}
			shapes.FillRoundRect(b, br, 3, c)
			shapes.Line(b, br.X+4, br.Y+2.5, br.X+br.W*0.45, br.Y+2.5, 2, s.WallLight)
			if h%11 == 3 {
				shapes.FillCircle(b, br.X+br.W*0.7, br.Y+br.H*0.6, 2.5, s.WallShade)
			}
		}
	}
	fadeRect(b, 0, bot, w, 4, resurrect64.Ink, 0.45)
	fadeRect(b, 0, bot+4, w, 6, resurrect64.Ink, 0.2)

	// Drainpipes down both edges, rounded by a lit stripe and a shaded
	// side, with brackets and joints.
	for _, px := range []float64{k(3), k(gfx.ScreenWidth - 7)} {
		pipe := shapes.R(px, bot, k(4), k(fenceTop)-bot)
		shapes.FillRect(b, pipe, resurrect64.Dusk)
		shapes.FillRect(b, shapes.R(px+3, bot, 3, pipe.H), s.Shade)
		shapes.FillRect(b, shapes.R(px+pipe.W-4, bot, 4, pipe.H), resurrect64.Shadow)
		for y := bot + k(10); y < k(fenceTop); y += k(24) {
			shapes.FillRoundRect(b, shapes.R(px-6, y, pipe.W+12, 6), 2, s.Shade)
			shapes.FillRect(b, shapes.R(px-6, y+4, pipe.W+12, 2), resurrect64.Shadow)
		}
	}

	for row := range 3 {
		for col := range 4 {
			r.buildWindow(b, windowLeft+col*windowPitchX, windowTop+row*windowPitchY, litWindows[row*4+col])
		}
	}
}

// drawMoon draws the moon round (cx, cy): a faint glow, the lit face
// shaded off to the right, and two craters.
func (r *SmoothRenderer) drawMoon(dst *ebiten.Image, cx, cy float64) {
	vector.FillCircle(dst, float32(cx), float32(cy), 34, faded(resurrect64.Fog, 0.08), true)
	vector.FillCircle(dst, float32(cx), float32(cy), 26, faded(resurrect64.Fog, 0.1), true)
	shapes.FillCircle(dst, cx, cy, 20, resurrect64.Fog)
	shapes.FillCircle(dst, cx-2, cy, 18, resurrect64.Mist)
	shapes.FillCircle(dst, cx-5, cy-1, 15, resurrect64.White)
	for _, c := range [][3]float64{{-7, 5, 3.5}, {3, -8, 2.5}, {6, 7, 2}} {
		shapes.FillCircle(dst, cx+c[0], cy+c[1], c[2], resurrect64.Fog)
		shapes.FillCircle(dst, cx+c[0]-0.8, cy+c[1]-0.8, c[2]*0.7, resurrect64.Mist)
	}
}

// buildWindow draws one closed window at (x, y) on the grid: a stone
// lintel with a keystone, the frame and mullion, the glass, dark or lit by
// a lamp inside, and a sill casting its shadow.
func (r *SmoothRenderer) buildWindow(b *ebiten.Image, x, y int, lit bool) {
	s := r.scheme
	wx, wy, ww, wh := k(x), k(y), k(windowW), k(windowH)
	lintel := shapes.R(wx-8, wy-12, ww+16, 9)
	shapes.FillRoundRect(b, lintel, 2, resurrect64.Dusk)
	shapes.FillRect(b, shapes.R(lintel.X+2, lintel.Y, lintel.W-4, 2), s.Shade)
	key := shapes.R(wx+ww/2-9, wy-15, 18, 13)
	shapes.FillRoundRect(b, key, 2, resurrect64.Dusk)
	shapes.FillRect(b, shapes.R(key.X+2, key.Y, key.W-4, 2), s.Shade)

	shapes.FillRoundRect(b, shapes.R(wx, wy, ww, wh), 3, s.Trim)
	shapes.FillRect(b, shapes.R(wx+7, wy+4, ww-14, wh-8), s.Outline)
	glass := shapes.R(wx+11, wy+8, ww-22, wh-16)
	if lit {
		r.drawLamp(b, glass, resurrect64.Honey, resurrect64.Amber)
	} else {
		shapes.FillRect(b, glass, s.Glass)
		fadeRect(b, glass.X, glass.Y, glass.W, glass.H/3, resurrect64.Indigo, 0.5)
		for _, gx := range []float64{glass.X + 14, glass.X + 90} {
			shapes.Line(b, gx, glass.Y+30, gx+22, glass.Y+8, 3, s.Glint)
			shapes.Line(b, gx+12, glass.Y+30, gx+24, glass.Y+18, 2, s.Glint)
		}
	}
	shapes.FillRect(b, shapes.R(wx+ww/2-4, wy+4, 8, wh-8), s.Trim)

	sill := shapes.R(wx-12, wy+wh, ww+24, 8)
	shapes.FillRoundRect(b, sill, 2, s.Shade)
	shapes.FillRect(b, shapes.R(sill.X+2, sill.Y, sill.W-4, 2), s.Trim)
	fadeRect(b, sill.X+4, sill.Y+sill.H, sill.W-8, 5, resurrect64.Ink, 0.4)
}

// drawLamp draws a lit window's glass: the lamplight warming to a glow at
// the bottom, and curtains drawn back to either side.
func (r *SmoothRenderer) drawLamp(dst *ebiten.Image, g shapes.Rect, light, glow int) {
	shapes.FillRect(dst, g, light)
	fadeRect(dst, g.X, g.Y+g.H-22, g.W, 22, glow, 0.5)
	fadeRect(dst, g.X, g.Y+g.H-12, g.W, 12, glow, 0.6)
	for _, side := range []float64{1, -1} {
		x0 := g.X
		if side < 0 {
			x0 = g.X + g.W
		}
		in := func(d float64) float64 { return x0 + side*d }
		// A curtain gathered at a tie halfway down.
		shapes.FillPoly(dst, resurrect64.Raspberry,
			in(0), g.Y, in(22), g.Y, in(14), g.Y+g.H*0.55, in(18), g.Y+g.H, in(0), g.Y+g.H)
		shapes.Line(dst, in(8), g.Y+2, in(6), g.Y+g.H*0.5, 2, resurrect64.Berry)
		shapes.Line(dst, in(6), g.Y+g.H*0.6, in(9), g.Y+g.H-1, 2, resurrect64.Berry)
		shapes.FillCircle(dst, in(13), g.Y+g.H*0.55, 2.5, resurrect64.Honey)
	}
	shapes.FillRect(dst, shapes.R(g.X, g.Y, g.W, 3), resurrect64.Berry) // the curtain rail's pelmet
}

// drawFlicker dims the lit windows whose lamps are flickering this tick,
// as the classic look does.
func (r *SmoothRenderer) drawFlicker(dst *ebiten.Image, tick int) {
	for i := range litWindows {
		if flickering(tick, i) {
			x, y := k(windowLeft+i%4*windowPitchX), k(windowTop+i/4*windowPitchY)
			r.drawLamp(dst, shapes.R(x+11, y+8, k(windowW)-22, k(windowH)-16), resurrect64.Caramel, resurrect64.Copper)
			shapes.FillRect(dst, shapes.R(x+k(windowW)/2-4, y+4, 8, k(windowH)-8), r.scheme.Trim)
		}
	}
}

// buildFront draws the fence, the ground, the litter and the trash cans.
func (r *SmoothRenderer) buildFront(cans []config.Can) {
	s, f := r.scheme, r.front
	w := float64(f.Bounds().Dx())
	f.Clear()

	// Weathered boards, each a little out of true at the top: a lit left
	// edge, grain and a knot, darkening towards the ground.
	bottom := k(fenceBottom + 1)
	for i, x := 0, 0.0; x < w; i, x = i+1, x+k(8) {
		top := k(fenceTop + (i*37+i/3)%4)
		tilt := float64((i*13)%3-1) * 2
		bw := k(7)
		shapes.FillRect(f, shapes.R(x+bw, top+8, k(1), bottom-top-8), s.Gap)
		shapes.FillPoly(f, s.Fence, x, top+3+tilt, x+3, top+tilt, x+bw-3, top-tilt, x+bw, top+3-tilt, x+bw, bottom, x, bottom)
		shapes.FillRect(f, shapes.R(x, top+4, 4, bottom-top-4), s.FenceLight)
		shapes.Line(f, x+1, top+3+tilt, x+bw-1, top+3-tilt, 3, s.Gap) // the cut top
		shapes.FillRect(f, shapes.R(x+bw-3, top+4, 3, bottom-top-4), resurrect64.Pine)
		g := float64((i * 7) % 9)
		shapes.StrokePath(f, path([]float64{x + 12, top + 18 + g, x + 14, top + 70 + g, x + 11, top + 140, x + 13, bottom - 20}), 1.5, resurrect64.Pine)
		shapes.StrokePath(f, path([]float64{x + 20, top + 40, x + 19, top + 110 + 2*g, x + 21, bottom - 40}), 1.5, resurrect64.Pine)
		ky := top + k((i*11)%40+14)
		shapes.FillPoly(f, resurrect64.Pine, ellipse(x+k(2)+4, ky+2, 5, 3, 0, 2*math.Pi, 16)...)
		shapes.FillCircle(f, x+k(2)+4, ky+2, 1.5, s.Gap)
	}
	for i := range 16 {
		fadeRect(f, 0, bottom-16+float64(i), w, 1, s.Gap, float64(i+1)/18)
	}
	// The rails behind the boards show as shadows, with a nail in each
	// board.
	for _, ry := range []int{128, 160} {
		y := k(ry)
		fadeRect(f, 0, y, w, 4, s.Gap, 0.8)
		fadeRect(f, 0, y+4, w, 3, s.Gap, 0.35)
		for x := k(3) + 2; x < w; x += k(8) {
			shapes.FillCircle(f, x, y-6, 2.5, s.Outline)
			shapes.FillCircle(f, x-0.7, y-6.7, 1, resurrect64.Fog)
		}
	}

	// The ground: cobbles bedded in dark mortar, lit along their tops, with
	// a lit edge along the fence.
	gy := k(groundTop)
	shapes.FillRect(f, shapes.R(0, gy, w, float64(f.Bounds().Dy())-gy), s.Outline)
	for y, row := gy+4, 0; y < float64(f.Bounds().Dy()); y, row = y+k(6), row+1 {
		for x := -k(7) * float64(row%2); x < w; x += k(14) {
			c := shapes.R(x+2, y+2, k(14)-4, k(6)-4)
			shapes.FillRoundRect(f, c, 8, s.GroundLight)
			shapes.FillRoundRect(f, c.Offset(0, 3), 8, s.Ground)
			h := (row*7 + int(x)) % 5
			shapes.FillCircle(f, c.X+c.W*(0.3+0.1*float64(h)), c.Y+c.H*0.6, 1.5, s.GroundLight)
		}
	}
	shapes.FillRect(f, shapes.R(0, gy, w, 4), s.Shade)
	fadeRect(f, 0, gy+4, w, 4, s.Outline, 0.5)

	r.drawGraffiti(f)
	r.drawFishbones(f, k(clutter[0].x), k(clutter[0].y))
	r.drawBone(f, k(clutter[1].x), k(clutter[1].y))
	r.drawTinCan(f, k(clutter[2].x), k(clutter[2].y))
	for _, c := range cans {
		r.drawCan(f, c)
	}
}

// drawFishbones draws a fish skeleton lying with its head to the right,
// in the 14×5 space the classic look gives it at (x, y).
func (r *SmoothRenderer) drawFishbones(dst *ebiten.Image, x, y float64) {
	c := resurrect64.Fog
	mid := y + 10
	shapes.Line(dst, x+8, mid, x+44, mid, 3, c)
	shapes.FillPoly(dst, c, x+8, mid, x, y+2, x+4, mid, x, y+18)
	for i := range 4 {
		rx := x + 16 + float64(i)*7
		shapes.Line(dst, rx, y+3, rx+3, mid, 2, c)
		shapes.Line(dst, rx, y+17, rx+3, mid, 2, c)
	}
	shapes.FillPoly(dst, c, x+42, y+3, x+52, y+6, x+55, mid, x+52, y+14, x+42, y+17)
	shapes.FillCircle(dst, x+49, y+8, 1.8, resurrect64.Ink)
}

// drawBone draws a bone in the 12×5 space at (x, y).
func (r *SmoothRenderer) drawBone(dst *ebiten.Image, x, y float64) {
	shapes.FillRoundRect(dst, shapes.R(x+7, y+6, 34, 8), 4, resurrect64.Mist)
	for _, e := range [][2]float64{{x + 6, y + 5}, {x + 6, y + 15}, {x + 42, y + 5}, {x + 42, y + 15}} {
		shapes.FillCircle(dst, e[0], e[1], 5, resurrect64.Mist)
	}
	shapes.FillRoundRect(dst, shapes.R(x+7, y+6, 34, 5), 3, resurrect64.White)
	shapes.FillCircle(dst, x+5, y+4, 3.5, resurrect64.White)
	shapes.FillCircle(dst, x+41, y+4, 3.5, resurrect64.White)
}

// drawTinCan draws a tin can lying on its side, its open end to the right,
// in the 10×5 space at (x, y).
func (r *SmoothRenderer) drawTinCan(dst *ebiten.Image, x, y float64) {
	body := shapes.R(x+2, y+2, 32, 16)
	shapes.FillRoundRect(dst, body, 3, resurrect64.Fog)
	shapes.FillRect(dst, shapes.R(body.X+2, body.Y+2, body.W-2, 4), resurrect64.Mist)
	shapes.FillRect(dst, shapes.R(body.X+2, body.Y+11, body.W-2, 5), resurrect64.Dusk)
	shapes.FillRect(dst, shapes.R(body.X+10, body.Y, 12, body.H), resurrect64.Blue)
	shapes.FillRect(dst, shapes.R(body.X+10, body.Y+11, 12, 5), resurrect64.Cobalt)
	shapes.FillPoly(dst, resurrect64.Lilac, ellipse(x+34, y+10, 5, 8.5, 0, 2*math.Pi, 20)...)
	shapes.FillPoly(dst, resurrect64.Ink, ellipse(x+34.5, y+10, 3, 6.5, 0, 2*math.Pi, 20)...)
}

// canStripes are the trash can's metal shaded round its body as the classic
// sprite is, from its left edge to its right, in 30ths of its width.
var canStripes = []struct {
	to  int
	idx int
}{
	{3, resurrect64.Lilac}, {4, resurrect64.Fog}, {7, resurrect64.Mist}, {10, resurrect64.Fog},
	{17, resurrect64.Lilac}, {22, resurrect64.Dusk}, {30, resurrect64.Shadow},
}

// drawCan draws an open trash can where the classic look's sprite sits,
// its rim level with Freddy's feet when he stands on it: a round metal
// body with raised hoops, and rubbish showing in its mouth.
func (r *SmoothRenderer) drawCan(dst *ebiten.Image, c config.Can) {
	h := 39
	if c.Tall {
		h = 47
	}
	x, y := k(c.X+canSpriteDX), k(CanTop(c)+canSpriteDY)
	cw := k(30)
	cx := x + cw/2
	rimY := y + 30
	rx, ry := cw/2, 14.0
	baseY := y + k(h) - ry

	// The body: a stripe of each shade, each cut to the curve of the base.
	outline := append([]float64{x, rimY}, ellipse(cx, baseY, rx, ry, math.Pi, 0, 32)...)
	outline = append(outline, x+cw, rimY)
	shapes.FillPoly(dst, resurrect64.Ink, grow(outline, cx, 3)...)
	from := 0
	for _, st := range canStripes {
		x0, x1 := x+cw*float64(from)/30, x+cw*float64(st.to)/30
		shapes.FillPoly(dst, st.idx, stripe(x0, x1, rimY, cx, baseY, rx, ry)...)
		from = st.to
	}
	// Raised hoops, curving round the body: a dark groove over a lit rib.
	for row := 12; row < h-6; row += 9 {
		hy := y + k(row) - 6
		shapes.StrokePath(dst, path(ellipse(cx, hy, rx-1, ry, math.Pi, 0, 32)), 4, resurrect64.Ink)
		shapes.StrokePath(dst, path(ellipse(cx, hy+4, rx-2, ry, math.Pi*0.97, math.Pi*0.03, 32)), 2.5, resurrect64.Mist)
	}
	// The mouth: a lit rim round the dark inside, with rubbish in it.
	shapes.FillPoly(dst, resurrect64.Ink, ellipse(cx, rimY, rx+2, ry+2, 0, 2*math.Pi, 40)...)
	shapes.FillPoly(dst, resurrect64.Fog, ellipse(cx, rimY, rx, ry, 0, 2*math.Pi, 40)...)
	shapes.FillPoly(dst, resurrect64.Mist, ellipse(cx, rimY-1.5, rx-3, ry-2.5, math.Pi, 2*math.Pi, 24)...)
	shapes.FillPoly(dst, resurrect64.Shadow, ellipse(cx, rimY+1, rx-6, ry-5, 0, 2*math.Pi, 40)...)
	shapes.FillPoly(dst, resurrect64.Emerald, x+30, rimY-1, x+44, rimY-7, x+47, rimY-3, x+33, rimY+3)
	shapes.FillPoly(dst, resurrect64.Coral, ellipse(cx+18, rimY+2, 10, 4, 0, 2*math.Pi, 16)...)
	shapes.FillPoly(dst, resurrect64.Coral, cx+27, rimY+2, cx+33, rimY-2, cx+33, rimY+6)
	shapes.FillPoly(dst, resurrect64.Mist, cx-10, rimY+4, cx-2, rimY-3, cx+4, rimY+5)
}

// stripe is the outline of the can's body between x0 and x1: from the rim
// down to the base, whose curve it follows.
func stripe(x0, x1, top, cx, baseY, rx, ry float64) []float64 {
	pts := []float64{x0, top, x1, top}
	const n = 8
	for i := range n + 1 {
		x := x1 + (x0-x1)*float64(i)/n
		d := (x - cx) / rx
		pts = append(pts, x, baseY+ry*math.Sqrt(max(0, 1-d*d)))
	}
	return pts
}

// grow is the polygon pts pushed d pixels further from x = cx and down,
// for an outline round it.
func grow(pts []float64, cx, d float64) []float64 {
	out := slices.Clone(pts)
	for i := 0; i+1 < len(out); i += 2 {
		if out[i] < cx {
			out[i] -= d
		} else {
			out[i] += d
		}
		if i > 1 && i < len(out)-2 {
			out[i+1] += d
		}
	}
	return out
}

// drawText writes s on the fence with each character in an 8×8 cell of
// the grid, as the classic look's font, so the lettering stays where it
// is: col in front of a shadow in shadow.
func (r *SmoothRenderer) drawText(dst *ebiten.Image, s string, x, y, col, shadow int, drop float64) {
	f := r.text
	for i, ch := range s {
		cx := k(x) + float64(i*letterPitch) + float64(letterPitch-f.CellWidth())/2
		cy := k(y) - 6
		f.Draw(dst, string(ch), int(cx+drop), int(cy+drop), shadow)
		f.Draw(dst, string(ch), int(cx), int(cy), col)
	}
}

// drawGraffiti sprays the decorative words onto the fence, with paint
// running down from some letters.
func (r *SmoothRenderer) drawGraffiti(dst *ebiten.Image) {
	for gi, g := range graffiti {
		r.drawText(dst, g.text, g.x, g.y, resurrect64.Coral, r.scheme.Outline, 3)
		for i := range len(g.text) {
			if (i+gi)%3 == 1 {
				dx := k(g.x) + float64(i*letterPitch) + 14
				top := k(g.y+gfx.GlyphSize) - 8
				l := k(3 + (i+gi)%3)
				shapes.Line(dst, dx, top, dx, top+l, 2.5, resurrect64.Coral)
				shapes.FillCircle(dst, dx, top+l, 3, resurrect64.Coral)
			}
		}
	}
}

// drawHUD writes the scores, lives and skill on the fence where the
// classic look does: chalk white with a dark shadow.
func (r *SmoothRenderer) drawHUD(dst *ebiten.Image, a *Alley, high int) {
	s := a.Session
	write := func(text string, x, y int) {
		r.drawText(dst, text, x, y, resurrect64.White, r.scheme.Outline, 3)
	}
	write("HI!", hudHighX-gfx.Width("HI! "), hudHighY)
	write(session.FormatScore(high), hudHighX, hudHighY)
	write(fmt.Sprintf("%d CAT", s.Lives), hudLivesX, hudLivesY)
	write(session.FormatScore(s.Score), hudScoreX, hudScoreY)
	write(string(s.Skill.Letter()), hudSkillX, hudSkillY)
}
