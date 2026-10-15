// Package alley is the hub screen (DESIGN §3.3): the building with its
// clotheslines, the fence and the trash cans, and Freddy moving among them.
// Its logic runs headless; drawing is in draw.go.
package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Alley holds the alley's state.
type Alley struct {
	Difficulty int
	Cans       []config.Can
	Lines      [3]Line
	Moving     int // the line that is scrolling
	Freddy     *entity.Freddy

	gap int // laundry gap chance for this difficulty
}

// New sets up the alley for a difficulty level, with Freddy walking in from
// the left.
func New(difficulty int, rng *rand.Rand) *Alley {
	d := config.ForLevel(difficulty)
	a := &Alley{
		Difficulty: difficulty,
		Cans:       d.Cans,
		Freddy:     entity.NewFreddy(),
		gap:        d.LaundryGapChance,
	}
	for i := range a.Lines {
		a.Lines[i] = newLine(i, a.gap, rng)
	}
	return a
}

// Update advances the alley by this frame's master ticks and Freddy steps.
func (a *Alley) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	for range ticks {
		a.scrollTick(rng)
	}
	for range freddySteps {
		a.Freddy.Update(in, a)
	}
}

// scrollTick moves the active clothesline one step (DESIGN §3.8,
// "Clotheslines").
func (a *Alley) scrollTick(rng *rand.Rand) {
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
		f.Y, f.Depth = config.FenceTopY, entity.OnFence
		return true
	}
	if top, ok := a.canUnder(f); ok {
		f.Y = top
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
