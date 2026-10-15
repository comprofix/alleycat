package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
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
	r.buildWalls()
	r.buildWindow(fishOpeningX, "curtain_teal")
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

	ink := gfx.AirColours[fb.AirStage()]
	if fb.Phase == Zapped {
		ink = gfx.White
		if fb.Zap%2 == 0 {
			ink = gfx.Blue
		}
	}
	fish, eel := r.sheet.Sprite("fish"), r.sheet.Sprite("eel")
	for i := range fb.Swimmers {
		sw := &fb.Swimmers[i]
		if sw.Off {
			continue
		}
		if IsEel(i) {
			drawTinted(dst, eel.Frame((fb.EelFrame+i)%4), sw.X, sw.Y, false, ink)
			continue
		}
		entity.DrawAt(dst, fish.Frame(fb.FishFrame^(i&1)), sw.X, sw.Y, sw.HDir < 0)
	}
	bubble := r.sheet.Sprite("bubble").Frame(0)
	for _, b := range fb.Bubbles {
		entity.DrawAt(dst, bubble, b.X+config.FreddyW/2, b.Y, false)
	}
	f := fb.Freddy
	drawTinted(dst, r.sheet.Sprite("freddy_swim").Frame(swimFrame(&f.Swim)), f.X, f.Y, f.Facing < 0, ink)
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
