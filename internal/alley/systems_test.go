package alley

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// tick runs n master ticks with no Freddy updates.
func tick(a *Alley, n int) {
	rng := newRNG()
	for range n {
		a.Update(1, 0, entity.Controls{}, rng)
	}
}

func TestWindowsOpenOnlyWhenFreddyIsHigh(t *testing.T) {
	a := standing(0) // Freddy on the floor
	tick(a, 500)
	if a.Window.Phase != 0 {
		t.Fatal("a window opened while Freddy was on the floor")
	}
	a.Freddy.Y, a.Freddy.Depth = config.FenceTopY, entity.OnFence
	opened := false
	for range 200 {
		tick(a, 1)
		if a.Window.Phase != 0 {
			opened = true
			break
		}
	}
	if !opened {
		t.Error("no window opened in 200 ticks with Freddy on the fence")
	}
}

func TestWindowPhasesAndHold(t *testing.T) {
	a := standing(0)
	a.Freddy.Y = config.FenceTopY
	a.Mice = [3]Mouse{{X: 300}, {X: 300}, {X: 300}} // out of the way
	a.Window = Window{Index: 11, Phase: config.WindowPhaseStart, hold: 45, throws: 0}
	a.Thrown = Thrown{}
	n := 0
	for a.Window.Phase > 0 && n < 1000 {
		a.updateWindow(newRNG())
		a.Tick++
		n++
	}
	// 14 opening steps, the hold, then 15 closing steps.
	if n < 14+45+14 || n > 14+45+16 {
		t.Errorf("window cycle took %d ticks, want about %d", n, 14+45+15)
	}
}

func TestEnteringAnOpenWindow(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Window = Window{Index: 5, Phase: 20}
	r := WindowRect(5)
	f.X, f.Y = r.X, r.Y+2
	f.Arc = physics.Arc{Dir: physics.Falling, Speed: 2}
	if !a.windowTouch() || a.Exit != ToRoom {
		t.Fatalf("falling into an open window: exit %d, want ToRoom", a.Exit)
	}
	a.Exit = NoExit
	f.Arc.Dir = physics.Rising
	a.windowTouch()
	if a.Exit != NoExit {
		t.Error("entered while rising")
	}
	a.Session.FeliciaWaiting = true
	f.Arc.Dir = physics.Falling
	a.windowTouch()
	if a.Exit != ToHeart {
		t.Error("Felicia's window should lead to the heart room")
	}
}

func TestThrownHitCostsALife(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	f.X, f.Y = 100, 40
	lives := a.Session.Lives
	a.Thrown = Thrown{Active: true, X: 100, Y: 40, Dir: 1, speed: config.ThrownSpeed, age: 1}
	a.checkThrownHit()
	if a.Session.Lives != lives-1 || a.Freeze != config.BonkTicks || a.Impact.Kind != Bonk {
		t.Errorf("after a hit: lives %d (was %d), freeze %d, impact %d", a.Session.Lives, lives, a.Freeze, a.Impact.Kind)
	}
	if !f.Knocked || f.Depth != entity.Behind || a.Thrown.Dir != -1 {
		t.Errorf("Freddy should be knocked behind the fence and the object bounce: knocked=%v depth=%d dir=%d",
			f.Knocked, f.Depth, a.Thrown.Dir)
	}
	a.checkThrownHit()
	if a.Session.Lives != lives-1 {
		t.Error("one object took two lives")
	}
}

func TestThrownSinksBehindFence(t *testing.T) {
	a := standing(0)
	a.Freddy.X = 0
	a.Thrown = Thrown{Active: true, X: 200, Y: 88, Kind: 0, Dir: 1, speed: config.ThrownSpeed, age: 1}
	for range 100 {
		a.updateThrown()
		if !a.Thrown.Active {
			return
		}
	}
	t.Error("the object never sank out of sight")
}

