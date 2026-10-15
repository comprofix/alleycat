// Package physics holds the original's movement primitives (DESIGN §3.8):
// rectangle overlap, the accumulator-based gravity, and the 8-pixel bands
// used to find the surface under a falling actor.
package physics

// Rect is an axis-aligned box: X, Y is the top-left corner.
type Rect struct {
	X, Y, W, H int
}

// Overlaps reports whether two boxes overlap. Touching edges count, as in
// the original (DESIGN §3.8, "Other facts").
func (r Rect) Overlaps(o Rect) bool {
	return r.X <= o.X+o.W && o.X <= r.X+r.W &&
		r.Y <= o.Y+o.H && o.Y <= r.Y+r.H
}

// Band rounds y down to its 8-pixel band. Surfaces are found by matching
// bands, so a fall of up to 8 px per update can never skip one.
func Band(y int) int { return y &^ 7 }

// Vertical motion directions.
const (
	Rising  = -1
	Still   = 0
	Falling = 1
)

// Arc is a vertical motion with the original's gravity: each update
// subtracts Step from an 8-bit accumulator, and each time it underflows the
// speed changes by 1, slowing a rise or speeding a fall (DESIGN §3.8).
type Arc struct {
	Dir   int   // Rising, Still or Falling
	Speed int   // pixels per update
	Step  uint8 // gravity step; larger is stronger
	Acc   uint8 // the accumulator
}

// Start begins a rise or fall. The accumulator starts at 1, so the first
// update always underflows.
func (a *Arc) Start(dir, speed int, step uint8) {
	a.Dir, a.Speed, a.Step, a.Acc = dir, speed, step, 1
}

// Gravity applies one update's gravity. A rise at speed 1 turns into a
// fall at speed 1; a fall speeds up to maxFall.
func (a *Arc) Gravity(maxFall int) {
	under := a.Step > a.Acc
	a.Acc -= a.Step
	if !under {
		return
	}
	switch a.Dir {
	case Rising:
		if a.Speed <= 1 {
			a.Dir = Falling
		} else {
			a.Speed--
		}
	case Falling:
		if a.Speed < maxFall {
			a.Speed++
		}
	}
}
