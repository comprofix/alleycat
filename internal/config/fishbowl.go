package config

// The fishbowl room and the underwater scene (DESIGN §3.4, "Fishbowl room
// in detail", "Underwater"; §3.8, "Swimming"). Times are master ticks;
// chances are out of 256.

// Fishbowl is the fishbowl room (DESIGN §3.4).
var Fishbowl = RoomLayout{
	StartX: 144, StartY: 96,
	WindowX: 160, WindowY: 96,
	BroomEvery: 1,
	Surfaces: []Surface{
		{120, 24, 8, false}, {152, 24, 16, false}, // left chair back, seat
		{136, 64, 40, false},                        // table top
		{152, 136, 16, false}, {120, 144, 8, false}, // middle chair seat, back
		{136, 216, 24, false},                                       // bowl table top
		{88, 128, 4, true}, {96, 128, 4, true}, {104, 128, 4, true}, // curtains, left
		{88, 176, 4, true}, {96, 176, 4, true}, {104, 176, 4, true}, // curtains, right
		{112, 272, 12, true}, {120, 272, 12, true}, // lamp shade
	},
}

// The bowl: touching its box dives Freddy in.
const (
	BowlX, BowlY = 228, 138
	BowlW, BowlH = 16, 10
	BowlFrames   = 4
	BowlEvery    = 6 // ticks per frame of the fish in the bowl
)

// Swimming, per Freddy update. Speeds are accumulator ÷ 8 sideways and
// ÷ 16 up and down.
const (
	UnderwaterStartX, UnderwaterStartY = 150, 4

	SwimHStart  = 16 // the sideways accumulator on entering, and its floor
	SwimHTurn   = 32 // after a new sideways direction
	SwimHAdd    = 3  // holding the same direction ...
	SwimHLimit  = 48 // ... while below this
	SwimHShift  = 3
	SwimVStart  = 16
	SwimVTurn   = 32 // after a new up/down direction, or turning to float up
	SwimVAdd    = 4
	SwimVLimit  = 64
	SwimVShift  = 4
	SwimMinY    = 2   // at the surface
	SwimTopY    = 3   // reaching Y ≤ 3 is surfacing
	SwimMaxY    = 179 // the bottom
	SinkSpeed   = 32  // drowning: 2 px per update, downward
	BubbleEvery = 6   // updates between bubbles while drowning
	BubbleStart = 5   // the first bubble's height above him ...
	BubbleRise  = 30  // ... growing by 30 a bubble while he lies on the bottom ...
	BubbleMax   = 200 // ... up to 200
)

// The fish and eels: 24 slots, 12 fish then 12 eels.
const (
	FishCount      = 12
	FishW, FishH   = 8, 6
	EelW, EelH     = 16, 2
	FishStep       = 4 // px sideways per step
	EelStep        = 2
	LaneDepth      = 24 // each slot moves within 24 px below its lane top
	SwimmerMaxX    = 302
	SwimmerStartX  = 256 // random start X is 0 to 255
	SwimmerTurn    = 17  // chance of a new random direction per step
	EelWakeSplitX  = 160 // a woken eel comes in at X 302 if Freddy is at X ≤ 160, else at 0
	ZapTicks       = 13
	ZapDX          = 8   // "ZAP!" is drawn at Freddy's X − 8 ...
	ZapMaxX        = 278 // ... within 0–278 ...
	ZapMaxY        = 180 // ... and at his Y, at most 180
	SurfaceColumns = 40  // 8-px wave columns along the surface ...
	SurfaceEvery   = 8   // ... one shifting every 8 ticks ...
	SurfaceSkipPx  = 16  // ... except within 16 px of Freddy at the surface
	SurfaceBandY   = 7   // Freddy is at the surface for the waves while Y ≤ 7
)

// FishLanes and EelLanes are each slot's lane top Y.
var (
	FishLanes = [FishCount]int{8, 8, 40, 40, 72, 72, 104, 104, 136, 136, 168, 168}
	EelLanes  = [FishCount]int{24, 24, 56, 56, 88, 88, 120, 120, 152, 152, 168, 168}
)

// AirStages are, by difficulty, the ticks since surfacing at which the air
// warning reaches stages 1, 2 and 3.
var AirStages = [8][3]int{
	{72, 144, 216}, {54, 108, 162}, {48, 96, 144}, {48, 96, 144},
	{48, 96, 144}, {48, 96, 144}, {46, 92, 138}, {44, 88, 132},
}
