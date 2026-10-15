package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
)

// Sleeping dogs room layout, measured from the reference for layout only
// (DESIGN §3.4, "Sleeping dogs room layout"). The art is our own.
const (
	dogsWallY0, dogsWallY1     = 16, 112
	dogsOpeningX, dogsOpeningY = 56, 48
	dogsChairLX, dogsChairRX   = 72, 216
	dogsChairY                 = 88
	dogsPictureX, dogsPictureY = 248, 40
	dogsLampX, dogsLampY       = 272, 72
	eatFrameFull               = 128 // the mouth closes from this accumulator value
)

// buildDogs draws the sleeping dogs room's backdrop: the taller walls, the
// window with its curtains, the chairs, the picture and the lamp.
func (r *Renderer) buildDogs() {
	b := r.back
	r.buildWalls(dogsWallY0, dogsWallY1)
	r.buildWindow(dogsOpeningX, dogsOpeningY, "curtain_teal")
	chair := r.sheet.Sprite("chair_teal").Frame(0)
	entity.DrawAt(b, chair, dogsChairLX, dogsChairY, false)
	entity.DrawAt(b, chair, dogsChairRX, dogsChairY, true)
	entity.DrawAt(b, r.sheet.Sprite("picture").Frame(0), dogsPictureX, dogsPictureY, false)
	entity.DrawAt(b, r.sheet.Sprite("lamp_teal").Frame(0), dogsLampX, dogsLampY, false)
}

// drawDogs draws the bowls and the sleeping dogs, row by row from the back.
// An empty bowl is gone.
func (r *Renderer) drawDogs(dst *ebiten.Image, d *DogsRoom) {
	bowl, dog := r.sheet.Sprite("milk_bowl"), r.sheet.Sprite("sleeper")
	for _, row := range []int{136, 152, 168} {
		for i, b := range config.FoodBowls {
			if level := d.MilkLevel(i); b.Y == row && level > 0 {
				entity.DrawAt(dst, bowl.Frame(level), b.X, b.Y+config.FoodBowlDY, false)
			}
		}
		for i, s := range d.Sleepers {
			if sl := config.Sleepers[i]; s.On && sl.Y == row {
				sr := SleeperRect(i)
				entity.DrawAt(dst, dog.Frame(s.Alert), sr.X, sr.Y, sl.Left)
			}
		}
	}
}

// drawEating draws Freddy's eating pose in place of Freddy, to one side of
// him as the bowl dictates, its mouth moving with the eating.
func (r *Renderer) drawEating(dst *ebiten.Image, d *DogsRoom) {
	f, b := d.Freddy, config.FoodBowls[d.EatBowl]
	x := f.X &^ 3
	if b.Left {
		x = max(x-config.EatPoseDX, 0)
	} else {
		x = min(x+config.EatPoseDX, config.EatPoseMaxX)
	}
	frame := 0
	if d.EatAcc >= eatFrameFull {
		frame = 1
	}
	entity.DrawAt(dst, r.sheet.Sprite("freddy_eat").Frame(frame), x, f.Y+config.EatPoseDY, !b.Left)
}
