package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// underwater returns a fishbowl room with Freddy just dived in.
func underwater(d int) *FishbowlRoom {
	r := newFishbowl(d, rng())
	r.dive(rng())
	return r
}

// clearWater puts every swimmer out of Freddy's way, at the far bottom
// right, keeping which are on.
func clearWater(r *FishbowlRoom) {
	for i := range r.Swimmers {
		r.Swimmers[i].X, r.Swimmers[i].Y = config.SwimmerMaxX, laneTop(i)
	}
	r.Freddy.X, r.Freddy.Y = 100, 60
}

// onFreddy puts swimmer i where Freddy is.
func onFreddy(r *FishbowlRoom, i int) {
	r.Swimmers[i].X, r.Swimmers[i].Y = r.Freddy.X+4, r.Freddy.Y+4
	r.Swimmers[i].Off = false
}

func TestTouchingTheBowlDivesIn(t *testing.T) {
	r := newFishbowl(0, rng())
	f := r.Freddy
	f.X, f.Y, f.Arc = 224, 136, physics.Arc{}
	r.Update(0, 1, entity.Controls{}, rng())
	if r.Phase != Underwater {
		t.Fatal("standing on the bowl table should dive him into the bowl")
	}
	if f.X != config.UnderwaterStartX || f.Y != config.UnderwaterStartY {
		t.Errorf("he should start at the surface, at %d,%d", f.X, f.Y)
	}
}

func TestTheBroomStaysInTheRoom(t *testing.T) {
	r := underwater(0)
	bx, by := r.Broom.X, r.Broom.Y
	ticks := r.Ticks
	r.Update(20, 0, entity.Controls{}, rng())
	if r.Broom.X != bx || r.Broom.Y != by {
		t.Error("the broom should not move while Freddy is underwater")
	}
	if r.Ticks != ticks+20 {
		t.Error("the room's time should keep counting underwater")
	}
}

func TestSleepingEelsByDifficulty(t *testing.T) {
	for d := range 8 {
		for seed := range uint64(20) {
			r := newFishbowl(d, rng())
			r.dive(rand.New(rand.NewPCG(seed, 7)))
			on := 0
			for i, s := range r.Swimmers {
				if !IsEel(i) && s.Off {
					t.Fatalf("difficulty %d: fish %d starts eaten", d, i)
				}
				if IsEel(i) && !s.Off {
					on++
				}
			}
			if want := config.ForLevel(d).ActiveEels; on != want {
				t.Fatalf("difficulty %d: %d eels awake, want %d", d, on, want)
			}
		}
	}
}

func TestSwimmersMoveInTwoHalves(t *testing.T) {
	r := underwater(7)
	before := r.Swimmers
	r.Update(1, 0, entity.Controls{}, rng())
	for i := range r.Swimmers {
		moved := r.Swimmers[i] != before[i]
		if want := i >= 1 && i <= 12 && !before[i].Off; moved != want {
			t.Errorf("first tick: slot %d moved %v", i, moved)
		}
	}
	before = r.Swimmers
	r.Update(1, 0, entity.Controls{}, rng())
	for i := range r.Swimmers {
		moved := r.Swimmers[i] != before[i]
		if want := (i == 0 || i >= 13) && !before[i].Off; moved != want {
			t.Errorf("second tick: slot %d moved %v", i, moved)
		}
	}
}

func TestSwimmersStayInTheirLanes(t *testing.T) {
	r := underwater(7)
	g := rand.New(rand.NewPCG(3, 3))
	for range 5000 {
		r.swimmersTick(g)
		for i, s := range r.Swimmers {
			top := laneTop(i)
			if s.X < 0 || s.X > config.SwimmerMaxX || s.Y < top || s.Y > top+config.LaneDepth {
				t.Fatalf("slot %d out of its lane at %d,%d", i, s.X, s.Y)
			}
		}
	}
}

func TestEatingAFishWakesAnEelAwayFromFreddy(t *testing.T) {
	for _, c := range []struct{ x, wantX, wantDir int }{{100, config.SwimmerMaxX, -1}, {200, 0, 1}} {
		r := underwater(0)
		clearWater(r)
		r.Freddy.X = c.x
		onFreddy(r, 3)
		r.contacts()
		if !r.Swimmers[3].Off || r.FishLeft != config.FishCount-1 {
			t.Fatalf("the fish should be eaten: off %v, %d left", r.Swimmers[3].Off, r.FishLeft)
		}
		woken := -1
		for i := swimmers - 1; i >= config.FishCount; i-- {
			if !r.Swimmers[i].Off && r.Swimmers[i].X == c.wantX && r.Swimmers[i].HDir == c.wantDir {
				woken = i
				break
			}
		}
		if woken < 0 {
			t.Errorf("Freddy at X %d: no eel woke at X %d heading %d", c.x, c.wantX, c.wantDir)
		}
	}
}

func TestAnEelZapsFreddy(t *testing.T) {
	r := underwater(0)
	clearWater(r)
	onFreddy(r, 20)
	r.contacts()
	if r.Phase != Zapped {
		t.Fatal("touching an eel should zap him")
	}
	for tick := 1; tick <= config.ZapTicks; tick++ {
		r.Update(1, 2, entity.Controls{}, rng())
		if failed := r.Out == Failed; failed != (tick == config.ZapTicks) {
			t.Fatalf("tick %d of ZAP!: out %d", tick, r.Out)
		}
	}
}

