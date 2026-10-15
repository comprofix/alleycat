package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
)

// Dog is the alley dog (DESIGN §3.8, "Alley systems"). It steps 8 px, with
// steps alternately 1 and 2 ticks apart.
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

func (a *Alley) updateDog(rng *rand.Rand) {
	d := &a.Dog
	if d.wait > 1 {
		d.wait--
		return
	}
	d.steps++
	d.wait = 1 + d.steps%2
	d.Frame++
	a.dogStep(rng)
	a.checkDogCatch()
}

func (a *Alley) dogStep(rng *rand.Rand) {
	d, f := &a.Dog, a.Freddy
	onFloor := f.Y >= config.FloorY && !f.Entering()

	if d.Exiting > 0 {
		if d.Exiting--; d.Exiting == 0 {
			if d.Capture != 0 {
				a.loseLife()
			}
			*d = Dog{wait: d.wait, steps: d.steps}
		}
		return
	}

	switch {
	case d.Capture != 0:
		if d.delay > 0 {
			d.delay--
			d.Dir = randDir(rng)
		} else {
			d.Dir = d.Capture
		}
	case !d.Active:
		if d.Entering == 0 {
			if !d.Forced && (!onFloor || rng.IntN(256) >= a.table.DogAppearChance) {
				return
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
			return
		}
		d.Active = true
		fallthrough
	default:
		d.Forced = false
		if onFloor && rng.IntN(256) <= a.table.DogPursueChance {
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
			a.dogAtEdge(onFloor)
		}
	case 1:
		if d.X += config.DogStepPx; d.X >= config.DogMaxX {
			d.X = config.DogMaxX
			a.dogAtEdge(onFloor)
		}
	}
}

// dogAtEdge: the dog waits at the edge while Freddy is on the floor, and
// otherwise leaves; after a catch it leaves once the fight is over.
func (a *Alley) dogAtEdge(onFloor bool) {
	d := &a.Dog
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

// checkDogCatch starts the fight when the dog reaches Freddy on the floor:
// Freddy vanishes and walks back in from the other side once the dog has
// gone; the life is lost when the dog leaves.
func (a *Alley) checkDogCatch() {
	d, f := &a.Dog, a.Freddy
	if d.Capture != 0 || !d.Out() || f.Y < config.DogCatchY || f.Entering() {
		return
	}
	if d.X+config.DogCatchAfter < f.X || max(d.X-config.DogCatchBefore, 0) > f.X {
		return
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
	f.Enter(f.X >= config.FreddyEnterSplitX)
}
