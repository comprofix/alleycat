package entity

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/physics"
)

// floorOnly is a world with nothing to stand on but the floor.
type floorOnly struct{}

func (floorOnly) Support(*Freddy) bool { return false }

// ledge is a world with one surface at Y across all X.
type ledge struct{ y int }

func (l ledge) Support(f *Freddy) bool {
	if f.Y >= 0 && physics.Band(f.Y) == l.y {
		f.Y = l.y
		return true
	}
	return false
}

// fence is a world with a fence top at Y = 96 that holds Freddy unless he
// is behind it.
type fence struct{}

func (fence) Support(f *Freddy) bool {
	if f.Depth == Behind || f.Y < 0 || physics.Band(f.Y) != 96 {
		return false
	}
	f.Y, f.Depth = 96, OnFence
	return true
}

// inAlley returns Freddy standing on the floor in the middle of the alley.
func inAlley() *Freddy {
	f := NewFreddy()
	f.enterWait, f.enterSlices = 0, 0
	f.X = 150
	return f
}

func run(f *Freddy, in Controls, w World, n int) {
	for range n {
		f.Update(in, w)
	}
}

func TestEnterFromLeft(t *testing.T) {
	f := NewFreddy()
	if !f.Hidden() {
		t.Fatal("Freddy should wait off-screen first")
	}
	run(f, Controls{}, floorOnly{}, config.FreddyEnterWait)
	if f.Hidden() {
		t.Fatal("still hidden after the wait")
	}
	for i := 1; i < config.FreddyEnterSlices; i++ {
		f.Update(Controls{}, floorOnly{})
		if px, ok := f.EnterSlice(); !ok || px != i*8 {
			t.Fatalf("slice %d: got %d px (entering %v), want %d", i, px, ok, i*8)
		}
	}
	f.Update(Controls{}, floorOnly{})
	if _, ok := f.EnterSlice(); ok || f.X != config.FreddyEnterLeftX+config.FreddyEnterStep {
		t.Errorf("after walking in: X=%d entering=%v, want X=8 and in", f.X, ok)
	}
}

func TestWalkAcceleratesToEight(t *testing.T) {
	f := inAlley()
	f.X = 8
	right := Controls{H: 1}
	f.Update(right, floorOnly{})
	if f.RunSpeed != config.FreddyWalkStartSpeed {
		t.Fatalf("speed after first step = %d, want 2", f.RunSpeed)
	}
	run(f, right, floorOnly{}, 40)
	if f.RunSpeed != config.FreddyWalkMaxSpeed {
		t.Errorf("speed after 41 updates = %d, want 8", f.RunSpeed)
	}
	f.Update(Controls{H: -1}, floorOnly{})
	if f.RunSpeed != config.FreddyWalkStartSpeed {
		t.Errorf("speed after reversing = %d, want 2", f.RunSpeed)
	}
}

func TestWalkClampsToAlley(t *testing.T) {
	f := inAlley()
	run(f, Controls{H: 1}, floorOnly{}, 200)
	if f.X != config.FreddyMaxX {
		t.Errorf("X = %d, want clamp %d", f.X, config.FreddyMaxX)
	}
	run(f, Controls{H: -1}, floorOnly{}, 200)
	if f.X != config.FreddyMinX {
		t.Errorf("X = %d, want clamp %d", f.X, config.FreddyMinX)
	}
}

func TestStandingJumpReturnsToFloor(t *testing.T) {
	f := inAlley()
	f.Update(Controls{V: -1}, floorOnly{})
	if !f.Airborne() || f.Arc.Step != (15-2)*16 {
		t.Fatalf("jump not started or wrong step %d", f.Arc.Step)
	}
	top := f.Y
	for f.Airborne() {
		f.Update(Controls{}, floorOnly{})
		top = min(top, f.Y)
	}
	if f.Y != config.FloorY || f.X != 150 {
		t.Errorf("landed at %d,%d, want 150,180", f.X, f.Y)
	}
	if top >= config.FloorY-20 {
		t.Errorf("standing jump only reached Y=%d", top)
	}
}

