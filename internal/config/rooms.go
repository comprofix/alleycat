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

// CheeseHoles are the cheese room's 16 holes, which are surfaces too
// (DESIGN §3.4, "Cheese room in detail"): 0–2 in the wedge's top are
// walkable, 3–15 in its face are perches.
var CheeseHoles = [CheeseHoleCount]Surface{
	{32, 80, 12, false}, {32, 104, 12, false}, {40, 120, 12, false},
	{72, 16, 12, true}, {64, 64, 12, true}, {80, 112, 12, true},
	{104, 32, 12, true}, {96, 72, 12, true}, {104, 192, 12, true},
	{120, 88, 12, true}, {136, 144, 12, true}, {128, 128, 12, true},
	{144, 40, 12, true}, {152, 96, 12, true}, {160, 24, 12, true},
	{160, 184, 12, true},
}

// Cheese is the cheese room (DESIGN §3.4): the furniture, then the holes
// from CheeseFirstHole on, then the wedge top. The holes come before the
// wedge top so that holes 0–2 are found first where the two overlap.
var Cheese = RoomLayout{
	StartX: 256, StartY: 96,
	WindowX: 264, WindowY: 96,
	BroomEvery: 1,
	Surfaces: append(append([]Surface{
		{120, 228, 8, false}, {152, 228, 16, false}, // chair back, seat
		{136, 256, 24, false},                                       // table top
		{88, 232, 4, true}, {96, 232, 4, true}, {104, 232, 4, true}, // curtains, left
		{88, 280, 4, true}, {96, 280, 4, true}, {104, 280, 4, true}, // curtains, right
	}, CheeseHoles[:]...),
		Surface{24, 24, 104, false}, Surface{32, 56, 88, false}, Surface{40, 88, 72, false}, // wedge top steps
		Surface{48, 120, 56, false}, Surface{56, 152, 40, false}, Surface{64, 184, 24, false},
	),
}

// CheeseFirstHole is the index of hole 0 in Cheese.Surfaces.
const CheeseFirstHole = 9

// The cheese room's mice and holes (DESIGN §3.4, "Cheese room in detail";
// §3.8, "Cheese room mice").
const (
	CheeseHoleCount   = 16
	CheeseTopHoles    = 3 // holes 0–2 are in the wedge's top
	CheeseMice        = 4
	CheeseMouseW      = 16
	CheeseMouseH      = 12
	CheeseMouseDX     = 8  // mouse X = hole left X + 8
	CheeseMouseDY     = -7 // mouse Y = band − 7 ...
	CheeseTopMouseDY  = 3  // ... or band + 3 in the top holes
	CheeseMouseShift  = 1  // emerging or ducking, it sits 1 px lower ...
	CheeseTopShift    = 3  // ... or 3 px in the top holes
	CheeseCycleTicks  = 3  // each mouse steps once every 3 ticks
	MouseHidden       = 20 // countdown ≥ 20: hidden in its hole
	MouseEmerge       = 18 // 19–18: emerging
	MouseOutMin       = 2  // 17–2: out; 1–0: ducking
	MouseStartRoll    = 16 // first countdown: 20 + 0–15
	MouseRestartRoll  = 8  // later countdowns: 20 + 0–7
	HoleAwayDist      = 50 // a new hole is re-rolled if this near Freddy ...
	HoleRerolls       = 32 // ... up to 32 times
	TransferLockout   = 12 // ticks from one transfer's start to the next
	TransferFrames    = 7  // one per tick
	TransferSwapAfter = 3  // Freddy moves to the far hole after the 3rd frame
	HoleArriveDX      = 8  // he arrives at the hole's left X + 8
	MouseIconX        = 20 // the first caught-mouse icon ...
	MouseIconY        = 2
	MouseIconStep     = 16 // ... and the next ones every 16 px
)

// CheeseRetreat is, by difficulty, how near Freddy (|dx| + |dy|) sends a
// mouse back into its hole.
var CheeseRetreat = [8]int{20, 80, 100, 120, 120, 140, 140, 140}

// CheesePairs are the two-way hole pairs by difficulty mod 4: going through
// one hole of a pair comes out of the other.
var CheesePairs = [4][8][2]int{
	{{0, 15}, {1, 6}, {2, 9}, {3, 11}, {4, 10}, {5, 12}, {7, 14}, {8, 13}},
	{{0, 14}, {1, 11}, {2, 6}, {3, 15}, {4, 12}, {5, 10}, {7, 9}, {8, 13}},
	{{0, 7}, {1, 9}, {2, 12}, {3, 13}, {4, 11}, {5, 15}, {6, 10}, {8, 14}},
	{{0, 2}, {1, 13}, {3, 14}, {4, 15}, {5, 6}, {7, 11}, {8, 12}, {9, 10}},
}

// The tally (DESIGN §3.6).
const (
	TallyTimeWindow     = 1350 // time bonus is 2 × max(0, 1350 − t) ...
	TallyDogsTimeWindow = 2700 // ... or max(0, 2700 − t) for the dogs room
	TallyRowTicks       = 2    // the count-up lights a row every 2 ticks
	TallyHoldTicks      = 30
	HereKittyTicks      = 32
)
