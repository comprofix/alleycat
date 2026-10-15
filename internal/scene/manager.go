package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/gfx"
)

// FadeFrames is the length of each half of a transition (fade out, then fade
// in), in 60 TPS game-loop frames: 0.25 s each way.
const FadeFrames = 15

type phase int

const (
	idle phase = iota
	fadingOut
	fadingIn
)

// Manager runs the current scene and fades between scenes. Scenes are not
// updated while a fade is in progress.
type Manager struct {
	cur, pending Scene
	phase        phase
	frame        int
	overlay      *ebiten.Image
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
		if ctx.next != nil {
			m.pending, ctx.next = ctx.next, nil
			m.phase, m.frame = fadingOut, 0
			// Every sound ends with its scene, as the original silences
			// the speaker at each transition (DESIGN §7.1).
			ctx.Audio.Stop()
			ctx.Audio.SetPaused(false)
		}
	case fadingOut:
		m.frame++
		if m.frame >= FadeFrames {
			m.cur, m.pending = m.pending, nil
			m.phase, m.frame = fadingIn, 0
		}
	case fadingIn:
		m.frame++
		if m.frame >= FadeFrames {
			m.phase, m.frame = idle, 0
		}
	}
	return nil
}

// Darkness is how far the screen is faded to black, from 0 to 1.
func (m *Manager) Darkness() float32 {
	switch m.phase {
	case fadingOut:
		return float32(m.frame) / FadeFrames
	case fadingIn:
		return 1 - float32(m.frame)/FadeFrames
	}
	return 0
}

// Draw renders the current scene and the fade overlay.
func (m *Manager) Draw(ctx *Context, dst *ebiten.Image) {
	m.cur.Draw(ctx, dst)
	d := m.Darkness()
	if d <= 0 {
		return
	}
	if m.overlay == nil {
		m.overlay = ebiten.NewImage(1, 1)
		m.overlay.Fill(gfx.Palette[gfx.Night])
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(gfx.ScreenWidth, gfx.ScreenHeight)
	op.ColorScale.ScaleAlpha(d)
	dst.DrawImage(m.overlay, op)
}
