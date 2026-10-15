package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// bookRoom returns a bookshelf room at difficulty d with Freddy standing at
// (x, y) and the broom out of the way.
func bookRoom(d, x, y int) *BookshelfRoom {
	r := newBookshelf(d, rng())
	f := r.Freddy
	f.X, f.Y, f.Arc = x, y, physics.Arc{}
	r.Broom.X, r.Broom.Y = 0, 0
	return r
}

func TestShelvesCatchFreddy(t *testing.T) {
	r := bookRoom(0, 200, 0)
	f := r.Freddy
	for shelf := 40; shelf <= 152; shelf += 16 {
		for k := range 8 {
			f.X, f.Y = 201, shelf+2+k
			if y, ok := r.ShelfAt(f); !ok || y != shelf+2 {
				t.Fatalf("Y %d: on shelf %v at %d, want %d", f.Y, ok, y, shelf+2)
			}
		}
		f.Y = shelf + 2 + 8 // between shelves
		if _, ok := r.ShelfAt(f); ok {
			t.Errorf("Y %d is between shelves", f.Y)
		}
	}
	f.Y = 42
	for _, x := range []int{160, 163, 284} {
		f.X = x
		if _, ok := r.ShelfAt(f); ok {
			t.Errorf("X %d is off the bookcase", x)
		}
	}
	// Falling onto the top shelf seats him there, X rounded down to 4.
	f.EnterRoom(203, 20, 0)
	for range 30 {
		r.Update(0, 1, entity.Controls{}, rng())
	}
	if f.Y != 42 || f.Seat == entity.NotSeated || f.X != 200 {
		t.Errorf("he should sit on the top shelf at 200,42, is at %d,%d seat %d", f.X, f.Y, f.Seat)
	}
}

func TestFernsOneATickThenClear(t *testing.T) {
	r := bookRoom(0, 200, 26) // touching the ferns at 192 and 224
	r.fernTick()
	if r.FernsLeft != 2 {
		t.Fatalf("one fern a tick, %d left", r.FernsLeft)
	}
	r.fernTick()
	r.Freddy.X = 250
	r.fernTick()
	if r.Out != Cleared {
		t.Errorf("the 3rd fern should clear the room, out %d, %d left", r.Out, r.FernsLeft)
	}
	r = bookRoom(0, 200, 60)
	r.fernTick()
	if r.FernsLeft != 3 {
		t.Error("ferns out of reach should stay")
	}
}

func TestSpiderTracksFreddyAlongTheTop(t *testing.T) {
	for d := range 8 {
		r := bookRoom(d, 20, config.FloorY)
		r.spiderStep()
		if want := config.SpiderStartX - config.SpiderSpeed[d]; r.Spider.X != want || r.Spider.Y != config.SpiderTopY {
			t.Errorf("difficulty %d: spider at %d,%d, want %d,%d", d, r.Spider.X, r.Spider.Y, want, config.SpiderTopY)
		}
	}
	r := bookRoom(0, 300, config.FloorY)
	r.Spider.X = 100
	r.spiderStep()
	if r.Spider.X != 102 {
		t.Errorf("the spider should head right towards Freddy, at %d", r.Spider.X)
	}
}

func TestSpiderDropsWhenLinedUpAndSpeedsUp(t *testing.T) {
	r := bookRoom(0, 283, config.FloorY) // 283 &^ 7 = 280, the spider's X
	var steps []int
	for range 14 {
		y := r.Spider.Y
		r.spiderStep()
		steps = append(steps, r.Spider.Y-y)
	}
	want := []int{2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5}
	for i := range want {
		if steps[i] != want[i] {
			t.Fatalf("drop steps %v, want %v", steps, want)
		}
	}
}

func TestSpiderClimbsBackWhenFreddyGetsAway(t *testing.T) {
	r := bookRoom(0, 283, config.FloorY)
	for range 5 {
		r.spiderStep()
	}
	r.Freddy.X = 283 - config.SpiderReach - 4 // more than 48 px away
	r.spiderStep()
	if r.Spider.State != -1 {
		t.Fatalf("the spider should turn back up, state %d", r.Spider.State)
	}
	for range 40 {
		r.spiderStep()
		if r.Spider.Y == config.SpiderTopY {
			break
		}
	}
	if r.Spider.Y != config.SpiderTopY || r.Spider.State != 0 {
		t.Errorf("the spider should be back at the top tracking, at Y %d state %d", r.Spider.Y, r.Spider.State)
	}
	// It also turns back on passing below his level.
	r = bookRoom(0, 283, config.FloorY)
	for range 5 {
		r.spiderStep()
	}
	r.Freddy.X, r.Freddy.Y = 310, r.Spider.Y+1 // beside it, just below its top
	r.spiderStep()
	if r.Spider.State != -1 {
		t.Errorf("the spider should turn back below Freddy's level, state %d", r.Spider.State)
	}
}

func TestTheBroomFreezesAndBlocksTheSpider(t *testing.T) {
	r := bookRoom(0, 20, config.FloorY)
	r.Broom.X, r.Broom.Y = r.Spider.X, r.Spider.Y
	r.spiderStep()
	if r.Spider.X != config.SpiderStartX {
		t.Error("the spider should not move while it touches the broom")
	}
	r.Broom.X = r.Spider.X - config.BroomW - 1 // just clear of it, in its way
	r.spiderStep()
	if r.Spider.X != config.SpiderStartX {
		t.Errorf("the spider should not move into the broom, at %d", r.Spider.X)
	}
}

