package scene

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/rooms"
	"github.com/comprofix/alleycat/internal/session"
)

func playContext() *Context {
	return &Context{
		Input:   input.New(&input.Script{}),
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
	rs, ok := step(ctx, p).(*RoomScene)
	if !ok || rs.room.Kind() != rooms.Heart || ctx.Session.FeliciaWaiting {
		t.Fatalf("Felicia's window: got %T, waiting=%v", rs, ctx.Session.FeliciaWaiting)
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

// heartScene enters Felicia's window and returns the heart room scene.
func heartScene(t *testing.T, ctx *Context) *RoomScene {
	t.Helper()
	p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
	p.alley.Exit = alley.ToHeart
	p.alley.Freddy.X, p.alley.Freddy.Y = 104, 50
	rs, ok := step(ctx, p).(*RoomScene)
	if !ok {
		t.Fatal("Felicia's window should lead to a room scene")
	}
	if _, ok := rs.room.(*rooms.HeartRoom); !ok {
		t.Fatalf("Felicia's window should lead to the heart room, got %T", rs.room)
	}
	return rs
}

func TestFallingOffTheHeartsCostsNoLife(t *testing.T) {
	ctx := playContext()
	rs := heartScene(t, ctx)
	rs.room.Shared().Out = rooms.Failed
	rej, ok := step(ctx, rs).(*Reject)
	if !ok {
		t.Fatal("falling off the hearts should play the rejection scene")
	}
	if _, ok := runUntil(t, ctx, rej).(*Play); !ok {
		t.Fatal("after the rejection Freddy should be back in the alley")
	}
	if ctx.Session.Lives != config.StartLives {
		t.Error("failing the heart room must not cost a life")
	}
}

func TestKissingFeliciaLevelsUp(t *testing.T) {
	ctx := playContext()
	s := ctx.Session
	s.RoomsCleared, s.LevelTicks = 3, 800
	d := s.Difficulty
	rs := heartScene(t, ctx)
	hr := rs.room.(*rooms.HeartRoom)
	hr.Out, hr.WithGift = rooms.Cleared, true
	kiss, ok := step(ctx, rs).(*Kiss)
	if !ok {
		t.Fatal("reaching Felicia should lead to the kiss")
	}
	if kiss.multiplier != 2*(2*d+1) {
		t.Errorf("multiplier %d, want %d (doubled with a gift)", kiss.multiplier, 2*(2*d+1))
	}
	if _, ok := runUntil(t, ctx, kiss).(*Play); !ok {
		t.Fatal("after the kiss Freddy should be back in the alley")
	}
	if want := kiss.bonus * kiss.multiplier; s.Score != want {
		t.Errorf("score %d, want the kiss bonus %d × %d", s.Score, kiss.bonus, kiss.multiplier)
	}
	if s.Difficulty != d+1 || s.Lives != config.StartLives+1 || s.RoomsCleared != 0 || s.LevelTicks != 0 {
		t.Errorf("after the kiss: difficulty %d, lives %d, rooms %d, level ticks %d", s.Difficulty, s.Lives, s.RoomsCleared, s.LevelTicks)
	}
	// The big heart dropped to the middle in 8 steps, and three bursts
	// followed, each of 8 hearts.
	hearts, small := 0, 0
	for _, st := range kiss.stamps {
		if st.cats {
			hearts++
		}
		if st.sprite == "heart_small" {
			small++
		}
	}
	if hearts != 8 || kiss.y != config.KissHeartStopY {
		t.Errorf("the big heart left %d copies and stopped at Y %d", hearts, kiss.y)
	}
	if small == 0 || (small-1)%8 != 0 {
		t.Errorf("the first burst left %d small hearts: one, then 8 a step", small)
	}
	if kiss.ticks < config.KissOpenTicks+config.KissTextTicks {
		t.Errorf("the kiss took %d ticks, too short", kiss.ticks)
	}
}

func TestDancingCatsByDifficulty(t *testing.T) {
	want := []int{0, 0, 4, 8, 16, 24, 26, 34}
	for d, n := range want {
		dance := newDance(d, func(*Context) Scene { return nil })
		switch {
		case n == 0 && dance != nil:
			t.Errorf("difficulty %d: no dancing below 2", d)
		case n > 0 && (dance == nil || len(dance.dancers) != n):
			t.Errorf("difficulty %d: want %d dancing cats", d, n)
		}
	}
}

func TestDanceLastsFortyTicks(t *testing.T) {
	ctx := playContext()
	dance := newDance(2, func(ctx *Context) Scene { return NewPlay(ctx) })
	for tick := 1; tick <= config.DanceTicks; tick++ {
		next := step(ctx, dance)
		if (next != nil) != (tick == config.DanceTicks) {
			t.Fatalf("tick %d: next %T", tick, next)
		}
	}
}

func TestKissToDifficultyTwoDances(t *testing.T) {
	ctx := playContext()
	ctx.Session.Difficulty = 1
	rs := heartScene(t, ctx)
	rs.room.Shared().Out = rooms.Cleared
	kiss := step(ctx, rs).(*Kiss)
	dance, ok := runUntil(t, ctx, kiss).(*Dance)
	if !ok {
		t.Fatal("a kiss up to difficulty 2 should lead to the dancing cats")
	}
	if _, ok := runUntil(t, ctx, dance).(*Play); !ok {
		t.Fatal("after the dancing cats Freddy should be back in the alley")
	}
}

func TestRejectionFallsThenHolds(t *testing.T) {
	ctx := playContext()
	rej := newReject(func(ctx *Context) Scene { return NewPlay(ctx) })
	fall := (config.RejectLandY + config.RejectFallStep - 1) / config.RejectFallStep
	for tick := 1; tick <= fall+config.RejectHoldTicks; tick++ {
		next := step(ctx, rej)
		if landed := tick >= fall; rej.landed != landed {
			t.Fatalf("tick %d: landed %v", tick, rej.landed)
		}
		if (next != nil) != (tick == fall+config.RejectHoldTicks) {
			t.Fatalf("tick %d: next %T", tick, next)
		}
	}
}
