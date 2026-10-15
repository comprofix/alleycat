// Package alley is the hub screen (DESIGN §3.3): the building with its
// clotheslines, windows and mice, the fence and the trash cans, the alley
// dog, and Freddy moving among them. Its logic runs headless; drawing is in
// draw.go.
package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
	"github.com/comprofix/alleycat/internal/session"
)

// Exit is what happens when Freddy goes in through a window.
type Exit int

const (
	NoExit  Exit = iota
	ToRoom       // an ordinary window
	ToHeart      // Felicia's window
)

// ImpactKind is what shows while the game is frozen.
type ImpactKind int

const (
	NoImpact ImpactKind = iota
	Bonk                // a thrown object hit Freddy
	Squeak              // a mouse knocked him off a line
)

// Impact is an impact word shown at a point while the game is frozen.
type Impact struct {
	Kind ImpactKind
	X, Y int
}

// Alley holds the alley's state.
type Alley struct {
	Difficulty int
	Cans       []config.Can
	Lines      [3]Line
	Moving     int // the line that is scrolling
	Freddy     *entity.Freddy
	Session    *session.Session

	Window   Window
	Thrown   Thrown
	Mice     [3]Mouse
	TrashCat TrashCat
	Dog      entity.Dog

	// Tick counts master ticks since the alley started.
	Tick int
	// Freeze is the ticks left of a freeze; Impact shows meanwhile.
	Freeze int
	Impact Impact
	// Exit is set when Freddy goes in through a window.
	Exit Exit
	// Sounds lists the sounds this update made, for the scene to play
	// (DESIGN §7.3). Each Update starts it afresh.
	Sounds []Sound

	gap        int // laundry gap chance for this difficulty
	table      config.Difficulty
	fenceSince int // tick Freddy last landed on the fence top
	mouseTurn  int
	lastCan    int
	canMark    bool // set landing on a can, cleared by the step-off click (DESIGN §7.3)
	catTouched bool // Freddy is touching the trash cat: its notes play once per touch
}

// New sets up the alley for a game, with Freddy walking in from the left.
func New(s *session.Session, rng *rand.Rand) *Alley {
	return newAlley(s, rng, entity.NewFreddy())
}

// NewReturning sets up the alley with Freddy back from a room at X, Y,
// falling behind the fence (DESIGN §3.8, "Returning from a room").
func NewReturning(s *session.Session, rng *rand.Rand, x, y int) *Alley {
	f := entity.NewFreddy()
	f.ReturnFromRoom(x, y)
	return newAlley(s, rng, f)
}

func newAlley(s *session.Session, rng *rand.Rand, f *entity.Freddy) *Alley {
	d := s.Table()
	a := &Alley{
		Difficulty: s.Difficulty,
		Cans:       d.Cans,
		Freddy:     f,
		Session:    s,
		gap:        d.LaundryGapChance,
		table:      d,
		lastCan:    -1,
	}
	for i := range a.Lines {
		a.Lines[i] = newLine(i, a.gap, rng)
	}
	a.initMice()
	return a
}

// Update advances the alley by this frame's master ticks and Freddy steps.
// While the game is frozen only the freeze counts down.
func (a *Alley) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	a.Sounds = a.Sounds[:0]
	for range ticks {
		if a.Exit != NoExit {
			return
		}
		if a.Freeze > 0 {
			if a.Freeze--; a.Freeze == 0 {
				a.Impact = Impact{}
			}
			continue
		}
		a.Tick++
		a.scrollTick(rng)
		a.updateWindow(rng)
		a.updateThrown()
		a.updateTrashCat(rng)
		a.updateMice(rng)
		a.updateDog(rng)
	}
	for range freddySteps {
		if a.Exit != NoExit || a.Freeze > 0 {
			return
		}
		a.Freddy.HoldEntry = a.Dog.Out()
		a.Freddy.Update(a.controls(in, rng), a)
		a.checkFreddy()
		a.checkThrownHit()
		a.checkDogCatch()
	}
}

// checkFreddy acts on what Freddy reported this step. Sinking behind the
// fence holds the game still for the crackle, as the original's sound
// routine does (DESIGN §7.3, §12). Leaping off after landing on a can clicks,
// unless the dog is out, which leaves the mark set.
func (a *Alley) checkFreddy() {
	f := a.Freddy
	if f.Sank {
		f.Sank = false
		a.Freeze = max(a.Freeze, config.FenceDropTicks)
		a.sound(SoundFenceDrop)
	}
	if f.Leapt {
		f.Leapt = false
		if a.canMark && !a.Dog.Active {
			a.canMark = false
			a.sound(SoundCanOff)
		}
	}
}

