package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/physics"
)

// Thrown is the object a window occupant throws (DESIGN §3.8, "Alley
// systems"). Only one is ever in flight.
type Thrown struct {
	Active   bool
	X, Y     int
	Kind     int // 0–3
	Dir      int // −1, 0 or 1
	speed    int
	age      int
	hit      bool // it has already hit Freddy
	straight bool // thrown at a fence camper: it stops over him and drops
}

// H is the object's height.
func (t *Thrown) H() int { return config.ThrownH[t.Kind] }

// Visible is how many rows show above the fence.
func (t *Thrown) Visible() int {
	return min(t.H(), config.ThrownHideY-t.Y)
}

func (t *Thrown) rect() physics.Rect {
	return physics.Rect{X: t.X, Y: t.Y, W: config.ThrownW, H: t.H()}
}

// launch throws an object from the open window, towards Freddy.
func (a *Alley) launch(rng *rand.Rand) {
	r := WindowRect(a.Window.Index)
	t := Thrown{
		Active:   true,
		X:        r.X + rng.IntN(config.ThrownXSpread),
		Y:        r.Y,
		Kind:     rng.IntN(len(config.ThrownH)),
		Dir:      -1,
		speed:    config.ThrownSpeed,
		age:      1,
		straight: a.Window.forced,
	}
	if t.X < a.Freddy.X {
		t.Dir = 1
	}
	a.Thrown = t
}

func (a *Alley) updateThrown() {
	t := &a.Thrown
	if !t.Active {
		return
	}
	if t.straight && physics.Band(t.X) == physics.Band(a.Freddy.X) {
		t.Dir = 0
	}
	t.age++
	t.X = min(max(t.X+t.Dir*t.speed, 0), config.ThrownMaxX)
	t.Y += t.age / 2
	if t.Visible() <= 0 {
		*t = Thrown{}
		return
	}
	a.checkThrownHit()
}

// checkThrownHit: a hit shows "BONK!" with the game frozen, costs a life
// and knocks Freddy down; the object bounces away (owner decision, DESIGN
// §3.3).
func (a *Alley) checkThrownHit() {
	t, f := &a.Thrown, a.Freddy
	if !t.Active || !t.rect().Overlaps(f.HitBox()) {
		return
	}
	f.KnockDown()
	if t.hit {
		return
	}
	t.hit = true
	a.Freeze = config.BonkTicks
	a.Impact = Impact{Kind: Bonk, X: min(max(f.X-config.BonkOffsetX, 0), config.BonkMaxX), Y: f.Y}
	a.sound(SoundBonk)
	a.loseLife()
	if t.Dir == -1 {
		t.Dir = 1
	} else {
		t.Dir = -1
	}
	t.speed = config.ThrownBounceSpeed
	t.age = 1
}
