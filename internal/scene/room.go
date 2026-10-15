package scene

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/debug"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/rooms"
)

// RoomScene runs a room (DESIGN §3.4) and sends Freddy on according to how
// it ends: the tally, "OOPS!", or a lost life.
type RoomScene struct {
	paused bool
	room   rooms.Room
	view   *rooms.Renderer
	x, y   int // where Freddy went in, in the alley
}

func newRoomScene(ctx *Context, room rooms.Room, x, y int) *RoomScene {
	return &RoomScene{room: room, view: rooms.NewRenderer(ctx.Rooms, ctx.Alley), x: x, y: y}
}

// Update implements Scene.
func (r *RoomScene) Update(ctx *Context) error {
	if handleGameKeys(ctx, &r.paused) || r.paused {
		return nil
	}
	r.room.Update(ctx.Ticks, ctx.FreddySteps, controls(ctx.Input), ctx.RNG)
	b, s := r.room.Shared(), ctx.Session
	switch b.Out {
	case rooms.Cleared:
		bonus := rooms.Bonus(r.room.Kind(), b.Ticks)
		s.AddScore(bonus)
		s.RoomsCleared++
		s.FeliciaWaiting = true
		ctx.Switch(newTally(bonus, rooms.TimeBonus(r.room.Kind(), b.Ticks), r.x, r.y))
	case rooms.LeftEarly:
		ctx.Switch(newResult(resultOops, returnTo(r.x, r.y)))
	case rooms.Failed:
		s.LoseLife()
		if b.ByDog {
			ctx.Switch(newHereKitty(afterLostLife))
		} else {
			ctx.Switch(newResult(resultLost, afterLostLife))
		}
	}
	return nil
}

// Draw implements Scene.
func (r *RoomScene) Draw(ctx *Context, dst *ebiten.Image) {
	r.view.Draw(dst, r.room)
	if debug.Enabled {
		r.view.DrawDebug(dst, r.room)
		ctx.Font.Draw(dst, fmt.Sprintf("TPS %.1f FPS %.1f", ebiten.ActualTPS(), ebiten.ActualFPS()), 4, 190, gfx.White, 1)
	}
	if r.paused {
		drawPaws(ctx, dst)
	}
}

// returnTo is the alley with Freddy back where he went in.
func returnTo(x, y int) func(*Context) Scene {
	return func(ctx *Context) Scene { return newReturnPlay(ctx, x, y) }
}

// afterLostLife: game over at 0 lives (straight to the title), otherwise
// Freddy walks back into the alley from the side (DESIGN §3.4, "How a room
// ends").
func afterLostLife(ctx *Context) Scene {
	if ctx.Session.Lives == 0 {
		ctx.HighScore = max(ctx.HighScore, ctx.Session.Score)
		return NewTitle()
	}
	return NewPlay(ctx)
}

// stubRoomTicks is how long a stub room shows before Freddy comes back out.
const stubRoomTicks = 36

// StubRoom stands in for a room until it is built (M5–M6). It names the
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
	ctx.Switch(newResult(resultOops, returnTo(r.x, r.y)))
	return nil
}

// Draw implements Scene.
func (r *StubRoom) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Navy])
	f := ctx.Font
	f.DrawCentred(dst, r.room.String(), 80, gfx.Yellow, 2)
	f.DrawCentred(dst, "(room not built yet)", 110, gfx.Silver, 1)
}

// resultWord is the word shown on leaving a room (DESIGN §3.7).
type resultWord int

const (
	resultOops resultWord = iota // left without losing a life
	resultLost                   // a life was lost: one of four words
)

// lostWords are picked at random when a life is lost (owner decision).
var lostWords = []string{"OUCH!", "ALAS!", "PITY!", "ARGH!"}

// resultInks are the colours the word may appear in (DESIGN §3.7).
var resultInks = []int{gfx.White, gfx.Cyan, gfx.Red}

// Result shows the result word for 28 ticks, then moves on (DESIGN §3.8,
// "Returning from a room").
type Result struct {
	word  resultWord
	text  string
	ink   int
	next  func(*Context) Scene
	ticks int
}

func newResult(word resultWord, next func(*Context) Scene) *Result {
	return &Result{word: word, next: next, ink: -1}
}

// Update implements Scene.
func (r *Result) Update(ctx *Context) error {
	if r.ink < 0 {
		r.ink = resultInks[ctx.RNG.IntN(len(resultInks))]
		r.text = "OOPS!"
		if r.word == resultLost {
			r.text = lostWords[ctx.RNG.IntN(len(lostWords))]
		}
	}
	r.ticks += ctx.Ticks
	if r.ticks >= config.ResultTicks {
		ctx.Switch(r.next(ctx))
	}
	return nil
}

// Draw implements Scene.
func (r *Result) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	if r.ink >= 0 {
		ctx.Font.DrawCentred(dst, r.text, 92, r.ink, 2)
	}
}