func TestNoSteeringInTheAir(t *testing.T) {
	f := inAlley()
	f.X = 20
	run(f, Controls{H: 1}, floorOnly{}, 30)
	f.Update(Controls{H: 1, V: -1}, floorOnly{})
	speed := f.RunSpeed
	if speed != config.FreddyWalkMaxSpeed-config.FreddyJumpSpeedCost {
		t.Fatalf("run speed after take-off = %d, want 6", speed)
	}
	x := f.X
	f.Update(Controls{H: -1}, floorOnly{})
	if f.X != x+speed {
		t.Errorf("moved to %d in the air holding Left, want %d", f.X, x+speed)
	}
}

func TestLandsOnlyWhileDescending(t *testing.T) {
	f := inAlley()
	w := ledge{y: 160} // passed through on the way up, landed on coming down
	f.Update(Controls{V: -1}, w)
	passed := false
	for f.Airborne() {
		if f.Arc.Dir == physics.Rising && physics.Band(f.Y) == 160 {
			passed = true
		}
		f.Update(Controls{}, w)
	}
	if !passed {
		t.Fatal("the jump never rose through the ledge's band")
	}
	if f.Y != 160 {
		t.Errorf("Y = %d, want to land on the ledge at 160", f.Y)
	}
}

func TestDropFromLedge(t *testing.T) {
	f := inAlley()
	f.Y = 96
	w := ledge{y: 96}
	f.Update(Controls{V: 1}, w)
	if f.Arc.Dir != physics.Falling || f.Arc.Step != config.FreddyDropStep {
		t.Fatalf("drop not started: %+v", f.Arc)
	}
	// It must not land back on the ledge it left, though it is still in
	// its band for the first updates.
	run(f, Controls{}, w, 3)
	if !f.Airborne() {
		t.Error("landed during the no-landing period")
	}
	for f.Airborne() {
		f.Update(Controls{}, w)
	}
	if f.Y != config.FloorY {
		t.Errorf("Y = %d, want the floor", f.Y)
	}
}

func TestDownOnFloorDoesNothing(t *testing.T) {
	f := inAlley()
	f.Update(Controls{V: 1}, floorOnly{})
	if f.Airborne() || f.Y != config.FloorY {
		t.Error("Down on the floor should not drop")
	}
}

func TestSeatedCannotWalk(t *testing.T) {
	f := inAlley()
	f.Y, f.Seat = 40, JustSeated
	f.Update(Controls{H: 1}, ledge{y: 40})
	if f.X != 150 || f.RunSpeed != config.FreddyLineRunSpeed {
		t.Errorf("seated: X=%d speed=%d, want 150 and 6", f.X, f.RunSpeed)
	}
	f.Update(Controls{H: 1, V: -1}, ledge{y: 40})
	if !f.Airborne() || f.Arc.Step != (15-6)*16 || f.RunSpeed != 4 {
		t.Errorf("jump from line: step %d speed %d, want 144 and 4", f.Arc.Step, f.RunSpeed)
	}
}

func TestTopOfScreen(t *testing.T) {
	f := inAlley()
	f.Y = 0
	f.Update(Controls{V: -1}, floorOnly{})
	for f.Arc.Dir == physics.Rising {
		f.Update(Controls{}, floorOnly{})
	}
	if f.Y < config.FreddyTopY {
		t.Errorf("Y = %d, above the limit %d", f.Y, config.FreddyTopY)
	}
}

func TestBehindFenceReenters(t *testing.T) {
	f := inAlley()
	f.X, f.Y, f.Depth = 200, 96, OnFence
	f.Update(Controls{V: -1}, fence{})
	if f.Depth != Behind {
		t.Fatal("jumping off the fence top should put Freddy behind it")
	}
	for range 200 {
		f.Update(Controls{}, fence{})
		if f.Hidden() {
			break
		}
	}
	if !f.Hidden() || f.X != config.FreddyEnterRightX || f.Depth != InFront {
		t.Errorf("after sinking behind the fence: hidden=%v X=%d depth=%d, want re-entry from the right",
			f.Hidden(), f.X, f.Depth)
	}
}

// Sinking out of sight behind the fence is reported for the pause.
func TestSankBehindFence(t *testing.T) {
	f := inAlley()
	f.Y = 60
	f.KnockDown()
	for i := 0; !f.Sank && i < 300; i++ {
		f.Update(Controls{}, floorOnly{})
	}
	if !f.Sank || !f.Hidden() {
		t.Fatalf("Sank %v, hidden %v: want him reset after sinking behind the fence", f.Sank, f.Hidden())
	}
}