func TestTheSpiderBites(t *testing.T) {
	r := bookRoom(0, config.SpiderStartX, config.SpiderTopY+4)
	r.Update(2, 0, entity.Controls{}, rng())
	if r.Bite == 0 {
		t.Fatal("the spider should bite Freddy")
	}
	r.Freddy.X, r.Freddy.Y = 200, 26 // touching a fern: too late
	for tick := 1; tick <= config.ByteTicks; tick++ {
		r.Update(1, 2, entity.Controls{}, rng())
		if r.FernsLeft != 3 {
			t.Fatal("no ferns while BYTE! plays")
		}
		if failed := r.Out == Failed; failed != (tick == config.ByteTicks) {
			t.Fatalf("tick %d of BYTE!: out %d", tick, r.Out)
		}
	}
	if r.Out != Failed || r.ByDog {
		t.Errorf("the bite should fail the room: out %d, by dog %v", r.Out, r.ByDog)
	}
}

func TestTheBroomIsHalfSpeedHere(t *testing.T) {
	r := newBookshelf(0, rng())
	moves := 0
	for range 40 {
		x, y := r.Broom.X, r.Broom.Y
		r.tick(rng())
		if r.Broom.X != x || r.Broom.Y != y {
			moves++
		}
	}
	if moves > 20 {
		t.Errorf("the broom moved %d times in 40 ticks, want at most 20", moves)
	}
}

func TestBooksFillTheBookcase(t *testing.T) {
	for seed := range uint64(20) {
		r := newBookshelf(0, rand.New(rand.NewPCG(seed, 8)))
		for row, cells := range r.Books {
			for col, k := range cells {
				switch {
				case col == 0 && k != BookcaseLeft, col == config.BookCols-1 && k != BookcaseRight:
					t.Fatalf("row %d: the ends should be side panels", row)
				case col > 0 && col < config.BookCols-1 && cells[col-1] == BookLeaning && k < BookUpright:
					t.Fatalf("row %d col %d: a leaning book needs a book beside it", row, col)
				case row%2 == 1 && col > 0 && col < config.BookCols-1 && k == BookShelf:
					t.Fatalf("row %d is a full row, but col %d is bare", row, col)
				}
			}
		}
	}
}

// bookshelfBot plays the room. For each fern in turn it climbs straight up
// the bookcase in the column under the fern, and jumps for it from the top
// shelf. The spider tracks Freddy along the top, so when it is near that
// column the bot first leads it away along the floor to the left, and climbs
// once it has dropped or is far enough off. Near the spider it gets down.
func bookshelfBot(r *BookshelfRoom) entity.Controls {
	f, s := r.Freddy, &r.Spider
	if f.Airborne() {
		return entity.Controls{}
	}
	col := -1
	for i := range r.Ferns {
		if !r.Ferns[i] {
			col = config.FernX[i] - 4
			break
		}
	}
	if col < 0 {
		return entity.Controls{}
	}
	if f.Y < config.FloorY { // on the bookcase
		if s.State == 0 && abs(s.X-f.X) < 64 {
			return entity.Controls{V: 1}
		}
		return entity.Controls{V: -1}
	}
	target := col
	if s.State == 0 && abs(s.X-col) < 110 {
		target = 20 // lead the spider away first
	}
	switch {
	case f.X < target-3:
		return entity.Controls{H: 1}
	case f.X > target+3:
		return entity.Controls{H: -1}
	case target == col:
		return entity.Controls{V: -1}
	}
	return entity.Controls{}
}

// TestBookshelfRoomCanBeCleared plays the room headless with a fixed seed,
// from falling in at the window to the 3rd fern.
func TestBookshelfRoomCanBeCleared(t *testing.T) {
	g := rand.New(rand.NewPCG(9, 9))
	r := newBookshelf(0, g)
	const limit = 3 * 60 * 182 / 10
	for tick := range limit {
		r.Update(1, 2, bookshelfBot(r), g)
		switch r.Out {
		case Cleared:
			t.Logf("cleared after %d ticks", tick+1)
			return
		case Playing:
		default:
			t.Fatalf("the room ended with %d after %d ticks (%d ferns left, Freddy %d,%d)", r.Out, tick+1, r.FernsLeft, r.Freddy.X, r.Freddy.Y)
		}
	}
	t.Fatalf("not cleared in %d ticks: %d ferns left, Freddy at %d,%d", limit, r.FernsLeft, r.Freddy.X, r.Freddy.Y)
}

// TestBookshelfSoak runs the room with random input to catch panics and
// stuck loops.
func TestBookshelfSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 10))
		r := newBookshelf(d, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1}
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newBookshelf(d, g)
			}
		}
	}
}

func TestSpiderLegsMoveEveryTwoSteps(t *testing.T) {
	r := bookRoom(0, 20, config.FloorY) // far off: the spider walks along the top
	var frames []int
	for range 8 {
		r.spiderStep()
		frames = append(frames, r.Spider.Frame)
	}
	want := []int{0, 1, 1, 0, 0, 1, 1, 0}
	for i := range want {
		if frames[i] != want[i] {
			t.Fatalf("frames %v, want %v", frames, want)
		}
	}
}
