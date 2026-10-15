package config

// The sleeping dogs room (DESIGN §3.4, "Sleeping dogs room in detail";
// §3.8, "Sleeping dogs: waking", "Eating"). Times are master ticks; chances
// are out of 256.

// Dogs is the sleeping dogs room (DESIGN §3.4).
var Dogs = RoomLayout{
	StartX: 64, StartY: 56,
	WindowX: 72, WindowY: 56,
	BroomEvery: 1,
	Surfaces: []Surface{
		{168, 0, 320, false}, {152, 0, 320, false}, {136, 0, 320, false}, // the floor rows
		{48, 32, 4, true}, {56, 32, 4, true}, {64, 32, 4, true}, // curtains, left
		{48, 80, 4, true}, {56, 80, 4, true}, {64, 80, 4, true}, // curtains, right
		{80, 72, 8, false}, {112, 72, 16, false}, // left chair back, seat
		{112, 216, 16, false}, {80, 224, 8, false}, // right chair seat, back
		{48, 248, 8, true},                       // picture
		{72, 272, 12, true}, {80, 272, 12, true}, // lamp shade
	},
}

// Sleeper is a sleeping dog's slot: X is its middle, Y its row (the Y
// Freddy stands at on that row), DrawX where its 40×13 picture starts:
// X − 20, except slot 3, which the original draws at 0.
type Sleeper struct {
	X, Y, DrawX int
	Left        bool // facing left
}

// Sleepers are the 12 dog slots; a random set of them is filled.
var Sleepers = [12]Sleeper{
	{44, 136, 24, true}, {124, 136, 104, false}, {196, 136, 176, true},
	{32, 152, 0, false}, {92, 152, 72, true}, {156, 152, 136, true}, {204, 152, 184, false}, {268, 152, 248, false},
	{52, 168, 32, true}, {132, 168, 112, false}, {188, 168, 168, false}, {292, 168, 272, true},
}

// FoodBowl is a bowl's place: X its left edge, Y the row Freddy eats it
// from. Left is set when his eating pose is drawn to his left.
type FoodBowl struct {
	X, Y int
	Left bool
}

// FoodBowls are the 12 bowls.
var FoodBowls = [12]FoodBowl{
	{8, 136, false}, {144, 136, false}, {160, 136, true},
	{40, 152, false}, {56, 152, true}, {120, 152, true}, {224, 152, false}, {288, 152, true},
	{16, 168, false}, {152, 168, false}, {208, 168, false}, {256, 168, true},
}

// The dogs and the eating.
const (
	SleeperW, SleeperH   = 40, 13
	DogCalmChance        = 57 // per check, an alert dog not near calms by 1
	DogAlertMax          = 2  // alert 2 and near again: it attacks
	FoodBowlW, FoodBowlH = 16, 8
	FoodBowlDY           = 8 // a bowl is drawn at its row's Y + 8
	EatReach             = 4 // he eats within 4 px of the bowl ...
	EatSlowReach         = 8 // ... walking up at speed 4 within 8 px
	EatWalkSpeed         = 4
	EatStep              = 48 // added to the eating accumulator each tick
	EatPoseDX            = 8  // the eating pose is drawn 8 px to one side ...
	EatPoseDY            = 3  // ... and 3 px down
	EatPoseMaxX          = 294
	MilkLevels           = 4 // a full bowl's picture; it steps down evenly to empty
)
