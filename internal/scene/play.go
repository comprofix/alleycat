package scene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/debug"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
)

// Play is the game scene. In M2 it is the alley with Freddy moving around
// it; it also handles paws mode, restart and menu.
type Play struct {
	paused bool
	alley  *alley.Alley
	view   *alley.Renderer
}

// NewPlay starts the alley at the session's difficulty.
func NewPlay(ctx *Context) *Play {
	return &Play{
		alley: alley.New(ctx.Session.Difficulty, ctx.RNG),
		view:  alley.NewRenderer(ctx.Alley),
	}
}

// Update implements Scene.
func (p *Play) Update(ctx *Context) error {
	in := ctx.Input
	switch {
	case in.JustPressed(input.Menu):
		ctx.Switch(NewSkillSelect())
		return nil
	case in.JustPressed(input.Restart):
		ctx.Session = session.New(ctx.Session.Skill)
		ctx.Switch(NewPlay(ctx))
		return nil
	case in.JustPressed(input.Pause):
		p.paused = !p.paused
	}
	if p.paused {
		return nil
	}
	p.alley.Update(ctx.Ticks, ctx.FreddySteps, controls(in), ctx.RNG)
	return nil
}

// controls reads the arrow keys. Opposite arrows held together cancel.
func controls(in *input.Input) entity.Controls {
	axis := func(neg, pos input.Action) int {
		switch {
		case in.Held(neg) && !in.Held(pos):
			return -1
		case in.Held(pos) && !in.Held(neg):
			return 1
		}
		return 0
	}
	return entity.Controls{H: axis(input.Left, input.Right), V: axis(input.Up, input.Down)}
}

// Draw implements Scene.
func (p *Play) Draw(ctx *Context, dst *ebiten.Image) {
	p.view.Draw(dst, p.alley)
	if debug.Enabled {
		p.view.DrawDebug(dst, p.alley)
		ctx.Font.Draw(dst, fmt.Sprintf("TPS %.1f FPS %.1f", ebiten.ActualTPS(), ebiten.ActualFPS()), 4, 190, gfx.White, 1)
	}

	if p.paused {
		f := ctx.Font
		dst.Fill(gfx.Palette[gfx.Night])
		f.DrawCentred(dst, "PAWS MODE", 84, gfx.Yellow, 2)
		f.DrawCentred(dst, "Esc to continue", 112, gfx.Silver, 1)
	}
}
