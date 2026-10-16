package scene

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

// The title cat walks in from the left, then wanders between the turn
// points, sometimes stopping to sit (DESIGN §4, "The title cat").
func TestTitleCatWanders(t *testing.T) {
	var c titleCat
	rng := rand.New(rand.NewPCG(1, 2))
	c.update(1, rng)
	if c.dir != 1 || c.x != config.TitleCatSpeed {
		t.Fatalf("start: dir %d x %d, want walking right", c.dir, c.x)
	}
	var sat, left, right bool
	for range 2000 {
		c.update(1, rng)
		if c.x < 0 || c.x > config.TitleCatTurnRightX+config.TitleCatSpeed {
			t.Fatalf("x %d off the fence", c.x)
		}
		switch c.dir {
		case 0:
			sat = true
		case -1:
			left = true
		case 1:
			right = true
		}
	}
	if !sat || !left || !right {
		t.Fatalf("sat %v left %v right %v, want all", sat, left, right)
	}
}
