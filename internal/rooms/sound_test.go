package rooms

import (
	"math/rand/v2"
	"testing"
)

// The patter stops for the room's one-off moments (DESIGN §7.2).
func TestHushed(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))

	fb := New(Fishbowl, 0, rng).(*FishbowlRoom)
	if Hushed(fb) {
		t.Error("a fresh fishbowl room is hushed")
	}
	fb.Phase = Zapped
	if !Hushed(fb) {
		t.Error("not hushed during ZAP")
	}

	bs := New(Bookshelf, 0, rng).(*BookshelfRoom)
	bs.Bite = 3
	if !Hushed(bs) {
		t.Error("not hushed during BYTE")
	}

	bc := New(Birdcage, 0, rng).(*BirdcageRoom)
	bc.Cage.Falling = true
	if !Hushed(bc) {
		t.Error("not hushed while the cage falls")
	}

	ch := New(Cheese, 0, rng)
	ch.Shared().Freddy.Gone = true
	if !Hushed(ch) {
		t.Error("not hushed during a dog fight")
	}
	ch.Shared().Freddy.Gone = false
	ch.Shared().Out = Failed
	if !Hushed(ch) {
		t.Error("not hushed after the room ended")
	}
}
