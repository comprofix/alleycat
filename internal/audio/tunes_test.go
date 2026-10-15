package audio

import (
	"math"
	"testing"
	"time"
)

func TestNoteHz(t *testing.T) {
	for name, want := range map[string]float64{"A4": 440, "C5": 523.25, "G#4": 415.30, "Bb3": 233.08, "A3": 220} {
		if got := noteHz(name); math.Abs(got-want) > 0.01 {
			t.Errorf("%s = %.2f Hz, want %.2f", name, got, want)
		}
	}
}

func TestTuneParse(t *testing.T) {
	n := tune("C5:2 r:1.5 E5:0.5")
	if len(n) != 3 || n[1].Hz != 0 || n[0].Dur != Ticks(2) || n[2].Dur != Ticks(0.5) {
		t.Errorf("parsed %v", n)
	}
	for _, bad := range []string{"C5", "C5:0", "H5:2", "C:2"} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%q parsed without complaint", bad)
				}
			}()
			tune(bad)
		}()
	}
}

func ticksOf(n []Note) float64 { return total(n).Seconds() / Ticks(1).Seconds() }

// Each tune fits its place in the game (DESIGN §7.4).
func TestTuneLengths(t *testing.T) {
	for _, c := range []struct {
		name   string
		n      []Note
		lo, hi float64 // ticks
	}{
		{"title", Title(), 780, 900},      // about 43–49 s, played once
		{"room left", RoomLeft(), 28, 28}, // 14 notes of 2 ticks
		{"failure", Failure(), 28, 28},
		{"kiss", Kiss(), 120, 140}, // a pass of the loop
		{"dance", Dance(), 40, 40},
	} {
		if got := ticksOf(c.n); got < c.lo-0.01 || got > c.hi+0.01 {
			t.Errorf("%s lasts %.1f ticks, want %.0f–%.0f", c.name, got, c.lo, c.hi)
		}
	}
	if w := Wipe(500 * time.Millisecond); total(w) < 500*time.Millisecond || w[1].Hz <= w[0].Hz {
		t.Errorf("wipe %v: want a rising figure lasting the change", w)
	}
}

func TestTally(t *testing.T) {
	if a, b := TallyNote(0, 2)[0], TallyNote(1, 2)[0]; b.Hz <= a.Hz || a.Dur != Ticks(2) {
		t.Errorf("tally notes %v then %v: want a rise, 2 ticks each", a, b)
	}
	tr := TallyTrill(3, 30)
	if len(tr) != 30 || tr[0].Hz == tr[1].Hz || tr[0].Hz != tr[2].Hz {
		t.Errorf("trill %v: want two notes alternating every tick", tr[:4])
	}
}
