// Package game implements ebiten.Game: it runs the scenes on the master
// clock, handles the window keys, and presents the 320×200 logical screen
// with the selected aspect mode.
package game

import (
	"fmt"
	"image/color"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/assets"
	"github.com/comprofix/alleycat/internal/clock"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/scene"
)

// TPS is the game-loop rate.
const TPS = 60

// Game is the top-level ebiten.Game.
type Game struct {
	screen   *ebiten.Image // the 320×200 logical screen
	prescale *ebiten.Image // integer upscale used for CRT presentation
	aspect   gfx.Aspect
	outW     int
	outH     int

	clock  *clock.Clock
	ctx    *scene.Context
	scenes *scene.Manager
}

// New loads the assets and starts at the title screen. seed initialises the
// game's random number generator.
func New(seed uint64) (*Game, error) {
	f, err := assets.FS.Open("font/font8x8.txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	glyphs, err := gfx.ParseFont(f)
	if err != nil {
		return nil, fmt.Errorf("font: %w", err)
	}
	ui, err := gfx.LoadSheet(assets.FS, "gen/ui")
	if err != nil {
		return nil, fmt.Errorf("ui sheet: %w", err)
	}
	alleySheet, err := gfx.LoadSheet(assets.FS, "gen/alley")
	if err != nil {
		return nil, fmt.Errorf("alley sheet: %w", err)
	}
	roomSheet, err := gfx.LoadSheet(assets.FS, "gen/rooms")
	if err != nil {
		return nil, fmt.Errorf("rooms sheet: %w", err)
	}

	g := &Game{
		screen: ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight),
		aspect: gfx.AspectSquare,
		clock:  clock.New(TPS),
		ctx: &scene.Context{
			Input:   input.New(input.Ebiten{}),
			RNG:     rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
			Font:    gfx.NewFont(glyphs),
			UI:      ui,
			Alley:   alleySheet,
			Rooms:   roomSheet,
			SoundOn: true,
		},
	}
	g.scenes = scene.NewManager(scene.NewTitle())
	return g, nil
}

// Update reads input, handles the window keys, advances the master clock
// and runs the current scene.
func (g *Game) Update() error {
	in := g.ctx.Input
	in.Update()

	if in.JustPressed(input.Quit) {
		return ebiten.Termination
	}
	if in.JustPressed(input.Fullscreen) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if in.JustPressed(input.Aspect) {
		g.aspect = g.aspect.Next()
	}
	if in.JustPressed(input.Sound) {
		g.ctx.SoundOn = !g.ctx.SoundOn
	}

	g.ctx.Ticks, g.ctx.FreddySteps = g.clock.Advance()
	return g.scenes.Update(g.ctx)
}

// Draw renders the scene to the logical screen and presents it.
func (g *Game) Draw(out *ebiten.Image) {
	g.scenes.Draw(g.ctx, g.screen)

	out.Fill(color.Black)
	r := gfx.Fit(g.aspect, g.outW, g.outH)
	if g.aspect == gfx.AspectSquare {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(r.W/gfx.ScreenWidth, r.H/gfx.ScreenHeight)
		op.GeoM.Translate(r.X, r.Y)
		op.Filter = ebiten.FilterNearest
		out.DrawImage(g.screen, op)
		return
	}

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
