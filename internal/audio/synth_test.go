package audio

import (
	"encoding/binary"
	"math"
	"testing"
	"time"

	"github.com/comprofix/alleycat/internal/config"
)

// render pulls d of mono samples from s.
func render(t *testing.T, s *Synth, d time.Duration) []float64 {
	t.Helper()
	n := samples(d)
	buf := make([]byte, 4*n)
	if got, err := s.Read(buf); err != nil || got != len(buf) {
		t.Fatalf("Read = %d, %v; want %d, nil", got, err, len(buf))
	}
	out := make([]float64, n)
	for i := range out {
		l := int16(binary.LittleEndian.Uint16(buf[4*i:]))
		r := int16(binary.LittleEndian.Uint16(buf[4*i+2:]))
		if l != r {
			t.Fatalf("sample %d: left %d != right %d", i, l, r)
		}
		out[i] = float64(l) / math.MaxInt16
	}
	return out
}

// pitch estimates the frequency of x from its rising zero crossings.
func pitch(x []float64) float64 {
	first, last, n := -1, -1, 0
	for i := 1; i < len(x); i++ {
		if x[i-1] < 0 && x[i] >= 0 {
			if first < 0 {
				first = i
			} else {
				n++
			}
			last = i
		}
	}
	if n == 0 {
		return 0
	}
	return float64(n) * config.AudioSampleRate / float64(last-first)
}

