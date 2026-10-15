package audio

import (
	"encoding/binary"
	"math"
	"sync"
	"time"

	"github.com/comprofix/alleycat/internal/clock"
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

	// Scheduling (see Now): commands wait in queue until the sample count
	// pos reaches their time. offset maps game time onto samples.
	queue   []command
	pos     int64
	timed   bool
	latency int64 // the current delay, in samples: it widens to cover the device's read-ahead
	stamp   int64
	offset  int64

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
	s.send(command{kind: playMusic, notes: notes, loop: loop})
}

// StopMusic stops the tune.
func (s *Synth) StopMusic() { s.send(command{kind: stopMusic}) }

// PlaySFX starts an effect, replacing any effect playing. The music is
// silent until it ends.
func (s *Synth) PlaySFX(notes []Note) { s.send(command{kind: playSFX, notes: notes}) }

// StopSFX ends the effect early, letting the music resume.
func (s *Synth) StopSFX() { s.send(command{kind: stopSFX}) }

// Stop ends the tune and the effect. The game calls it at every scene change,
// as the original silences the speaker at each transition (DESIGN §7.1).
func (s *Synth) Stop() { s.send(command{kind: stopAll}) }

// Now sets the game time, in master ticks, that the commands sent after it
// belong to. Once it has been called, each command plays at its own time
// plus a fixed delay, sample-exact, rather than whenever the audio buffer is
// next filled: so the patter's steps fall exactly 3 ticks apart (DESIGN
// §7). Without it, as in tests, commands apply at once.
func (s *Synth) Now(ticks float64) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	const (
		perTick = config.AudioSampleRate / clock.TickHz
		base    = config.AudioScheduleMs * config.AudioSampleRate / 1000
		most    = config.AudioScheduleMaxMs * config.AudioSampleRate / 1000
	)
	at := int64(ticks * perTick)
	if !s.timed {
		s.latency = base
	}
	// Anchor the game clock to the audio clock. The device reads audio ahead
	// in chunks by an amount we cannot know, so if a command would already
	// be late, widen the delay to cover it and re-anchor; this settles after
	// a moment. If the two clocks drift far apart, re-anchor too.
	ahead := at + s.offset - s.pos
	if !s.timed || ahead < 0 || ahead > 4*s.latency {
		// A little late: the device reads further ahead than allowed for.
		// Very late means the game stalled; that needs no wider delay.
		if s.timed && ahead < 0 && -ahead <= most {
			s.latency = min(s.latency-ahead+base/2, most)
		}
		s.offset = s.pos + s.latency - at
		s.timed = true
	}
	s.stamp = at + s.offset
}

// send queues a command at the current game time, or applies it at once if
// the game has not set a time.
func (s *Synth) send(c command) {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.timed {
		s.apply(c)
		return
	}
	c.at = s.stamp
	if n := len(s.queue); n > 0 {
		c.at = max(c.at, s.queue[n-1].at) // keep the order they were sent in
	}
	s.queue = append(s.queue, c)
}

// apply carries out a command on the tracks.
func (s *Synth) apply(c command) {
	switch c.kind {
	case playMusic:
		s.music.start(c.notes, c.loop)
	case stopMusic:
		s.music.stop()
	case playSFX:
		s.sfx.start(c.notes, false)
	case stopSFX:
		s.sfx.stop()
	case stopAll:
		s.music.stop()
		s.sfx.stop()
	}
}

// command is a queued change to the tracks.
type command struct {
	at    int64 // the sample it applies at
	kind  commandKind
	notes []Note
	loop  bool
}

type commandKind int

const (
	playMusic commandKind = iota
	stopMusic
	playSFX
	stopSFX
	stopAll
)

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
		for len(s.queue) > 0 && s.queue[0].at <= s.pos {
			s.apply(s.queue[0])
			s.queue = s.queue[1:]
		}
		s.pos++

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
			// A note that starts from silence starts at the top of its
			// wave, as the speaker's timer restarts high. A change of
			// pitch mid-sound keeps the wave going: restarting it on every
			// note of a fast warble buzzes at the note rate.
			if fresh && s.gain == 0 {
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
