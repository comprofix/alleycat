package rooms

import (
	"math/rand/v2"
	"testing"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
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

// hasSound reports whether the room made sound k, and its argument.
func hasSound(b *Base, k SoundKind) (bool, int) {
	for _, s := range b.Sounds {
		if s.Kind == k {
			return true, s.Arg
		}
	}
	return false, 0
}

// Each room's moments record their sounds for the scene (DESIGN §7.3).
func TestRoomSounds(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	check := func(name string, b *Base, k SoundKind) {
		t.Helper()
		if ok, _ := hasSound(b, k); !ok {
			t.Errorf("%s: sounds %v, want kind %d", name, b.Sounds, k)
		}
		b.Sounds = b.Sounds[:0]
	}

	ch := New(Cheese, 0, rng).(*CheeseRoom)
	ch.Mice[0].Caught, ch.Mice[0].Count = false, 0
	ch.tryCatch(0)
	check("cheese catch", &ch.Base, SoundMouseCaught)

	bc := New(Birdcage, 0, rng).(*BirdcageRoom)
	bc.Freddy.X, bc.Freddy.Y = bc.Cage.X, bc.Cage.Y
	bc.cageTick()
	check("cage push", &bc.Base, SoundCagePush)
	bc.cageTick()
	if ok, _ := hasSound(&bc.Base, SoundCagePush); ok {
		t.Error("the cage push repeated while the push was pending")
	}
	bc.Cage.Y = 140
	bc.fall(rng)
	if ok, y := hasSound(&bc.Base, SoundCageFall); !ok || y != 140 {
		t.Errorf("cage fall: sounds %v, want the fall at Y 140", bc.Sounds)
	}

	fb := New(Fishbowl, 0, rng).(*FishbowlRoom)
	fb.Swimmers[0].Off = false
	fb.Swimmers[0].X, fb.Swimmers[0].Y = fb.Freddy.X, fb.Freddy.Y
	fb.contacts()
	check("fish eaten", &fb.Base, SoundFishEaten)
	for want := range 2 {
		fb.bubbleWait = 1
		fb.bubble()
		if ok, n := hasSound(&fb.Base, SoundAirWarning); !ok || n != want {
			t.Errorf("bubble %d: sounds %v, want the air warning numbered %d", want, fb.Sounds, want)
		}
		fb.Sounds = fb.Sounds[:0]
	}
	fb.zap()
	check("zap", &fb.Base, SoundZap)

	bs := New(Bookshelf, 0, rng).(*BookshelfRoom)
	bs.Freddy.X, bs.Freddy.Y = config.FernX[0], config.FernY
	bs.fernTick()
	check("fern", &bs.Base, SoundFernEaten)
	bs.bite()
	check("bite", &bs.Base, SoundByte)

	dg := New(Dogs, 0, rng).(*DogsRoom)
	for i := range dg.Sleepers {
		if dg.Sleepers[i].On {
			s := config.Sleepers[i]
			dg.Freddy.X, dg.Freddy.Y = s.X, s.Y
			dg.Sleepers[i].Alert = config.DogAlertMax - 1
			break
		}
	}
	dg.checkWait = dg.table.DogWakeCheckTicks
	dg.wakeTick(rng)
	check("dog wakes", &dg.Base, SoundDogWake)
	b := config.FoodBowls[0]
	dg.Freddy.X, dg.Freddy.Y, dg.Freddy.Gone = b.X, b.Y, false
	dg.Portions[0] = 1
	for range 50 {
		if dg.eatTick(entity.Controls{Act: true}); dg.Portions[0] == 0 {
			break
		}
	}
	check("bowl empty", &dg.Base, SoundBowlEmpty)
}

// The heart room's bumps and gifts record their sounds.
func TestHeartSounds(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	r := NewHeart(0, 1, rng)
	f := r.Freddy
	r.Arrow = Arrow{Active: true, X: f.X, Y: f.Y}
	r.arrowHits()
	if ok, _ := hasSound(&r.Base, SoundArrowHit); !ok {
		t.Errorf("arrow hit: sounds %v", r.Sounds)
	}
	r.Sounds = r.Sounds[:0]
	r.Cats[0].X, r.Cats[0].Y = f.X, f.Y
	r.catTouch(0)
	if ok, _ := hasSound(&r.Base, SoundRivalBump); !ok {
		t.Errorf("rival bump: sounds %v", r.Sounds)
	}
	r.Sounds = r.Sounds[:0]
	g := r.Gifts[0]
	f.X, f.Y = g.X, g.Y
	r.justDropped = -1
	r.giftTick(entity.Controls{})
	if ok, _ := hasSound(&r.Base, SoundGiftPick); !ok || r.Carrying != 0 {
		t.Fatalf("gift pick: sounds %v, carrying %d", r.Sounds, r.Carrying)
	}
	r.Sounds = r.Sounds[:0]
	r.putDown()
	if ok, _ := hasSound(&r.Base, SoundGiftDrop); !ok {
		t.Errorf("gift drop: sounds %v", r.Sounds)
	}
}

// Every room starts its sound list afresh on each Update.
func TestRoomSoundsReset(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	rooms := []Room{NewHeart(0, 1, rng)}
	for _, k := range []Kind{Birdcage, Cheese, Fishbowl, Dogs, Bookshelf} {
		rooms = append(rooms, New(k, 0, rng))
	}
	marker := Sound{Kind: SoundZap, Arg: -999}
	for _, r := range rooms {
		b := r.Shared()
		b.Sounds = append(b.Sounds[:0], marker)
		r.Update(1, 0, entity.Controls{}, rng)
		for _, s := range b.Sounds {
			if s == marker {
				t.Errorf("%T kept last update's sounds", r)
			}
		}
	}
}
