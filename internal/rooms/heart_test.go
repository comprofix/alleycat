package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// heartRoom returns a heart room with every heart intact, no arrow about
// to hit, and Freddy standing on row at X x.
func heartRoom(d, rooms, row, x int) *HeartRoom {
	r := newHeart(d, rooms, rng())
	r.Broken = [len(config.HeartRows)][config.HeartCells]bool{}
	f := r.Freddy
	f.X, f.Y, f.Arc = x, config.HeartRows[row]+config.HeartStandDY, physics.Arc{}
	return r
}

// clearCats sends every rival away for good and parks Felicia at the far
// left, out of the way.
func clearCats(r *HeartRoom) {
	for i := range config.Felicia {
		r.Cats[i].Away, r.Cats[i].AwaySince = true, 1<<30
	}
	r.Cats[config.Felicia].X = config.CatMinX
}

func TestHeartSetUp(t *testing.T) {
	for seed := range uint64(30) {
		r := newHeart(int(seed%8), int(seed%11), rand.New(rand.NewPCG(seed, 4)))
		for col := range config.HeartCells {
			if r.Broken[0][col] {
				t.Fatalf("seed %d: the bottom row should be whole", seed)
			}
		}
		gifts := 0
		for _, g := range r.Gifts {
			if g.On {
				gifts++
			}
		}
		if want := min(max(int(seed%11), 1), config.GiftMax); gifts != want {
			t.Fatalf("seed %d: %d gifts for %d rooms, want %d", seed, gifts, seed%11, want)
		}
		if !r.Cats[0].Away {
			t.Fatal("the bottom rival should start away")
		}
		for i, c := range r.Cats {
			if c.Y != config.HeartRows[i]+config.HeartCatDY || (i > 0 && (c.X < 96 || c.X > 223)) {
				t.Fatalf("seed %d: cat %d at %d,%d", seed, i, c.X, c.Y)
			}
		}
		if f := r.Freddy; !f.NoFloor || f.MinX != config.HeartMinX || f.MaxX != config.HeartMaxX {
			t.Fatal("Freddy should have no floor and the heart room's limits")
		}
	}
}

func TestStandingOnlyOnIntactHearts(t *testing.T) {
	r := heartRoom(0, 1, 2, 100)
	f := r.Freddy
	col := (f.X+config.HeartStandDX)/16 - 2
	if !r.Support(f) {
		t.Fatal("an intact heart should hold him")
	}
	r.Broken[2][col] = true
	if r.Support(f) {
		t.Fatal("a broken heart should not hold him")
	}
	// Past the last heart the row still holds him.
	f.X = config.HeartMaxX
	if !r.Support(f) {
		t.Error("columns past the last heart count as intact")
	}
	// Between rows nothing does.
	f.X, f.Y = 100, config.HeartRows[2]+config.HeartStandDY+9
	if r.Support(f) {
		t.Error("nothing between the rows")
	}
}

func TestNoFloorFallingOffFailsWithNoLife(t *testing.T) {
	r := heartRoom(0, 1, 0, 100)
	clearCats(r)
	r.Arrow.Active, r.Arrow.Y = true, config.ArrowMaxY // gone next tick, far off
	r.Broken[0][(100+config.HeartStandDX)/16-2] = true
	for range 40 {
		r.Update(0, 1, entity.Controls{}, rng())
		if r.Out != Playing {
			break
		}
	}
	if r.Out != Failed || r.ByDog {
		t.Fatalf("falling through a broken bottom heart should fail the room: out %d", r.Out)
	}
	if r.Freddy.Y < config.HeartFallOffY {
		t.Errorf("he should have fallen off the bottom, at Y %d", r.Freddy.Y)
	}
}

