package audio

import (
	"math"
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

func newTestPatter() *Patter { return NewPatter(rand.New(rand.NewPCG(1, 2))) }

// run ticks p n times in state s and returns the click pitch at each tick,
// 0 where nothing starts.
func runPatter(p *Patter, s PatterState, n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		if c := p.Tick(s); c != nil {
			out[i] = c[0].Hz
		}
	}
	return out
}

func steps(hz []float64) []int {
	var at []int
	for i, h := range hz {
		if h > 0 {
			at = append(at, i)
		}
	}
	return at
}

// The patter clicks only on its step ticks: every 3 in the alley, every 2
// in rooms.
func TestPatterStepRate(t *testing.T) {
	for _, c := range []struct {
		mode  PatterMode
		every int
	}{{Alley, config.PatterAlley.Every}, {Room, config.PatterRoom.Every}} {
		at := steps(runPatter(newTestPatter(), PatterState{Mode: c.mode}, 600))
		if len(at) < 50 {
			t.Errorf("mode %d: only %d clicks in 600 ticks", c.mode, len(at))
		}
		for _, i := range at {
			if i%c.every != 0 {
				t.Errorf("mode %d: click at tick %d, want steps every %d", c.mode, i, c.every)
				break
			}
		}
	}
}

func TestPatterClickIsShort(t *testing.T) {
	p := newTestPatter()
	for range 100 {
		if c := p.Tick(PatterState{Mode: Alley}); c != nil {
			if len(c) != 1 || c[0].Dur.Seconds()*1000 > 4 {
				t.Fatalf("click = %+v, want one note of a few ms", c)
			}
			return
		}
	}
	t.Fatal("no click in 100 ticks")
}

// A falling object clicks every tick, its pitch from its height [asm].
func TestPatterFalling(t *testing.T) {
	p := newTestPatter()
	for y := 20; y < 180; y += 10 {
		c := p.Tick(PatterState{Mode: Alley, Falling: true, FallY: y})
		want := config.FallClickK / float64(y+config.FallClickY0)
		if c == nil || math.Abs(c[0].Hz-want) > 0.01 {
			t.Fatalf("Y %d: click %v, want %.1f Hz", y, c, want)
		}
	}
	// Falling only applies in the alley.
	if hz := runPatter(p, PatterState{Mode: Room, Falling: true, FallY: 50}, 1)[0]; hz == config.FallClickK/82.0 {
		t.Error("falling-object click in a room")
	}
}

// The broom bounce clicks every tick with a rising pitch, from the same start
// each bounce [asm].
func TestPatterBounce(t *testing.T) {
	p := newTestPatter()
	hz := runPatter(p, PatterState{Mode: Alley, Bouncing: true}, 10)
	if hz[0] != config.BounceStartHz {
		t.Errorf("first bounce click %.0f Hz, want %d", hz[0], config.BounceStartHz)
	}
	for i := 1; i < len(hz); i++ {
		if hz[i] <= hz[i-1] {
			t.Fatalf("bounce clicks %v: want one a tick, rising", hz)
		}
	}
	runPatter(p, PatterState{Mode: Alley}, 5)
	if hz := runPatter(p, PatterState{Mode: Alley, Bouncing: true}, 1)[0]; hz != config.BounceStartHz {
		t.Errorf("next bounce starts at %.0f Hz, want %d", hz, config.BounceStartHz)
	}
}

// A dog on screen plays the 3-click alarm every 3 ticks, every tick while
// Freddy is bouncing, and beats a falling object [asm].
func TestPatterDogAlarm(t *testing.T) {
	p := newTestPatter()
	hz := runPatter(p, PatterState{Mode: Alley, DogOut: true, Falling: true, FallY: 50}, 9)
	at := steps(hz)
	if len(at) != 3 || at[0] != 0 || at[1] != 3 || at[2] != 6 {
		t.Fatalf("alarm clicks at %v, want every 3 ticks", at)
	}
	if hz[0] != alarmPitches[0] || hz[3] != alarmPitches[1] || hz[6] != alarmPitches[2] {
		t.Errorf("alarm pitches %v, want the 3-click figure", hz)
	}
	bouncing := runPatter(p, PatterState{Mode: Room, DogOut: true, Bouncing: true}, 4)
	if len(steps(bouncing)) != 4 {
		t.Errorf("alarm while bouncing %v, want a click every tick", bouncing)
	}
}

