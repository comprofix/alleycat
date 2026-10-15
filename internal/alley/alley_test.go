package alley

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/session"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

// newSession is a game at the given difficulty.
func newSession(difficulty int) *session.Session {
	s := session.New(config.Kitten)
	s.Difficulty = difficulty
	return s
}

// standing returns an alley with Freddy fully walked in.
func standing(difficulty int) *Alley {
	a := New(newSession(difficulty), newRNG())
	for a.Freddy.Hidden() || entering(a.Freddy) {
		a.Freddy.Update(entity.Controls{}, a)
	}
	return a
}

func entering(f *entity.Freddy) bool { _, ok := f.EnterSlice(); return ok }

func TestCanSupport(t *testing.T) {
	a := standing(0) // cans: 24 S, 64 T, ...
	f := a.Freddy
	cases := []struct {
		x, y   int
		wantOK bool
		wantY  int
	}{
		{64, 136, true, config.TallCanY},
		{86, 140, true, config.TallCanY},  // X + 22 is the last
		{87, 140, false, 0},               // one past
		{63, 136, false, 0},               // before the can
		{30, 143, true, config.ShortCanY}, // short can at 24
		{64, 150, false, 0},               // below the band
	}
	for _, c := range cases {
		f.X, f.Y = c.x, c.y
		ok := a.Support(f)
		if ok != c.wantOK || (ok && f.Y != c.wantY) {
			t.Errorf("at %d,%d: support %v Y=%d, want %v Y=%d", c.x, c.y, ok, f.Y, c.wantOK, c.wantY)
		}
	}
}

func TestFenceSupportAndDepth(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	f.X, f.Y = 200, 101
	if !a.Support(f) || f.Y != config.FenceTopY || f.Depth != entity.OnFence {
		t.Fatalf("fence: Y=%d depth=%d, want 96 on the fence", f.Y, f.Depth)
	}
	f.Y, f.Depth = 99, entity.Behind
	if a.Support(f) {
		t.Error("the fence top must not hold Freddy from behind")
	}
}

func TestLineSupportSnapsAndSeats(t *testing.T) {
	a := standing(0)
	l := &a.Lines[1]
	for i := range l.Pieces {
		l.Pieces[i] = Piece{}
	}
	l.Offset = 0
	l.Pieces[5] = Piece{Small, NoLaundry} // laundry over X 80–87
	f := a.Freddy
	f.X, f.Y = 75, 44 // probe X + 10 = 85
	if !a.Support(f) || f.Y != 40 || f.X != 72 || f.Seat != entity.JustSeated {
		t.Errorf("got X=%d Y=%d seat=%d, want 72,40 seated", f.X, f.Y, f.Seat)
	}
	f.X, f.Y, f.Seat = 80, 44, entity.NotSeated // probe 90: empty half
	if a.Support(f) {
		t.Error("supported over an empty half")
	}
}

func TestOnlyOneLineMoves(t *testing.T) {
	a := standing(3)
	rng := newRNG()
	moved := map[int]bool{}
	for range 400 {
		m := a.Moving
		a.scrollTick(rng)
		moved[m] = true
		for i := range a.Lines {
			if i != a.Moving && !a.Lines[i].AtBoundary() {
				t.Fatalf("line %d stopped mid-piece", i)
			}
		}
	}
	if len(moved) < 2 {
		t.Errorf("only lines %v moved in 400 ticks", moved)
	}
}

func TestLineDirections(t *testing.T) {
	rng := newRNG()
	for row, want := range []int{1, -1, 1} {
		l := newLine(row, 0, rng)
		marker := Piece{Small + 2, BigLeft} // never generated
		l.Pieces[3] = marker
		for range 4 {
			l.step(0, rng)
		}
		got := 0
		for i, p := range l.Pieces {
			if p == marker {
				got = i - 3
			}
		}
		if got != want {
			t.Errorf("line %d: piece moved %d, want %d", row, got, want)
		}
	}
}

func TestCarriedOffTheEnd(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Moving = 0 // top line moves right
	f.X, f.Y, f.Seat = config.FreddyMaxX-2, config.LineY[0], entity.Seated
	a.step(&a.Lines[0], newRNG())
	if f.Seat != entity.NotSeated || !f.Airborne() || f.RunSpeed != config.FreddyLineFallRunSpeed {
		t.Errorf("should fall off the line end: seat=%d airborne=%v", f.Seat, f.Airborne())
	}
	f.X, f.Y, f.Seat = 100, config.LineY[0], entity.Seated
	f.Arc.Dir = 0
	a.step(&a.Lines[0], newRNG())
	if f.X != 104 {
		t.Errorf("carried to X=%d, want 104", f.X)
	}
}

func TestAirborneFreddyHoldsLine(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Moving = 1
	l := &a.Lines[1]
	l.step(a.gap, newRNG()) // now mid-piece
	off := l.Offset
	f.Y = 40
	f.Arc.Dir = 1 // falling through the middle band
	a.scrollTick(newRNG())
	if l.Offset != off || a.Moving != 1 {
		t.Error("the middle line moved while Freddy was in the air in its band")
	}
}

func TestLaundryGapsGrowWithDifficulty(t *testing.T) {
	count := func(gap int) int {
		rng, empty := newRNG(), 0
		for range 10000 {
			if newPiece(2, gap, rng) == (Piece{}) {
				empty++
			}
		}
		return empty
	}
	easy, hard := count(config.ForLevel(0).LaundryGapChance), count(config.ForLevel(7).LaundryGapChance)
	if easy >= hard {
		t.Errorf("empty pieces: %d at difficulty 0, %d at 7; want more at 7", easy, hard)
	}
	// Bottom line, difficulty 0: a gap (6/256), no fill (1/256 of the rest),
	// or two missing small items (160/256 × 1/16 of the rest): about 6.4%.
	if easy < 500 || easy > 800 {
		t.Errorf("empty pieces at difficulty 0 = %d of 10000, want about 640", easy)
	}
}

// The alley stands still after Freddy sinks behind the fence or thumps onto
// the floor (DESIGN §12).
func TestPauses(t *testing.T) {
	a := standing(0)
	a.Freddy.Sank = true
	a.checkPauses()
	if a.Freeze != config.FenceDropTicks || a.Freddy.Sank {
		t.Errorf("after sinking: freeze %d, Sank %v; want %d, cleared", a.Freeze, a.Freddy.Sank, config.FenceDropTicks)
	}

	a = standing(0)
	a.Freddy.Thumped = true
	a.checkPauses()
	if a.Freeze != config.FloorThumpTicks || a.Freddy.Thumped {
		t.Errorf("after a thump: freeze %d, Thumped %v; want %d, cleared", a.Freeze, a.Freddy.Thumped, config.FloorThumpTicks)
	}

	// Sitting on a clothesline marks him.
	a = standing(0)
	f := a.Freddy
	for i := range a.Lines {
		for x := 8; x < 290; x += 8 {
			f.X, f.Y = x, config.LineY[i]
			f.OnLine = false
			if a.Support(f) {
				if !f.OnLine {
					t.Fatal("sitting on a line did not set the mark")
				}
				return
			}
		}
	}
	t.Fatal("found no line to sit on")
}
