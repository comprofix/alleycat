package config

// The heart room (DESIGN §3.5, "Heart room in detail"). Times are master
// ticks; chances are out of 256.

// HeartRows are the 7 row lines, bottom first. Freddy stands at a line +
// HeartStandDY, the cats at a line + HeartCatDY, and the hearts are drawn
// from a line + HeartDrawDY.
var HeartRows = [7]int{176, 152, 128, 104, 80, 56, 32}

// Hearts.
const (
	HeartStandDY   = 5
	HeartCatDY     = 3
	HeartDrawDY    = 15
	HeartCols      = 16 // drawn hearts per row, 16 px apart from HeartX
	HeartX         = 32
	HeartW, HeartH = 16, 8
	HeartCells     = 18 // columns 16 and 17, past the last heart, count as intact
	HeartStandDX   = 7  // Freddy's heart is column (X + 7) ÷ 16 − 2
)

// Freddy in the heart room.
const (
	HeartStartX, HeartStartY = 152, 175
	HeartMinX, HeartMaxX     = 36, 270
	HeartFallOffY            = 198 // falling this far is falling off the bottom
)

// The arrows.
const (
	ArrowW, ArrowH   = 16, 8
	ArrowSpawnRoll   = 26 // 0–15 from a side, 16–25 from the top
	ArrowSides       = 16
	ArrowLeftX       = 12
	ArrowRightX      = 288
	ArrowSideDY      = 8 // side heights are the table + 8
	ArrowTopY        = 6
	ArrowTopDX       = 4 // top X is the table + 4
	ArrowTopRight    = 5 // the first 5 top places head right
	ArrowMaxY        = 191
	ArrowMaxX        = 300
	ArrowGrowTicks   = 40 // its 6 frames play over the first 40 ticks
	ArrowFrames      = 6
	ArrowCellDY      = 8 // it flips the heart under it when (Y − 8) rounded down to 8 is a row line
	ArrowHitSpeed    = 2
	ArrowHitStep     = 0x20
	ArrowHitNoLand   = 8
	ArrowCellCols    = 16
	ArrowCellOffsetX = 2 // column X ÷ 16 − 2
)

// ArrowSideY and ArrowTopX are the arrows' spawn tables.
var (
	ArrowSideY = [8]int{0, 24, 48, 72, 96, 120, 144, 168}
	ArrowTopX  = [10]int{0, 32, 64, 96, 128, 160, 192, 224, 256, 288}
)

// The cats: 6 rivals (slots 0–5, bottom row first) and Felicia (slot 6).
const (
	HeartCatSlots      = 7
	Felicia            = 6
	CatW, CatH         = 24, 12
	CatMinX, CatMaxX   = 37, 266
	CatStartX          = 96 // rivals start at a random X from 96 ...
	CatStartRoll       = 128
	CatBumpStep        = 4 // after bumping Freddy, 4 px a step ...
	CatBumpSteps       = 4 // ... for 4 steps, then it turns back
	CatStartChance     = 17
	CatStopChance      = 9
	CatAlign           = 8  // chasing compares X rounded down to 8
	FeliciaComeY       = 40 // above this Freddy draws Felicia to him
	CatReturnLeftX     = 36 // a lured rival comes back at X 36 ...
	CatReturnRightX    = 264
	CatReturnSplitX    = 160 // ... if Freddy is right of 160, else at 264
	CatFrames          = 6
	CatKnockSpeed      = 4
	CatKnockStep       = 8
	CatKnockNoLand     = 4
	BottomCatLureTimes = 2 // the bottom-row rival stays away twice as long
)

// Gifts.
const (
	GiftW, GiftH = 24, 15
	GiftMax      = 8
	GiftMaxX     = 263
)

// GiftX are the gifts' starting places on the bottom row.
var GiftX = [GiftMax]int{96, 200, 64, 232, 32, 264, 128, 168}

// The rejection scene (DESIGN §3.5, "Failing").
const (
	RejectX                  = 148 // the head falls here, whatever Freddy's X
	RejectHeadW, RejectHeadH = 24, 12
	RejectFallStep           = 12  // px a tick, from the top ...
	RejectLandY              = 168 // ... until it reaches this
	RejectFaceW, RejectFaceH = 48, 17
	RejectHoldTicks          = 18 // the face and bubble then stay this long
)