func TestEelsAreHarmlessWhileDrowning(t *testing.T) {
	r := underwater(0)
	clearWater(r)
	r.Drowning = true
	onFreddy(r, 20)
	r.contacts()
	if r.Phase != Underwater || !r.Swimmers[20].Off {
		t.Errorf("a drowning Freddy should brush the eel away: phase %d, off %v", r.Phase, r.Swimmers[20].Off)
	}
}

func TestTwelveFishClearTheRoom(t *testing.T) {
	r := underwater(0)
	for i := range config.FishCount {
		clearWater(r)
		for j := config.FishCount; j < swimmers; j++ {
			r.Swimmers[j].Off = true
		}
		onFreddy(r, i)
		r.contacts()
		if cleared := r.Out == Cleared; cleared != (i == config.FishCount-1) {
			t.Fatalf("after fish %d: out %d", i, r.Out)
		}
	}
}

func TestAirRunsOut(t *testing.T) {
	for d := range 8 {
		r := underwater(d)
		for i := range r.Swimmers {
			r.Swimmers[i].Off = true
		}
		tb := config.ForLevel(d)
		stages := 0
		drownedAt := -1
		for r.Out == Playing && r.Ticks < 1000 {
			r.Update(1, 2, entity.Controls{V: 1}, rng())
			if s := r.AirStage(); s > stages {
				if r.AirTicks() != config.AirStages[d][s-1] {
					t.Errorf("difficulty %d: stage %d at %d ticks", d, s, r.AirTicks())
				}
				stages = s
			}
			if r.Drowning && drownedAt < 0 {
				drownedAt = r.AirTicks()
			}
		}
		if stages != 3 || drownedAt != tb.DrowningTicks || r.Out != Failed || r.AirTicks() != tb.AirFailTicks {
			t.Errorf("difficulty %d: stages %d, drowning at %d, out %d at %d", d, stages, drownedAt, r.Out, r.AirTicks())
		}
	}
}

func TestSurfacingRefillsTheAir(t *testing.T) {
	r := underwater(0)
	for i := range r.Swimmers {
		r.Swimmers[i].Off = true
	}
	for range 60 {
		r.Update(1, 2, entity.Controls{V: 1}, rng())
	}
	for range 100 {
		r.Update(1, 2, entity.Controls{V: -1}, rng())
	}
	if a := r.AirTicks(); a > 2 {
		t.Errorf("swimming up to the surface should refill his air, %d ticks since", a)
	}
}

// fishbowlBot plays the room: it walks to the bowl, then goes for the
// nearest fish, dodging eels in the way, and surfaces when the air warning
// reaches stage 2.
func fishbowlBot(r *FishbowlRoom) entity.Controls {
	f := r.Freddy
	if r.Phase == InRoom {
		in := entity.Controls{H: 1}
		if f.Standing() && f.Y == config.FloorY && f.X >= 184 && f.X <= 200 {
			in.V = -1
		}
		return in
	}
	if r.AirStage() >= 2 || (f.Y <= config.SwimTopY && r.AirTicks() > 2) {
		return entity.Controls{V: -1}
	}
	best, bestD := -1, 1<<30
	for i := range config.FishCount {
		if s := r.Swimmers[i]; !s.Off {
			if d := abs(s.X-f.X-8) + abs(s.Y-f.Y-4); d < bestD {
				best, bestD = i, d
			}
		}
	}
	if best < 0 {
		return entity.Controls{}
	}
	s := r.Swimmers[best]
	in := entity.Controls{H: sign(s.X - f.X - 8), V: sign(s.Y - f.Y - 4)}
	// Dodge: an eel near his path makes him hold back and move off its row.
	ahead := physics.Rect{X: f.X - 16, Y: f.Y - 10, W: config.FreddyHitW + 32, H: config.FreddyHitH + 20}
	for i := config.FishCount; i < swimmers; i++ {
		if !r.Swimmers[i].Off && r.SwimmerRect(i).Overlaps(ahead) {
			e := r.Swimmers[i]
			in.H = sign(f.X - e.X)
			in.V = 1
			if e.Y > f.Y+4 {
				in.V = -1
			}
		}
	}
	return in
}

// TestFishbowlRoomCanBeCleared plays the room headless with a fixed seed,
// from falling in at the window to the 12th fish.
func TestFishbowlRoomCanBeCleared(t *testing.T) {
	g := rand.New(rand.NewPCG(0, 2))
	r := newFishbowl(0, g)
	const limit = 3 * 60 * 182 / 10
	for tick := range limit {
		r.Update(1, 2, fishbowlBot(r), g)
		switch r.Out {
		case Cleared:
			t.Logf("cleared after %d ticks", tick+1)
			return
		case Playing:
		default:
			t.Fatalf("the room ended with %d after %d ticks (phase %d, %d fish left)", r.Out, tick+1, r.Phase, r.FishLeft)
		}
	}
	t.Fatalf("not cleared in %d ticks: phase %d, %d fish left, Freddy at %d,%d", limit, r.Phase, r.FishLeft, r.Freddy.X, r.Freddy.Y)
}

// TestFishbowlSoak runs the room with random input to catch panics and
// stuck loops, diving in each time.
func TestFishbowlSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 4))
		r := newFishbowl(d, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1}
			}
			if frame%2000 == 0 && r.Phase == InRoom {
				r.dive(g)
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newFishbowl(d, g)
			}
		}
	}
}
