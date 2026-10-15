package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

func rng() *rand.Rand { return rand.New(rand.NewPCG(9, 9)) }

// standingIn returns a birdcage room with Freddy landed on the floor at x.
func standingIn(x int) *BirdcageRoom {
	r := newBirdcage(0, rng())
	f := r.Freddy
	f.X, f.Y = x, config.FloorY
	f.Arc = physics.Arc{}
	return r
}

func TestBonus(t *testing.T) {
	cases := []struct {
		k    Kind
		t    int
		want int
	}{
		{Birdcage, 248, 3500 + 2*(1350-248)}, // R1: 5704
		{Birdcage, 691, 4818},                // R1: 4818
		{Cheese, 1161, 1878},                 // R1: 1878
		{Birdcage, 5000, 3500},
		{Dogs, 100, 2500 + 2600},
	}
	for _, c := range cases {
		if got := Bonus(c.k, c.t); got != c.want {
			t.Errorf("Bonus(%v, %d) = %d, want %d", c.k, c.t, got, c.want)
		}
	}
}

func TestEnteringFallsFromTheStart(t *testing.T) {
	r := newBirdcage(0, rng())
	f := r.Freddy
	if f.X != 240 || f.Y != 96 || f.Arc.Dir != physics.Falling {
		t.Fatalf("start: %d,%d dir %d", f.X, f.Y, f.Arc.Dir)
	}
	for range 200 {
		r.Update(0, 1, entity.Controls{}, rng())
	}
	if r.Out != Playing {
		t.Fatalf("simply falling in should not end the room (out %d)", r.Out)
	}
	if !f.Standing() {
		t.Error("Freddy should have landed")
	}
}

func TestLeavingByTheWindow(t *testing.T) {
	r := standingIn(100)
	f := r.Freddy
	f.X, f.Y = 244, 90
	f.Arc = physics.Arc{Dir: physics.Falling, Speed: 1}
	if r.Support(f) || r.Out != LeftEarly {
		t.Errorf("falling into the window: out %d, want LeftEarly", r.Out)
	}
}

func TestTableAndPerches(t *testing.T) {
	r := standingIn(100)
	f := r.Freddy
	f.X, f.Y = 140, 138
	if !r.Support(f) || f.Y != 136 || f.Seat != entity.NotSeated {
		t.Errorf("table: Y=%d seat=%d", f.Y, f.Seat)
	}
	f.X, f.Y, f.Seat = 19, 114, entity.NotSeated
	if !r.Support(f) || f.Y != 112 || f.Seat != entity.JustSeated || f.X != 16 {
		t.Errorf("lamp perch: %d,%d seat=%d", f.X, f.Y, f.Seat)
	}
}

func TestPushingTheCageOffFreesTheBird(t *testing.T) {
	r := standingIn(100)
	f := r.Freddy
	r.Broom.X, r.Broom.Y = 0, 0 // keep the broom away from the table
	pushes := 0
	for !r.Cage.Broken && pushes < 20 {
		// Stand Freddy just left of the cage on the table, walking right.
		f.X, f.Y, f.Dir = r.Cage.X-20, config.TableBand, 1
		f.Arc = physics.Arc{}
		r.cageTick() // touch: pending
		r.cageTick() // move
		pushes++
		for r.Cage.Falling && !r.Cage.Broken {
			r.fall(rng())
		}
	}
	if !r.Cage.Broken {
		t.Fatalf("cage never fell (X=%d)", r.Cage.X)
	}
	if r.Cage.X <= config.CageSafeMaxX {
		t.Errorf("cage fell at X=%d, inside the table", r.Cage.X)
	}
	if r.Bird.X != r.Cage.X {
		t.Error("the bird should start where the cage broke")
	}
}

func TestCatchingTheBirdClearsTheRoom(t *testing.T) {
	r := standingIn(100)
	r.Cage.Broken = true
	r.Bird = Bird{X: 104, Y: config.FloorY + 2, waypoint: 0xff}
	r.checkBird()
	if r.Out != Cleared {
		t.Errorf("out %d, want Cleared", r.Out)
	}
}

func TestBirdStaysInTheRoom(t *testing.T) {
	r := standingIn(10)
	r.Cage.Broken = true
	r.Bird = Bird{X: 160, Y: 100, VX: 1, waypoint: 0xff}
	g := rng()
	for range 5000 {
		r.birdStep(g)
		b := r.Bird
		if b.X < config.BirdMinX || b.X > config.BirdMaxX || b.Y < config.BirdMinY || b.Y > config.BirdMaxY {
			t.Fatalf("bird left the room: %d,%d", b.X, b.Y)
		}
		if r.Out != Playing {
			r.Out = Playing
			r.Freddy.X = 10
		}
	}
}

func TestDogCatchFailsTheRoom(t *testing.T) {
	r := standingIn(150)
	r.Dog = entity.Dog{X: 140, Dir: 1, Active: true}
	if !r.Dog.CheckCatch(r.Freddy) {
		t.Fatal("the dog should catch Freddy on the floor")
	}
	r.Freddy.Gone = true
	for range 400 {
		r.tick(rng())
		if r.Out != Playing {
			break
		}
	}
	if r.Out != Failed || !r.ByDog {
		t.Errorf("out %d byDog %v, want a dog failure", r.Out, r.ByDog)
	}
}

func TestBroomSweepsFootprints(t *testing.T) {
	r := standingIn(100)
	f := r.Freddy
	for x := 40; x < 120; x += 2 {
		f.X, f.Dir = x, 1
		r.Broom.Footprint(f)
	}
	marks := 0
	for _, m := range r.Broom.Marks {
		marks += m
	}
	if marks == 0 {
		t.Fatal("walking left no footprints")
	}
	f.X, f.Y = 300, 40 // well away
	g := rng()
	for range 3000 {
		r.Broom.Tick(f, &r.Dog, 0, 248, 96, g)
	}
	left := 0
	for _, m := range r.Broom.Marks {
		left += m
	}
	if left >= marks {
		t.Errorf("the broom swept nothing: %d marks before, %d after", marks, left)
	}
}

func TestBroomBouncesTowardsTheWindow(t *testing.T) {
	r := standingIn(100)
	f := r.Freddy
	f.X, f.Y = 60, config.FloorY
	f.Bounce(248, 96)
	if f.Dir != 1 || f.Arc.Dir != physics.Rising || f.Arc.Speed != config.BounceSpeed {
		t.Errorf("bounce: dir %d arc %+v", f.Dir, f.Arc)
	}
	f.Arc.Dir = physics.Still
	f.Bouncing = true
	f.Bounce(248, 96)
	if f.Arc.Dir != physics.Still {
		t.Error("a second bounce before landing should do nothing")
	}
}

// TestSoak runs the room with random input to catch panics and stuck loops.
func TestSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 1))
		r := newBirdcage(d, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1}
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newBirdcage(d, g)
			}
		}
	}
}
