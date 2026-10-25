package alley

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// DrawDebug outlines the collision boxes for debug builds. Surfaces are
// shown as the zone Freddy's top-left corner must fall in to stand on them
// (DESIGN §3.8, "Alley surfaces"); Freddy's own box is outlined, with his
// top-left corner marked.
func (r *Renderer) DrawDebug(dst *ebiten.Image, a *Alley) {
	box := func(x, y, w, h, idx int) {
		vector.StrokeRect(dst, float32(x)+0.5, float32(y)+0.5, float32(w)-1, float32(h)-1, 1, gfx.Palette[idx], false)
	}
	f := a.Freddy

	// Floor and fence top span the alley.
	box(f.MinX, config.FloorY, f.MaxX-f.MinX+1, 1, resurrect64.Seafoam)
	box(0, config.FenceTopY, gfx.ScreenWidth, 8, resurrect64.Honey)

	// Cans: X from the can's X to X + 22, Y within the can's 8-px band.
	for _, c := range a.Cans {
		box(c.X, CanTop(c), config.CanReach+8, 8, resurrect64.Coral)
	}

	// Laundry: each supporting 8-px half, offset by the X + 10 probe.
	for i := range a.Lines {
		l := &a.Lines[i]
		for p, piece := range l.Pieces {
			for h, k := range piece {
				if k == NoLaundry {
					continue
				}
				x := l.Offset + p*config.LinePiecePx + h*config.LinePiecePx/2 - config.LineProbeX
				box(x, config.LineY[i], 8, 8, resurrect64.Cerise)
			}
		}
	}

	// Alley systems (DESIGN §3.8, "Alley systems").
	if a.Window.Phase > 0 {
		r := WindowRect(a.Window.Index)
		box(r.X, r.Y, r.W, r.H, resurrect64.Blue)
	}
	if t := &a.Thrown; t.Active {
		r := t.rect()
		box(r.X, r.Y, r.W, r.H, resurrect64.Honey)
	}
	for i, m := range a.Mice {
		r := mouseRect(i, m.X)
		box(r.X, r.Y, r.W, r.H, resurrect64.Fog)
	}
	if c := &a.TrashCat; c.Phase > 0 && c.H > 0 {
		r := c.rect()
		box(r.X, r.Y, r.W, r.H, resurrect64.Coral)
	}
	if d := &a.Dog; d.Out() && d.Capture == 0 {
		// The zone Freddy's X must be in to be caught, at catching height.
		x := max(d.X-config.DogCatchBefore, 0)
		box(x, config.DogCatchY, d.X+config.DogCatchAfter-x+1, config.FloorY-config.DogCatchY+1, resurrect64.Cerise)
	}

	if !f.Hidden() {
		b := f.HitBox()
		box(b.X, b.Y, b.W, b.H, resurrect64.Leaf)
		vector.FillRect(dst, float32(f.X), float32(f.Y), 1, 1, gfx.Palette[resurrect64.White], false)
	}
}
