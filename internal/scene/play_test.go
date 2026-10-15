package scene

import (
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/rooms"
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

	var next Scene
	switch room := step(ctx, p).(type) {
	case *StubRoom:
		next = runUntil(t, ctx, room)
	case *RoomScene:
		room.room.Shared().Out = rooms.LeftEarly
		next = step(ctx, room)
	default:
		t.Fatalf("entering a window should go to a room, got %T", room)
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

// runUntil steps s until it asks for another scene.
func runUntil(t *testing.T, ctx *Context, s Scene) Scene {
	t.Helper()
	for range 500 {
		if next := step(ctx, s); next != nil {
			return next
		}
	}
	t.Fatalf("%T never moved on", s)
	return nil
}

func birdcageScene(ctx *Context) (*RoomScene, *rooms.BirdcageRoom) {
	r := rooms.New(rooms.Birdcage, 0, ctx.RNG)
	return &RoomScene{room: r, x: 104, y: 50}, r.(*rooms.BirdcageRoom)
}

func TestClearingARoomGoesToTallyAndFelicia(t *testing.T) {
	ctx := playContext()
	rs, r := birdcageScene(ctx)
	r.Out = rooms.Cleared
	r.Ticks = 248 // R1's 5704
	tally, ok := step(ctx, rs).(*Tally)
	if !ok {
		t.Fatal("a cleared room should go to the tally")
	}
	s := ctx.Session
	if s.Score != 5704 || s.RoomsCleared != 1 || !s.FeliciaWaiting {
		t.Errorf("after clearing: score %d, rooms %d, Felicia %v", s.Score, s.RoomsCleared, s.FeliciaWaiting)
	}
	back, ok := runUntil(t, ctx, tally).(*Play)
	if !ok || back.alley.Freddy.X != 104 {
		t.Errorf("after the tally: %T, want the alley with Freddy back at the window", back)
	}
}

func TestDogInARoomCostsALife(t *testing.T) {
	ctx := playContext()
	rs, r := birdcageScene(ctx)
	r.Out, r.ByDog = rooms.Failed, true
	kitty, ok := step(ctx, rs).(*HereKitty)
	if !ok {
		t.Fatal("a dog catch should play HERE KITTY")
	}
	if ctx.Session.Lives != config.StartLives-1 {
		t.Errorf("lives %d, want %d", ctx.Session.Lives, config.StartLives-1)
	}
	back, ok := runUntil(t, ctx, kitty).(*Play)
	if !ok {
		t.Fatal("after HERE KITTY: want the alley")
	}
	if !back.alley.Freddy.Entering() {
		t.Error("after a failed room Freddy should walk back in from the side")
	}
}

func TestLosingTheLastLifeInARoomEndsTheGame(t *testing.T) {
	ctx := playContext()
	ctx.Session.Lives = 1
	rs, r := birdcageScene(ctx)
	r.Out, r.ByDog = rooms.Failed, true
	kitty := step(ctx, rs).(*HereKitty)
	if _, ok := runUntil(t, ctx, kitty).(*Title); !ok {
		t.Error("losing the last life should go back to the title")
	}
}

func TestLeavingARoomEarlyIsOops(t *testing.T) {
	ctx := playContext()
	rs, r := birdcageScene(ctx)
	r.Out = rooms.LeftEarly
	res, ok := step(ctx, rs).(*Result)
	if !ok || res.word != resultOops || ctx.Session.Lives != config.StartLives {
		t.Fatalf("leaving early: %T lives %d", res, ctx.Session.Lives)
	}
	if _, ok := runUntil(t, ctx, res).(*Play); !ok {
		t.Error("after OOPS!: want the alley")
	}
}