// An alarm countdown plays n alarm clicks on consecutive ticks, then the
// patter carries on.
func TestPatterAlarmCountdown(t *testing.T) {
	p := newTestPatter()
	runPatter(p, PatterState{Mode: Room}, 7)
	p.Alarm(5)
	hz := runPatter(p, PatterState{Mode: Room}, 5)
	for i, h := range hz {
		if h != alarmPitches[i%3] && h != alarmPitches[(i+1)%3] && h != alarmPitches[(i+2)%3] {
			t.Fatalf("countdown clicks %v, want 5 alarm clicks in a row", hz)
		}
	}
	if p.alarm != 0 {
		t.Errorf("countdown left %d", p.alarm)
	}
}

// Quiet plays nothing and does not use up an alarm countdown.
func TestPatterQuiet(t *testing.T) {
	p := newTestPatter()
	p.Alarm(3)
	if at := steps(runPatter(p, PatterState{Mode: Quiet, DogOut: true, Bouncing: true}, 30)); len(at) != 0 {
		t.Errorf("clicks at %v while quiet", at)
	}
	if p.alarm != 3 {
		t.Errorf("countdown %d after quiet ticks, want 3", p.alarm)
	}
}

func TestPatterDeterministic(t *testing.T) {
	a := runPatter(newTestPatter(), PatterState{Mode: Alley}, 500)
	b := runPatter(newTestPatter(), PatterState{Mode: Alley}, 500)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("tick %d: %.0f vs %.0f with the same seed", i, a[i], b[i])
		}
	}
}

func TestNilPatter(t *testing.T) {
	var p *Patter
	p.Alarm(3)
	if p.Tick(PatterState{Mode: Alley}) != nil {
		t.Error("nil patter clicked")
	}
}

// Patter clicks are the original's tiny ticks: the bar's first step longer
// [asm] (DESIGN §7.2).
func TestPatterClickLengths(t *testing.T) {
	for _, c := range []struct {
		mode PatterMode
		pp   config.PatterParams
	}{{Alley, config.PatterAlley}, {Room, config.PatterRoom}} {
		p := newTestPatter()
		lengths := map[float64]bool{}
		for range 2000 {
			if n := p.Tick(PatterState{Mode: c.mode}); n != nil {
				lengths[math.Round(n[0].Dur.Seconds()*1e5)/100] = true
			}
		}
		for _, want := range []float64{c.pp.ClickMs, config.DownbeatMs} {
			if !lengths[want] {
				t.Errorf("mode %d: no %.2f ms click among %v", c.mode, want, lengths)
			}
		}
	}
}

// The patter's density matches the reference capture: about a third of the
// steps click in the alley and half in the rooms (DESIGN §7.2).
func TestPatterDensity(t *testing.T) {
	for _, c := range []struct {
		mode   PatterMode
		lo, hi float64
	}{{Alley, 0.33, 0.45}, {Room, 0.45, 0.56}} {
		p := newTestPatter()
		n := 3000 / params(c.mode).Every
		got := float64(len(steps(runPatter(p, PatterState{Mode: c.mode}, 3000)))) / float64(n)
		if got < c.lo || got > c.hi {
			t.Errorf("mode %d: %.0f%% of steps click, want %.0f–%.0f%%", c.mode, got*100, c.lo*100, c.hi*100)
		}
	}
}

// The pitch moves along the sequence or a sweep on every step, so two
// steps in a row never repeat a pitch within a phrase.
func TestPatterPitchMoves(t *testing.T) {
	p := newTestPatter()
	hz := runPatter(p, PatterState{Mode: Room}, 2*8*config.PatterRoom.BarsPerPhrase)
	distinct := map[float64]bool{}
	for _, h := range hz {
		if h > 0 {
			distinct[h] = true
		}
	}
	if len(distinct) < 4 {
		t.Errorf("only %d distinct pitches over two phrases", len(distinct))
	}
}