func TestArrowsSpawnFromSidesAndTop(t *testing.T) {
	r := heartRoom(0, 1, 0, 100)
	g := rand.New(rand.NewPCG(7, 7))
	sides := 0
	for range 2600 {
		r.fireArrow(g)
		a := r.Arrow
		switch {
		case a.Y == config.ArrowTopY:
			if a.X%32 != config.ArrowTopDX || (a.Dir == 1) != (a.X < 160) {
				t.Fatalf("top arrow at X %d heading %d", a.X, a.Dir)
			}
		case a.X == config.ArrowLeftX && a.Dir == 1, a.X == config.ArrowRightX && a.Dir == -1:
			sides++
			if (a.Y-config.ArrowSideDY)%24 != 0 {
				t.Fatalf("side arrow at Y %d", a.Y)
			}
		default:
			t.Fatalf("arrow at %d,%d heading %d", a.X, a.Y, a.Dir)
		}
	}
	if sides < 1450 || sides > 1750 { // 16 in 26 of 2600 is 1600
		t.Errorf("%d side arrows of 2600, want about 1600", sides)
	}
}

func TestArrowsFlipEachHeartOnceAndGo(t *testing.T) {
	r := heartRoom(0, 1, 6, 250) // Freddy on the top row, out of the way
	clearCats(r)
	r.Arrow = Arrow{Active: true, X: 40, Y: config.HeartRows[3] + 2, Dir: 1, lastCell: -1}
	start := r.Broken
	for range 200 {
		r.arrowTick(rng())
		if !r.Arrow.Active {
			break
		}
	}
	flips := 0
	for row := range r.Broken {
		for col := range r.Broken[row] {
			if r.Broken[row][col] != start[row][col] {
				flips++
			}
		}
	}
	if flips == 0 {
		t.Fatal("the arrow should flip the hearts it passes")
	}
	// A second arrow on the same path flips them all back.
	r.Arrow = Arrow{Active: true, X: 40, Y: config.HeartRows[3] + 2, Dir: 1, lastCell: -1}
	for range 200 {
		r.arrowTick(rng())
		if !r.Arrow.Active {
			break
		}
	}
	if r.Broken != start {
		t.Error("each heart should flip once per pass: a second pass should undo the first")
	}
}

func TestAnArrowKnocksFreddyDown(t *testing.T) {
	r := heartRoom(0, 1, 3, 100)
	f := r.Freddy
	r.Arrow = Arrow{Active: true, X: f.X, Y: f.Y, Dir: 1, lastCell: -1}
	r.arrowTick(rng())
	if r.Arrow.Active || f.Arc.Dir != physics.Falling || f.Arc.Speed != config.ArrowHitSpeed {
		t.Fatalf("the arrow should knock him down and go: active %v, arc %+v", r.Arrow.Active, f.Arc)
	}
	r.arrowTick(rng())
	if !r.Arrow.Active {
		t.Error("a new arrow should fly at once")
	}
}

func TestCatsMoveInThreeGroups(t *testing.T) {
	r := heartRoom(0, 1, 6, 250)
	var groups [][]int
	for range 6 {
		before := r.catSlot
		var g []int
		for s := before; ; {
			if s++; s >= config.HeartCatSlots {
				s = 0
			}
			g = append(g, s)
			if s == 0 || s == 1 || s == 4 {
				break
			}
		}
		r.catsTick(rng())
		groups = append(groups, g)
	}
	want := [][]int{{1}, {2, 3, 4}, {5, 6, 0}, {1}, {2, 3, 4}, {5, 6, 0}}
	for i := range want {
		if len(groups[i]) != len(want[i]) || groups[i][0] != want[i][0] {
			t.Fatalf("groups %v, want %v", groups, want)
		}
	}
}

func TestAnArrowFreezesACat(t *testing.T) {
	r := heartRoom(0, 1, 0, 250)
	c := &r.Cats[3]
	c.X, c.Dir = 150, 1
	r.Arrow = Arrow{Active: true, X: c.X, Y: c.Y, Dir: 1, lastCell: -1}
	r.catStep(3, rng())
	if c.X != 150 {
		t.Errorf("a cat under the arrow should not move, at %d", c.X)
	}
}

