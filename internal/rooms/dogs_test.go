package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// dogsRoom returns a sleeping dogs room at difficulty d with only the given
// dog slots filled and Freddy standing at (x, y).
func dogsRoom(d, x, y int, slots ...int) *DogsRoom {
	r := newDogs(d, rng())
	r.Sleepers = [len(config.Sleepers)]SleepingDog{}
	for _, s := range slots {
		r.Sleepers[s].On = true
	}
	f := r.Freddy
	f.X, f.Y, f.Arc = x, y, physics.Arc{}
	r.Broom.X, r.Broom.Y = 0, 0 // out of the way
	return r
}

var alt = entity.Controls{Act: true}

func TestDogsAndPortionsByDifficulty(t *testing.T) {
	for d := range 8 {
		tb := config.ForLevel(d)
		for seed := range uint64(20) {
			r := newDogs(d, rand.New(rand.NewPCG(seed, 5)))
			on := 0
			for _, s := range r.Sleepers {
				if s.On {
					on++
				}
			}
			if on != tb.SleepingDogs {
				t.Fatalf("difficulty %d: %d dogs, want %d", d, on, tb.SleepingDogs)
			}
			for i, p := range r.Portions {
				if p != tb.FoodPortions {
					t.Fatalf("difficulty %d: bowl %d holds %d, want %d", d, i, p, tb.FoodPortions)
				}
			}
		}
	}
}

func TestStayingNearADogWakesItThenItAttacks(t *testing.T) {
	for d := range 8 {
		every := config.ForLevel(d).DogWakeCheckTicks + 1
		r := dogsRoom(d, 124, 136, 1)
		for tick := 1; tick <= 3*every; tick++ {
			r.Update(1, 0, entity.Controls{}, rng())
			alert, attacked := r.Sleepers[1].Alert, r.Dog.Capture != 0
			switch {
			case tick == 3*every:
				if !attacked || !r.Freddy.Gone {
					t.Fatalf("difficulty %d: the dog should attack at tick %d", d, tick)
				}
				if r.Sleepers[1].On {
					t.Fatalf("difficulty %d: the dog should leave its place to attack", d)
				}
			case attacked:
				t.Fatalf("difficulty %d: attacked early, at tick %d", d, tick)
			case alert != tick/every:
				t.Fatalf("difficulty %d: alert %d at tick %d", d, alert, tick)
			}
		}
		if !r.Woke {
			t.Errorf("difficulty %d: waking should be signalled", d)
		}
	}
}

func TestNearMeansTheSameRowAndWithinTheWakeDistance(t *testing.T) {
	for d := range 8 {
		dist := config.ForLevel(d).DogWakeDistance
		r := dogsRoom(d, 124+dist, 136, 1)
		if !r.near(1) {
			t.Errorf("difficulty %d: %d px away should be near", d, dist)
		}
		r.Freddy.X = 124 - dist - 1
		if r.near(1) {
			t.Errorf("difficulty %d: %d px away should not be near", d, dist+1)
		}
		r.Freddy.X, r.Freddy.Y = 124, 152
		if r.near(1) {
			t.Errorf("difficulty %d: another row should not be near", d)
		}
	}
}

func TestDogsCalmDownWhenFreddyIsAway(t *testing.T) {
	r := dogsRoom(0, 290, 168, 1)
	r.Sleepers[1].Alert = 2
	g := rand.New(rand.NewPCG(2, 2))
	for range 2000 {
		prev := r.Sleepers[1].Alert
		r.Update(1, 0, entity.Controls{}, g)
		if r.Sleepers[1].Alert > prev {
			t.Fatal("a dog should not grow more alert with Freddy away")
		}
	}
	if r.Sleepers[1].Alert != 0 {
		t.Errorf("the dog should have settled, alert %d", r.Sleepers[1].Alert)
	}
}

func TestAWakingDogCatchesFreddy(t *testing.T) {
	r := dogsRoom(0, 124, 136, 1)
	r.Sleepers[1].Alert = 2
	for range 1000 {
		r.Update(1, 2, entity.Controls{}, rng())
		if r.Out != Playing {
			break
		}
	}
	if r.Out != Failed || !r.ByDog {
		t.Fatalf("the capture should fail the room with HERE KITTY: out %d, by dog %v", r.Out, r.ByDog)
	}
}

func TestEatingTakesFiveAndAThirdTicksAPortion(t *testing.T) {
	r := dogsRoom(2, 144, 136) // 3 portions
	for tick := 1; tick <= 20; tick++ {
		r.Update(1, 0, alt, rng())
		if r.Portions[1] == 0 {
			if tick != 16 {
				t.Fatalf("3 portions took %d ticks, want 16", tick)
			}
			return
		}
	}
	t.Fatal("the bowl was not emptied")
}

func TestAltWalksFreddyToTheNearestBowl(t *testing.T) {
	r := dogsRoom(0, 84, 152)
	for range 40 {
		r.Update(1, 2, alt, rng())
		if r.Eating {
			break
		}
	}
	if !r.Eating || r.EatBowl != 4 || abs(r.Freddy.X-56) >= config.EatReach {
		t.Fatalf("he should walk to bowl 4 at X 56 and eat: eating %v bowl %d at X %d", r.Eating, r.EatBowl, r.Freddy.X)
	}
	// Equally near two bowls, he heads for the lower-numbered one.
	r = dogsRoom(0, 48, 152)
	r.eatTick(alt)
	if r.walk != -1 {
		t.Errorf("between bowls 3 and 4 he should head for bowl 3, walk %d", r.walk)
	}
}

