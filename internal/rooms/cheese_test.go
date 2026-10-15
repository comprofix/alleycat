package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// inHole returns a cheese room with Freddy settled in hole h.
func inHole(d, h int) *CheeseRoom {
	r := newCheese(d, rng())
	putInHole(r, h)
	return r
}

func putInHole(r *CheeseRoom, h int) {
	s, f := config.CheeseHoles[h], r.Freddy
	f.X, f.Y = s.X+config.HoleArriveDX, s.Band
	f.Arc = physics.Arc{}
	f.Seat = entity.NotSeated
	if s.Perch {
		f.Seat = entity.Seated
	}
}

// parkMice hides every mouse far from the action, in holes 12–15.
func parkMice(r *CheeseRoom) {
	for i := range r.Mice {
		r.Mice[i] = Mouse{Hole: 12 + i, Count: 100}
	}
}

func TestCheesePairsJoinEveryHoleOnce(t *testing.T) {
	for d := range 8 {
		r := newCheese(d, rng())
		seen := map[int]bool{}
		for h := range config.CheeseHoleCount {
			p := r.Partner(h)
			if p == h || r.Partner(p) != h {
				t.Errorf("difficulty %d: hole %d leads to %d, which leads to %d", d, h, p, r.Partner(p))
			}
			seen[p] = true
		}
		if len(seen) != config.CheeseHoleCount {
			t.Errorf("difficulty %d: only %d holes are reachable", d, len(seen))
		}
	}
}

func TestEveryHoleIsFound(t *testing.T) {
	r := newCheese(0, rng())
	for h := range config.CheeseHoleCount {
		putInHole(r, h)
		if got := r.HoleAt(r.Freddy); got != h {
			t.Errorf("Freddy in hole %d: HoleAt = %d", h, got)
		}
	}
	r.Freddy.X, r.Freddy.Y = 150, config.FloorY
	if got := r.HoleAt(r.Freddy); got != -1 {
		t.Errorf("on the floor: HoleAt = %d", got)
	}
}

func TestLandingInAHoleSeatsFreddy(t *testing.T) {
	r := newCheese(0, rng())
	parkMice(r)
	s, f := config.CheeseHoles[9], r.Freddy
	f.EnterRoom(s.X+config.HoleArriveDX, s.Band-20, 0)
	for range 20 {
		r.Update(0, 1, entity.Controls{}, rng())
	}
	if f.Seat == entity.NotSeated || r.HoleAt(f) != 9 {
		t.Fatalf("Freddy should sit in hole 9: seat %d, at %d,%d", f.Seat, f.X, f.Y)
	}
}

func TestTransferMovesFreddyToThePairedHole(t *testing.T) {
	r := inHole(3, 0) // at difficulty 3, hole 0 leads to hole 2; both are in the top
	parkMice(r)
	f, alt := r.Freddy, entity.Controls{Act: true}
	start := config.CheeseHoles[0]

	r.Update(1, 0, alt, rng())
	if !r.Transfer.Active || !f.Gone {
		t.Fatal("Alt in a hole should start a transfer and hide Freddy")
	}
	for frame := 1; frame < config.TransferFrames-1; frame++ {
		r.Update(1, 0, entity.Controls{}, rng())
		if r.Transfer.Frame != frame || !r.Transfer.Active {
			t.Fatalf("tick %d: frame %d, active %v", frame, r.Transfer.Frame, r.Transfer.Active)
		}
		moved := f.X != start.X+config.HoleArriveDX
		if moved != (frame >= config.TransferSwapAfter) {
			t.Fatalf("frame %d: moved %v", frame, moved)
		}
		// Freddy is frozen while hidden.
		r.Update(0, 2, entity.Controls{H: 1}, rng())
	}
	r.Update(1, 0, entity.Controls{}, rng())
	if r.Transfer.Active || f.Gone {
		t.Fatal("the transfer should be over after 7 frames")
	}
	r.Update(0, 1, entity.Controls{}, rng())
	if got := r.HoleAt(f); got != 2 {
		t.Fatalf("Freddy should be in hole 2, is in %d (%d,%d)", got, f.X, f.Y)
	}
}