func TestRivalsChaseByDifficulty(t *testing.T) {
	for d := range 8 {
		turns := 0
		for seed := range uint64(400) {
			r := heartRoom(d, 1, 0, 40) // Freddy below them all, to the left
			c := &r.Cats[3]
			c.X, c.Dir = 200, 1
			r.chase(3, rand.New(rand.NewPCG(seed, uint64(d))))
			if c.Dir == -1 {
				turns++
			}
		}
		want := 400 * (config.ForLevel(d).RivalChaseChance + 1) / 256
		if turns < want-want/3-8 || turns > want+want/3+8 {
			t.Errorf("difficulty %d: turned %d of 400, want about %d", d, turns, want)
		}
	}
	// Level with Freddy, a chasing cat stops.
	r := heartRoom(7, 1, 0, 200)
	r.Cats[3].X, r.Cats[3].Dir = 203, 1
	for range 50 {
		r.chase(3, rng())
	}
	if r.Cats[3].Dir != 0 {
		t.Error("a cat level with Freddy should stop when it chases")
	}
}

func TestFeliciaShiesAwayThenComes(t *testing.T) {
	r := heartRoom(7, 1, 3, 100)
	fe := &r.Cats[config.Felicia]
	fe.X = 200
	turnedAway := false
	for range 50 {
		r.chase(config.Felicia, rng())
		if fe.Dir == 1 {
			turnedAway = true
		}
		if fe.Dir == -1 {
			t.Fatal("Felicia should not come to him while he is low down")
		}
	}
	if !turnedAway {
		t.Fatal("Felicia should walk away from him")
	}
	r.Freddy.Y = config.HeartRows[6] + config.HeartStandDY // the top row
	r.chase(config.Felicia, rng())
	if fe.Dir != -1 {
		t.Error("once he is near the top Felicia should come to him")
	}
}

func TestBumpingARivalKnocksFreddyDown(t *testing.T) {
	r := heartRoom(0, 1, 2, 100)
	c := &r.Cats[2]
	c.X, c.Dir = 110, -1
	if !r.catTouch(2) {
		t.Fatal("touching a rival should register")
	}
	f := r.Freddy
	if f.Arc.Dir != physics.Falling || f.Arc.Speed != config.CatKnockSpeed || c.Bump != config.CatBumpSteps || c.Dir != 1 {
		t.Fatalf("bump: arc %+v, cat bump %d dir %d", f.Arc, c.Bump, c.Dir)
	}
	f.X, f.Y = 20, 181 // out of its way
	for range config.CatBumpSteps {
		r.catStep(2, rng())
	}
	if c.Dir != -1 {
		t.Errorf("after backing off the cat should turn back, dir %d", c.Dir)
	}
	if r.Out != Playing {
		t.Error("bumping a rival costs nothing")
	}
}

func TestTouchingFeliciaClears(t *testing.T) {
	r := heartRoom(0, 1, 6, 100)
	r.Cats[config.Felicia].X = 104
	r.Carrying = 2
	r.catTouch(config.Felicia)
	if r.Out != Cleared || !r.WithGift {
		t.Fatalf("Felicia should clear the room, with the gift noted: out %d, gift %v", r.Out, r.WithGift)
	}
}

