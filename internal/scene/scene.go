// Package scene provides the scene system with fade transitions (DESIGN §4,
// §6) and the front-end scenes.
package scene

import (
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

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
	Input   *input.Input
	RNG     *rand.Rand
	Session *session.Session // nil outside a game
	Font    *gfx.Font
	UI      *gfx.Sheet
	Alley   *gfx.Sheet // alley sprites: Freddy, laundry, litter
	SoundOn bool

	// Clock output for the current frame (DESIGN §3.8).
	Ticks       int // master ticks (18.2 Hz)
	FreddySteps int // Freddy updates (2 per tick)

	next Scene
}

// Switch asks the manager to fade to s at the end of this frame.
func (c *Context) Switch(s Scene) { c.next = s }
