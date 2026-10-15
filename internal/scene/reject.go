package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
)

// Reject is the heart room's rejection scene (DESIGN §3.5, "Failing"): on
// a black screen Freddy's head falls from the top to the bottom, then
// becomes his face with a speech bubble of cartoon symbols. No life is
// lost. A falling tone plays as the head drops, then two tones alternate
// under the bubble (DESIGN §7.3).
type Reject struct {
	y       int
	held    int
	landed  bool
	sounded bool // the falling tone has started
	next    func(*Context) Scene
}

func newReject(next func(*Context) Scene) *Reject {
	return &Reject{next: next}
}

// Update implements Scene.
func (r *Reject) Update(ctx *Context) error {
	if !r.sounded {
		r.sounded = true
		steps := (config.RejectLandY + config.RejectFallStep - 1) / config.RejectFallStep
		if n := ctx.FX.RejectFall(steps); n != nil {
			ctx.Audio.PlaySFX(n)
		}
	}
	for range ctx.Ticks {
		switch {
		case !r.landed:
			if r.y += config.RejectFallStep; r.y >= config.RejectLandY {
				r.landed = true
				if n := ctx.FX.Hi(config.RejectHoldTicks); n != nil {
					ctx.Audio.PlaySFX(n)
				}
			}
		default:
			if r.held++; r.held >= config.RejectHoldTicks {
				ctx.Switch(r.next(ctx))
				return nil
			}
		}
	}
	return nil
}

// Draw implements Scene.
func (r *Reject) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	if r.landed {
		entity.DrawAt(dst, ctx.Rooms.Sprite("reject_face").Frame(0), config.RejectX, config.RejectLandY, false)
		return
	}
	entity.DrawAt(dst, ctx.Rooms.Sprite("freddy_head").Frame(0), config.RejectX, r.y, false)
}
