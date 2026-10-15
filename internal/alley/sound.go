package alley

// Sound is a moment the alley makes a sound for (DESIGN §7.3). The dog
// fight is not one: the scene follows the dog's capture for that.
type Sound int

const (
	SoundMouseCaught Sound = iota
	SoundSqueak            // a mouse knocked Freddy off a line
	SoundBonk              // a thrown object hit him
	SoundTrashCat          // he touched the trash cat
	SoundCanOn             // he landed on a trash can
	SoundCanOff            // he leapt off after landing on one
	SoundFenceDrop         // he sank behind the fence and was reset
)
