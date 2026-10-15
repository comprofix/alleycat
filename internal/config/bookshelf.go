package config

// The bookshelf room (DESIGN §3.4, "Bookshelf room in detail"). Times are
// master ticks.

// Bookshelf is the bookshelf room (DESIGN §3.4).
var Bookshelf = RoomLayout{
	StartX: 88, StartY: 96,
	WindowX: 96, WindowY: 96,
	BroomEvery: 2,
	Ladder:     &Ladder{MinX: 164, MaxX: 280, TopY: 40, BottomY: 152, Gap: 16, Snap: 2},
	Surfaces: []Surface{
		{112, 16, 12, true}, {120, 16, 12, true}, // left lamp shade
		{112, 136, 12, true}, {120, 136, 12, true}, // right lamp shade
		{120, 48, 8, false}, {152, 48, 16, false}, // left chair back, seat
		{152, 96, 16, false}, {120, 104, 8, false}, // right chair seat, back
		{88, 56, 12, true}, {96, 56, 12, true}, {104, 56, 12, true}, // curtains, left
		{88, 104, 12, true}, {96, 104, 12, true}, {104, 104, 12, true}, // curtains, right
	},
}

// The ferns on top of the bookcase.
var FernX = [3]int{192, 224, 256}

const (
	FernY        = 24
	FernW, FernH = 16, 16
)

// The spider.
const (
	SpiderW, SpiderH = 24, 14
	SpiderStartX     = 280
	SpiderTopY       = 8 // it waits and tracks Freddy at this Y
	SpiderEvery      = 2 // ticks between steps
	SpiderDropMin    = 2 // drop and climb steps run 2, 3, 4, 5 px ...
	SpiderDropEvery  = 4 // ... rising every 4 steps, then repeating
	SpiderAlign      = 8 // it drops when its X and Freddy's agree, rounded down to 8
	SpiderReach      = 48
	SpiderMaxY       = 159
	SpiderFrameEvery = 2 // steps per animation frame
	ByteTicks        = 9
	ByteDX, ByteDY   = 12, 4 // "BYTE!" is drawn at Freddy's X − 12 and Y − 4 ...
	ByteMaxX         = 270   // ... with X within 0–270
)

// SpiderSpeed is the spider's sideways step along the top, by difficulty.
var SpiderSpeed = [8]int{2, 4, 6, 8, 10, 12, 14, 16}

// The bookcase, as drawn: 15 × 16 cells of 8 px.
const (
	BookcaseX, BookcaseY = 168, 40
	BookCols, BookRows   = 15, 16
	BookCell             = 8
)
