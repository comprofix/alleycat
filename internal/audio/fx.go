package audio

import (
	"math/rand/v2"
	"time"

	"github.com/comprofix/alleycat/internal/config"
)

// FX builds the sound effects (DESIGN §7.3). The shapes and timings follow
// the original; the exact pitches are approximate. Its own RNG keeps audio
// from disturbing the game's random sequence. A nil *FX returns no notes.
type FX struct {
	rng *rand.Rand
}

// NewFX returns an effect builder drawing on rng.
func NewFX(rng *rand.Rand) *FX { return &FX{rng: rng} }

// blip is a two-tone blip (DESIGN §7.1): the first tone lasts until the next
// tick, half a tick on average, and the second one tick.
func blip(a, b float64) []Note {
	return []Note{{Hz: a, Dur: Ticks(0.5)}, {Hz: b, Dur: Ticks(1)}}
}

func msDur(ms float64) time.Duration { return time.Duration(ms * float64(time.Millisecond)) }

// MouseCaught: a two-tone blip up.
func (fx *FX) MouseCaught() []Note {
	if fx == nil {
		return nil
	}
	return blip(1193, 1591)
}

// Squeak: a mouse knocks Freddy off a line. A random high chirp each tick
// for 8 ticks, from 8 steps between about 1.9 and 2.3 kHz.
func (fx *FX) Squeak() []Note {
	if fx == nil {
		return nil
	}
	n := make([]Note, config.SqueakTicks)
	for i := range n {
		n[i] = Note{Hz: 1900 + float64(fx.rng.IntN(8))*60, Dur: Ticks(1)}
	}
	return n
}

// Bonk: a thrown object hits Freddy. A jittering high warble, a new pitch
// every quarter of a millisecond, sliding down over the 10 ticks the splat
// shows.
func (fx *FX) Bonk() []Note {
	if fx == nil {
		return nil
	}
	const step = 0.25 // ms
	total := Ticks(config.BonkTicks).Seconds() * 1000
	var n []Note
	for t := 0.0; t < total; t += step {
		slide := t / total // 0 → 1
		centre := 7000 - 4000*slide
		n = append(n, Note{Hz: centre * (0.8 + 0.4*fx.rng.Float64()), Dur: msDur(step)})
	}
	return n
}

// TrashCat: touching the cat that pops out of a can. Three descending notes
// of about 20 ms.
func (fx *FX) TrashCat() []Note {
	if fx == nil {
		return nil
	}
	return []Note{{Hz: 597, Dur: msDur(20)}, {Hz: 447, Dur: msDur(20)}, {Hz: 335, Dur: msDur(20)}}
}

// CanOn and CanOff: one click of about 20 ms, landing on a trash can and
// leaving it.
func (fx *FX) CanOn() []Note {
	if fx == nil {
		return nil
	}
	return []Note{{Hz: 1165, Dur: msDur(20)}}
}

func (fx *FX) CanOff() []Note {
	if fx == nil {
		return nil
	}
	return []Note{{Hz: 1308, Dur: msDur(20)}}
}

// FenceDrop: Freddy sinks behind the fence. Two silent ticks, then the
// speaker switched on and off at random, the switches spreading further
// apart, until 7 ticks have passed.
func (fx *FX) FenceDrop() []Note {
	if fx == nil {
		return nil
	}
	n := []Note{{Dur: Ticks(2)}}
	crackle := Ticks(config.FenceDropTicks-2).Seconds() * 1000
	gap := 0.05 // ms
	for t := 0.0; t < crackle; t += gap {
		// A 1 Hz note restarts high, so a short one is a plain push of the
		// speaker: a crackle, not a tone.
		hz := 0.0
		if fx.rng.IntN(2) == 0 {
			hz = 1
		}
		n = append(n, Note{Hz: hz, Dur: msDur(gap)})
		gap = min(gap*1.04, 3)
	}
	return n
}

// DogFight: a dog catches Freddy. The original's opening: a rising stepped
// sweep, one step a tick, wrapping round. It repeats until StopSFX ends it
// with the fight. The original follows it with a high warble that the owner
// found shrill, so it is left out (DESIGN §7.3).
func (fx *FX) DogFight() []Note {
	if fx == nil {
		return nil
	}
	const steps = 600 // about 33 s, longer than any fight
	n := make([]Note, steps)
	hz := 1700.0
	for i := range n {
		n[i] = Note{Hz: hz, Dur: Ticks(1)}
		if hz *= 1.25; hz > 4600 {
			hz = 1700
		}
	}
	return n
}