func TestOnlyBowlsOnHisRowCount(t *testing.T) {
	r := dogsRoom(0, 290, 168) // the row-168 bowls are far away; none above count
	r.Portions = [12]int{1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0}
	r.eatTick(alt)
	if r.walk != 0 || r.Eating {
		t.Errorf("with no food on his row Alt should do nothing: walk %d, eating %v", r.walk, r.Eating)
	}
}

func TestLettingGoOfAltStopsEating(t *testing.T) {
	r := dogsRoom(3, 144, 136)
	r.Update(3, 0, alt, rng())
	if !r.Eating || r.EatAcc == 0 {
		t.Fatal("he should be eating")
	}
	r.Update(1, 0, entity.Controls{}, rng())
	if r.Eating || r.EatAcc != 0 {
		t.Error("letting go should stop eating and reset the accumulator")
	}
	r.Update(3, 0, alt, rng())
	r.Freddy.Bouncing = true
	r.Update(1, 0, alt, rng())
	if r.Eating {
		t.Error("the broom's bounce should stop eating")
	}
}

func TestHoldingAltCarriesOnToTheNextBowl(t *testing.T) {
	r := dogsRoom(0, 144, 136)
	for range 200 {
		r.Update(1, 2, alt, rng())
		if r.Portions[0] == 0 {
			break
		}
	}
	if r.Portions[1] != 0 || r.Portions[2] != 0 || r.Portions[0] != 0 {
		t.Errorf("holding Alt should empty every bowl on the row: %v", r.Portions[:3])
	}
}

func TestTheLastBowlClearsTheRoom(t *testing.T) {
	r := dogsRoom(0, 144, 136)
	r.Portions = [12]int{0, 1}
	r.BowlsLeft = 1
	for range 10 {
		r.Update(1, 0, alt, rng())
	}
	if r.Out != Cleared {
		t.Fatalf("emptying the last bowl should clear the room, out %d", r.Out)
	}
}

// dogsBot plays the room: it holds Alt to eat along its row. A fully awake
// dog near it, or a row with no food left, sends it to another row: down a
// row, or up by jumping from the front row or the floor.
func dogsBot(r *DogsRoom) entity.Controls {
	f := r.Freddy
	if !f.Standing() {
		return entity.Controls{}
	}
	move := func() entity.Controls {
		below := false
		for i, b := range config.FoodBowls {
			if r.Portions[i] > 0 && b.Y > f.Y {
				below = true
			}
		}
		if below && f.Y < config.FloorY {
			return entity.Controls{V: 1}
		}
		return entity.Controls{V: -1}
	}
	for i, s := range r.Sleepers {
		if s.On && s.Alert >= config.DogAlertMax && r.near(i) {
			return move()
		}
	}
	for i, b := range config.FoodBowls {
		if r.Portions[i] > 0 && b.Y == f.Y {
			return alt
		}
	}
	return move()
}

// TestDogsRoomCanBeCleared plays the room headless with a fixed seed, from
// falling in at the window to the last bowl.
func TestDogsRoomCanBeCleared(t *testing.T) {
	g := rand.New(rand.NewPCG(12, 3))
	r := newDogs(0, g)
	const limit = 3 * 60 * 182 / 10
	for tick := range limit {
		r.Update(1, 2, dogsBot(r), g)
		switch r.Out {
		case Cleared:
			t.Logf("cleared after %d ticks", tick+1)
			return
		case Playing:
		default:
			t.Fatalf("the room ended with %d after %d ticks (%d bowls left)", r.Out, tick+1, r.BowlsLeft)
		}
	}
	t.Fatalf("not cleared in %d ticks: %d bowls left, Freddy at %d,%d", limit, r.BowlsLeft, r.Freddy.X, r.Freddy.Y)
}

// TestDogsSoak runs the room with random input to catch panics and stuck
// loops.
func TestDogsSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 6))
		r := newDogs(d, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1, Act: g.IntN(2) == 0}
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newDogs(d, g)
			}
		}
	}
}

func TestBowlsStartFullAndStepDownToGone(t *testing.T) {
	for d := range 8 {
		r := newDogs(d, rng())
		start := config.ForLevel(d).FoodPortions
		prev := r.MilkLevel(0)
		if prev != config.MilkLevels {
			t.Fatalf("difficulty %d: a fresh bowl should look full, level %d", d, prev)
		}
		for p := start - 1; p >= 0; p-- {
			r.Portions[0] = p
			level := r.MilkLevel(0)
			if level >= prev || (p == 0) != (level == 0) {
				t.Fatalf("difficulty %d, %d of %d portions left: level %d after %d", d, p, start, level, prev)
			}
			prev = level
		}
	}
}

func TestNoDogLiesOnABowl(t *testing.T) {
	// Touching edges are fine: the original packs the dogs right beside
	// their bowls. Shrinking both boxes by 1 leaves only real overlaps.
	for i := range config.Sleepers {
		dog := SleeperRect(i)
		dog.W, dog.H = dog.W-1, dog.H-1
		for j, b := range config.FoodBowls {
			bowl := physics.Rect{X: b.X, Y: b.Y + config.FoodBowlDY, W: config.FoodBowlW - 1, H: config.FoodBowlH - 1}
			if dog.Overlaps(bowl) {
				t.Errorf("dog %d (X %d–%d) lies on bowl %d (X %d–%d)", i, dog.X, dog.X+dog.W, j, bowl.X, bowl.X+bowl.W)
			}
		}
	}
}
