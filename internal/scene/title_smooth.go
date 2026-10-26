package scene

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/shapes"
)

// The title in the new look (DESIGN §2.1): the classic look's night scene
// and layout, drawn in code at four times the size.

// tk is a position on the 320×200 grid in the new look's screen pixels.
func tk(v int) float64 { return float64(v * entity.K) }

// The new look's lettering on the title: the name, as big as the classic
// look's three-times font, and the lines under it.
const (
	titleNameSize = 160
	titleLineSize = 40
)

// titleSmooth is what the title draws with in the new look, render caches.
type titleSmooth struct {
	back         *ebiten.Image
	name         *ebiten.Image // the name in white, to be laid in its bands
	line         *gfx.VectorFont
	freddy       entity.SmoothFreddy
	canvas       entity.Canvas
	nameX        int
	nameW, nameH int
	// nameTop and nameBottom are the first and last rows the letters fill.
	nameTop, nameBottom int
}

func newTitleSmooth() *titleSmooth {
	t := &titleSmooth{line: gfx.NewVectorFont(titleLineSize)}
	t.buildBack()
	big := gfx.NewVectorFont(titleNameSize)
	t.nameW, t.nameH = big.Width(titleName), big.LineHeight()
	t.nameX = big.CentreX(titleName)
	t.name = ebiten.NewImage(t.nameW, t.nameH)
	big.Draw(t.name, titleName, 0, 0, resurrect64.White)
	t.nameTop, t.nameBottom = inkRows(t.name)
	return t
}

// inkRows are the first and last rows of img with anything drawn in them.
func inkRows(img *ebiten.Image) (top, bottom int) {
	b := img.Bounds()
	px := make([]byte, 4*b.Dx()*b.Dy())
	img.ReadPixels(px)
	top = -1
	for y := range b.Dy() {
		for x := range b.Dx() {
			if px[4*(y*b.Dx()+x)+3] > 0 {
				if top < 0 {
					top = y
				}
				bottom = y
				break
			}
		}
	}
	return max(top, 0), bottom
}

// fadeRectS fills a rectangle with palette colour c at opacity a.
func fadeRectS(dst *ebiten.Image, x, y, w, h float64, c int, a float64) {
	col := gfx.RGBA(c)
	f := color.RGBA{
		R: uint8(float64(col.R) * a), G: uint8(float64(col.G) * a),
		B: uint8(float64(col.B) * a), A: uint8(255 * a),
	}
	vector.FillRect(dst, float32(x), float32(y), float32(w), float32(h), f, false)
}

// buildBack draws the night scene behind the title, where the classic look
// has it: the sky lightening towards the rooftops, stars, the moon, the
// town's rooftops with a few lamps lit, and the fence along the bottom.
func (t *titleSmooth) buildBack() {
	b := ebiten.NewImage(gfx.ScreenWidth*entity.K, gfx.ScreenHeight*entity.K)
	w := float64(b.Bounds().Dx())
	b.Fill(gfx.Palette[resurrect64.Ink])
	// The sky, from Ink at the top through Shadow to Navy.
	const steps = 32
	for i := range steps {
		y := tk(84) + float64(i)*tk(40)/steps
		fadeRectS(b, 0, y, w, tk(40)/steps+1, resurrect64.Shadow, float64(i+1)/steps)
	}
	shapes.FillRect(b, shapes.R(0, tk(124), w, tk(gfx.ScreenHeight-124)), resurrect64.Shadow)
	for i := range steps {
		y := tk(124) + float64(i)*tk(24)/steps
		fadeRectS(b, 0, y, w, tk(24)/steps+1, resurrect64.Navy, float64(i+1)/steps)
	}
	shapes.FillRect(b, shapes.R(0, tk(148), w, tk(gfx.ScreenHeight-148)), resurrect64.Navy)
	for i, s := range titleStars {
		x, y := tk(s[0])+2, tk(s[1])+2
		if i%3 == 0 {
			shapes.Line(b, x-4, y, x+4, y, 1, resurrect64.Fog)
			shapes.Line(b, x, y-4, x, y+4, 1, resurrect64.Fog)
			shapes.FillCircle(b, x, y, 2, resurrect64.White)
		} else {
			shapes.FillCircle(b, x, y, 1.5, resurrect64.Fog)
		}
	}
	alley.DrawMoon(b, tk(262)+20, tk(14)+20)

	// The rooftops: a chimney on each, the roofline catching the moon, and
	// lamps lit in some windows, as the classic look picks them.
	fence := tk(titleFenceY)
	for i, r := range titleRoofs {
		x, rw, top := tk(r.x), tk(r.w), tk(r.top)
		shapes.FillRect(b, shapes.R(x, top, rw, fence-top+8), resurrect64.Ink)
		cx := x + rw/3
		shapes.FillRect(b, shapes.R(cx, top-12, 8, 13), resurrect64.Ink)
		shapes.FillRect(b, shapes.R(cx-2, top-13, 12, 3), resurrect64.Ink)
		shapes.FillRect(b, shapes.R(x, top, rw, 3), resurrect64.Shadow)
		shapes.FillRect(b, shapes.R(cx-2, top-13, 12, 1.5), resurrect64.Shadow)
		for wy := r.top + 4; wy < titleFenceY-4; wy += 6 {
			for wx := r.x + 3; wx < r.x+r.w-3; wx += 6 {
				if (wx*7+wy*3+i)%5 == 0 {
					g := shapes.R(tk(wx), tk(wy)-2, tk(2), tk(2)+2)
					shapes.FillRoundRect(b, g, 1.5, resurrect64.Honey)
					shapes.FillRect(b, shapes.R(g.X, g.Y+g.H-4, g.W, 4), resurrect64.Amber)
				}
			}
		}
	}

	// The fence: boards a little out of true at the top, lit on the left
	// edge and shaded on the right, darkening towards the ground.
	bottom := float64(b.Bounds().Dy())
	for i, x := 0, 0.0; x < w; i, x = i+1, x+tk(8) {
		top := tk(titleFenceY + (i*37+i/3)%3)
		tilt := float64((i*13)%3-1) * 2
		bw := tk(7)
		shapes.FillRect(b, shapes.R(x+bw, top+8, tk(1), bottom-top-8), resurrect64.Ink)
		shapes.FillPoly(b, resurrect64.DeepTeal, x, top+3+tilt, x+3, top+tilt, x+bw-3, top-tilt, x+bw, top+3-tilt, x+bw, bottom, x, bottom)
		shapes.FillRect(b, shapes.R(x, top+4, 4, bottom-top-4), resurrect64.Teal)
		shapes.Line(b, x+1, top+3+tilt, x+bw-1, top+3-tilt, 3, resurrect64.Ink)
		shapes.FillRect(b, shapes.R(x+bw-3, top+4, 3, bottom-top-4), resurrect64.Pine)
	}
	for i := range 16 {
		fadeRectS(b, 0, bottom-tk(8)+float64(i)*tk(8)/16, w, tk(8)/16+1, resurrect64.Ink, float64(i+1)/24)
	}
	t.back = b
}

