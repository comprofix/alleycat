package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Heart room layout, from the reference for layout only (DESIGN §3.5,
// "Heart room layout"). The art is our own.
const (
	cherubW, cherubH = 32, 24
	cherubSideY      = 24 // the side columns start below the top row
)

// buildHeart draws the heart room's backdrop: the green ground and the
// border of cherubs along the top and down both sides. The hearts change,
// so they are drawn each frame.
func (r *Renderer) buildHeart() {
	b := r.back
	b.Fill(gfx.Palette[r.scheme.Wall])
	// A faint lattice of tiny hearts behind the platforms (M8f).
	for y, row := 4, 0; y < gfx.ScreenHeight; y, row = y+12, row+1 {
		for x := 4 + (row%2)*12; x < gfx.ScreenWidth; x += 24 {
			fill(b, x, y, 1, 1, resurrect64.Leaf)
			fill(b, x+2, y, 1, 1, resurrect64.Leaf)
			fill(b, x+1, y+1, 1, 1, resurrect64.Forest) // the point in shade
		}
	}
	cherub := r.sheet.Sprite("cherub").Frame(0)
	for x := 0; x < gfx.ScreenWidth; x += cherubW {
		entity.DrawAt(b, cherub, x, 0, x >= gfx.ScreenWidth/2)
	}
	for y := cherubSideY; y < gfx.ScreenHeight; y += cherubH {
		entity.DrawAt(b, cherub, 0, y, false)
		entity.DrawAt(b, cherub, gfx.ScreenWidth-cherubW, y, true)
	}
}

// drawHeart draws the heart room: the hearts, the gifts, the cats, the
// arrow and Freddy, blue while he carries a gift.
func (r *Renderer) drawHeart(dst *ebiten.Image, h *HeartRoom) {
	if r.back == nil {
		r.back = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.scheme = gfx.HeartScheme
		r.buildHeart()
	}
	dst.DrawImage(r.back, nil)

	heart := r.sheet.Sprite("heart")
	for row, line := range config.HeartRows {
		for col := range config.HeartCols {
			frame := 0
			if h.Broken[row][col] {
				frame = 1
			}
			entity.DrawAt(dst, heart.Frame(frame), config.HeartX+col*config.HeartW, line+config.HeartDrawDY, false)
		}
	}
	gift := r.sheet.Sprite("gift").Frame(0)
	for _, g := range h.Gifts {
		if g.On {
			entity.DrawAt(dst, gift, g.X, g.Y, false)
		}
	}
	for i := range h.Cats {
		set := "rival"
		if i == config.Felicia {
			set = "felicia"
		}
		r.drawCat(dst, &h.Cats[i], set)
	}
	if a := &h.Arrow; a.Active {
		entity.DrawAt(dst, r.sheet.Sprite("arrow").Frame(a.Frame()), a.X, a.Y, a.Dir < 0)
	}
	f := h.Freddy
	switch {
	case f.Gone:
	case h.Carrying >= 0:
		entity.DrawAt(dst, entity.FreddyImageIn(r.actors, f, "gift"), f.X, f.Y, f.Facing < 0)
		entity.DrawDust(dst, r.actors, f)
	default:
		entity.DrawFreddy(dst, r.actors, f)
	}
}

// drawCat draws a heart room cat with Freddy's own frames recoloured, from
// the sprite set ("rival", "felicia" or "silver"), so the cats are his size
// and look like him; bumped, it shows his jumping pose (DESIGN §8).
func (r *Renderer) drawCat(dst *ebiten.Image, c *HeartCat, set string) {
	if c.Away {
		return
	}
	img := r.actors.Sprite(set + "_walk").Frame(c.Frame)
	if c.Bump > 0 {
		img = r.actors.Sprite(set + "_jump").Frame(0)
	}
	entity.DrawAt(dst, img, c.X, c.Y, c.Dir < 0)
}

// drawHeartDebug outlines the heart room's row lines, cats, gifts, arrow,
// Freddy and the fall-off line.
func drawHeartDebug(dst *ebiten.Image, h *HeartRoom) {
	box := func(x, y, w, hh, idx int) {
		vector.StrokeRect(dst, float32(x)+0.5, float32(y)+0.5, float32(w)-1, float32(hh)-1, 1, gfx.Palette[idx], false)
	}
	for _, line := range config.HeartRows {
		box(config.HeartX, line+config.HeartStandDY, config.HeartCols*config.HeartW, 1, resurrect64.Honey)
	}
	box(0, config.HeartFallOffY, gfx.ScreenWidth, 1, resurrect64.Cerise)
	for i := range h.Cats {
		if !h.Cats[i].Away {
			cr := h.CatRect(i)
			box(cr.X, cr.Y, cr.W, cr.H, resurrect64.White)
		}
	}
	for g := range h.Gifts {
		if h.Gifts[g].On {
			gr := h.GiftRect(g)
			box(gr.X, gr.Y, gr.W, gr.H, resurrect64.Seafoam)
		}
	}
	if h.Arrow.Active {
		ar := h.ArrowRect()
		box(ar.X, ar.Y, ar.W, ar.H, resurrect64.Coral)
	}
	if f := h.Freddy; !f.Gone {
		hb := f.HitBox()
		box(hb.X, hb.Y, hb.W, hb.H, resurrect64.Leaf)
	}
}

// DrawHeartKiss draws the heart room as it stands at the kiss, recoloured
// to the kiss colours (DESIGN §3.5, "The kiss"): the room is frozen, Freddy
// and Felicia are in the big heart, and the rivals turn silver.
func (r *Renderer) DrawHeartKiss(dst *ebiten.Image, h *HeartRoom) {
	s := gfx.KissScheme
	if r.kiss == nil {
		r.kiss = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.kiss.Fill(gfx.Palette[s.Wall])
		cherub := r.sheet.Sprite("cherub_kiss").Frame(0)
		for x := 0; x < gfx.ScreenWidth; x += cherubW {
			entity.DrawAt(r.kiss, cherub, x, 0, x >= gfx.ScreenWidth/2)
		}
		for y := cherubSideY; y < gfx.ScreenHeight; y += cherubH {
			entity.DrawAt(r.kiss, cherub, 0, y, false)
			entity.DrawAt(r.kiss, cherub, gfx.ScreenWidth-cherubW, y, true)
		}
	}
	dst.DrawImage(r.kiss, nil)
	heart := r.sheet.Sprite("heart_kiss")
	for row, line := range config.HeartRows {
		for col := range config.HeartCols {
			frame := 0
			if h.Broken[row][col] {
				frame = 1
			}
			entity.DrawAt(dst, heart.Frame(frame), config.HeartX+col*config.HeartW, line+config.HeartDrawDY, false)
		}
	}
	gift := r.sheet.Sprite("gift").Frame(0)
	for _, g := range h.Gifts {
		if g.On {
			entity.DrawAt(dst, gift, g.X, g.Y, false)
		}
	}
	for i := range config.Felicia {
		r.drawCat(dst, &h.Cats[i], "silver")
	}
}
