package entity

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/physics"
)

// Drawing for the shared actors. It reads state only. Sprites come from the
// alley sheet, which holds Freddy and the dog (DESIGN §8).

func freddyFrame(sheet *gfx.Sheet, f *Freddy) *ebiten.Image {
	switch {
	case f.Seat != NotSeated:
		return sheet.Sprite("freddy_sit").Frame(0)
	case f.Arc.Dir == physics.Rising:
		return sheet.Sprite("freddy_jump").Frame(0)
	case f.Arc.Dir == physics.Falling:
		return sheet.Sprite("freddy_jump").Frame(1)
	case f.Dir != 0:
		return sheet.Sprite("freddy_walk").Frame(f.WalkFrame)
	}
	return sheet.Sprite("freddy_stand").Frame(0)
}

// DrawFreddy draws Freddy. While he walks in he shows at the screen edge;
// behind the alley fence, rows from config.FenceHideY down are hidden.
func DrawFreddy(dst *ebiten.Image, sheet *gfx.Sheet, f *Freddy) {
	if f.Hidden() || f.Gone {
		return
	}
	img := freddyFrame(sheet, f)
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	x := f.X
	if px, entering := f.EnterSlice(); entering {
		if px == 0 {
			return
		}
		x = EdgeX(f.Facing > 0, px, w)
	}
	if f.Depth == Behind {
		vis := config.FenceHideY - f.Y
		if vis <= 0 {
			return
		}
		if vis < h {
			img = img.SubImage(image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Min.Y+vis)).(*ebiten.Image)
		}
	}
	DrawAt(dst, img, x, f.Y, f.Facing < 0)
}

// DrawDog draws the dog: running, or the fight cloud after a catch, sliding
// in or out at the screen edge. coat picks the running sprite: "dog_run"
// (white, the alley dog) or "dog_run_orange" (the room dog).
func DrawDog(dst *ebiten.Image, sheet *gfx.Sheet, d *Dog, coat string) {
	if !d.Out() {
		return
	}
	img := sheet.Sprite(coat).Frame(d.Frame)
	flip := d.Dir < 0
	if d.Capture != 0 {
		img = sheet.Sprite("dog_fight").Frame(0)
		flip = d.Frame%2 == 1
	}
	x := d.X
	w := img.Bounds().Dx()
	switch {
	case d.Entering > 0:
		x = EdgeX(d.Dir > 0, (config.DogEdgeSteps+1-d.Entering)*8, w)
	case d.Exiting > 0:
		x = EdgeX(d.X < gfx.ScreenWidth/2, (d.Exiting-1)*8, w)
	}
	y := config.DogY
	if d.Y != 0 {
		y = d.Y
	}
	DrawAt(dst, img, x, y, flip)
}

// DrawAt draws img with its top-left at x, y, mirrored when flip is set.
func DrawAt(dst, img *ebiten.Image, x, y int, flip bool) {
	op := &ebiten.DrawImageOptions{}
	if flip {
		op.GeoM.Scale(-1, 1)
		op.GeoM.Translate(float64(img.Bounds().Dx()), 0)
	}
	op.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(img, op)
}

// EdgeX places a sprite of width w so that px of it show at the screen edge
// it is entering or leaving by.
func EdgeX(fromLeft bool, px, w int) int {
	if fromLeft {
		return px - w
	}
	return gfx.ScreenWidth - px
}
