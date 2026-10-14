// Package game implements ebiten.Game: it renders the 320×200 logical
// screen offscreen and presents it with the selected aspect mode.
package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/testcard"
)

// Game is the top-level ebiten.Game.
type Game struct {
	screen   *ebiten.Image // the 320×200 logical screen
	prescale *ebiten.Image // integer upscale used for CRT presentation
	aspect   gfx.Aspect
	outW     int
	outH     int
}

// New returns a Game showing the M0 palette test card.
func New() *Game {
	g := &Game{
		screen: ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight),
		aspect: gfx.AspectSquare,
	}
	g.screen.WritePixels(testcard.Build().Pix)
	return g
}

// Update handles the window keys. Alt+Enter or F11 toggles fullscreen;
// A toggles the aspect mode (test card only, until the options menu exists).
func (g *Game) Update() error {
	alt := ebiten.IsKeyPressed(ebiten.KeyAlt)
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) || (alt && inpututil.IsKeyJustPressed(ebiten.KeyEnter)) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.aspect = g.aspect.Next()
	}
	return nil
}

// Draw presents the logical screen on the output surface.
func (g *Game) Draw(out *ebiten.Image) {
	out.Fill(color.Black)
	r := gfx.Fit(g.aspect, g.outW, g.outH)

	if g.aspect == gfx.AspectSquare {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(r.W/gfx.ScreenWidth, r.H/gfx.ScreenHeight)
		op.GeoM.Translate(r.X, r.Y)
		op.Filter = ebiten.FilterNearest
		out.DrawImage(g.screen, op)
	} else {
		// Upscale by an integer factor with nearest-neighbour, then resample
		// linearly to the non-integer 4:3 size so rows stay even and sharp.
		k := gfx.PrescaleFactor(r)
		w, h := gfx.ScreenWidth*k, gfx.ScreenHeight*k
		if g.prescale == nil || g.prescale.Bounds().Dx() != w || g.prescale.Bounds().Dy() != h {
			if g.prescale != nil {
				g.prescale.Deallocate()
			}
			g.prescale = ebiten.NewImage(w, h)
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(k), float64(k))
		op.Filter = ebiten.FilterNearest
		g.prescale.DrawImage(g.screen, op)

		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(r.W/float64(w), r.H/float64(h))
		op.GeoM.Translate(r.X, r.Y)
		op.Filter = ebiten.FilterLinear
		out.DrawImage(g.prescale, op)
	}

	ebitenutil.DebugPrint(out, fmt.Sprintf(
		"Aspect: %s (A to toggle)   Fullscreen: Alt+Enter / F11   Output: %dx%d",
		g.aspect, g.outW, g.outH))
}

// LayoutF makes the output surface match the window in device pixels, so
// scaling is done by Draw rather than by Ebitengine.
func (g *Game) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	s := ebiten.Monitor().DeviceScaleFactor()
	w, h := outsideWidth*s, outsideHeight*s
	g.outW, g.outH = int(w), int(h)
	return w, h
}

// Layout is required by ebiten.Game; LayoutF takes precedence.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	panic("unreachable: LayoutF is used")
}