// The kiss (DESIGN §3.5, "The kiss").
const (
	KissOpenTicks  = 10  // the big heart shows alone this long ...
	KissTextTicks  = 72  // ... then the bonus and multiplier this long (measured, 4 s)
	KissFlashTicks = 4   // the text bar flashes every 4 ticks
	KissStepTicks  = 2   // then the heart drops a step every 2 ticks
	KissHeartY     = 20  // it starts here ...
	KissHeartStep  = 8   // ... dropping 8 px a step ...
	KissHeartStopY = 84  // ... until it reaches this
	KissCentreX    = 128 // sliding towards the middle by
	KissSlideMax   = 13  // min(|X − 128| ÷ 8, 13) + 2 a step
	KissSlideAdd   = 2
	KissMaxX       = 278 // it starts at Freddy's X (at most 278) − 16, rounded down to 16
	KissStartDX    = 16
	KissHeartDX    = 4  // the heart is drawn 4 px right of its X ...
	KissCatsDX     = 12 // ... with the cats 12 px further right ...
	KissCatsDY     = 6  // ... and 6 down
	KissScoreY     = 56 // the bonus, and below it the multiplier bar
	KissBarY       = 80
)

// KissBurst is one of the three bursts of hearts that end the kiss: 8
// hearts fly out from Start in the 8 directions, Step a move, every Every
// ticks, leaving trails, until one reaches a limit.
type KissBurst struct {
	Sprite         string
	StepX, StepY   int
	StartX, StartY int
	MaxX, MaxY     int
	Every          int
}

// KissBursts are the bursts in order: small, middle and big hearts.
var KissBursts = [3]KissBurst{
	{"heart_small", 16, 10, 152, 96, 303, 191, 1},
	{"heart_mid", 32, 20, 144, 92, 287, 183, 2},
	{"heart_big", 56, 35, 132, 84, 263, 167, 3},
}

// KissDirs are the 8 burst directions: up, up-right, right, down-right,
// down, down-left, left, up-left.
var KissDirs = [8][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}

// The dancing-cats interlude after a kiss, when the new difficulty is 2 or
// more (DESIGN §3.5, "The dancing-cats interlude").
const (
	DanceMinDifficulty = 2
	DanceTicks         = 40 // it lasts 40 ticks ...
	DanceFrameTicks    = 5  // ... the cats changing step every 5
)

// Dancer is one dancing cat: its top-left, and whether it starts on the
// other step.
type Dancer struct {
	X, Y int
	Off  bool
}

// danceLine is the four-cat line; the ring, the frame and the inner group
// build the bigger arrangements. Places are the original's layout.
var (
	danceLine    = []Dancer{{96, 84, false}, {128, 84, false}, {160, 84, false}, {192, 84, false}}
	danceRing    = []Dancer{{128, 10, false}, {160, 10, false}, {64, 46, false}, {96, 46, false}, {192, 46, false}, {224, 46, false}, {32, 82, false}, {64, 82, false}, {224, 82, false}, {256, 82, false}, {64, 118, false}, {96, 118, false}, {192, 118, false}, {224, 118, false}, {128, 154, false}, {160, 154, false}}
	danceInner   = []Dancer{{128, 46, true}, {160, 46, true}, {96, 82, true}, {128, 82, true}, {160, 82, true}, {192, 82, true}, {128, 118, true}, {160, 118, true}}
	danceFrame   = []Dancer{{0, 10, true}, {32, 10, true}, {64, 10, true}, {96, 10, true}, {288, 10, true}, {256, 10, true}, {224, 10, true}, {192, 10, true}, {0, 46, true}, {32, 46, true}, {288, 46, true}, {256, 46, true}, {0, 82, true}, {288, 82, false}, {0, 118, true}, {32, 118, true}, {288, 118, true}, {256, 118, true}, {0, 154, true}, {32, 154, true}, {64, 154, true}, {96, 154, true}, {288, 154, true}, {256, 154, true}, {224, 154, true}, {192, 154, true}}
	danceDiamond = []Dancer{{128, 36, false}, {160, 36, false}, {96, 84, false}, {128, 84, false}, {160, 84, false}, {192, 84, false}, {128, 132, false}, {160, 132, false}}
)

// DanceLayouts are the dancers by the new difficulty: none below 2; a line
// of 4; a diamond of 8; a ring of 16; the ring and an inner group (24); a
// frame round the screen (26); the frame and the inner group (34).
var DanceLayouts = [MaxDifficulty + 1][]Dancer{
	nil, nil, danceLine, danceDiamond, danceRing,
	append(append([]Dancer{}, danceRing...), danceInner...),
	danceFrame,
	append(append([]Dancer{}, danceFrame...), danceInner...),
}
