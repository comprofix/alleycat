package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/rooms"
)

// stubRoomTicks is how long a stub room shows before Freddy comes back out.
const stubRoomTicks = 36

// StubRoom stands in for a room until it is built (M4–M6). It names the
// room and sends Freddy back out of the window he came in by, as if he had
// left without clearing it.
type StubRoom struct {
	room  rooms.Kind
	x, y  int // where Freddy went in
	ticks int
}

func newStubRoom(room rooms.Kind, x, y int) *StubRoom {
	return &StubRoom{room: room, x: x, y: y}
}

// Update implements Scene.
func (r *StubRoom) Update(ctx *Context) error {
	r.ticks += ctx.Ticks
	if r.ticks < stubRoomTicks {
		return nil
	}
	if r.room == rooms.Heart {
		// Failing the heart room costs no life and shows no word; its
		// rejection scene comes in M6 (DESIGN §3.5).
		ctx.Switch(newReturnPlay(ctx, r.x, r.y))
		return nil
	}
	ctx.Switch(newResult(resultOops, r.x, r.y))
	return nil
}

// Draw implements Scene.
func (r *StubRoom) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Navy])
	f := ctx.Font
	f.DrawCentred(dst, r.room.String(), 80, gfx.Yellow, 2)
	f.DrawCentred(dst, "(room not built yet)", 110, gfx.Silver, 1)
}

// resultWord is the word shown on returning to the alley (DESIGN §3.7).
type resultWord int

const (
	resultOops resultWord = iota // left without losing a life
)

var resultText = [...]string{"OOPS!"}

// resultInks are the colours the word may appear in (DESIGN §3.7).
var resultInks = []int{gfx.White, gfx.Cyan, gfx.Red}

// Result shows the result word for 28 ticks, then returns Freddy to the
// alley (DESIGN §3.8, "Returning from a room").
type Result struct {
	word  resultWord
	ink   int
	x, y  int
	ticks int
}

func newResult(word resultWord, x, y int) *Result {
	return &Result{word: word, x: x, y: y, ink: -1}
}

// Update implements Scene.
func (r *Result) Update(ctx *Context) error {
	if r.ink < 0 {
		r.ink = resultInks[ctx.RNG.IntN(len(resultInks))]
	}
	r.ticks += ctx.Ticks
	if r.ticks >= config.ResultTicks {
		ctx.Switch(newReturnPlay(ctx, r.x, r.y))
	}
	return nil
}

// Draw implements Scene.
func (r *Result) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	if r.ink >= 0 {
		ctx.Font.DrawCentred(dst, resultText[r.word], 92, r.ink, 2)
	}
}
