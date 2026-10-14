package scene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
)

// Play is the placeholder game scene for M1. It shows the session and the
// master clock running, and handles paws mode, restart and menu.
type Play struct {
	paused      bool
	ticks       int
	freddySteps int
}

// NewPlay returns the placeholder game scene.
func NewPlay() *Play { return &Play{} }

// Update implements Scene.
func (p *Play) Update(ctx *Context) error {
	in := ctx.Input
	switch {
	case in.JustPressed(input.Menu):
		ctx.Switch(NewSkillSelect())
		return nil
	case in.JustPressed(input.Restart):
		ctx.Session = session.New(ctx.Session.Skill)
		ctx.Switch(NewPlay())
		return nil
	case in.JustPressed(input.Pause):
		p.paused = !p.paused
	}
	if p.paused {
		return nil
	}
	p.ticks += ctx.Ticks
	p.freddySteps += ctx.FreddySteps
	return nil
}

// Draw implements Scene.
func (p *Play) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Navy])
	f := ctx.Font
	s := ctx.Session
	f.DrawCentred(dst, "GAME PLACEHOLDER", 24, gfx.Yellow, 1)
	f.Draw(dst, fmt.Sprintf("Skill: %s (%c)", s.Skill, s.Skill.Letter()), 24, 52, gfx.White, 1)
	f.Draw(dst, fmt.Sprintf("Difficulty: %d   Lives: %d", s.Difficulty, s.Lives), 24, 64, gfx.White, 1)
	f.Draw(dst, fmt.Sprintf("Score: %s", session.FormatScore(s.Score)), 24, 76, gfx.White, 1)

	f.Draw(dst, fmt.Sprintf("Master ticks: %d", p.ticks), 24, 100, gfx.Cyan, 1)
	f.Draw(dst, fmt.Sprintf("Freddy steps: %d", p.freddySteps), 24, 112, gfx.Cyan, 1)
	sound := "off"
	if ctx.SoundOn {
		sound = "on"
	}
	f.Draw(dst, "Sound: "+sound, 24, 124, gfx.Cyan, 1)

	f.Draw(dst, "Esc paws  Ctrl+R restart  Ctrl+M menu", 8, 176, gfx.Silver, 1)

	if p.paused {
		dst.Fill(gfx.Palette[gfx.Night])
		f.DrawCentred(dst, "PAWS MODE", 84, gfx.Yellow, 2)
		f.DrawCentred(dst, "Esc to continue", 112, gfx.Silver, 1)
	}
}
