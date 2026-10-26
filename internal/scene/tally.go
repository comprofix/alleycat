package scene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Tally screen layout: columns of cat faces either side, rows 8 px apart
// from the bottom up (DESIGN §3.6, "The tally screen").
var tallyColumns = []int{16, 32, 272, 288}

const (
	tallyBottomY = 176
	tallyRowH    = 8
)

// Tally is the room-clear screen: the cat faces light up from the bottom, as
// high as the time bonus reaches, then the bonus shows and the screen
// flashes (DESIGN §3.6).
type Tally struct {
	bonus int
	stopY int // the highest row that lights
	lit   int // rows lit so far
	ticks int
	held  int
	x, y  int // where Freddy went in, in the alley

	art *tallyArt // the new look's pictures; a render cache
}

func newTally(bonus, timeBonus, x, y int) *Tally {
	stop := ((config.TallyDogsTimeWindow - timeBonus) >> 4) & 0xf0
	return &Tally{bonus: bonus, stopY: stop, x: x, y: y}
}

func (t *Tally) rows() int { return (tallyBottomY-t.stopY)/tallyRowH + 1 }

// Update implements Scene.
func (t *Tally) Update(ctx *Context) error {
	for range ctx.Ticks {
		if t.lit < t.rows() {
			if t.ticks++; t.ticks%config.TallyRowTicks == 0 {
				t.lit++
				// Each row plays the next note of a rising series (DESIGN §7.4).
				ctx.Audio.PlayMusic(audio.TallyNote(t.lit-1, config.TallyRowTicks), false)
			}
			continue
		}
		if t.held == 0 {
			// Then a trill while the screen flashes.
			ctx.Audio.PlayMusic(audio.TallyTrill(t.lit-1, config.TallyHoldTicks), false)
		}
		if t.held++; t.held >= config.TallyHoldTicks {
			ctx.Switch(newReturnPlay(ctx, t.x, t.y))
			return nil
		}
	}
	return nil
}

// Draw implements Scene: in the look dst is the size of.
func (t *Tally) Draw(ctx *Context, dst *ebiten.Image) {
	if smoothScreen(dst) {
		t.drawSmooth(dst)
		return
	}
	bg := resurrect64.Ink
	if t.lit >= t.rows() && t.held/4%2 == 1 {
		bg = resurrect64.Navy // the flashing at the end
	}
	dst.Fill(gfx.Palette[bg])
	face := ctx.Rooms.Sprite("tally_face")
	for row, y := 0, tallyBottomY; y >= 0; row, y = row+1, y-tallyRowH {
		frame := 1
		if row < t.lit {
			frame = 0
		}
		for _, x := range tallyColumns {
			entity.DrawAt(dst, face.Frame(frame), x, y, false)
		}
	}
	if t.lit >= t.rows() {
		// The bonus, big and shadowed (M8g), centred between the columns.
		text := fmt.Sprintf("%04d", t.bonus%10000)
		ctx.PixelFont.DrawShadowed(dst, text, gfx.CentreX(text, 2), t.stopY, resurrect64.Honey, resurrect64.Ink, 2)
	}
}

// HereKitty is the dog-capture scene in a room: a dog standing on one hind
// leg beckons, and the words appear, slanting (DESIGN §3.4, "HERE KITTY").
// Our own art.
type HereKitty struct {
	ticks  int
	played bool
	next   func(*Context) Scene

	art *kittyArt // the new look's pictures; a render cache
}

func newHereKitty(next func(*Context) Scene) *HereKitty { return &HereKitty{next: next} }

// Update implements Scene.
func (h *HereKitty) Update(ctx *Context) error {
	if !h.played {
		h.played = true
		ctx.Audio.PlayMusic(audio.Failure(), false) // DESIGN §7.4
	}
	if h.ticks += ctx.Ticks; h.ticks >= config.HereKittyTicks {
		ctx.Switch(h.next(ctx))
	}
	return nil
}

// Draw implements Scene: in the look dst is the size of.
func (h *HereKitty) Draw(ctx *Context, dst *ebiten.Image) {
	if smoothScreen(dst) {
		h.drawSmooth(dst)
		return
	}
	dst.Fill(gfx.Palette[resurrect64.Ink])
	frame := h.ticks / 4 // 8 frames of 4 ticks
	entity.DrawAt(dst, ctx.Rooms.Sprite("kitty_dog").Frame(frame%2), kittyDogX, kittyDogY, false)
	for _, w := range kittyWords {
		if frame >= w.frame {
			drawSlanted(ctx, dst, w.text, w.x, w.y)
		}
	}
}

// HERE KITTY layout, from the HD video for layout only. The words step
// evenly down and to the right (owner request, 2026-10-16).
const kittyDogX, kittyDogY = 144, 66

var kittyWords = []struct {
	text  string
	x, y  int
	frame int
}{
	{"HERE", 160, 50, 2},
	{"KITTY", 180, 72, 4},
	{"KITTY", 200, 94, 5},
	{"KITTY", 220, 116, 6},
}

// drawSlanted writes text in cyan with each letter a little higher than the
// last, like the original's hand lettering.
func drawSlanted(ctx *Context, dst *ebiten.Image, text string, x, y int) {
	const rise = 3
	for i, r := range text {
		ctx.PixelFont.Draw(dst, string(r), x+i*8, y-i*rise, resurrect64.Seafoam, 1)
	}
}
