// Package game implements ebiten.Game: it runs the scenes on the master
// clock, handles the window keys, and presents the 320×200 logical screen
// with the selected aspect mode.
package game

import (
	"fmt"
	"image/color"
	"log"
	"math/rand/v2"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	eaudio "github.com/hajimehoshi/ebiten/v2/audio"

	"github.com/comprofix/alleycat/assets"
	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/clock"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/scene"
	"github.com/comprofix/alleycat/internal/score"
)

// Game is the top-level ebiten.Game.
type Game struct {
	preset   config.Preset
	screen   *ebiten.Image // the internal screen
	prescale *ebiten.Image // whole-number upscale used for CRT presentation
	outW     int
	outH     int

	clock  *clock.Clock
	ctx    *scene.Context
	scenes *scene.Manager

	speaker *eaudio.Player // kept alive while the game runs

	// The saved files (DESIGN §9).
	dir       string // where they live; empty saves nothing
	settings  config.Settings
	savedHigh int
}

// Saved is what was loaded from the saved files, and where they live
// (DESIGN §9). An empty Dir saves nothing.
type Saved struct {
	Settings config.Settings
	High     int
	Dir      string
}

// New loads the assets and starts at the title screen. seed initialises the
// game's random number generator.
func New(seed uint64, saved Saved) (*Game, error) {
	font, err := gfx.LoadFont(assets.FS, "font/font8x8.txt")
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

	synth := audio.NewSynth()
	speaker, err := audio.Start(synth)
	if err != nil {
		return nil, fmt.Errorf("audio: %w", err)
	}

	set := saved.Settings
	p := config.Project
	g := &Game{
		preset:    p,
		screen:    ebiten.NewImage(p.ScreenWidth, p.ScreenHeight),
		clock:     clock.New(config.TPS),
		speaker:   speaker,
		dir:       saved.Dir,
		settings:  set,
		savedHigh: saved.High,
		ctx: &scene.Context{
			Input:     input.New(&input.Ebiten{}),
			RNG:       rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)),
			Fonts:     gfx.NewFonts(font),
			PixelFont: gfx.NewPixelFont(font),
			UI:        ui,
			Alley:     alleySheet,
			Rooms:     roomSheet,
			Audio:     synth,
			// The patter and the effects have their own RNGs so audio never
			// shifts the game's random sequence (DESIGN §7.2).
			Patter:    audio.NewPatter(rand.New(rand.NewPCG(seed^0x5eed0a0d10, seed))),
			FX:        audio.NewFX(rand.New(rand.NewPCG(seed^0xeffec75, seed+1))),
			Sound:     scene.SoundSettings{On: set.SoundOn, Muted: set.Muted, Volume: set.Volume, Soften: set.Soften},
			HighScore: saved.High,
		},
	}
	g.ctx.ApplySound()
	g.scenes = scene.NewManager(scene.NewTitle())
	return g, nil
}

// Update reads input, handles the window keys, advances the master clock
// and runs the current scene.
func (g *Game) Update() error {
	in := g.ctx.Input
	// The pointer is turned into screen pixels through where present will
	// show the screen.
	sw, sh := g.preset.ScreenWidth, g.preset.ScreenHeight
	r := gfx.Fit(g.settings.ScaleMode, sw, sh, g.outW, g.outH)
	in.SetView(input.View{X: r.X, Y: r.Y, W: r.W, H: r.H, Width: sw, Height: sh})
	in.Update()

	if in.JustPressed(input.Quit) || ebiten.IsWindowBeingClosed() {
		g.saveSettings() // keeps the window size
		return ebiten.Termination
	}
	if in.JustPressed(input.Fullscreen) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
		g.saveSettings()
	}
	if in.JustPressed(input.ScaleMode) {
		g.settings.ScaleMode = g.preset.NextMode(g.settings.ScaleMode)
		g.saveSettings()
	}
	if in.JustPressed(input.Sound) {
		g.ctx.Sound.On = !g.ctx.Sound.On
		g.ctx.ApplySound()
		g.saveSettings()
	}

	g.ctx.Ticks, g.ctx.FreddySteps = g.clock.Advance()
	g.ctx.TickCount = g.clock.Total()
	g.ctx.Audio.Now(g.clock.Time())
	err := g.scenes.Update(g.ctx)
	if g.ctx.HighScore > g.savedHigh {
		g.savedHigh = g.ctx.HighScore
		if g.dir != "" {
			if err := score.SaveHigh(filepath.Join(g.dir, config.ScoresFile), g.savedHigh); err != nil {
				log.Printf("saving the high score: %v", err)
			}
		}
	}
	return err
}

// saveSettings writes the current settings (DESIGN §9). A failure is
// logged, never fatal.
func (g *Game) saveSettings() {
	s := &g.settings
	s.Fullscreen = ebiten.IsFullscreen()
	if w, h := ebiten.WindowSize(); w > 0 && h > 0 {
		s.WindowW, s.WindowH = w, h
	}
	snd := g.ctx.Sound
	s.SoundOn, s.Muted, s.Volume, s.Soften = snd.On, snd.Muted, snd.Volume, snd.Soften
	if g.dir == "" {
		return
	}
	if err := config.SaveSettings(filepath.Join(g.dir, config.SettingsFile), *s); err != nil {
		log.Printf("saving settings: %v", err)
	}
}

// Draw renders the scene to the internal screen and presents it.
func (g *Game) Draw(out *ebiten.Image) {
	g.scenes.Draw(g.ctx, g.screen)
	g.present(out)
}

// present scales the internal screen onto the output surface in the
// current scale mode (DESIGN §2.1).
func (g *Game) present(out *ebiten.Image) {
	sw, sh := g.preset.ScreenWidth, g.preset.ScreenHeight
	out.Fill(color.Black)
	r := gfx.Fit(g.settings.ScaleMode, sw, sh, g.outW, g.outH)
	if g.settings.ScaleMode != config.ScaleCRT {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(r.W/float64(sw), r.H/float64(sh))
		op.GeoM.Translate(r.X, r.Y)
		op.Filter = ebiten.FilterNearest
		if g.settings.ScaleMode == config.ScaleSmooth {
			op.Filter = ebiten.FilterLinear
		}
		out.DrawImage(g.screen, op)
		return
	}

	// Upscale by a whole number with nearest-neighbour, then resample
	// smoothly to the 4:3 size so rows stay even and sharp.
	k := gfx.PrescaleFactor(r, sw, sh)
	w, h := sw*k, sh*k
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
