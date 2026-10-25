package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Fishbowl room layout, measured from the reference for layout only (DESIGN
// §3.4, "Fishbowl room layout"). The art is our own.
const (
	fishOpeningX                 = 144
	fishChairLX, fishChairMX     = 24, 136
	fishTableX                   = 64
	bowlTableX                   = 216
	fishLampX                    = 272
	bowlDrawX, bowlDrawY         = 226, 134
	surfaceDepth                 = 4 // the surface band is 4–6 px deep
	rippleCell                   = 16
	swimFrameUp, swimFrameDown   = 4, 6
	swimFrameTurn, swimFrameSink = 8, 9
)

// buildFishbowl draws the fishbowl room's backdrop: the walls, the window
// with its curtains, the chairs, the table, the bowl table and the lamp.
// The bowl itself is drawn each frame.
func (r *Renderer) buildFishbowl() {
	b := r.back
	r.buildWalls(wallY0, wallY1)
	r.buildWindow(fishOpeningX, openingY, "curtain_teal")
	chair := r.sheet.Sprite("chair_teal").Frame(0)
	entity.DrawAt(b, chair, fishChairLX, furnitureY, false)
	entity.DrawAt(b, r.sheet.Sprite("table_teal").Frame(0), fishTableX, tableY, false)
	entity.DrawAt(b, chair, fishChairMX, furnitureY, true)
	entity.DrawAt(b, r.sheet.Sprite("pedestal_teal").Frame(0), bowlTableX, pedestalY, false)
	entity.DrawAt(b, r.sheet.Sprite("lamp_teal").Frame(0), fishLampX, lampY, false)
}

// buildWater draws the underwater backdrop: the water and a fixed scatter
// of ripple marks. The surface band is drawn each frame.
func (r *Renderer) buildWater() {
	s := gfx.UnderwaterScheme
	b := r.water
	b.Fill(gfx.Palette[s.Water])
	// Deeper water darkens gently (M8f): a quarter of the pixels from row
	// 120, half of them near the bottom.
	for y := 120; y < gfx.ScreenHeight; y++ {
		for x := (y / 2) % 2 * 2; x < gfx.ScreenWidth; x += 4 {
			if y%2 == 0 {
				fill(b, x+(y/2)%2, y, 1, 1, resurrect64.Navy)
			}
			if y >= 170 && y%2 == 1 {
				fill(b, x+2, y, 1, 1, resurrect64.Navy)
			}
		}
	}
	// Faint light rays slanting down from the surface.
	for _, rx := range []int{40, 112, 196, 268} {
		for y := 8; y < 90; y++ {
			if y%3 == 0 {
				fill(b, rx+y/3, y, 2, 1, resurrect64.Blue)
			}
		}
	}
	// The sandy bottom, below Freddy's deepest dive, with pebbles and weed.
	const sandY = 194
	fill(b, 0, sandY, gfx.ScreenWidth, gfx.ScreenHeight-sandY, resurrect64.Honey)
	for x := 3; x < gfx.ScreenWidth; x += 11 {
		fill(b, x, sandY+1+x%4, 1, 1, resurrect64.Coral)
		if x%3 == 0 {
			fill(b, x+4, sandY+2, 2, 1, resurrect64.Lilac)
		}
	}
	for _, wx := range []int{18, 74, 150, 233, 297} {
		for y := sandY - 1; y > sandY-14-wx%9; y-- {
			fill(b, wx+(y/3)%2, y, 1, 1, resurrect64.Emerald)
		}
	}
	for y := 16; y < gfx.ScreenHeight-4; y += rippleCell {
		for x := 4; x < gfx.ScreenWidth-8; x += rippleCell {
			if h := (x*5 + y*11) % 13; h < 2 {
				fill(b, x+h*3, y+h, 2, 1, s.WallShade)
				fill(b, x+h*3+2, y+h-1, 2, 1, s.WallShade)
				fill(b, x+h*3+4, y+h, 2, 1, s.WallShade)
			}
		}
	}
}

// drawTinted draws a single-colour sprite recoloured to palette entry idx.
func drawTinted(dst, img *ebiten.Image, x, y int, flip bool, idx int) {
	var cm colorm.ColorM
	cm.Scale(0, 0, 0, 1)
	c := gfx.Palette[idx]
	cm.Translate(float64(c.R)/255, float64(c.G)/255, float64(c.B)/255, 0)
	op := &colorm.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(img.Bounds().Dx()), 0)
	}
	op.GeoM.Translate(float64(x), float64(y))
	colorm.DrawImage(dst, img, cm, op)
}

