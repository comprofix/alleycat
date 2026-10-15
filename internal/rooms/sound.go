package rooms

// Hushed reports whether a one-off moment holds the room still, so the
// background patter stops (DESIGN §7.2): ZAP, BYTE, the cage falling, a dog
// fight, or the room having ended.
func Hushed(r Room) bool {
	b := r.Shared()
	if b.Out != Playing || b.Freddy.Gone {
		return true
	}
	switch r := r.(type) {
	case *FishbowlRoom:
		return r.Phase == Zapped
	case *BookshelfRoom:
		return r.Bite > 0
	case *BirdcageRoom:
		return r.Cage.Falling && !r.Cage.Broken && !r.Dog.Out()
	}
	return false
}

// SoundKind is a moment a room makes a sound for (DESIGN §7.3). The dog
// fight is not one: the scene follows the dog's capture for that.
type SoundKind int

const (
	SoundMouseCaught SoundKind = iota // cheese
	SoundCagePush                     // birdcage
	SoundCageFall                     // a tick of the cage falling; Arg is its Y
	SoundBirdChirp                    // the bird changes course
	SoundFishEaten                    // fishbowl
	SoundAirWarning                   // a drowning bubble; Arg counts them, from 0
	SoundZap                          // an eel
	SoundFernEaten                    // bookshelf
	SoundByte                         // the spider bites
	SoundBowlEmpty                    // sleeping dogs
	SoundDogWake                      // a sleeping dog wakes
	SoundRivalBump                    // heart room
	SoundArrowHit
	SoundGiftPick
	SoundGiftDrop
)

// Sound is one sound a room made, with its argument where it has one.
type Sound struct {
	Kind SoundKind
	Arg  int
}

// sound records a sound for the scene to play.
func (b *Base) sound(k SoundKind) { b.Sounds = append(b.Sounds, Sound{Kind: k}) }

// soundArg records a sound with an argument.
func (b *Base) soundArg(k SoundKind, arg int) {
	b.Sounds = append(b.Sounds, Sound{Kind: k, Arg: arg})
}
