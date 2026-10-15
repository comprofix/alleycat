package scene

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/debug"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/rooms"
	"github.com/comprofix/alleycat/internal/session"
)

// Play is the alley scene. It runs the alley, sends Freddy into a room when
// he goes in through a window, and ends the game when the lives run out. It
// also handles paws mode, restart and menu.
type Play struct {
	paused bool
	alley  *alley.Alley
	view   *alley.Renderer
}

// NewPlay starts the alley at the session's difficulty, with Freddy walking
// in from the left.
func NewPlay(ctx *Context) *Play {
	return &Play{
		alley: alley.New(ctx.Session, ctx.RNG),
		view:  alley.NewRenderer(ctx.Alley, ctx.Font),
	}
}

// newReturnPlay is the alley with Freddy back from a room at x, y.
func newReturnPlay(ctx *Context, x, y int) *Play {
	return &Play{
		alley: alley.NewReturning(ctx.Session, ctx.RNG, x, y),
		view:  alley.NewRenderer(ctx.Alley, ctx.Font),
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
	a := p.alley
	a.Update(ctx.Ticks, ctx.FreddySteps, controls(in), ctx.RNG)

	// At 0 lives the game ends at once, straight back to the title (DESIGN
	// §3.8, "Lives and game over").
	if ctx.Session.Lives == 0 {
		ctx.HighScore = max(ctx.HighScore, ctx.Session.Score)
		ctx.Switch(NewTitle())
		return nil
	}

	if a.Exit != alley.NoExit {
		s := ctx.Session
		room := rooms.Heart
		if a.Exit == alley.ToHeart {
			s.FeliciaWaiting = false
		} else {
			room = s.Rooms.Pick(s.Difficulty, ctx.RNG)
		}
		ctx.Switch(newStubRoom(room, a.Freddy.X, a.Freddy.Y))
	}
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
	p.view.Draw(dst, p.alley, ctx.HighScore)
	if debug.Enabled {
		p.view.DrawDebug(dst, p.alley)
		ctx.Font.Draw(dst, fmt.Sprintf("TPS %.1f FPS %.1f", ebiten.ActualTPS(), ebiten.ActualFPS()), 4, 190, gfx.White, 1)
	}

	if p.paused {
		// The alley stays in view, frozen, under a small banner (owner
		// request, 2026-10-15).
		f := ctx.Font
		const msg = "PAWS MODE - Esc to continue"
		w := gfx.Width(msg)
		x := (gfx.ScreenWidth - w) / 2
		dst.SubImage(image.Rect(x-4, 0, x+w+4, 12)).(*ebiten.Image).Fill(gfx.Palette[gfx.Night])
		f.Draw(dst, msg, x, 2, gfx.Yellow, 1)
	}
}
