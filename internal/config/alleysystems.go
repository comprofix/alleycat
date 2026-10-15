package config

// Alley systems (DESIGN §3.8, "Alley systems"). Times are master ticks;
// chances are out of 256.

// FreddyHitW and FreddyHitH are Freddy's box for windows, thrown objects,
// mice and the trash cat, whatever his pose.
const (
	FreddyHitW = 24
	FreddyHitH = 14
)

// Windows.
const (
	WindowCols       = 4
	WindowRows       = 3
	WindowX0         = 24 // first column's opening
	WindowPitchX     = 80
	WindowY0         = 24 // top row's opening
	WindowPitchY     = 32
	WindowW          = 32
	WindowH          = 15
	WindowRoll       = 16 // a roll of 12–15 opens nothing
	WindowMaxFreddyY = 96 // windows open only while Freddy is at or above this
	WindowEnterMaxY  = 96 // he enters only above this

	WindowPhaseStart    = 29 // phases count down; 29–16 opening
	WindowPhaseHoldEnd  = 16 // the hold timer starts here
	WindowPhaseHold     = 15 // fully open
	WindowEnterMinPhase = 5
	WindowEnterMaxPhase = 24

	FenceCampTicks = 72 // on the fence this long forces a top-row window

	// A mouse within this X range of a window holds it still.
	WindowMouseBefore = 16
	WindowMouseAfter  = 32
)

// Thrown objects.
const (
	ThrownW           = 16
	ThrownXSpread     = 16 // launch X is the window's X + 0..15
	ThrownMaxX        = 302
	ThrownSpeed       = 4   // px a tick sideways
	ThrownBounceSpeed = 12  // after bouncing off Freddy
	ThrownHideY       = 106 // rows from here down are behind the fence
	BonkTicks         = 10  // the game freezes while "BONK!" shows
	BonkOffsetX       = 12  // the impact is drawn this far left of Freddy
	BonkMaxX          = 270
	KnockFallSpeed    = 2
)

// ThrownH is the height of each kind of thrown object.
var ThrownH = [4]int{9, 6, 12, 12}

// ThrownMouseH is the thrown object's height when testing it against a
// mouse.
const ThrownMouseH = 12

// Mice.
const (
	MouseW, MouseH     = 16, 8
	MouseMaxX          = 302
	MouseRightStartX   = 300
	MouseSpeedHigh     = 8 // while Freddy is at Y ≤ 96
	MouseSpeedLow      = 4
	MouseSpeedFast     = 12 // charging, or holding a window
	MouseChangeHigh    = 24
	MouseChangeWalking = 40
	MouseChangeStopped = 16
	MouseRowReach      = 16 // Freddy is "in its row" from mouse Y to mouse Y + 16
	MouseRespawnTicks  = 54
	MouseIgnoreAboveY  = -12 // touches above this height do nothing
	SqueakTicks        = 8   // the game freezes while the squeak shows
	MouseKnockNoLand   = 17
)

// MouseY is each mouse's top: walking on the top, middle and bottom lines.
var MouseY = [3]int{0, 32, 64}

// MousePoints is what each mouse is worth.
var MousePoints = [3]int{90, 70, 50}

// MouseChargeChance is, by difficulty, the chance a mouse charges a seated
// Freddy (inclusive).
var MouseChargeChance = [8]int{8, 32, 64, 128, 160, 192, 208, 240}

// Trash-can cat.
const (
	TrashCatW         = 16
	TrashCatPhases    = 27
	TrashCatRisePhase = 14 // phases 26–14 rise, 13–0 sink
	TrashCatChance    = 6  // per tick when Freddy isn't on a can (≤ 5)
	TrashCatXOffset   = 6
	TrashCatXSpread   = 8
	TrashCatBaseBelow = 3  // base Y is the can's band + 3
	TrashCatRiseTop   = 15 // rising: top = base + phase − 15
	TrashCatSinkTop   = 12 // sinking: top = base + 12 − phase
)

// The alley dog.
const (
	DogW, DogH     = 32, 15
	DogY           = 177
	DogStepPx      = 8
	DogMaxX        = 286
	DogEdgeSteps   = 3 // steps to come into or go out of view
	DogCatchY      = 163
	DogCatchBefore = 24 // catch when dog X − 24 ≤ Freddy X ≤ dog X + 32
	DogCatchAfter  = 32
	DogFightMaxX   = 279
	DogSideSplitX  = 160
)

// Returning from a room (DESIGN §3.8, "Returning from a room").
const (
	ResultTicks    = 28
	ReturnFallStep = 32
)

// Pauses held by the original's sound routines (DESIGN §7.1, §12; owner:
// added to be faithful).
const (
	FenceDropTicks = 7 // Freddy sinks behind the fence and is reset
)