func (a *Alley) sound(s Sound) { a.Sounds = append(a.Sounds, s) }

// controls applies the trash cat's bump: standing on a can it touches,
// Freddy drops off in a random direction (DESIGN §3.8, "Alley systems").
func (a *Alley) controls(in entity.Controls, rng *rand.Rand) entity.Controls {
	f := a.Freddy
	if f.Airborne() || f.Seat != entity.NotSeated || f.Entering() || f.Y >= config.FloorY {
		return in
	}
	if !a.TrashCat.touches(f) {
		a.catTouched = false
		return in
	}
	// Its notes play once for each touch (DESIGN §7.3).
	if !a.catTouched {
		a.catTouched = true
		a.sound(SoundTrashCat)
	}
	return entity.Controls{H: entity.RandDir(rng), V: 1}
}

// loseLife takes a life, if any are left.
func (a *Alley) loseLife() { a.Session.LoseLife() }

// scrollTick moves the active clothesline one step (DESIGN §3.8,
// "Clotheslines"). The lines rest while a thrown object is in flight or
// Freddy has been knocked down.
func (a *Alley) scrollTick(rng *rand.Rand) {
	if a.Thrown.Active || a.Freddy.Knocked {
		return
	}
	l := &a.Lines[a.Moving]
	switch {
	case a.blocks(a.Moving):
		if !l.AtBoundary() {
			return
		}
		cur := a.Moving
		a.Moving = a.pickLine(rng, func(r int) bool { return r != cur })
	case l.AtBoundary() && rng.IntN(256) < config.LineSwitchChance:
		a.Moving = a.pickLine(rng, func(r int) bool { return !a.blocks(r) })
	}
	a.step(&a.Lines[a.Moving], rng)
}

func (a *Alley) pickLine(rng *rand.Rand, ok func(int) bool) int {
	for {
		if r := rng.IntN(len(a.Lines)); ok(r) {
			return r
		}
	}
}

// blocks reports whether Freddy, in the air within a line's band, holds that
// line still.
func (a *Alley) blocks(row int) bool {
	f := a.Freddy
	b := config.LineBand[row]
	return f.Airborne() && f.Y >= b.Lo && f.Y < b.Hi
}

// step scrolls a line and carries Freddy if he is sitting on it.
func (a *Alley) step(l *Line, rng *rand.Rand) {
	l.step(a.gap, rng)
	f := a.Freddy
	if f.Seat == entity.NotSeated || f.Y != config.LineY[l.Row] {
		return
	}
	x := f.X + l.Dir*config.LineStepPx
	if x > f.MaxX || x < f.MinX {
		f.FallOffLine()
		return
	}
	f.X = x
}

// Support implements entity.World with the alley's surfaces (DESIGN §3.8,
// "Alley surfaces").
func (a *Alley) Support(f *entity.Freddy) bool {
	if f.Y < 0 {
		return false
	}
	band := physics.Band(f.Y)
	if band == config.FenceTopY {
		if f.Depth == entity.Behind {
			return false
		}
		if f.Depth != entity.OnFence {
			a.fenceSince = a.Tick
		}
		f.Y, f.Depth = config.FenceTopY, entity.OnFence
		return true
	}
	if top, ok := a.canUnder(f); ok {
		f.Y = top
		if !a.canMark {
			a.canMark = true
			if !a.Dog.Active {
				a.sound(SoundCanOn)
			}
		}
		return true
	}
	for i, y := range config.LineY {
		if band == y && a.Lines[i].Supports(f.X) {
			f.Y = y
			f.X = physics.Band(f.X)
			f.Seat = entity.JustSeated
			return true
		}
	}
	return false
}

// canUnder returns the standing Y of the can under Freddy, if any.
func (a *Alley) canUnder(f *entity.Freddy) (int, bool) {
	band := physics.Band(f.Y + 2)
	for _, c := range a.Cans {
		top := CanTop(c)
		if band == top+2 && physics.Band(f.X) >= c.X && physics.Band(f.X-config.CanReach) <= c.X {
			return top, true
		}
	}
	return 0, false
}

// CanTop is the Y Freddy stands at on a can.
func CanTop(c config.Can) int {
	if c.Tall {
		return config.TallCanY
	}
	return config.ShortCanY
}
