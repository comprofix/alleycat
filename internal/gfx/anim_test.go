package gfx

import "testing"

func TestAnimatorCycles(t *testing.T) {
	a := Animator{TicksPerFrame: 3, Frames: 2}
	var got []int
	for range 7 {
		got = append(got, a.Frame())
		a.Advance(1)
	}
	want := []int{0, 0, 0, 1, 1, 1, 0}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("frames = %v, want %v", got, want)
		}
	}
}

func TestAnimatorZeroValueIsStill(t *testing.T) {
	var a Animator
	a.Advance(10)
	if a.Frame() != 0 {
		t.Error("zero-value animator should not move")
	}
}