func TestGiftsPickUpPutDownAndLure(t *testing.T) {
	r := heartRoom(0, 2, 0, config.GiftX[0]+4)
	clearCats(r)
	r.giftTick(entity.Controls{})
	if r.Carrying != 0 || r.Gifts[0].On {
		t.Fatalf("touching a gift should pick it up: carrying %d", r.Carrying)
	}
	// Alt puts it down on the nearest row, right where he is.
	f := r.Freddy
	f.X, f.Y = 150, config.HeartRows[3]+config.HeartStandDY+2
	r.giftTick(entity.Controls{Act: true})
	if r.Carrying != -1 || !r.Gifts[0].On || r.Gifts[0].Y != config.HeartRows[3] || r.Gifts[0].X != 148 {
		t.Fatalf("put down: carrying %d, gift %+v", r.Carrying, r.Gifts[0])
	}
	// He does not pick it straight back up.
	r.giftTick(entity.Controls{})
	if r.Carrying != -1 {
		t.Error("the gift just put down should stay until he steps off it")
	}
	// Not on top of another gift.
	r.Gifts[0].On = false
	r.Carrying = 0
	f.X, f.Y = config.GiftX[1], config.HeartRows[0]+config.HeartStandDY
	r.putDown()
	if r.Carrying != 0 {
		t.Error("a gift may not go down on another")
	}
	// A rival walking into a gift takes it and goes away for its delay.
	r = heartRoom(0, 1, 6, 250)
	c := &r.Cats[1]
	r.Gifts[0] = Gift{X: 150, Y: config.HeartRows[1], On: true}
	c.X, c.Away = 140, false
	c.Y = config.HeartRows[1] + config.HeartCatDY
	r.lure(1)
	if r.Gifts[0].On || !c.Away {
		t.Fatal("the rival should take the gift and go")
	}
	r.Ticks = c.AwaySince + r.table.RivalHideTicks
	r.Freddy.X = 250
	r.catStep(1, rng())
	if c.Away || c.X != config.CatReturnLeftX {
		t.Errorf("after its delay the rival should come back on the far side: away %v at %d", c.Away, c.X)
	}
	// Felicia takes a gift but stays.
	fe := &r.Cats[config.Felicia]
	r.Gifts[1] = Gift{X: fe.X, Y: config.HeartRows[6], On: true}
	r.lure(config.Felicia)
	if r.Gifts[1].On || fe.Away {
		t.Error("Felicia should take the gift and stay")
	}
}

// heartBot plays the room: it jumps up row by row to Felicia, waiting for
// an intact heart above it and a clear moment.
func heartBot(r *HeartRoom) entity.Controls {
	f := r.Freddy
	if f.Airborne() || f.Gone {
		return entity.Controls{}
	}
	row := RowAt((f.Y - config.HeartStandDY) &^ 7)
	if row < 0 {
		return entity.Controls{}
	}
	if row == len(config.HeartRows)-1 {
		return entity.Controls{H: sign(r.Cats[config.Felicia].X - f.X)}
	}
	// Find the nearest column with an intact heart above, and go there.
	best, bestD := -1, 1<<30
	for col := range config.HeartCols {
		if !r.Broken[row+1][col] {
			x := config.HeartX + col*16 - 4
			if d := abs(x - f.X); d < bestD {
				best, bestD = x, d
			}
		}
	}
	if best < 0 {
		return entity.Controls{}
	}
	if bestD > 6 {
		return entity.Controls{H: sign(best - f.X)}
	}
	return entity.Controls{V: -1}
}

// TestHeartRoomCanBeCleared plays the room headless with a fixed seed,
// from Freddy's arrival to Felicia.
func TestHeartRoomCanBeCleared(t *testing.T) {
	g := rand.New(rand.NewPCG(23, 11))
	r := newHeart(0, 2, g)
	const limit = 3 * 60 * 182 / 10
	for tick := range limit {
		r.Update(1, 2, heartBot(r), g)
		switch r.Out {
		case Cleared:
			t.Logf("cleared after %d ticks", tick+1)
			return
		case Playing:
		default:
			t.Fatalf("the room ended with %d after %d ticks (Freddy %d,%d)", r.Out, tick+1, r.Freddy.X, r.Freddy.Y)
		}
	}
	t.Fatalf("not cleared in %d ticks: Freddy at %d,%d", limit, r.Freddy.X, r.Freddy.Y)
}

// TestHeartSoak runs the room with random input to catch panics and stuck
// loops.
func TestHeartSoak(t *testing.T) {
	for d := range 8 {
		g := rand.New(rand.NewPCG(uint64(d), 12))
		r := newHeart(d, d+1, g)
		in := entity.Controls{}
		for frame := range 20000 {
			if frame%15 == 0 {
				in = entity.Controls{H: g.IntN(3) - 1, V: g.IntN(3) - 1, Act: g.IntN(4) == 0}
			}
			r.Update(frame%3%2, frame%2, in, g)
			if r.Out != Playing {
				r = newHeart(d, d+1, g)
			}
		}
	}
}
