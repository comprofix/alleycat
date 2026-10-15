package rooms

import (
	"math/rand/v2"
	"testing"
)

func TestNeverThreeInARow(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	var p Picker
	var prev [2]Kind
	for i := range 5000 {
		k := p.Pick(i%8, rng)
		if i >= 2 && k == prev[0] && k == prev[1] {
			t.Fatalf("pick %d: %v three times running", i, k)
		}
		prev[1], prev[0] = prev[0], k
	}
}

func TestPoolsWeightTheChoice(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	var p Picker
	count := map[Kind]int{}
	for range 8000 {
		count[p.Pick(1, rng)]++ // difficulty 1: bookshelf and birdcage
	}
	if count[Bookshelf]+count[Birdcage] < 5000 {
		t.Errorf("difficulty 1 picks: %v; want mostly bookshelf and birdcage", count)
	}
	if count[Heart] != 0 {
		t.Error("the heart room is never picked at random")
	}
}
