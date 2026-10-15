package scene

import (
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
)

type noKeys struct{}

func (noKeys) AppendPressedKeys(k []ebiten.Key) []ebiten.Key { return k }

func playContext() *Context {
	return &Context{
		Input:   input.New(noKeys{}),
		RNG:     rand.New(rand.NewPCG(1, 2)),
		Session: session.New(config.Kitten),
	}
}

// step runs a scene for one frame with one master tick and takes its
// requested switch.
func step(ctx *Context, s Scene) Scene {
	ctx.Input.Update()
	ctx.Ticks, ctx.FreddySteps = 1, 1
	if err := s.Update(ctx); err != nil {
		panic(err)
	}
	next := ctx.next
	ctx.next = nil
	return next
}

func TestWindowLeadsToRoomAndBack(t *testing.T) {
	ctx := playContext()
	p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
	p.alley.Exit = alley.ToRoom
	p.alley.Freddy.X, p.alley.Freddy.Y = 104, 50

	room, ok := step(ctx, p).(*StubRoom)
	if !ok {
		t.Fatal("entering a window should go to a room")
	}
	var next Scene
	for range 100 {
		if next = step(ctx, room); next != nil {
			break
		}
	}
	res, ok := next.(*Result)
	if !ok || res.word != resultOops {
		t.Fatalf("leaving the room should show OOPS!, got %T", next)
	}
	for range 100 {
		if next = step(ctx, res); next != nil {
			break
		}
	}
	back, ok := next.(*Play)
	if !ok {
		t.Fatalf("after the result word: got %T, want the alley", next)
	}
	if f := back.alley.Freddy; f.X != 104 || f.Y != 50 {
		t.Errorf("Freddy back at %d,%d, want 104,50", f.X, f.Y)
	}
	if ctx.Session.Lives != config.StartLives {
		t.Error("leaving a room early must not cost a life")
	}
}

func TestFeliciaWindowLeadsToHeartRoom(t *testing.T) {
	ctx := playContext()
	ctx.Session.FeliciaWaiting = true
	p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
	p.alley.Exit = alley.ToHeart
	room, ok := step(ctx, p).(*StubRoom)
	if !ok || room.room.String() != "FELICIA" || ctx.Session.FeliciaWaiting {
		t.Fatalf("Felicia's window: got %T, waiting=%v", room, ctx.Session.FeliciaWaiting)
	}
}

func TestGameOverGoesToTitle(t *testing.T) {
	ctx := playContext()
	ctx.Session.Score = 500
	p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
	ctx.Session.Lives = 0
	if _, ok := step(ctx, p).(*Title); !ok {
		t.Fatal("at 0 lives the game should go back to the title")
	}
	if ctx.HighScore != 500 {
		t.Errorf("high score %d, want 500", ctx.HighScore)
	}
}
