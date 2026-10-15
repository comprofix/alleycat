package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
)

// updateDog runs the alley dog (DESIGN §3.8, "Alley systems").
func (a *Alley) updateDog(rng *rand.Rand) {
	switch a.Dog.Tick(a.Freddy, a.table, rng) {
	case entity.DogCaught:
		a.caught()
	case entity.DogFightOver:
		a.loseLife()
	}
}

// checkDogCatch is checked after Freddy moves too.
func (a *Alley) checkDogCatch() {
	if a.Dog.CheckCatch(a.Freddy) {
		a.caught()
	}
}

// caught: Freddy vanishes and walks back in from the other side once the
// dog has gone; the life is lost when the dog leaves.
func (a *Alley) caught() {
	f := a.Freddy
	f.Enter(f.X >= config.FreddyEnterSplitX)
}