func TestHoldingAltRepeatsAfterTheLockout(t *testing.T) {
	r := inHole(3, 0)
	parkMice(r)
	alt := entity.Controls{Act: true}
	var starts []int
	for range 40 {
		was := r.Transfer.Active
		r.Update(1, 2, alt, rng())
		if r.Transfer.Active && !was {
			starts = append(starts, r.Ticks)
		}
	}
	if len(starts) < 3 {
		t.Fatalf("holding Alt should keep going through holes, starts %v", starts)
	}
	for i := 1; i < len(starts); i++ {
		if d := starts[i] - starts[i-1]; d != config.TransferLockout {
			t.Errorf("transfers %d ticks apart, want %d (starts %v)", d, config.TransferLockout, starts)
		}
	}
}

func TestNoTransferWithoutAHoleOrAlt(t *testing.T) {
	r := inHole(0, 9)
	parkMice(r)
	r.transferTick(entity.Controls{})
	if r.Transfer.Active {
		t.Error("no Alt, no transfer")
	}
	r.Freddy.X, r.Freddy.Y, r.Freddy.Seat = 150, config.FloorY, entity.NotSeated
	r.transferTick(entity.Controls{Act: true})
	if r.Transfer.Active {
		t.Error("Alt on the floor should do nothing")
	}
}

func TestNoTransferTouchingTheBroom(t *testing.T) {
	r := inHole(0, 9)
	parkMice(r)
	r.Broom.X, r.Broom.Y = r.Freddy.X, r.Freddy.Y-10
	r.transferTick(entity.Controls{Act: true})
	if r.Transfer.Active {
		t.Error("touching the broom should block a transfer")
	}
	r.Broom.X, r.Broom.Y = 0, 160
	r.Freddy.Bouncing = true
	r.transferTick(entity.Controls{Act: true})
	if r.Transfer.Active {
		t.Error("being bounced should block a transfer")
	}
}

func TestMiceStartHiddenInDifferentHoles(t *testing.T) {
	for seed := range uint64(200) {
		r := newCheese(int(seed%8), rand.New(rand.NewPCG(seed, 3)))
		used := map[int]bool{}
		for _, m := range r.Mice {
			if m.Count < config.MouseHidden || m.Count >= config.MouseHidden+config.MouseStartRoll {
				t.Fatalf("seed %d: start countdown %d", seed, m.Count)
			}
			if used[m.Hole] {
				t.Fatalf("seed %d: two mice in hole %d", seed, m.Hole)
			}
			used[m.Hole] = true
		}
	}
}

func TestNewHolesAvoidFreddyAndOtherMice(t *testing.T) {
	r := newCheese(0, rng())
	f := r.Freddy
	f.X, f.Y = 60, 100 // near holes 6 and 7
	g := rand.New(rand.NewPCG(5, 5))
	for range 2000 {
		r.pickHole(0, g)
		h := r.Mice[0].Hole
		if r.holeTaken(0, h) {
			t.Fatalf("mouse 0 took hole %d, which another mouse uses", h)
		}
		if x, y := mouseAt(h); r.near(x, y) < config.HoleAwayDist {
			t.Fatalf("hole %d is within %d of Freddy", h, config.HoleAwayDist)
		}
	}
}

func TestEachMouseStepsOnceEveryThreeTicks(t *testing.T) {
	r := newCheese(0, rng())
	parkMice(r)
	for range 30 {
		r.miceTick(rng())
	}
	for i, m := range r.Mice {
		if m.Count != 90 {
			t.Errorf("mouse %d: countdown %d after 30 ticks, want 90", i, m.Count)
		}
	}
}

func TestCountdownPhases(t *testing.T) {
	r := newCheese(0, rng())
	parkMice(r)
	r.Freddy.X, r.Freddy.Y = 290, config.FloorY
	m := &r.Mice[0]
	m.Hole, m.Count = 9, config.MouseHidden
	var seen []bool
	for range 22 {
		r.mouseStep(0, rng())
		seen = append(seen, m.Visible())
	}
	// 19 down to 0 visible (20 steps), then a new countdown from 20 or more.
	for i, v := range seen {
		if want := i < 20; v != want {
			t.Fatalf("step %d: visible %v, want %v (steps %v)", i, v, want, seen)
		}
	}
}

