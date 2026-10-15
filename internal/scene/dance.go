package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
)

// Dance is the dancing-cats interlude after a kiss (DESIGN §3.5): on a
// black screen the cats dance for 40 ticks, changing step every 5, half of
// them out of step. The arrangement grows with the new difficulty.
type Dance struct {
	dancers []config.Dancer
	ticks   int
	next    func(*Context) Scene
}

// newDance returns the interlude for the new difficulty, or nil below 2.
func newDance(difficulty int, next func(*Context) Scene) *Dance {
	d := config.DanceLayouts[min(max(difficulty, 0), config.MaxDifficulty)]
	if len(d) == 0 {
		return nil
	}
	return &Dance{dancers: d, next: next}
}

// Update implements Scene.
func (d *Dance) Update(ctx *Context) error {
	if d.ticks += ctx.Ticks; d.ticks >= config.DanceTicks {
		ctx.Switch(d.next(ctx))
	}
	return nil
}

// Draw implements Scene.
func (d *Dance) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	cat := ctx.Rooms.Sprite("dance_cat")
	step := (d.ticks / config.DanceFrameTicks) % 2
	for _, c := range d.dancers {
		frame := step
		if c.Off {
			frame ^= 1
		}
		entity.DrawAt(dst, cat.Frame(frame), c.X, c.Y, false)
	}
}
