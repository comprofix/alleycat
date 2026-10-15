package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// TrashCat is the cat that pops out of a trash can (DESIGN §3.8, "Alley
// systems"). Phase counts down from 27: it rises, then sinks back.
type TrashCat struct {
	Phase int
	X     int
	Top   int // top row while showing
	H     int // rows showing
	base  int
}

func (t *TrashCat) rect() physics.Rect {
	return physics.Rect{X: t.X, Y: t.Top, W: config.TrashCatW, H: t.H}
}

func (t *TrashCat) touches(f *entity.Freddy) bool {
	return t.Phase > 0 && t.H > 0 && t.rect().Overlaps(f.HitBox())
}

func (a *Alley) updateTrashCat(rng *rand.Rand) {
	t, f := &a.TrashCat, a.Freddy
	if t.touches(f) {
		a.Dog.Forced = true
		return
	}
	if t.Phase == 0 {
		onCan := f.Y == config.TallCanY || f.Y == config.ShortCanY
		if !onCan && rng.IntN(256) >= config.TrashCatChance {
			return
		}
		c := a.pickCan(rng)
		t.X = c.X + config.TrashCatXOffset + rng.IntN(config.TrashCatXSpread)
		t.base = CanTop(c) + 2 + config.TrashCatBaseBelow
		t.Phase = config.TrashCatPhases
	}
	t.Phase--
	if t.Phase >= config.TrashCatRisePhase {
		t.Top = t.base + t.Phase - config.TrashCatRiseTop
		t.H = config.TrashCatPhases - t.Phase
	} else {
		t.Top = t.base + config.TrashCatSinkTop - t.Phase
		t.H = t.Phase
	}
	if t.touches(f) {
		a.Dog.Forced = true
	}
}

// pickCan picks a random can, never the same one twice running.
func (a *Alley) pickCan(rng *rand.Rand) config.Can {
	for {
		if i := rng.IntN(len(a.Cans)); i != a.lastCan {
			a.lastCan = i
			return a.Cans[i]
		}
	}
}
