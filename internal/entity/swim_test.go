package entity

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

func swimmer(y int) *Freddy {
	f := NewFreddy()
	f.StartSwim()
	f.Y = y
	return f
}

func TestSwimTopSpeedIsCappedByTheAccumulator(t *testing.T) {
	for _, c := range []struct{ maxX, want int }{{4, 4}, {6, 6}, {12, 6}} {
		f := swimmer(100)
		top := 0
		for range 40 {
			x := f.X
			f.SwimUpdate(Controls{H: 1}, c.maxX, 4)
			top = max(top, f.X-x)
		}
		if top != c.want {
			t.Errorf("cap %d: top speed %d, want %d", c.maxX, top, c.want)
		}
	}
}

func TestSwimGlidesThenStops(t *testing.T) {
	f := swimmer(100)
	f.X = config.FreddyMinX
	for range 20 {
		f.SwimUpdate(Controls{H: 1}, 12, 4)
	}
	var moved []int
	for range 60 {
		x := f.X
		f.SwimUpdate(Controls{}, 12, 4)
		moved = append(moved, f.X-x)
	}
	if moved[0] != 6 {
		t.Errorf("he should glide on at full speed at first, moved %d", moved[0])
	}
	for i := 1; i < len(moved); i++ {
		if moved[i] > moved[i-1] {
			t.Fatalf("the glide should only slow: %v", moved)
		}
	}
	if last := moved[len(moved)-1]; last != 0 || f.Dir != 0 {
		t.Errorf("the glide should end in a stop: %v, dir %d", moved, f.Dir)
	}
}

func TestSwimFloatsUpWithNoArrows(t *testing.T) {
	f := swimmer(150)
	for range 60 {
		f.SwimUpdate(Controls{}, 12, 4)
	}
	if f.Y >= 150-40 {
		t.Errorf("he should float up about 1 px an update, Y %d after 60", f.Y)
	}
}

func TestSwimVerticalCapAndLimits(t *testing.T) {
	for _, maxY := range []int{3, 4} {
		f := swimmer(10)
		top := 0
		for range 80 {
			y := f.Y
			f.SwimUpdate(Controls{V: 1}, 12, maxY)
			top = max(top, f.Y-y)
		}
		if top != maxY || f.Y != config.SwimMaxY {
			t.Errorf("cap %d: top speed %d, Y %d", maxY, top, f.Y)
		}
	}
	f := swimmer(100)
	for range 80 {
		f.SwimUpdate(Controls{V: -1, H: -1}, 12, 4)
	}
	if f.Y != config.SwimMinY || f.X != config.FreddyMinX {
		t.Errorf("he should stop at the surface and the side, at %d,%d", f.X, f.Y)
	}
}

func TestSurfacingNeedsTheUpwardPose(t *testing.T) {
	// Drifting up while swimming hard sideways: the sideways pose, no air.
	f := swimmer(20)
	for range 80 {
		if f.SwimUpdate(Controls{H: 1}, 12, 4) {
			t.Fatal("drifting to the top sideways should not refill his air")
		}
		if f.X >= f.MaxX {
			f.X = 40
		}
	}
	// Swimming up: he surfaces.
	f = swimmer(20)
	for range 20 {
		if f.SwimUpdate(Controls{V: -1}, 12, 4) {
			return
		}
	}
	t.Fatal("swimming up to the top should refill his air")
}

func TestSinkingIgnoresInput(t *testing.T) {
	f := swimmer(100)
	f.Sink()
	for range 10 {
		y := f.Y
		f.SwimUpdate(Controls{V: -1, H: 1}, 12, 4)
		if f.Y-y != 2 || f.Swim.Pose != SwimSink {
			t.Fatalf("sinking: moved %d, pose %d", f.Y-y, f.Swim.Pose)
		}
	}
}
