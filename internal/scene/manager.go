package scene

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// FadeTicks is the length of each half of a transition (fade out, then fade
// in): a quarter of a second each way.
const FadeTicks = config.TPS / 4

type phase int

const (
	idle phase = iota
	fadingOut
	fadingIn
)

// Manager runs the current scene and fades between scenes, as the Game
// Skeleton's. Scenes are not updated while a fade is in progress.
type Manager struct {
	cur, pending Scene
	phase        phase
	tick         int
}

// NewManager starts with s, fading in from black.
func NewManager(s Scene) *Manager {
	return &Manager{cur: s, phase: fadingIn}
}

// Current returns the active scene.
func (m *Manager) Current() Scene { return m.cur }

// Update advances the current scene or the transition.
func (m *Manager) Update(ctx *Context) error {
	switch m.phase {
	case idle:
		if err := m.cur.Update(ctx); err != nil {
			return err
		}
		if ctx.next == nil {
			break
		}
		if ctx.cut {
			m.cur = ctx.next
		} else {
			m.pending = ctx.next
			m.phase, m.tick = fadingOut, 0
			// Every sound ends with its scene, as the original silences
			// the speaker at each transition (DESIGN §7.1).
			ctx.Audio.Stop()
			ctx.Audio.SetPaused(false)
			// The rising figure plays through the fade out and back in.
			ctx.Audio.PlaySFX(audio.Wipe(2 * FadeTicks * time.Second / config.TPS))
		}
		ctx.next, ctx.cut = nil, false
	case fadingOut:
		m.tick++
		if m.tick >= FadeTicks {
			m.cur, m.pending = m.pending, nil
			m.phase, m.tick = fadingIn, 0
		}
	case fadingIn:
		m.tick++
		if m.tick >= FadeTicks {
			m.phase, m.tick = idle, 0
		}
	}
	return nil
}

// Darkness is how far the screen is faded to black, from 0 to 1.
func (m *Manager) Darkness() float32 {
	switch m.phase {
	case fadingOut:
		return float32(m.tick) / FadeTicks
	case fadingIn:
		return 1 - float32(m.tick)/FadeTicks
	}
	return 0
}

// Draw renders the current scene and the fade.
func (m *Manager) Draw(ctx *Context, dst *ebiten.Image) {
	drawScene(ctx, m.cur, dst)
	if d := m.Darkness(); d > 0 {
		Dim(dst, resurrect64.Ink, d)
	}
}

// Dim covers dst with palette colour c at the given opacity (0–1).
func Dim(dst *ebiten.Image, c int, opacity float32) {
	col := gfx.RGBA(c)
	col.R = uint8(float32(col.R) * opacity)
	col.G = uint8(float32(col.G) * opacity)
	col.B = uint8(float32(col.B) * opacity)
	col.A = uint8(255 * opacity)
	vector.FillRect(dst, 0, 0, float32(dst.Bounds().Dx()), float32(dst.Bounds().Dy()), col, false)
}
