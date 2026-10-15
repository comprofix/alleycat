package audio

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/comprofix/alleycat/internal/clock"
	"github.com/comprofix/alleycat/internal/config"
)

// onsets returns the samples where sound starts after at least 5 ms of
// silence.
func onsets(x []float64) []int {
	var at []int
	last := -1 << 30
	for i, v := range x {
		if v > 0.05 || v < -0.05 {
			if i-last > samples(ms(5)) {
				at = append(at, i)
			}
			last = i
		}
	}
	return at
}

// Patter-style clicks every 3 ticks start exactly 3 ticks apart, however the
// audio happens to be read: the game runs at 60 frames a second and the
// audio is read ahead in uneven chunks, as a device does (DESIGN §7).
func TestScheduledClicksAreEvenlySpaced(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	c := clock.New(60)
	chunks := []int{441, 2205, 1100, 2205, 700, 1800}
	var x []float64
	wall, read := 0, 0
	for frame := range 180 {
		ticks, _ := c.Advance()
		s.Now(c.Time())
		for i := range ticks {
			if k := c.Total() - int64(ticks) + 1 + int64(i); k%3 == 0 {
				s.Now(float64(k))
				s.PlayMusic([]Note{{Hz: 2000, Dur: ms(1)}}, false)
			}
		}
		wall += config.AudioSampleRate / 60
		for read < wall+2205 {
			n := chunks[(frame+read)%len(chunks)]
			buf := make([]byte, 4*n)
			s.Read(buf)
			for j := range n {
				x = append(x, float64(int16(binary.LittleEndian.Uint16(buf[4*j:])))/math.MaxInt16)
			}
			read += n
		}
	}
	// Skip the first second, while the delay settles to the read-ahead.
	at := onsets(x[config.AudioSampleRate:])
	if len(at) < 10 {
		t.Fatalf("onsets %v, want a click every 3 ticks for 2 s", at)
	}
	perTick := config.AudioSampleRate / clock.TickHz
	for i := 1; i < len(at); i++ {
		if gap := float64(at[i] - at[i-1]); math.Abs(gap-3*perTick) > 2 {
			t.Errorf("clicks %d and %d are %.0f samples apart, want %.0f", i-1, i, gap, 3*perTick)
		}
	}
}

// If the game stalls and the audio runs on, the next sound re-anchors and
// plays within the delay rather than being dropped or bunched.
func TestScheduleReanchors(t *testing.T) {
	s := NewSynth()
	s.SetSoften(false)
	s.Now(0)
	render(t, s, ms(2000)) // the game stalls for 2 s
	s.Now(1)
	s.PlaySFX([]Note{{Hz: 2000, Dur: ms(1)}})
	at := onsets(render(t, s, ms(200)))
	if len(at) != 1 || at[0] > samples(ms(config.AudioScheduleMs))+2 {
		t.Errorf("after a stall: onsets %v, want one click within the delay", at)
	}
}

// Commands sent at the same moment apply in the order they were sent.
func TestScheduleKeepsOrder(t *testing.T) {
	s := NewSynth()
	s.Now(5)
	s.PlaySFX([]Note{{Hz: 2000, Dur: ms(500)}})
	s.Stop()
	x := render(t, s, ms(300))
	if !toneless(x) {
		t.Error("a Stop sent after PlaySFX at the same time did not win")
	}
}