// draw draws the title in the new look.
func (t *titleSmooth) draw(dst *ebiten.Image, ti *Title) {
	dst.DrawImage(t.back, nil)
	t.drawCat(dst, &ti.cat)

	// The name: a deep drop shadow, an ink outline, then the letters in
	// three bands, light at the top, as the classic look's logo.
	x, y := float64(t.nameX), tk(titleLogoY)-float64(t.nameTop)
	lay := func(dx, dy float64, c, from, to int) {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(x+dx, y+dy+float64(from))
		op.ColorScale.ScaleWithColor(gfx.RGBA(c))
		dst.DrawImage(t.name.SubImage(image.Rect(0, from, t.nameW, to)).(*ebiten.Image), op)
	}
	lay(12, 12, resurrect64.Grape, 0, t.nameH)
	for _, d := range [12][2]float64{{-5, 0}, {5, 0}, {0, -5}, {0, 5}, {-4, -3}, {4, -3}, {-4, 3}, {4, 3}, {-3, -4}, {3, -4}, {-3, 4}, {3, 4}} {
		lay(d[0], d[1], resurrect64.Ink, 0, t.nameH)
	}
	h := t.nameBottom - t.nameTop + 1
	b1, b2 := t.nameTop+h*3/10, t.nameTop+h*7/10
	lay(0, 0, resurrect64.Honey, 0, b1)
	lay(0, 0, resurrect64.Coral, b1, b2)
	lay(0, 0, resurrect64.Copper, b2, t.nameH)

	f := t.line
	f.DrawShadowed(dst, titleLine, f.CentreX(titleLine), int(tk(titleLogoY+34)), resurrect64.Fog, resurrect64.Ink)
	py, px := int(tk(titlePromptY)), f.CentreX(titlePrompt)
	f.DrawShadowed(dst, titlePrompt, px, py, resurrect64.White, resurrect64.Ink)
	if ti.blink.Frame() == 0 { // the pointer, blinking
		cx, cy := float64(px-28), float64(py+f.LineHeight()/2)
		shapes.FillPoly(dst, resurrect64.Ink, cx-6, cy-10, cx+10, cy+2, cx-6, cy+14)
		shapes.FillPoly(dst, resurrect64.White, cx-8, cy-12, cx+8, cy, cx-8, cy+12)
	}
}

// drawCat draws Freddy on the fence top, walking or sat looking out, with
// the Lilac rim the classic look gives him here to lift him off the dark
// rooftops.
func (t *titleSmooth) drawCat(dst *ebiten.Image, c *titleCat) {
	if c.dir != 0 {
		t.freddy.DrawCat(dst, c.x, titleFenceY-config.FreddyH+1, c.facing < 0, entity.CatWalking, c.frame, entity.FreddyCoat, resurrect64.Lilac)
		return
	}
	const sitH = 14
	pic := entity.Picture{X: c.x, Y: titleFenceY - sitH + 1, W: 24, H: sitH, Flip: c.facing < 0, Outline: resurrect64.Lilac}
	t.canvas.Draw(dst, pic, func(p entity.Pen) { entity.DrawSitting(p, entity.FreddyCoat) })
}
