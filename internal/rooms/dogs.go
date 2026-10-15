package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// SleepingDog is one of the 12 dog slots (DESIGN §3.4, "Sleeping dogs room
// in detail"). Alert is 0 (asleep), 1 (stirring) or 2 (awake).
type SleepingDog struct {
	On    bool
	Alert int
}

// DogsRoom: eat from every bowl without waking the sleeping dogs.
type DogsRoom struct {
	Base
	Sleepers  [len(config.Sleepers)]SleepingDog
	Portions  [len(config.FoodBowls)]int
	BowlsLeft int
	// Eating is set while Freddy eats from bowl EatBowl.
	Eating  bool
	EatBowl int
	EatAcc  uint8
	// Woke is set when a dog's alert reaches 2, for the wake-up sound.
	Woke bool

	checkWait int
	walk      int // a step towards a bowl, for the next Freddy update
}

func newDogs(difficulty int, rng *rand.Rand) *DogsRoom {
	r := &DogsRoom{Base: newBase(config.Dogs, difficulty, false, rng)}
	for n := r.table.SleepingDogs; n > 0; {
		if s := &r.Sleepers[rng.IntN(len(r.Sleepers))]; !s.On {
			s.On = true
			n--
		}
	}
	for i := range r.Portions {
		r.Portions[i] = r.table.FoodPortions
	}
	r.BowlsLeft = len(r.Portions)
	return r
}

// Kind implements Room.
func (r *DogsRoom) Kind() Kind { return Dogs }

// Update implements Room.
func (r *DogsRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	r.Sounds = r.Sounds[:0]
	for range ticks {
		if r.Out != Playing {
			return
		}
		r.tick(rng)
		if r.Dog.Out() && r.Dog.Tick(r.Freddy, r.table, rng) == entity.DogFightOver {
			r.ByDog = true
			r.end(Failed)
			return
		}
		r.wakeTick(rng)
		r.eatTick(in)
	}
	for range freddySteps {
		if r.Out != Playing {
			return
		}
		step := in
		if r.walk != 0 {
			step.H, step.V = r.walk, 0
			r.walk = 0
		}
		r.step(step)
	}
}

// SleeperRect is dog slot i's box as drawn.
func SleeperRect(i int) physics.Rect {
	s := config.Sleepers[i]
	return physics.Rect{X: s.DrawX, Y: s.Y, W: config.SleeperW, H: config.SleeperH}
}

// near reports whether Freddy is near dog slot i: on its row and within
// the wake distance.
func (r *DogsRoom) near(i int) bool {
	s, f := config.Sleepers[i], r.Freddy
	return f.Y == s.Y && abs(s.X-f.X) <= r.table.DogWakeDistance
}

// wakeTick checks the dogs once the check interval has passed: a dog Freddy
// is near grows more alert and, already awake, attacks; the others may calm
// down (DESIGN §3.8, "Sleeping dogs: waking").
func (r *DogsRoom) wakeTick(rng *rand.Rand) {
	if r.Dog.Capture != 0 {
		return
	}
	if r.checkWait++; r.checkWait <= r.table.DogWakeCheckTicks {
		return
	}
	r.checkWait = 0
	for i := len(r.Sleepers) - 1; i >= 0; i-- {
		s := &r.Sleepers[i]
		if !s.On {
			continue
		}
		switch {
		case r.near(i) && s.Alert >= config.DogAlertMax:
			r.attack(i)
			return
		case r.near(i):
			if s.Alert++; s.Alert == config.DogAlertMax {
				r.Woke = true
				r.sound(SoundDogWake)
			}
		case s.Alert > 0 && rng.IntN(256) < config.DogCalmChance:
			s.Alert--
		}
	}
}

// attack: dog i gets up from its place and catches Freddy where he is.
func (r *DogsRoom) attack(i int) {
	f := r.Freddy
	r.Sleepers[i].On = false
	r.stopEating()
	f.Gone = true
	r.Dog.CaptureAt(f.X, f.Y)
}

// eatTick runs the eating once a tick: with Alt held Freddy walks to the
// nearest bowl with food on his row, then eats from it (DESIGN §3.4,
// "Eating").
func (r *DogsRoom) eatTick(in entity.Controls) {
	f := r.Freddy
	if !in.Act || f.Bouncing || f.Gone {
		r.stopEating()
		return
	}
	bowl, dist, dir := -1, 1<<30, 0
	for i := len(r.Portions) - 1; i >= 0; i-- {
		b := config.FoodBowls[i]
		if r.Portions[i] == 0 || b.Y != f.Y {
			continue
		}
		if d := abs(f.X - b.X); d <= dist {
			bowl, dist, dir = i, d, 1
			if f.X >= b.X {
				dir = -1
			}
		}
	}
	if bowl < 0 {
		r.stopEating()
		return
	}
	if dist >= config.EatReach {
		r.stopEating()
		if dist <= config.EatSlowReach {
			f.RunSpeed = config.EatWalkSpeed
		}
		f.Dir, r.walk = dir, dir
		return
	}
	r.Eating, r.EatBowl = true, bowl
	acc := r.EatAcc
	if r.EatAcc += config.EatStep; r.EatAcc >= acc {
		return
	}
	if r.Portions[bowl]--; r.Portions[bowl] > 0 {
		return
	}
	r.sound(SoundBowlEmpty)
	f.Dir, r.walk = 0, 0
	if r.BowlsLeft--; r.BowlsLeft == 0 {
		r.end(Cleared)
	}
}

// MilkLevel is how full bowl i looks, 0–4: every bowl starts full and steps
// down evenly as its portions go; 0 is empty, and the bowl is gone (DESIGN
// §3.4, "Sleeping dogs room in detail").
func (r *DogsRoom) MilkLevel(i int) int {
	start := r.table.FoodPortions
	return (r.Portions[i]*config.MilkLevels + start - 1) / start
}

func (r *DogsRoom) stopEating() {
	r.Eating, r.EatAcc = false, 0
}
