package scene

import (
	"image"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
)

// Title is the title screen (M8g): a night sky over the town, Freddy
// wandering the fence, and the name. Before the first game it moves on to
// skill select, by itself after a long wait or at any key. After a game it
// loops its tune, and any key replays at the same skill (DESIGN §4, "No
// attract mode").
type Title struct {
	blink  gfx.Animator
	played bool // the title theme has started
	ticks  int  // master ticks since the theme started
	cat    titleCat

	back *ebiten.Image // the night scene, drawn once; a render cache
}

// NewTitle returns the title scene.
func NewTitle() *Title {
	// The pointer blinks about twice a second (9 master ticks per frame).
	return &Title{blink: gfx.Animator{TicksPerFrame: 9, Frames: 2}}
}

// Update implements Scene.
func (t *Title) Update(ctx *Context) error {
	// The theme plays once each time the title shows (DESIGN §7.4).
	if !t.played {
		t.played = true
		ctx.Audio.PlayMusic(audio.Title(), false)
	}
	t.blink.Advance(ctx.Ticks)
	t.cat.update(ctx.Ticks, ctx.RNG)
	t.ticks += ctx.Ticks
	switch {
	case ctx.Input.AnyKeyJustPressed() && ctx.Played:
		startGame(ctx, ctx.LastSkill)
	case ctx.Input.AnyKeyJustPressed():
		ctx.Switch(NewSkillSelect())
	case !ctx.Played && t.ticks > config.TitleHoldTicks+config.TitleFirstExtra:
		ctx.Switch(NewSkillSelect())
	case ctx.Played && t.ticks >= config.TitleHoldTicks+config.TitleRepeatExtra:
		// The tune starts again on the next update.
		t.played, t.ticks = false, 0
	}
	return nil
}

// titleCat is Freddy wandering along the title fence on his own, walking
// and stopping to sit (DESIGN §4, "The title cat").
type titleCat struct {
	x, dir, facing int
	frame          int // walk-cycle frame, 0–5
	wait           int // master ticks until the next choice
}

// update moves him one step per master tick.
func (c *titleCat) update(ticks int, rng *rand.Rand) {
	for range ticks {
		switch {
		case c.x <= config.TitleCatTurnLeftX:
			c.dir = 1
		case c.x >= config.TitleCatTurnRightX:
			c.dir = -1
		case c.wait <= 0:
			c.wait = config.TitleCatEvery
			c.dir = 0
			if b := rng.IntN(256); b <= config.TitleCatStopAbove {
				c.dir = 1
				if b%2 == 0 {
					c.dir = -1
				}
			}
		}
		if c.dir != 0 {
			c.facing = c.dir
			c.x += c.dir * config.TitleCatSpeed
			c.frame = (c.frame + 1) % 6
		}
		c.wait--
	}
}

// draw draws him on the fence top with the slate rim that lifts him off
// the dark rooftops, as in the alley.
func (c *titleCat) draw(dst *ebiten.Image, sheet *gfx.Sheet) {
	img := sheet.Sprite("freddy_sit").Frame(0)
	if c.dir != 0 {
		img = sheet.Sprite("freddy_walk").Frame(c.frame)
	}
	y := titleFenceY - img.Bounds().Dy() + 1
	flip := c.facing < 0
	for _, d := range [4][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}} {
		entity.DrawFlat(dst, img, c.x+d[0], y+d[1], flip, gfx.Slate)
	}
	entity.DrawAt(dst, img, c.x, y, flip)
}

// The title card's layout (our own art).
const (
	titleFenceY = 176 // the fence top
	titleLogoY  = 36
)

// titleRoofs are the town's rooftops: left edge, width and top.
var titleRoofs = []struct{ x, w, top int }{
	{0, 28, 150}, {28, 18, 142}, {46, 34, 154}, {80, 14, 146}, {94, 30, 152},
	{124, 22, 140}, {146, 36, 156}, {182, 16, 145}, {198, 30, 151}, {228, 20, 143},
	{248, 34, 155}, {282, 16, 147}, {298, 22, 152},
}

var titleStars = [][2]int{{14, 8}, {40, 22}, {66, 4}, {92, 30}, {118, 12}, {150, 6}, {173, 26},
	{204, 10}, {229, 34}, {258, 18}, {281, 7}, {306, 28}, {22, 52}, {290, 60}, {58, 70}, {270, 96}}

func fillRect(dst *ebiten.Image, x, y, w, h, idx int) {
	dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(gfx.Palette[idx])
}

// buildBack draws the night scene behind the title.
func (t *Title) buildBack(ctx *Context) {
	b := ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
	// The sky in bands, darkest at the top.
	fillRect(b, 0, 0, gfx.ScreenWidth, 118, gfx.Night)
	fillRect(b, 0, 118, gfx.ScreenWidth, 16, gfx.Charcoal)
	fillRect(b, 0, 134, gfx.ScreenWidth, gfx.ScreenHeight-134, gfx.Navy)
	for i, s := range titleStars {
		idx := gfx.Silver
		if i%3 == 0 {
			idx = gfx.White
		}
		fillRect(b, s[0], s[1], 1, 1, idx)
	}
	entity.DrawAt(b, ctx.Alley.Sprite("moon").Frame(0), 262, 14, false)
	for i, r := range titleRoofs {
		fillRect(b, r.x, r.top, r.w, titleFenceY-r.top, gfx.Night)
		fillRect(b, r.x+r.w/3, r.top-3, 2, 3, gfx.Night) // a chimney
		for wy := r.top + 4; wy < titleFenceY-4; wy += 6 {
			for wx := r.x + 3; wx < r.x+r.w-3; wx += 6 {
				if (wx*7+wy*3+i)%5 == 0 {
					fillRect(b, wx, wy, 2, 2, gfx.Yellow)
				}
			}
		}
	}
	// The fence along the bottom, its boards ragged at the top.
	for i, x := 0, 0; x < gfx.ScreenWidth; i, x = i+1, x+8 {
		top := titleFenceY + (i*37+i/3)%3
		fillRect(b, x, top, 7, gfx.ScreenHeight-top, gfx.Teal)
		fillRect(b, x, top+1, 1, gfx.ScreenHeight-top, gfx.Slate)
		fillRect(b, x+7, top+2, 1, gfx.ScreenHeight-top, gfx.Charcoal)
	}
	t.back = b
}

// Draw implements Scene.
func (t *Title) Draw(ctx *Context, dst *ebiten.Image) {
	if t.back == nil {
		t.buildBack(ctx)
	}
	dst.DrawImage(t.back, nil)
	t.cat.draw(dst, ctx.Alley)
	f := ctx.Font
	const name = "ALLEY CAT"
	x := gfx.CentreX(name, 3)
	f.Draw(dst, name, x+3, titleLogoY+3, gfx.Plum, 3) // a deep drop shadow
	f.DrawOutlined(dst, name, x, titleLogoY, gfx.Orange, gfx.Night, 3)
	f.DrawShadowed(dst, "a faithful remake", gfx.CentreX("a faithful remake", 1), titleLogoY+34, gfx.Silver, gfx.Night, 1)

	const prompt = "PRESS ANY KEY"
	y := 104
	f.DrawShadowed(dst, prompt, gfx.CentreX(prompt, 1), y, gfx.White, gfx.Night, 1)
	ptr := ctx.UI.Sprite("pointer")
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(gfx.CentreX(prompt, 1)-ptr.W-4), float64(y))
	dst.DrawImage(ptr.Frame(t.blink.Frame()), op)
}
