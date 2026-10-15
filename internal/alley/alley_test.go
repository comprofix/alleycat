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

// has reports whether the last update made sound s.
func has(a *Alley, s Sound) bool {
	for _, got := range a.Sounds {
		if got == s {
			return true
		}
	}
	return false
}

// Sinking behind the fence holds the alley still for 7 ticks and crackles
// (DESIGN §7.3, §12).
func TestFenceDrop(t *testing.T) {
	a := standing(0)
	a.Freddy.Sank = true
	a.Sounds = a.Sounds[:0]
	a.checkFreddy()
	if a.Freeze != config.FenceDropTicks || a.Freddy.Sank || !has(a, SoundFenceDrop) {
		t.Errorf("freeze %d, Sank %v, sounds %v; want %d, cleared, the crackle",
			a.Freeze, a.Freddy.Sank, a.Sounds, config.FenceDropTicks)
	}
}

// Landing on a can clicks once; leaping off clicks and clears the mark. With
// the dog out neither clicks and the mark stays (DESIGN §7.3).
func TestCanClicks(t *testing.T) {
	a := standing(0) // a tall can at X 64
	f := a.Freddy
	f.X, f.Y = 64, 136
	if !a.Support(f) || !has(a, SoundCanOn) {
		t.Fatalf("landing on a can: sounds %v, want the click", a.Sounds)
	}
	a.Sounds = a.Sounds[:0]
	a.Support(f)
	if has(a, SoundCanOn) {
		t.Error("the landing click repeated while standing on the can")
	}
	f.Leapt = true
	a.checkFreddy()
	if !has(a, SoundCanOff) || a.canMark {
		t.Errorf("leaping off: sounds %v, mark %v; want the click and the mark cleared", a.Sounds, a.canMark)
	}

	a = standing(0)
	f = a.Freddy
	a.Dog.Active = true
	f.X, f.Y = 64, 136
	a.Support(f)
	f.Leapt = true
	a.checkFreddy()
	if len(a.Sounds) != 0 || !a.canMark {
		t.Errorf("with the dog out: sounds %v, mark %v; want silence and the mark kept", a.Sounds, a.canMark)
	}
}

// Each Update starts the sound list afresh.
func TestSoundsReset(t *testing.T) {
	a := standing(0)
	a.Sounds = append(a.Sounds, SoundBonk)
	a.Update(1, 0, entity.Controls{}, newRNG())
	if has(a, SoundBonk) {
		t.Error("sounds carried over from the last update")
	}
}

// Catching a mouse in the air blips; a mouse touching a seated Freddy
// squeaks (DESIGN §7.3).
func TestMouseSounds(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	f.X, f.Y = a.Mice[1].X, config.MouseY[1]
	a.Mice[1].Caught = false
	a.Sounds = a.Sounds[:0]
	if !a.mouseTouch(1) || !has(a, SoundMouseCaught) {
		t.Errorf("catching a mouse: sounds %v, want the blip", a.Sounds)
	}

	a = standing(0)
	f = a.Freddy
	f.X, f.Y = a.Mice[1].X, config.MouseY[1]
	f.Seat = entity.Seated
	a.Mice[1].Caught = false
	a.Sounds = a.Sounds[:0]
	if !a.mouseTouch(1) || !has(a, SoundSqueak) {
		t.Errorf("knocked off a line: sounds %v, want the squeak", a.Sounds)
	}
}

// A thrown object's hit BONKs once (DESIGN §7.3).
func TestBonkSound(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Thrown = Thrown{Active: true, X: f.X, Y: f.Y, Dir: 1}
	a.checkThrownHit()
	if !has(a, SoundBonk) {
		t.Fatalf("hit: sounds %v, want BONK", a.Sounds)
	}
	a.Sounds = a.Sounds[:0]
	a.checkThrownHit()
	if has(a, SoundBonk) {
		t.Error("BONK repeated for the same object")
	}
}

// The trash cat's notes play once per touch.
func TestTrashCatSound(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	f.X, f.Y = 64, config.TallCanY
	a.TrashCat = TrashCat{Phase: 10, X: f.X, Top: f.Y, H: 8}
	a.controls(entity.Controls{}, newRNG())
	a.controls(entity.Controls{}, newRNG())
	n := 0
	for _, s := range a.Sounds {
		if s == SoundTrashCat {
			n++
		}
	}
	if n != 1 {
		t.Errorf("trash cat notes played %d times for one touch, want 1", n)
	}
	a.TrashCat = TrashCat{}
	a.controls(entity.Controls{}, newRNG())
	a.TrashCat = TrashCat{Phase: 10, X: f.X, Top: f.Y, H: 8}
	a.controls(entity.Controls{}, newRNG())
	if a.Sounds[len(a.Sounds)-1] != SoundTrashCat {
		t.Error("a new touch did not play the notes again")
	}
}
