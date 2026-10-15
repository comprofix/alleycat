package entity

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
)

// Dog is the chasing dog: the alley dog, and the room dog in the rooms
// (DESIGN §3.4, §3.8 "Alley systems"). It steps 8 px, with steps
// alternately 1 and 2 ticks apart.
type Dog struct {
	X        int
	Dir      int // −1 or 1
	Active   bool
	Entering int // steps left coming into view
	Exiting  int // steps left going out of view
	// Capture is the direction the dog runs off in after a catch, or 0.
	Capture int
	// Forced calls the dog in at once (the trash cat).
	Forced bool
	Frame  int

	delay int // fight steps left before it runs off
	wait  int // ticks to the next step
	steps int
}

// Out reports whether the dog is on screen or coming or going.
func (d *Dog) Out() bool { return d.Active || d.Entering > 0 || d.Exiting > 0 }

// RandDir is −1 or 1 at random.
func RandDir(rng *rand.Rand) int {
	if rng.IntN(2) == 0 {
		return -1
	}
	return 1
}

// DogEvent is what a dog step did that its owner must act on.
type DogEvent int

const (
	DogNothing   DogEvent = iota
	DogCaught             // the fight has started: Freddy must vanish
	DogFightOver          // the dog has gone after a fight: a life is lost
)

// Tick runs the dog for one master tick.
func (d *Dog) Tick(f *Freddy, t config.Difficulty, rng *rand.Rand) DogEvent {
	if d.wait > 1 {
		d.wait--
		return DogNothing
	}
	d.steps++
	d.wait = 1 + d.steps%2
	d.Frame++
	if d.step(f, t, rng) {
		return DogFightOver
	}
	if d.CheckCatch(f) {
		return DogCaught
	}
	return DogNothing
}

func (d *Dog) step(f *Freddy, t config.Difficulty, rng *rand.Rand) (fightOver bool) {
	onFloor := f.Y >= config.FloorY && !f.Entering() && !f.Gone

	if d.Exiting > 0 {
		if d.Exiting--; d.Exiting == 0 {
			fightOver = d.Capture != 0
			*d = Dog{wait: d.wait, steps: d.steps}
		}
		return fightOver
	}

	switch {
	case d.Capture != 0:
		if d.delay > 0 {
			d.delay--
			d.Dir = RandDir(rng)
		} else {
			d.Dir = d.Capture
		}
	case !d.Active:
		if d.Entering == 0 {
			if !d.Forced && (!onFloor || rng.IntN(256) >= t.DogAppearChance) {
				return false
			}
			d.Dir = -1
			if f.X >= config.DogSideSplitX {
				d.Dir = 1
			}
			d.Entering = config.DogEdgeSteps + 1
		}
		d.Entering--
		d.X = 0
		if d.Dir < 0 {
			d.X = config.DogMaxX
		}
		if d.Entering > 0 {
			return false
		}
		d.Active = true
		fallthrough
	default:
		d.Forced = false
		if onFloor && rng.IntN(256) <= t.DogPursueChance {
			d.Dir = 1
			if d.X > f.X {
				d.Dir = -1
			}
		}
	}

	switch d.Dir {
	case -1:
		if d.X -= config.DogStepPx; d.X < 0 {
			d.X = 0
			d.atEdge(onFloor)
		}
	case 1:
		if d.X += config.DogStepPx; d.X >= config.DogMaxX {
			d.X = config.DogMaxX
			d.atEdge(onFloor)
		}
	}
	return false
}

// atEdge: the dog waits at the edge while Freddy is on the floor, and
// otherwise leaves; after a catch it leaves once the fight is over.
func (d *Dog) atEdge(onFloor bool) {
	if d.Capture != 0 {
		if d.delay == 0 {
			d.Exiting = config.DogEdgeSteps + 1
		}
		return
	}
	if !onFloor {
		d.Exiting = config.DogEdgeSteps + 1
	}
}

// CheckCatch starts the fight when the dog reaches Freddy on the floor, and
// reports whether it did. The caller decides what becomes of Freddy.
func (d *Dog) CheckCatch(f *Freddy) bool {
	if d.Capture != 0 || !d.Out() || f.Y < config.DogCatchY || f.Entering() || f.Gone {
		return false
	}
	if d.X+config.DogCatchAfter < f.X || max(d.X-config.DogCatchBefore, 0) > f.X {
		return false
	}
	mid := min((d.X+f.X)/2, config.DogFightMaxX)
	d.X = mid
	dist := config.DogSideSplitX + 1 - mid
	d.Capture = -1
	if mid > config.DogSideSplitX {
		dist = mid - (config.DogSideSplitX - 1)
		d.Capture = 1
	}
	d.delay = dist / config.DogStepPx
	d.Active, d.Entering, d.Exiting = true, 0, 0
	return true
}
