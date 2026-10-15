package config

// Alley surfaces, as the Y Freddy stands at (DESIGN §3.8, "Alley surfaces").
const (
	FloorY     = 180
	FenceTopY  = 96
	TallCanY   = 134
	ShortCanY  = 142
	CanReach   = 15 // X − 15, rounded down to 8, must not pass the can's X
	LineProbeX = 10 // Freddy's X + 10 is tested against the laundry
)

// LineY is the Y Freddy sits at on each clothesline, top to bottom.
var LineY = [3]int{8, 40, 72}

// LineBand is the Y range, per line, where Freddy in the air pauses that
// line (DESIGN §3.8, "Clotheslines"). Lo is inclusive, Hi exclusive.
var LineBand = [3]struct{ Lo, Hi int }{{-4, 25}, {28, 57}, {60, 89}}

// FenceHideY is the first row hidden behind the fence when Freddy is behind
// it (DESIGN §3.8, "Fence depth").
const FenceHideY = 105

// Clothesline scrolling (DESIGN §3.8, "Clotheslines").
const (
	LineStepPx       = 4  // per step, one step per master tick
	LinePiecePx      = 16 // laundry is generated a piece at a time
	LineSwitchChance = 65 // /256, per completed piece
)

// LineDir is each line's direction: the middle line moves the other way.
var LineDir = [3]int{1, -1, 1}

// Laundry generation (DESIGN §3.8, "Clotheslines"). Chances are out of 256.
var LineFillAbove = [3]int{128, 48, 0} // a piece is filled when a roll is above this

const (
	LaundryBigChance    = 24 // big item over both halves
	LaundryMediumChance = 72 // medium item over both halves
	LaundrySmallKinds   = 3  // plus "missing", each 1 in 4
)