func TestShyMiceDuckWithinTheRetreatDistance(t *testing.T) {
	for d := range 8 {
		dist := config.CheeseRetreat[d]
		for _, c := range []struct {
			count, near, want int
		}{
			{11, dist - 1, 1},  // out: ducks
			{11, dist, 10},     // just outside: carries on
			{19, dist - 1, 0},  // emerging: dives back
			{25, dist - 1, 24}, // hidden: unaffected
			{2, dist - 1, 1},   // last step out
		} {
			r := newCheese(d, rng())
			parkMice(r)
			m := &r.Mice[0]
			m.Hole, m.Count = 14, c.count
			x, y := mouseAt(14)
			r.Freddy.X, r.Freddy.Y = x+c.near, y // right of the mouse, level with it
			r.mouseStep(0, rng())
			if m.Count != c.want {
				t.Errorf("difficulty %d, countdown %d, Freddy %d away: countdown %d, want %d",
					d, c.count, c.near, m.Count, c.want)
			}
		}
	}
}

func TestBroomFreezesAMouseButFreddyCanCatchIt(t *testing.T) {
	r := newCheese(0, rng())
	parkMice(r)
	r.Freddy.X, r.Freddy.Y = 290, config.FloorY
	m := &r.Mice[0]
	m.Hole, m.Count = 9, 10
	mr := r.MouseRect(0)
	r.Broom.X, r.Broom.Y = mr.X, mr.Y-10
	r.mouseStep(0, rng())
	if m.Count != 10 {
		t.Fatalf("a mouse under the broom should be frozen, countdown %d", m.Count)
	}
	r.Freddy.X, r.Freddy.Y = mr.X, mr.Y
	r.mouseStep(0, rng())
	if !m.Caught {
		t.Fatal("Freddy should catch a frozen mouse")
	}
}

func TestCatchingFourMiceClearsTheRoom(t *testing.T) {
	r := newCheese(0, rng())
	for i := range r.Mice {
		m := &r.Mice[i]
		m.Hole, m.Count = 3+i*3, 10
		mr := r.MouseRect(i)
		r.Freddy.X, r.Freddy.Y = mr.X, mr.Y
		r.mouseStep(i, rng())
		if !m.Caught || r.Caught != i+1 {
			t.Fatalf("mouse %d: caught %v, total %d", i, m.Caught, r.Caught)
		}
		if i < 3 && r.Out != Playing {
			t.Fatal("the room should go on until the 4th mouse")
		}
		r.mouseStep(i, rng()) // a caught mouse stays caught
	}
	if r.Out != Cleared || r.Caught != 4 {
		t.Fatalf("4 mice should clear the room: out %d, caught %d", r.Out, r.Caught)
	}
}

func TestHiddenMiceCannotBeCaught(t *testing.T) {
	r := newCheese(0, rng())
	parkMice(r)
	m := &r.Mice[0]
	m.Hole, m.Count = 9, 30
	mr := r.MouseRect(0)
	r.Freddy.X, r.Freddy.Y = mr.X, mr.Y
	r.mouseStep(0, rng())
	if m.Caught {
		t.Fatal("a hidden mouse should not be caught")
	}
}

// TestCheeseRoomCanBeCleared plays the room headless with a fixed seed.
// Freddy starts in a hole and goes through it whenever a mouse shows in
// the hole it leads to; the room must be cleared within 3 minutes.
func TestCheeseRoomCanBeCleared(t *testing.T) {
	g := rand.New(rand.NewPCG(1984, 1))
	r := newCheese(0, g)
	putInHole(r, 4)
	const limit = 3 * 60 * 182 / 10 // 3 minutes of ticks
	for tick := range limit {
		in := entity.Controls{}
		if h := r.HoleAt(r.Freddy); h >= 0 {
			to := r.Partner(h)
			for _, m := range r.Mice {
				if m.Hole == to && m.Visible() && m.Count >= 8 {
					in.Act = true
				}
			}
		}
		r.Update(1, 2, in, g)
		if r.Out == Cleared {
			t.Logf("cleared after %d ticks", tick+1)
			return
		}
		if r.Out != Playing {
			t.Fatalf("the room ended with %d after %d ticks", r.Out, tick+1)
		}
	}
	t.Fatalf("not cleared in %d ticks: caught %d, Freddy at %d,%d", limit, r.Caught, r.Freddy.X, r.Freddy.Y)
}

// TestCheeseSoak runs the room with random input to catch panics and stuck
// loops.
func TestCheeseSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 2))
		r := newCheese(d, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1, Act: g.IntN(2) == 0}
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newCheese(d, g)
			}
		}
	}
}
