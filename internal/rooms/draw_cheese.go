package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
)

// Cheese room layout, measured from the reference for layout only (DESIGN
// §3.4, "Cheese room layout"). The art is our own.
const (
	wedgeX0, wedgeY0, wedgeY1 = 24, 32, 168
	cheeseOpeningX            = 248
	cheeseChairX              = 228
	pedestalX, pedestalY      = 256, 144
	pitCell                   = 8 // one pit at most per 8×8 cell of the face
)

// wedgeEdge is the wedge's right edge, (Y, X) points joined by straight
// lines: flat at the top, a slope down to the right, then the face.
var wedgeEdge = [][2]int{{32, 120}, {72, 208}, {88, 224}, {160, 224}, {168, 216}}

// wedgeRight is the wedge's right edge at row y.
func wedgeRight(y int) int {
	for i := 1; i < len(wedgeEdge); i++ {
		a, b := wedgeEdge[i-1], wedgeEdge[i]
		if y <= b[0] {
			return a[1] + (b[1]-a[1])*(y-a[0])/(b[0]-a[0])
		}
	}
	return wedgeEdge[len(wedgeEdge)-1][1]
}

// holeDrawAt is where hole h's 16×16 opening is drawn.
func holeDrawAt(h int) (x, y int) {
	s := config.CheeseHoles[h]
	if h < config.CheeseTopHoles {
		return s.X + 8, s.Band
	}
	return s.X + 8, s.Band - 8
}

// buildCheese draws the cheese room's backdrop: the walls, the wedge with
// its pits and holes, the window with its curtains, the chair and the
// pedestal table.
func (r *Renderer) buildCheese() {
	s, b := r.scheme, r.back
	r.buildWalls()
	r.buildWindow(cheeseOpeningX)
	entity.DrawAt(b, r.sheet.Sprite("chair").Frame(0), cheeseChairX, furnitureY, false)
	entity.DrawAt(b, r.sheet.Sprite("pedestal").Frame(0), pedestalX, pedestalY, false)

	for y := wedgeY0; y < wedgeY1; y++ {
		x1 := wedgeRight(y)
		fill(b, wedgeX0, y, x1-wedgeX0, 1, s.Cheese)
		fill(b, x1-1, y, 1, 1, s.Hole)
	}
	fill(b, wedgeX0, wedgeY0, wedgeRight(wedgeY0)-wedgeX0, 1, s.Hole)
	// A fixed scatter of small pits; drawing never uses the game's RNG.
	for y := wedgeY0 + 4; y < wedgeY1-4; y += pitCell {
		for x := wedgeX0 + 4; x < wedgeRight(y)-6; x += pitCell {
			if h := (x*7 + y*13) % 11; h < 3 {
				fill(b, x+h, y+h, 2, 1, s.Hole)
			}
		}
	}
	for h := range config.CheeseHoleCount {
		name := "hole_face"
		if h < config.CheeseTopHoles {
			name = "hole_top"
		}
		x, y := holeDrawAt(h)
		entity.DrawAt(b, r.sheet.Sprite(name).Frame(0), x, y, false)
	}
}

// drawCheese draws the mice in their holes, Freddy going through a hole,
// and the caught mice at the top of the screen.
func (r *Renderer) drawCheese(dst *ebiten.Image, c *CheeseRoom) {
	mouse := r.sheet.Sprite("cheese_mouse")
	for i := range c.Mice {
		m := &c.Mice[i]
		if !m.Visible() {
			continue
		}
		frame := 2 + m.Frame
		switch m.Count {
		case config.MouseHidden - 1, 0:
			frame = 0
		case config.MouseEmerge, 1:
			frame = 1
		}
		mr := c.MouseRect(i)
		entity.DrawAt(dst, mouse.Frame(frame), mr.X, mr.Y, false)
	}
	if t := &c.Transfer; t.Active {
		h := t.From
		if t.Frame >= config.TransferSwapAfter {
			h = t.To
		}
		x, y := holeDrawAt(h)
		entity.DrawAt(dst, r.sheet.Sprite("freddy_hole").Frame(t.Frame), x, y, false)
	}
	icon := r.sheet.Sprite("mouse_icon").Frame(0)
	for i := range c.Caught {
		entity.DrawAt(dst, icon, config.MouseIconX+i*config.MouseIconStep, config.MouseIconY, false)
	}
}
