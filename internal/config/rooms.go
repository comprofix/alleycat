package config

// The room framework (DESIGN §3.4, "Room framework in detail"). Times are
// master ticks; chances are out of 256.

// Entering a room or coming back to the alley.
const (
	RoomEnterStep   = 64 // gravity step of the fall into the room
	RoomEnterNoLand = 10 // updates before he can land (also on returning)
)

// Leaving by the window: Freddy's box against a 12×16 box at
// (window X − 4, window Y − 8).
const (
	ExitW, ExitH   = 12, 16
	ExitDX, ExitDY = 4, 8
)

// Surface is one surface in a room: Freddy stands on it when his Y,
// rounded down to 8, equals Band, his X rounded down to 8 is at least X,
// and his X − Reach (not below 0) rounded down to 4 is at most X. A Perch
// works like a clothesline: he sits and can only jump or drop off.
type Surface struct {
	Band, X, Reach int
	Perch          bool
}

// RoomLayout is a room's fixed geometry.
type RoomLayout struct {
	StartX, StartY   int
	WindowX, WindowY int
	Surfaces         []Surface
	BroomEvery       int // ticks between broom steps
}

// Birdcage is the birdcage room (DESIGN §3.4).
var Birdcage = RoomLayout{
	StartX: 240, StartY: 96,
	WindowX: 248, WindowY: 96,
	BroomEvery: 1,
	Surfaces: []Surface{
		{112, 16, 12, true}, {120, 16, 12, true}, // lamp shade
		{120, 56, 8, false}, {152, 56, 16, false}, // left chair back, seat
		{96, 88, 8, true},                           // picture
		{136, 128, 40, false},                       // table top
		{152, 208, 16, false}, {120, 216, 8, false}, // middle chair
		{152, 248, 16, false}, {120, 256, 8, false}, // right chair
		{88, 208, 8, true}, {96, 208, 8, true}, {104, 208, 8, true}, // curtains, left
		{88, 256, 8, true}, {96, 256, 8, true}, {104, 256, 8, true}, // curtains, right
	},
}

// The broom (DESIGN §3.4, "The broom in detail").
const (
	BroomW, BroomH     = 16, 30
	BroomStartX        = 0
	BroomStartY        = 160
	BroomStepX         = 8
	BroomStepY         = 2
	BroomMaxX          = 304
	BroomFloorY        = 165
	BroomTargetDY      = 20 // distance is measured to the broom's Y + 20
	BroomColumns       = 40
	BroomFrames        = 12 // its swing: 7 poses there and back
	BroomSweepFrame    = 3  // it sweeps a mark on the middle pose
	BroomTimeDivisor   = 8  // ticks in the room ÷ 8 reduce the distance
	FootprintMax       = 4
	FootprintY         = 192
	FootprintColumnOff = 12 // column = (X + 12) ÷ 8
)

// The broom's bounce.
const (
	BounceSpeed       = 8
	BounceNoLand      = 16
	BounceMinStrength = 48
)

// The birdcage and the bird (DESIGN §3.4, "The birdcage", "The bird").
const (
	CageW, CageH       = 24, 16
	CageStartX         = 144
	CageStartY         = 134
	CagePush           = 8
	CageSafeMinX       = 120 // it falls once it leaves X 120–168
	CageSafeMaxX       = 168
	CageFallStep       = 5
	CageFloorY         = 164
	CageBrokenW        = 32
	CageBrokenH        = 17
	CageBroomNearY     = 102 // the broom is near the table at Y ≥ 102 ...
	CageBroomBefore    = 20  // ... and within cage X − 20 ...
	CageBroomAfter     = 28  // ... to cage X + 28
	CageSeparateMax    = 32  // pushes back to clear Freddy
	CageSeparateY      = 3
	TableBand          = 136 // the table top, where the broom knocks him off
	CageKnockNoLand    = 16
	BirdW, BirdH       = 8, 5
	BirdStepX          = 4
	BirdStepY          = 2
	BirdMinX, BirdMaxX = 0, 309
	BirdMinY, BirdMaxY = 48, 167
	BirdExtraEvery     = 4  // every 4th step a second step follows in the same tick
	BirdCheckChance    = 49 // per step: check Freddy, or change direction
	BirdLeaveChance    = 7  // per step, abandon a waypoint while flying to it
	BirdBoredChance    = 17 // per step, leave a waypoint it has reached
	BirdWaypointRoll   = 256
)

// BirdFlee is, by difficulty, how near Freddy (|dx| + 2|dy|) makes the bird
// flee.
var BirdFlee = [8]int{46, 92, 140, 140, 160, 170, 180, 180}

// BirdWaypoints are the perches the bird may fly to.
var BirdWaypoints = [11][2]int{
	{24, 106}, {220, 122}, {96, 82}, {64, 122}, {272, 80}, {64, 154},
	{260, 158}, {268, 122}, {248, 98}, {248, 80}, {224, 80},
}

// The tally (DESIGN §3.6).
const (
	TallyTimeWindow     = 1350 // time bonus is 2 × max(0, 1350 − t) ...
	TallyDogsTimeWindow = 2700 // ... or max(0, 2700 − t) for the dogs room
	TallyRowTicks       = 2    // the count-up lights a row every 2 ticks
	TallyHoldTicks      = 30
	HereKittyTicks      = 32
)