// drawLit draws a sprite drawn in Ink and White as the palette entries dark
// and light: every channel is mapped from the source's red, Ink going to
// dark and White to light, so the sprite's shading survives recolouring.
func drawLit(dst, img *ebiten.Image, x, y int, flip bool, dark, light int) {
	var cm colorm.ColorM
	r0 := float64(gfx.Palette[resurrect64.Ink].R) / 255
	d, l := gfx.Palette[dark], gfx.Palette[light]
	for i, c := range [3][2]uint8{{d.R, l.R}, {d.G, l.G}, {d.B, l.B}} {
		a, b := float64(c[0])/255, float64(c[1])/255
		k := (b - a) / (1 - r0)
		for j := range 3 {
			cm.SetElement(i, j, 0)
		}
		cm.SetElement(i, 0, k)
		cm.SetElement(i, 4, a-r0*k)
	}
	op := &colorm.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(img.Bounds().Dx()), 0)
	}
	op.GeoM.Translate(float64(x), float64(y))
	colorm.DrawImage(dst, img, cm, op)
}

// drawBowl draws the bowl with its fish on the bowl table.
func (r *Renderer) drawBowl(dst *ebiten.Image, fb *FishbowlRoom) {
	entity.DrawAt(dst, r.sheet.Sprite("bowl").Frame(fb.BowlFrame), bowlDrawX, bowlDrawY, false)
}

// drawUnderwater draws the underwater scene: the water, the surface, the
// fish and eels, Freddy, his bubbles and "ZAP!".
func (r *Renderer) drawUnderwater(dst *ebiten.Image, fb *FishbowlRoom) {
	if r.water == nil {
		r.water = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
		r.buildWater()
	}
	dst.DrawImage(r.water, nil)
	s := gfx.UnderwaterScheme
	for c, f := range fb.Surface {
		for half := range 2 {
			h := surfaceDepth + (c*3+f+half*2)%3
			fill(dst, c*8+half*4, 0, 4, h, s.Trim)
		}
	}

	ink, lit := gfx.AirColours[fb.AirStage()], gfx.AirLights[fb.AirStage()]
	if fb.Phase == Zapped {
		ink, lit = resurrect64.White, resurrect64.White
		if fb.Zap%2 == 0 {
			ink, lit = resurrect64.Cobalt, resurrect64.Blue
		}
	}
	fish, eel := r.sheet.Sprite("fish"), r.sheet.Sprite("eel")
	for i := range fb.Swimmers {
		sw := &fb.Swimmers[i]
		if sw.Off {
			continue
		}
		if IsEel(i) {
			drawLit(dst, eel.Frame((fb.EelFrame+i)%4), sw.X, sw.Y, false, ink, lit)
			continue
		}
		entity.DrawAt(dst, fish.Frame(fb.FishFrame^(i&1)), sw.X, sw.Y, sw.HDir < 0)
	}
	bubble := r.sheet.Sprite("bubble").Frame(0)
	for _, b := range fb.Bubbles {
		entity.DrawAt(dst, bubble, b.X+config.FreddyW/2, b.Y, false)
	}
	f := fb.Freddy
	drawLit(dst, r.sheet.Sprite("freddy_swim").Frame(swimFrame(&f.Swim)), f.X, f.Y, f.Facing < 0, ink, lit)
	if fb.Phase == Zapped {
		entity.DrawAt(dst, r.actors.Sprite("zap").Frame(0), fb.ZapX, fb.ZapY, false)
	}
}

// swimFrame is the freddy_swim frame for a pose.
func swimFrame(s *entity.Swim) int {
	switch s.Pose {
	case entity.SwimUp:
		return swimFrameUp + s.Frame
	case entity.SwimDown:
		return swimFrameDown + s.Frame
	case entity.SwimTurn:
		return swimFrameTurn
	case entity.SwimSink:
		return swimFrameSink
	}
	return s.Frame
}
