package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Mouse is one of the three mice, walking on top of its clothesline
// (DESIGN §3.8, "Alley systems"). Mouse i is on line i.
type Mouse struct {
	X      int
	Dir    int // −1, 0 (stopped) or 1
	Caught bool
	Frame  int // walk-cycle frame
	caught int // tick it was caught
}

func mouseRect(i, x int) physics.Rect {
	return physics.Rect{X: x, Y: config.MouseY[i], W: config.MouseW, H: config.MouseH}
}

// startSide is where mice come in: the side away from Freddy.
func (a *Alley) startSide() (x, dir int) {
	if a.Freddy.X > config.DogSideSplitX {
		return 0, 1
	}
	return config.MouseRightStartX, -1
}

func (a *Alley) initMice() {
	x, dir := a.startSide()
	for i := range a.Mice {
		a.Mice[i] = Mouse{X: x, Dir: dir}
	}
}

// updateMice moves one mouse per tick, in turn.
func (a *Alley) updateMice(rng *rand.Rand) {
	f := a.Freddy
	if f.Knocked {
		return
	}
	a.mouseTurn = (a.mouseTurn + 1) % len(a.Mice)
	i := a.mouseTurn
	m := &a.Mice[i]
	if a.thrownTouchesMouse(i) || a.mouseTouch(i) {
		return
	}
	if m.Caught {
		if a.Tick-m.caught < config.MouseRespawnTicks {
			return
		}
		m.X, m.Dir = a.startSide()
		m.Caught = false
	}

	speed := config.MouseSpeedLow
	switch {
	case a.Window.HeldByMouse:
		speed = config.MouseSpeedFast
	case f.Y <= config.WindowMaxFreddyY:
		speed = config.MouseSpeedHigh
	}

	y := config.MouseY[i]
	switch {
	case i == a.Moving && m.Dir == 0:
		m.Dir = entity.RandDir(rng)
	case f.Seat != entity.NotSeated && y <= f.Y && f.Y <= y+config.MouseRowReach &&
		rng.IntN(256) <= config.MouseChargeChance[a.Difficulty]:
		speed = config.MouseSpeedFast
		m.Dir = -1
		if m.X < f.X {
			m.Dir = 1
		}
	default:
		chance := config.MouseChangeHigh
		if f.Y > config.WindowMaxFreddyY {
			chance = config.MouseChangeStopped
			if m.Dir != 0 {
				chance = config.MouseChangeWalking
			}
		}
		if rng.IntN(256) <= chance {
			if m.Dir != 0 {
				m.Dir = 0
			} else {
				m.Dir = entity.RandDir(rng)
			}
		}
	}

	switch m.Dir {
	case 1:
		if m.X += speed; m.X > config.MouseMaxX {
			m.X, m.Dir = config.MouseMaxX, -1
		}
	case -1:
		if m.X -= speed; m.X < 0 {
			m.X, m.Dir = 0, 1
		}
	}
	if m.Dir != 0 {
		m.Frame++
	}
	if !a.thrownTouchesMouse(i) {
		a.mouseTouch(i)
	}
}

func (a *Alley) thrownTouchesMouse(i int) bool {
	t := &a.Thrown
	if !t.Active {
		return false
	}
	r := physics.Rect{X: t.X, Y: t.Y, W: config.ThrownW, H: config.ThrownMouseH}
	return r.Overlaps(mouseRect(i, a.Mice[i].X))
}

// mouseTouch handles mouse i touching Freddy: a seated Freddy is knocked off
// his line with a squeak; otherwise he catches it. It reports a touch.
func (a *Alley) mouseTouch(i int) bool {
	f, m := a.Freddy, &a.Mice[i]
	if f.Hidden() || !mouseRect(i, m.X).Overlaps(f.HitBox()) {
		return false
	}
	if m.Caught || f.Y < config.MouseIgnoreAboveY {
		return true
	}
	if f.Seat != entity.NotSeated {
		f.KnockOffLine()
		a.Freeze = config.SqueakTicks
		x := m.X
		if x >= 16 {
			x -= 4
		}
		a.Impact = Impact{Kind: Squeak, X: x, Y: config.MouseY[i]}
		a.sound(SoundSqueak)
		return true
	}
	m.Caught, m.caught, m.Dir = true, a.Tick, 1
	a.Session.AddScore(config.MousePoints[i])
	a.sound(SoundMouseCaught)
	return true
}
