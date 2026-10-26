// Package scene provides the scene system with fade transitions (DESIGN §4,
// §6) and the front-end scenes.
package scene

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
)

// Scene is one screen of the game.
type Scene interface {
	// Update advances the scene by one game-loop frame.
	Update(ctx *Context) error
	// Draw renders onto the 320×200 logical screen. It must not change state.
	Draw(ctx *Context, dst *ebiten.Image)
}

// Context is the shared state handed to scenes.
type Context struct {
	Input     *input.Input
	RNG       *rand.Rand
	Session   *session.Session // nil outside a game
	Fonts     gfx.Fonts        // the menus' fonts
	PixelFont *gfx.PixelFont   // the pixel look's lettering
	UI        *gfx.Sheet
	Alley     *gfx.Sheet    // alley sprites: Freddy, laundry, litter
	Rooms     *gfx.Sheet    // room sprites: furniture, broom, cage, bird
	Audio     *audio.Synth  // nil in headless tests; every call is then ignored
	Patter    *audio.Patter // the background patter (DESIGN §7.2); nil in tests too
	FX        *audio.FX     // the sound effects (DESIGN §7.3); nil in tests too
	// Settings are the player's saved choices (DESIGN §9). The options
	// menu changes them and calls SettingsChanged; the game applies and
	// saves them.
	Settings *config.Settings
	Scheme   gfx.Scheme // the menus' colours
	// Preset is the look being shown (DESIGN §2.1): its screen size, scale
	// modes and art. The game sets it at start and when the look changes.
	Preset config.Preset

	// HighScore is the best score this run; it updates when a game ends
	// (DESIGN §3.8, "Lives and game over"). Saving it is M9 work.
	HighScore int

	// LastSkill is the skill last played, where the skill menu opens.
	LastSkill config.Skill

	// Clock output for the current frame (DESIGN §3.8).
	Ticks       int   // master ticks (18.2 Hz)
	TickCount   int64 // master ticks so far, this frame's included
	FreddySteps int   // Freddy updates (2 per tick)

	pixel           *ebiten.Image // the pixel scenes' canvas in the new look; a render cache
	next            Scene
	cut             bool
	quit            bool
	settingsChanged bool
}

// native is a scene that draws at the screen's own size in either look:
// the menus, and the scenes with the new look's art. Every other scene
// draws the pixel art on the 320×200 grid.
type native interface{ native() }

// drawScene draws s onto the screen dst. A pixel scene shown in the new
// look is enlarged four times, pixels kept square and sharp, until its new
// art is drawn (DESIGN §2.1).
func drawScene(ctx *Context, s Scene, dst *ebiten.Image) {
	if _, ok := s.(native); ok {
		s.Draw(ctx, dst)
		return
	}
	drawPixels(ctx, dst, func(l *ebiten.Image) { s.Draw(ctx, l) })
}

// drawPixels has draw draw pixel art on the 320×200 grid onto dst: straight
// onto it in the classic look, or in the new look on a 320×200 canvas laid
// over it enlarged four times.
func drawPixels(ctx *Context, dst *ebiten.Image, draw func(*ebiten.Image)) {
	if dst.Bounds().Dx() == gfx.ScreenWidth {
		draw(dst)
		return
	}
	if ctx.pixel == nil {
		ctx.pixel = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
	}
	ctx.pixel.Clear()
	draw(ctx.pixel)
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(dst.Bounds().Dx())/gfx.ScreenWidth, float64(dst.Bounds().Dy())/gfx.ScreenHeight)
	op.Filter = ebiten.FilterNearest
	dst.DrawImage(ctx.pixel, op)
}

// Switch asks the manager to fade to s at the end of this frame.
func (c *Context) Switch(s Scene) { c.next, c.cut = s, false }

// Cut asks the manager to change to s at the end of this frame without a
// fade or a pause in the sound, e.g. to open or close the pause menu.
func (c *Context) Cut(s Scene) { c.next, c.cut = s, true }

// Quit asks the game to save and close.
func (c *Context) Quit() { c.quit = true }

// Quitting reports whether a scene asked to quit.
func (c *Context) Quitting() bool { return c.quit }

// SettingsChanged records that a scene changed the settings, so the game
// applies and saves them.
func (c *Context) SettingsChanged() { c.settingsChanged = true }

// TakeSettingsChanged reports whether the settings changed since the last
// call, and clears the flag.
func (c *Context) TakeSettingsChanged() bool {
	ch := c.settingsChanged
	c.settingsChanged = false
	return ch
}

// ApplySound sends the sound settings to the synth (DESIGN §9). Ctrl+S
// flips Sound, as in the original. Sound plays only when it is on and not
// muted, and keeps its timing either way.
func (c *Context) ApplySound() {
	s := c.Settings
	c.Audio.SetEnabled(s.SoundOn && !s.Muted)
	c.Audio.SetVolume(s.Volume)
	c.Audio.SetSoften(s.Soften)
}
