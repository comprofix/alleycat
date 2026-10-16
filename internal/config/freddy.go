package config

// Freddy's movement, in pixels per Freddy update (2 per master tick)
// (DESIGN §3.8, "Freddy" and "Freddy in detail").
const (
	FreddyWalkStartSpeed = 2 // run speed on landing, standing still or reversing
	FreddyWalkMaxSpeed   = 8
	FreddyWalkAccelEvery = 4 // updates per +1 run speed

	FreddyJumpSpeed     = 8 // rising speed at take-off
	FreddyJumpSpeedCost = 2 // run speed lost on take-off (not below 2)
	FreddyMaxFallSpeed  = 4

	FreddyDropStep   = 32 // gravity step for Down or walking off a surface
	FreddyDropNoLand = 8  // updates before he can land after a drop

	FreddyClampFallSpeed = 2 // fall speed after hitting the side clamp mid-air

	// Riding off the end of a clothesline (DESIGN §3.8, "Clotheslines").
	FreddyLineFallStep     = 24
	FreddyLineFallNoLand   = 17
	FreddyLineFallRunSpeed = 1
	FreddyLineRunSpeed     = 6 // run speed while sitting on a line

	FreddyMinX = 8
	FreddyMaxX = 290
	FreddyTopY = -50 // highest a jump can carry him
)

// Freddy's sprite boxes (DESIGN §3.8): X and Y are the top-left corner.
const (
	FreddyW        = 24
	FreddyH        = 11 // walking, standing and in the air
	FreddySittingH = 14
)

// The title cat wandering the fence (DESIGN §4, "The title cat").
const (
	TitleCatSpeed      = 4   // px per master tick, no run-up
	TitleCatTurnLeftX  = 32  // at or below this he walks right
	TitleCatTurnRightX = 288 // at or above this he walks left
	TitleCatEvery      = 18  // master ticks between choices
	TitleCatStopAbove  = 160 // a random byte above this stops him
)

// Re-entering the alley (DESIGN §3.8, "Re-entering the alley").
const (
	FreddyEnterWait   = 12 // updates before he appears
	FreddyEnterSlices = 3  // updates spent coming into view, 8 px at a time
	FreddyEnterStep   = 8  // the first step into the alley
	FreddyEnterLeftX  = 0
	FreddyEnterRightX = 296
	FreddyEnterSplitX = 160 // he re-enters on the left if his last X was below this
)
