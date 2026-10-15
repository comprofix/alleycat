package physics

import "testing"

func TestOverlapsCountsTouching(t *testing.T) {
	a := Rect{0, 0, 10, 10}
	cases := []struct {
		b    Rect
		want bool
	}{
		{Rect{5, 5, 10, 10}, true},
		{Rect{10, 0, 5, 5}, true}, // touching right edge
		{Rect{0, 10, 5, 5}, true}, // touching bottom edge
		{Rect{11, 0, 5, 5}, false},
		{Rect{0, -6, 5, 5}, false},
	}
	for _, c := range cases {
		if got := a.Overlaps(c.b); got != c.want {
			t.Errorf("%v overlaps %v = %v, want %v", a, c.b, got, c.want)
		}
		if got := c.b.Overlaps(a); got != c.want {
			t.Errorf("%v overlaps %v = %v, want %v (reversed)", c.b, a, got, c.want)
		}
	}
}

func TestBand(t *testing.T) {
	for y, want := range map[int]int{0: 0, 7: 0, 8: 8, 101: 96, 103: 96, 104: 104} {
		if got := Band(y); got != want {
			t.Errorf("Band(%d) = %d, want %d", y, got, want)
		}
	}
}

// rise simulates a jump from speed 8 and returns the total height and the
// number of updates to the top.
func rise(step uint8) (height, updates int) {
	var a Arc
	a.Start(Rising, 8, step)
	for a.Dir == Rising {
		a.Gravity(4)
		if a.Dir != Rising {
			break
		}
		height += a.Speed
		updates++
	}
	return height, updates
}

func TestRunningJumpMatchesReference(t *testing.T) {
	// A full-speed jump uses step (15−8)×16 = 112. R1 measured a rise of
	// about 60 px peaking 12–13 video frames after take-off (DESIGN
	// §11.1.1): 16 updates at 36.4/s is 13.2 frames at 30 fps.
	h, n := rise(112)
	if h != 61 || n != 16 {
		t.Errorf("running jump: %d px in %d updates, want 61 px in 16", h, n)
	}
	sh, _ := rise(208) // standing jump, run speed 2
	if sh >= h {
		t.Errorf("standing jump %d px should be lower than running jump %d px", sh, h)
	}
}

func TestFallSpeedCapped(t *testing.T) {
	var a Arc
	a.Start(Falling, 1, 255)
	for range 100 {
		a.Gravity(4)
	}
	if a.Speed != 4 {
		t.Errorf("fall speed = %d, want cap 4", a.Speed)
	}
}

func TestRiseTurnsToFallAtSpeedOne(t *testing.T) {
	var a Arc
	a.Start(Rising, 1, 32)
	a.Gravity(4) // first update always underflows
	if a.Dir != Falling || a.Speed != 1 {
		t.Errorf("got dir %d speed %d, want falling at 1", a.Dir, a.Speed)
	}
}
