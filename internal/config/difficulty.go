package config

// Difficulty limits (DESIGN §3.1, §3.6).
const (
	MinDifficulty = 0
	MaxDifficulty = 7
)

// Can is one trash can in the alley (DESIGN §3.3).
type Can struct {
	X    int
	Tall bool // top surface at Y=134 if tall, Y=142 if short
}

// Difficulty is every value that depends on the difficulty level.
// Ticks are master ticks (18.2 Hz); chances are out of 256.
type Difficulty struct {
	Cans []Can // §3.3

	DogAppearChance int // §3.4, chasing dog appears (per check)
	DogPursueChance int // §3.4, dog turns toward Freddy

	WindowHoldTicks int // §3.8, fully-open hold (owner-adjusted for 4–7)

	SleepingDogs      int // §3.4
	FoodPortions      int // §3.4, per bowl
	DogWakeCheckTicks int // §3.8
	DogWakeDistance   int // §3.8, pixels

	ActiveEels    int // §3.8, of 12 slots
	DrowningTicks int // §3.8, helpless sinking starts
	AirFailTicks  int // §3.8, room fails
	SwimMaxX      int // §3.8, px per Freddy step
	SwimMaxY      int // §3.8, px per Freddy step

	MouseRetreatDistance int // §3.8, cheese room, |dx|+|dy|

	RivalChaseChance int // §3.5
	RivalHideTicks   int // §3.5, after taking a gift (doubled for the bottom-row cat)

	Multiplier int // §3.6, 2d+1 (doubled when carrying a gift)

	LaundryGapChance int // §3.8, /256 chance a new laundry piece is empty
}

var (
	cans0    = []Can{{24, false}, {64, true}, {104, false}, {152, false}, {240, false}, {280, true}}
	cans1    = []Can{{24, true}, {56, true}, {96, false}, {160, false}, {272, true}}
	cansEven = []Can{{24, false}, {64, true}, {160, false}, {264, true}}
	cansOdd  = []Can{{32, false}, {88, true}, {280, false}}
)

// Difficulties is indexed by difficulty level 0–7.
var Difficulties = [MaxDifficulty + 1]Difficulty{
	{cans0, 2, 16, 45, 6, 1, 18, 40, 2, 288, 360, 4, 3, 20, 16, 540, 1, 6},
	{cans1, 4, 32, 36, 6, 2, 16, 50, 4, 216, 288, 6, 3, 80, 32, 360, 3, 20},
	{cansEven, 8, 48, 27, 7, 3, 15, 60, 6, 192, 264, 8, 4, 100, 48, 270, 5, 32},
	{cansOdd, 12, 64, 18, 7, 4, 14, 70, 8, 192, 264, 10, 4, 120, 64, 180, 7, 46},
	{cansEven, 16, 80, 15, 8, 4, 13, 85, 9, 192, 264, 12, 4, 120, 96, 175, 9, 46},
	{cansOdd, 24, 96, 12, 8, 4, 12, 80, 10, 192, 264, 12, 4, 140, 128, 170, 11, 52},
	{cansEven, 32, 112, 10, 9, 4, 11, 85, 10, 184, 256, 12, 4, 140, 192, 160, 13, 58},
	{cansOdd, 64, 128, 8, 9, 4, 10, 90, 11, 176, 248, 12, 4, 140, 192, 140, 15, 64},
}

// ForLevel returns the table for a difficulty level, clamped to 0–7.
func ForLevel(level int) Difficulty {
	return Difficulties[max(MinDifficulty, min(MaxDifficulty, level))]
}