func TestCatchingAMouse(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Mice[0] = Mouse{X: 100}
	f.X, f.Y = 96, 0
	f.Arc = physics.Arc{Dir: physics.Rising, Speed: 3}
	if !a.mouseTouch(0) || !a.Mice[0].Caught || a.Session.Score != 90 {
		t.Fatalf("catch: caught=%v score=%d, want 90", a.Mice[0].Caught, a.Session.Score)
	}
	f.X, f.Y = 200, config.FloorY // away from the mouse
	a.Tick += config.MouseRespawnTicks - 1
	a.mouseTurn = len(a.Mice) - 1
	a.updateMice(newRNG())
	if !a.Mice[0].Caught {
		t.Error("the mouse came back too soon")
	}
	a.Tick++
	a.mouseTurn = len(a.Mice) - 1
	a.updateMice(newRNG())
	if a.Mice[0].Caught {
		t.Error("the mouse did not come back after 54 ticks")
	}
}

func TestMouseKnocksSeatedFreddyOff(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	a.Mice[1] = Mouse{X: 100}
	f.X, f.Y, f.Seat = 96, config.LineY[1], entity.Seated
	a.mouseTouch(1)
	if f.Seat != entity.NotSeated || !f.Airborne() || a.Freeze != config.SqueakTicks || a.Mice[1].Caught {
		t.Errorf("seat %d airborne %v freeze %d caught %v", f.Seat, f.Airborne(), a.Freeze, a.Mice[1].Caught)
	}
}

func TestTrashCatCallsTheDog(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	c := a.Cans[1] // the tall can at 64
	f.X, f.Y = c.X+2, config.TallCanY
	a.lastCan = -1
	// Standing on a can starts the cat at once.
	a.updateTrashCat(newRNG())
	if a.TrashCat.Phase == 0 {
		t.Fatal("the trash cat did not start with Freddy on a can")
	}
	a.TrashCat = TrashCat{Phase: 10, X: f.X, Top: f.Y + 5, H: 8}
	a.updateTrashCat(newRNG())
	if !a.Dog.Forced {
		t.Error("touching the trash cat should call the dog")
	}
	in := a.controls(entity.Controls{}, newRNG())
	if in.V != 1 || in.H == 0 {
		t.Errorf("Freddy on the can should be bumped off: %+v", in)
	}
}

func TestDogCatchCostsALife(t *testing.T) {
	a := standing(0)
	f := a.Freddy
	f.X = 150
	a.Dog = entity.Dog{X: 140, Dir: 1, Active: true}
	lives := a.Session.Lives
	a.checkDogCatch()
	if a.Dog.Capture == 0 || !f.Hidden() {
		t.Fatal("the dog should catch Freddy on the floor")
	}
	for range 400 {
		a.updateDog(newRNG())
		if !a.Dog.Out() {
			break
		}
	}
	if a.Session.Lives != lives-1 {
		t.Errorf("lives %d after the fight, want %d", a.Session.Lives, lives-1)
	}
	if a.Dog.Out() {
		t.Error("the dog never left")
	}
}

func TestDogOnlyAppearsWhileFreddyIsOnTheFloor(t *testing.T) {
	a := standing(7) // highest appear chance
	a.Freddy.Y = config.FenceTopY
	for range 500 {
		a.updateDog(newRNG())
	}
	if a.Dog.Out() {
		t.Error("the dog came out while Freddy was up on the fence")
	}
}

// TestSoak runs every difficulty with random input to catch panics, stuck
// loops and impossible states.
func TestSoak(t *testing.T) {
	for d := range 8 {
		rng := newRNG()
		a := New(newSession(d), rng)
		a.Session.Lives = 1000
		in := entity.Controls{}
		for frame := range 30000 {
			if frame%20 == 0 {
				in = entity.Controls{H: rng.IntN(3) - 1, V: rng.IntN(3) - 1}
			}
			a.Update(frame%3%2, frame%2, in, rng)
			if a.Exit != NoExit {
				a.Exit = NoExit
			}
			f := a.Freddy
			if f.Y < config.FreddyTopY || f.Y > config.FloorY || f.X < 0 || f.X > config.FreddyEnterRightX {
				t.Fatalf("difficulty %d frame %d: Freddy at %d,%d", d, frame, f.X, f.Y)
			}
		}
	}
}