func peak(x []float64) float64 {
	p := 0.0
	for _, v := range x {
		p = max(p, math.Abs(v))
	}
	return p
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func near(got, want, tol float64) bool { return math.Abs(got-want) <= tol }

// swing is half the peak-to-peak range: a centred tone's amplitude.
func swing(x []float64) float64 {
	lo, hi := 0.0, 0.0
	for _, v := range x {
		lo, hi = min(lo, v), max(hi, v)
	}
	return (hi - lo) / 2
}

// toneless reports whether x holds no tone: at most the speaker's slow
// settling, with no sample-to-sample jumps.
func toneless(x []float64) bool {
	for i := 1; i < len(x); i++ {
		if math.Abs(x[i]-x[i-1]) > 0.003 {
			return false
		}
	}
	return true
}

// full is a tone's amplitude at the default volume once it has settled.
const full = config.BeeperAmplitude * config.DefaultVolume

func TestTicks(t *testing.T) {
	if got := Ticks(1); !near(got.Seconds(), 0.05493, 0.0001) {
		t.Errorf("Ticks(1) = %v, want about 54.9ms", got)
	}
}

func TestPitch(t *testing.T) {
	for _, soften := range []bool{false, true} {
		s := NewSynth()
		s.SetSoften(soften)
		s.PlayMusic([]Note{{440, time.Second}}, false)
		x := render(t, s, time.Second)
		settled := x[samples(ms(300)):]
		if got := pitch(settled); !near(got, 440, 1) {
			t.Errorf("soften %v: pitch = %.1f Hz, want 440", soften, got)
		}
		if a := swing(settled); a > full*1.15 {
			t.Errorf("soften %v: amplitude %.3f is above the configured level", soften, a)
		}
		if p := peak(x); p > 2*full*1.05 {
			t.Errorf("soften %v: onset peak %.3f is above the one-sided maximum", soften, p)
		}
	}
}

func TestNoClickAtStart(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlaySFX([]Note{{1000, ms(100)}})
	x := render(t, s, ms(100))
	if math.Abs(x[0]) > full*0.05 {
		t.Errorf("first sample %.3f: want a ramp from silence", x[0])
	}
	// One-sided at first: the wave runs from 0 up to twice the settled level.
	if p := peak(x[samples(ms(5)):samples(ms(20))]); !near(p, 2*full, 2*full*0.15) {
		t.Errorf("peak after attack = %.3f, want about %.3f", p, 2*full)
	}
}

func TestRestAndEndAreSilent(t *testing.T) {
	s := NewSynth()
	s.PlaySFX([]Note{{800, ms(50)}, {0, ms(50)}})
	x := render(t, s, ms(200))
	if !toneless(x[samples(ms(55)):samples(ms(100))]) {
		t.Error("tone in the rest")
	}
	if !toneless(x[samples(ms(105)):]) {
		t.Error("tone after the end")
	}
}

// An effect pre-empts the music, which then resumes where it paused.
func TestSFXPreemptsAndMusicResumes(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlayMusic([]Note{{500, ms(100)}, {1500, ms(100)}}, false)
	render(t, s, ms(50))

	s.PlaySFX([]Note{{3000, ms(100)}})
	during := render(t, s, ms(100))
	if got := pitch(during[100:]); !near(got, 3000, 15) {
		t.Errorf("during the effect pitch = %.0f, want 3000", got)
	}

	// 50 ms of the first note were left.
	rest := render(t, s, ms(40))
	if got := pitch(rest[100:]); !near(got, 500, 10) {
		t.Errorf("after the effect pitch = %.0f, want the paused note at 500", got)
	}
	render(t, s, ms(20))
	second := render(t, s, ms(80))
	if got := pitch(second); !near(got, 1500, 15) {
		t.Errorf("then pitch = %.0f, want the second note at 1500", got)
	}
}

func TestLoop(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlayMusic([]Note{{400, ms(30)}, {1200, ms(30)}}, true)
	render(t, s, ms(65)) // into the second pass
	if got := pitch(render(t, s, ms(20))); !near(got, 400, 20) {
		t.Errorf("looped pitch = %.0f, want 400", got)
	}
	s.StopMusic()
	render(t, s, ms(10))
	if !toneless(render(t, s, ms(20))) {
		t.Error("tone after StopMusic")
	}
}

// Sound off silences the output but keeps the timing (DESIGN §7.1).
func TestDisabledKeepsTiming(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlayMusic([]Note{{600, time.Second}}, false)
	s.SetEnabled(false)
	s.PlaySFX([]Note{{2000, ms(100)}})
	if p := peak(render(t, s, ms(150))); p > 0.001 {
		t.Errorf("sound off peak = %.4f, want silence", p)
	}
	s.SetEnabled(true)
	if got := pitch(render(t, s, ms(400))[samples(ms(300)):]); !near(got, 600, 10) {
		t.Errorf("after sound on pitch = %.0f, want the music (the effect ran out while off)", got)
	}
}

func TestVolume(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.SetVolume(0.5)
	s.PlayMusic([]Note{{500, time.Second}}, false)
	want := config.BeeperAmplitude * 0.5
	if a := swing(render(t, s, ms(400))[samples(ms(300)):]); !near(a, want, want*0.1) {
		t.Errorf("amplitude at volume 0.5 = %.3f, want %.3f", a, want)
	}
	s.SetVolume(0)
	render(t, s, ms(10))
	if !toneless(render(t, s, ms(20))) {
		t.Error("tone at volume 0")
	}
}

// "Soften beeper" takes the edge off: less energy in the sample-to-sample
// steps, at the same pitch.
func TestSoftenSmooths(t *testing.T) {
	rough := func(soften bool) float64 {
		s := NewSynth()
		s.SetSoften(soften)
		s.PlayMusic([]Note{{1000, time.Second}}, false)
		x := render(t, s, ms(200))[1000:]
		sum := 0.0
		for i := 1; i < len(x); i++ {
			d := x[i] - x[i-1]
			sum += d * d
		}
		return sum
	}
	if hard, soft := rough(false), rough(true); soft > hard*0.6 {
		t.Errorf("soften roughness %.3f vs %.3f: want a clear reduction", soft, hard)
	}
}

// Very short notes, as in the background clicks (DESIGN §7.2), still sound.
func TestShortClick(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlaySFX([]Note{{2000, ms(4)}, {0, ms(20)}})
	if p := peak(render(t, s, ms(10))); p < config.BeeperAmplitude*config.DefaultVolume*0.5 {
		t.Errorf("4 ms click peak = %.3f, want it audible", p)
	}
}

func TestEmptyTables(t *testing.T) {
	s := NewSynth()
	s.PlayMusic(nil, true)
	s.PlaySFX([]Note{})
	s.PlayMusic([]Note{{440, 0}}, true) // no length at all: stops, doesn't hang
	if p := peak(render(t, s, ms(20))); p > 0.001 {
		t.Errorf("peak = %.4f, want silence", p)
	}
}

func TestNilSynth(t *testing.T) {
	var s *Synth
	s.PlayMusic([]Note{{440, ms(10)}}, false)
	s.PlaySFX([]Note{{440, ms(10)}})
	s.StopMusic()
	s.StopSFX()
	s.SetEnabled(false)
	s.SetVolume(1)
	s.SetSoften(false)
}

// A patter tick of a fifth of a millisecond is a sharp one-sided pulse at
// full height, like the original speaker's (DESIGN §7): no ramp, starting at
// the top of the wave, nothing below zero.
func TestTinyClickIsSharp(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlaySFX([]Note{{3000, 210 * time.Microsecond}})
	x := render(t, s, ms(2))
	if x[1] < full {
		t.Errorf("second sample %.3f: want the click at full height at once", x[1])
	}
	if p := peak(x); !near(p, 2*full, 2*full*0.1) {
		t.Errorf("0.21 ms click peak = %.3f, want about %.3f", p, 2*full)
	}
	for i, v := range x {
		if v < -0.01 {
			t.Fatalf("sample %d = %.3f: a click should not swing below zero", i, v)
		}
	}
	if !toneless(x[samples(ms(1)):]) {
		t.Error("click rings on after it ends")
	}
}

// A long tone settles around zero, as in the original capture.
func TestToneCentres(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.PlayMusic([]Note{{500, time.Second}}, false)
	x := render(t, s, ms(600))[samples(ms(400)):]
	sum := 0.0
	for _, v := range x {
		sum += v
	}
	if m := sum / float64(len(x)); math.Abs(m) > full*0.05 {
		t.Errorf("mean %.4f after 400 ms: want the tone centred", m)
	}
}
