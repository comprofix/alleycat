package audio

import (
	"math"
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

// pushes builds speaker pushes and rests: on for each true slot, off for
// each false one, each slot lasting slotMs. A pushed slot after a rest
// restarts the wave at the top, so a 1 Hz note is a plain push of the
// speaker: a buzz or a crackle rather than a tone.
func pushes(pattern []bool, slotMs float64) []Note {
	n := make([]Note, len(pattern))
	for i, on := range pattern {
		hz := 0.0
		if on {
			hz = 1
		}
		n[i] = Note{Hz: hz, Dur: msDur(slotMs)}
	}
	return n
}

// CagePush: bumping the birdcage. A two-tone blip, a slight rise.
func (fx *FX) CagePush() []Note {
	if fx == nil {
		return nil
	}
	return blip(390, 411)
}

// CageFall is one tick of the falling cage at height y: its pitch falls as
// the cage drops, K ÷ y Hz [asm] (DESIGN §7.3).
func (fx *FX) CageFall(y int) []Note {
	if fx == nil {
		return nil
	}
	return []Note{{Hz: config.CageFallK / float64(max(y, 1)), Dur: Ticks(1)}}
}

// BirdChirp: the bird changes course. A very high two-tone blip, about 4–7
// kHz at random, with a slight drop.
func (fx *FX) BirdChirp() []Note {
	if fx == nil {
		return nil
	}
	hz := 4000 + 3000*fx.rng.Float64()
	return blip(hz, hz*0.9)
}

// FishEaten: a two-tone blip up.
func (fx *FX) FishEaten() []Note {
	if fx == nil {
		return nil
	}
	return blip(795, 1125)
}

// AirWarning is the n-th drowning bubble, from 0: a low two-tone blip, lower
// each time (DESIGN §7.3). The scene adds the alarm that follows.
func (fx *FX) AirWarning(n int) []Note {
	if fx == nil {
		return nil
	}
	f := math.Pow(0.93, float64(n))
	return blip(259*f, 244*f)
}

// Zap: touching an eel. 13 ticks of a buzz from speaker pushes, its pulses
// widening so it brightens and rises.
func (fx *FX) Zap() []Note {
	if fx == nil {
		return nil
	}
	const period = 3.0 // ms, a 333 Hz buzz
	total := Ticks(config.ZapTicks).Seconds() * 1000
	var n []Note
	for t := 0.0; t < total; t += period {
		duty := 0.1 + 0.5*t/total
		n = append(n, Note{Hz: 1, Dur: msDur(period * duty)}, Note{Dur: msDur(period * (1 - duty))})
	}
	return n
}

// FernEaten: a two-tone blip up.
func (fx *FX) FernEaten() []Note {
	if fx == nil {
		return nil
	}
	return blip(390, 520)
}

// Byte: the spider bites. 9 ticks of a patterned, gritty buzz from speaker
// pushes.
func (fx *FX) Byte() []Note {
	if fx == nil {
		return nil
	}
	pattern := []bool{true, false, true, true, false, false, true, false, true, true, true, false, false, true, false, false}
	slots := int(Ticks(config.ByteTicks).Seconds() * 1000 / 0.5)
	p := make([]bool, slots)
	for i := range p {
		p[i] = pattern[i%len(pattern)]
	}
	return pushes(p, 0.5)
}

// BowlEmpty: a two-tone blip up.
func (fx *FX) BowlEmpty() []Note {
	if fx == nil {
		return nil
	}
	return blip(519, 651)
}

// DogWake: a sleeping dog wakes. A 2-tick burst of noise from random speaker
// pushes. The scene adds the alarm that follows.
func (fx *FX) DogWake() []Note {
	if fx == nil {
		return nil
	}
	p := make([]bool, int(Ticks(2).Seconds()*1000/0.1))
	for i := range p {
		p[i] = fx.rng.IntN(2) == 0
	}
	return pushes(p, 0.1)
}

// RivalBump and ArrowHit: two-tone blips down. GiftPick goes up and
// GiftDrop down.
func (fx *FX) RivalBump() []Note { return fx.pair(362, 256) }
func (fx *FX) ArrowHit() []Note  { return fx.pair(511, 362) }
func (fx *FX) GiftPick() []Note  { return fx.pair(1193, 1419) }
func (fx *FX) GiftDrop() []Note  { return fx.pair(1193, 1004) }

func (fx *FX) pair(a, b float64) []Note {
	if fx == nil {
		return nil
	}
	return blip(a, b)
}

// Smooch: the moment the big heart appears in the kiss. Three fast sweeps,
// about 0.55 s: up, up again, then down.
func (fx *FX) Smooch() []Note {
	if fx == nil {
		return nil
	}
	var n []Note
	sweep := func(from, to float64, steps int) {
		for i := range steps {
			n = append(n, Note{Hz: from + (to-from)*float64(i)/float64(steps-1), Dur: msDur(15)})
		}
	}
	sweep(2400, 6000, 10)
	sweep(2400, 4000, 10)
	sweep(4000, 1500, 17)
	return n
}

// RejectFall: Freddy's head drops on the rejection screen. A falling tone,
// a step a tick for the steps of the drop, from about 4.6 kHz to 360 Hz.
func (fx *FX) RejectFall(steps int) []Note {
	if fx == nil {
		return nil
	}
	n := make([]Note, steps)
	for i := range n {
		// Even steps in period, as the original's are.
		period := 1/4600.0 + (1/360.0-1/4600.0)*float64(i)/float64(max(steps-1, 1))
		n[i] = Note{Hz: 1 / period, Dur: Ticks(1)}
	}
	return n
}

// Hi: the rejection's speech bubble. Two tones alternating every tick.
func (fx *FX) Hi(ticks int) []Note {
	if fx == nil {
		return nil
	}
	n := make([]Note, ticks)
	for i := range n {
		n[i] = Note{Hz: 390, Dur: Ticks(1)}
		if i%2 == 1 {
			n[i].Hz = 411
		}
	}
	return n
}
