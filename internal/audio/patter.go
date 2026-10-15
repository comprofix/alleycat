package audio

import (
	"math/rand/v2"
	"time"

	"github.com/comprofix/alleycat/internal/config"
)

// PatterMode is where the background patter is playing (DESIGN §7.2).
type PatterMode int

const (
	Quiet PatterMode = iota // a one-off moment holds the game still
	Alley
	Room
)

// PatterState is what the patter needs to know each tick.
type PatterState struct {
	Mode     PatterMode
	Bouncing bool // the broom has bounced Freddy and he has not landed yet
	DogOut   bool // a dog is on screen
	Falling  bool // a thrown object is in the air
	FallY    int  // its height on screen
}

// Our own rhythms, one bar per byte: a bit per step of the 8-step bar,
// first step in the top bit. Each row is a phrase: 4 bars in the alley, 8 in
// the rooms, the last bar fuller to lead into the next phrase.
var (
	alleyRhythms = [][]uint8{
		{0xE9, 0x9B, 0xE9, 0xDD},
		{0xB6, 0xB6, 0xD6, 0xF7},
		{0x9B, 0xE5, 0x9B, 0xFB},
		{0xD5, 0xD5, 0xB7, 0xEF},
	}
	roomRhythms = [][]uint8{
		{0xCD, 0xC9, 0xCD, 0xDB, 0xCD, 0xC9, 0xCD, 0xFD},
		{0x96, 0x96, 0x9E, 0x96, 0x96, 0x96, 0x9E, 0xDF},
		{0xE6, 0xB3, 0xE6, 0xB3, 0xE6, 0xB3, 0xE7, 0xF3},
		{0x6D, 0x6D, 0x7D, 0x6D, 0x6D, 0x6D, 0x7D, 0xFD},
	}
)

// patterSequence is our own 8-note pitch sequence; the patter moves one
// note along it every step, sounding or not.
var patterSequence = [8]float64{3200, 1200, 2400, 600, 4800, 1600, 800, 2400}

// alarmPitches is our own 3-click alarm figure.
var alarmPitches = [3]float64{1800, 3000, 1200}

// Patter is the background click patter that runs in the alley and the
// rooms (DESIGN §7.2). Call Tick once per master tick; it returns the click
// to start, if any. It has its own RNG so it never disturbs the game's.
type Patter struct {
	rng *rand.Rand

	since int // ticks since the last step
	alarm int // alarm countdown, in steps
	ring  int // next alarm pitch

	bouncing bool
	bounceHz float64

	mode    PatterMode // the mode the phrase belongs to
	step    int        // within the bar, 0–7
	bar     int        // within the phrase
	phrase  int        // row of the rhythm table
	seq     int        // position in patterSequence
	sweep   bool       // this phrase sweeps up instead of following the sequence
	sweepHz float64
	fills   int // fill clicks still to come
}

// NewPatter returns a patter drawing on rng.
func NewPatter(rng *rand.Rand) *Patter {
	return &Patter{rng: rng, since: 1 << 30}
}

// Alarm starts the alarm countdown: n steps of the alarm figure, one a tick
// (DESIGN §7.3: the air warning and a dog waking).
func (p *Patter) Alarm(n int) {
	if p != nil {
		p.alarm = max(p.alarm, n)
	}
}

// Tick advances one master tick and returns the click to start now, or nil.
func (p *Patter) Tick(s PatterState) []Note {
	if p == nil {
		return nil
	}
	starting := s.Bouncing && !p.bouncing
	p.bouncing = s.Bouncing
	if starting {
		p.bounceHz = config.BounceStartHz
	}
	if s.Mode == Quiet {
		return nil
	}

	params := params(s.Mode)
	p.since++
	every := params.Every
	switch {
	case s.DogOut || p.alarm > 0:
		every = config.PatterAlarmEvery
	case s.Falling && s.Mode == Alley:
		every = config.PatterFallEvery
	}
	if !s.Bouncing && p.alarm == 0 && p.since < every {
		return nil
	}
	p.since = 0

	switch {
	case s.DogOut:
		return p.alarmClick()
	case p.alarm > 0:
		p.alarm--
		return p.alarmClick()
	case s.Falling && s.Mode == Alley:
		return click(config.FallClickK/float64(max(s.FallY, 0)+config.FallClickY0), config.FallClickMs)
	case s.Bouncing:
		hz := p.bounceHz
		if p.bounceHz *= config.BounceRise; p.bounceHz > config.BounceTopHz {
			p.bounceHz = config.BounceStartHz
		}
		return click(hz, config.BounceClickMs)
	}
	return p.patterStep(s.Mode, params)
}

func params(m PatterMode) config.PatterParams {
	if m == Room {
		return config.PatterRoom
	}
	return config.PatterAlley
}

func (p *Patter) alarmClick() []Note {
	hz := alarmPitches[p.ring]
	p.ring = (p.ring + 1) % len(alarmPitches)
	return click(hz, config.AlarmClickMs)
}

// patterStep plays one step of the bar (DESIGN §7.2): a click where the
// rhythm has one, a fill click in a rest while fills remain, otherwise
// nothing. The first step of a bar is a longer click.
func (p *Patter) patterStep(mode PatterMode, pp config.PatterParams) []Note {
	if mode != p.mode {
		p.mode, p.step, p.bar = mode, 0, 0
		p.newPhrase(pp)
	}

	hz := p.pitch()
	rhythm := p.rhythms()[p.phrase][p.bar]
	ms := pp.ClickMs
	if p.step == 0 {
		ms = config.DownbeatMs
	}
	hit := rhythm&(0x80>>p.step) != 0

	if p.step++; p.step == 8 {
		p.step = 0
		if p.bar++; p.bar == pp.BarsPerPhrase {
			p.bar = 0
			p.newPhrase(pp)
		}
	}

	switch {
	case hit:
		return click(hz, ms)
	case p.fills > 0:
		p.fills--
		return click(hz, pp.FillMs)
	}
	return nil
}

// pitch is this step's click pitch: the next note of the sequence, or the
// next of the sweep. Either moves on every step.
func (p *Patter) pitch() float64 {
	if p.sweep {
		hz := p.sweepHz
		if p.sweepHz *= config.PatterSweepRise; p.sweepHz > config.BounceTopHz {
			p.sweepHz = patterSequence[1]
		}
		return hz
	}
	hz := patterSequence[p.seq]
	p.seq = (p.seq + 1) % len(patterSequence)
	return hz
}

func (p *Patter) rhythms() [][]uint8 {
	if p.mode == Room {
		return roomRhythms
	}
	return alleyRhythms
}

// newPhrase picks the next phrase's rhythm row, whether it sweeps, and
// whether fill clicks follow.
func (p *Patter) newPhrase(pp config.PatterParams) {
	p.phrase = p.rng.IntN(len(p.rhythms()))
	p.sweep = p.rng.IntN(256) < config.PatterSweepChance
	p.sweepHz = 1500 + float64(p.rng.IntN(1000))
	if p.rng.IntN(256) < pp.FillChance {
		p.fills = p.rng.IntN(8)
	}
}

func click(hz, ms float64) []Note {
	return []Note{{Hz: hz, Dur: time.Duration(ms * float64(time.Millisecond))}}
}
