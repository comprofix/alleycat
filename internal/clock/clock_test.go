package clock

import (
	"math"
	"testing"
)

func TestTickRateOverTime(t *testing.T) {
	c := New(60)
	ticks, steps := 0, 0
	const frames = 60 * 100 // 100 seconds
	for range frames {
		tk, st := c.Advance()
		ticks += tk
		steps += st
	}
	wantTicks := TickHz * 100
	if math.Abs(float64(ticks)-wantTicks) > 1 {
		t.Errorf("ticks in 100 s = %d, want about %.1f", ticks, wantTicks)
	}
	if math.Abs(float64(steps)-2*wantTicks) > 1 {
		t.Errorf("Freddy steps in 100 s = %d, want about %.1f", steps, 2*wantTicks)
	}
}

func TestAtMostOnePerFrameAt60TPS(t *testing.T) {
	c := New(60)
	for i := range 6000 {
		tk, st := c.Advance()
		if tk > 1 || st > 1 {
			t.Fatalf("frame %d: ticks=%d steps=%d, want at most 1 each", i, tk, st)
		}
	}
}

func TestFreddyStepsAreEvenlySpaced(t *testing.T) {
	// At 36.4 steps/s and 60 TPS, gaps between steps are 1 or 2 frames.
	c := New(60)
	last := -1
	for i := range 6000 {
		_, st := c.Advance()
		if st == 0 {
			continue
		}
		if last >= 0 {
			if gap := i - last; gap < 1 || gap > 2 {
				t.Fatalf("gap of %d frames between Freddy steps at frame %d", gap, i)
			}
		}
		last = i
	}
}

// Total counts master ticks, and Time runs smoothly between them.
func TestTotalAndTime(t *testing.T) {
	c := New(60)
	prev := c.Time()
	sum := 0
	for range 600 {
		ticks, _ := c.Advance()
		sum += ticks
		if now := c.Time(); now <= prev {
			t.Fatalf("Time went from %v to %v: want it always moving on", prev, now)
		} else {
			prev = now
		}
	}
	if c.Total() != int64(sum) {
		t.Errorf("Total %d, want %d", c.Total(), sum)
	}
	if got := c.Time(); got < float64(sum) || got >= float64(sum)+1 {
		t.Errorf("Time %v, want between %d and %d", got, sum, sum+1)
	}
}
