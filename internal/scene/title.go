package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/gfx"
)

// Title is the placeholder title screen. Any key goes to skill select.
type Title struct {
	blink gfx.Animator
}

// NewTitle returns the title scene.
func NewTitle() *Title {
	// The pointer blinks about twice a second (9 master ticks per frame).
	return &Title{blink: gfx.Animator{TicksPerFrame: 9, Frames: 2}}
}

// Update implements Scene.
func (t *Title) Update(ctx *Context) error {
	t.blink.Advance(ctx.Ticks)
	if ctx.Input.AnyKeyJustPressed() {
		ctx.Switch(NewSkillSelect())
	}
	return nil
}

// Draw implements Scene.
func (t *Title) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	ctx.Font.DrawCentred(dst, "ALLEY CAT", 56, gfx.Orange, 3)
	ctx.Font.DrawCentred(dst, "a faithful remake", 92, gfx.Silver, 1)

	const prompt = "PRESS ANY KEY"
	y := 140
	ctx.Font.DrawCentred(dst, prompt, y, gfx.White, 1)
	ptr := ctx.UI.Sprite("pointer")
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64((gfx.ScreenWidth-gfx.Width(prompt))/2-ptr.W-4), float64(y))
	dst.DrawImage(ptr.Frame(t.blink.Frame()), op)
}
