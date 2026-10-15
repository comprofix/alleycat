package audio

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/comprofix/alleycat/internal/config"
)

// Synth is the beeper: one square-wave voice fed by two tracks. A sound
// effect pre-empts the music, which pauses and resumes where it left off
// once the effect ends (DESIGN §7).
//
// The game thread queues tracks; the audio thread pulls samples through
// Read. All methods are safe for concurrent use, and a nil *Synth ignores
// every call, so headless tests can run scenes without audio.
type Synth struct {
	mu sync.Mutex

	music, sfx track

	enabled bool
	volume  float64
	soften  bool

	// Voice state, owned by Read.
	hz    float64 // pitch of the last sounding note, kept through its release
	phase float64 // 0–1
	gain  float64 // 0–1, slewed for attack and release
	slew  float64 // gain step per sample, set by the last sounding note
	lp    float64 // low-pass filter state
	hpIn  float64 // high-pass filter state: last input and output
	hpOut float64
}

// NewSynth returns an enabled synth at the default volume with "Soften
// beeper" on.
func NewSynth() *Synth {
	return &Synth{enabled: true, volume: config.DefaultVolume, soften: true}
}

// PlayMusic starts a tune from its beginning, replacing any tune playing.
// A looping tune repeats until stopped. The table must not be modified
// while it plays.
func (s *Synth) PlayMusic(notes []Note, loop bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.music.start(notes, loop)
}

// StopMusic stops the tune.
func (s *Synth) StopMusic() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.music.stop()
}

// PlaySFX starts an effect, replacing any effect playing. The music is
// silent until it ends.
func (s *Synth) PlaySFX(notes []Note) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sfx.start(notes, false)
}

// StopSFX ends the effect early, letting the music resume.
func (s *Synth) StopSFX() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sfx.stop()
}

// SetEnabled turns sound on or off (Ctrl+S). Tracks keep their timing while
// sound is off, as in the original (DESIGN §7.1).
func (s *Synth) SetEnabled(on bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = on
}

// SetVolume sets the output level, 0–1.
func (s *Synth) SetVolume(v float64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.volume = min(max(v, 0), 1)
}

// SetSoften turns the "Soften beeper" low-pass filter on or off.
func (s *Synth) SetSoften(on bool) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.soften = on
}

// Read fills p with 16-bit little-endian stereo samples at
// config.AudioSampleRate. The stream never ends.
func (s *Synth) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	const (
		sr        = float64(config.AudioSampleRate)
		maxHz     = sr * 0.45 // keep the band-limiting step valid
		slewSamps = config.BeeperSlewMs * sr / 1000
	)
	lpCoef := 1 - math.Exp(-2*math.Pi*config.SoftenCutoffHz/sr)
	hpCoef := math.Exp(-2 * math.Pi * config.SpeakerHighPassHz / sr)
	sharpSamps := config.SharpNoteMs * sr / 1000

	n := len(p) / 4
	for i := range n {
		var hz float64
		var length int // samples in the sounding note
		var fresh bool // the note starts on this sample
		switch {
		case s.sfx.active():
			hz, length, fresh = s.sfx.hz(), s.sfx.length(), s.sfx.take()
			s.sfx.step()
		case s.music.active():
			hz, length, fresh = s.music.hz(), s.music.length(), s.music.take()
			s.music.step()
		}

		target := 0.0
		if hz > 0 && s.enabled {
			s.hz = min(hz, maxHz)
			// Each note starts at the top of its wave, as the speaker's
			// timer restarts high on every new note.
			if fresh {
				s.phase = 0
			}
			// Short notes, such as the patter's clicks, switch on and off
			// at once, as the speaker does; longer ones ramp.
			s.slew = 1
			if float64(length) >= sharpSamps {
				s.slew = 1 / max(1, slewSamps)
			}
			target = 1
		}
		if s.gain < target {
			s.gain = min(s.gain+s.slew, target)
		} else if s.gain > target {
			s.gain = max(s.gain-s.slew, target)
		}

		// The speaker is driven or not: the wave runs from 0 to the peak,
		// never below. A slow high-pass then centres long tones while short
		// clicks stay one-sided pulses (DESIGN §7).
		u := 0.0
		if s.gain > 0 {
			dt := s.hz / sr
			u = (square(s.phase, dt) + 1) * s.gain * config.BeeperAmplitude * s.volume
			s.phase += dt
			if s.phase >= 1 {
				s.phase -= 1
			}
		}
		s.hpOut = u - s.hpIn + hpCoef*s.hpOut
		s.hpIn = u
		x := s.hpOut
		if s.soften {
			s.lp += lpCoef * (x - s.lp)
			x = s.lp
		} else {
			s.lp = x
		}

		v := uint16(int16(math.Round(min(max(x, -1), 1) * math.MaxInt16)))
		binary.LittleEndian.PutUint16(p[4*i:], v)
		binary.LittleEndian.PutUint16(p[4*i+2:], v)
	}
	return 4 * n, nil
}

// square is a 50% duty square wave at the given phase, band-limited with
// polyBLEP so high notes don't alias into harshness. dt is the phase step
// per sample.
func square(phase, dt float64) float64 {
	v := -1.0
	if phase < 0.5 {
		v = 1
	}
	half := phase + 0.5
	if half >= 1 {
		half -= 1
	}
	return v + polyBLEP(phase, dt) - polyBLEP(half, dt)
}

// polyBLEP is the correction for a unit step at phase 0.
func polyBLEP(t, dt float64) float64 {
	switch {
	case t < dt:
		t /= dt
		return t + t - t*t - 1
	case t > 1-dt:
		t = (t - 1) / dt
		return t*t + t + t + 1
	}
	return 0
}

// track plays a note table, one sample at a time.
type track struct {
	notes []Note
	loop  bool
	i     int  // current note
	left  int  // samples left in it
	fresh bool // the current note has not sounded yet
}

func (t *track) active() bool { return t.notes != nil }

func (t *track) hz() float64 { return t.notes[t.i].Hz }

// length is the current note's length in samples.
func (t *track) length() int { return samples(t.notes[t.i].Dur) }

// take reports whether the current note has just started, once.
func (t *track) take() bool {
	f := t.fresh
	t.fresh = false
	return f
}

func (t *track) start(notes []Note, loop bool) {
	t.notes, t.loop, t.i = notes, loop, 0
	if len(notes) == 0 {
		t.stop()
		return
	}
	t.left = samples(notes[0].Dur)
	t.fresh = true
	t.skipEmpty()
}

func (t *track) stop() { t.notes = nil }

// step advances one sample.
func (t *track) step() {
	t.left--
	t.skipEmpty()
}

// skipEmpty moves past finished notes, looping or ending the track. A
// looping table with no length at all is stopped rather than spun on.
func (t *track) skipEmpty() {
	for wrapped := false; t.left <= 0; {
		t.i++
		if t.i == len(t.notes) {
			if !t.loop || wrapped {
				t.stop()
				return
			}
			t.i, wrapped = 0, true
		}
		t.left = samples(t.notes[t.i].Dur)
		t.fresh = true
	}
}

func samples(d time.Duration) int {
	return int(math.Round(d.Seconds() * config.AudioSampleRate))
}
