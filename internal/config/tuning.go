package config

// Session rules (DESIGN §3.6).
const (
	StartLives = 3
	MaxLives   = 9
	KissBase   = 4000 // the kiss bonus is 4000 + max(0, 10800 − T) ÷ 2 (DESIGN §3.6)
	KissWindow = 10800
)

// Timing results from the R1 reference check (DESIGN §11.1.1). All agree
// with the §3.8 model; the clock itself lives in package clock.
const (
	// EatingStepsPerTick: 3-portion bowl took 0.8–0.9 s in the video
	// (2:45–2:47), matching one eating step per master tick.
	EatingStepsPerTick = 1

	// BroomStepPx per master tick: video 152 px/s vs 146 px/s (0:09.7–0:10.4).
	BroomStepPx = 8

	// ChasingDogStepPx per dog step, 2 steps every 3 ticks: video 101 px/s
	// vs 97 px/s (0:10.6–0:11.5).
	ChasingDogStepPx = 8

	// RivalCatStepPx per rival update, one update per cat every 3 ticks:
	// video 47.8 px/s vs 48.5 px/s (0:38.0–0:39.9).
	RivalCatStepPx = 8

	// Arrow movement per master tick: video 3.25/1.25 px per frame vs
	// 3.03/1.21 (0:36.0–0:36.8).
	ArrowStepX = 5
	ArrowStepY = 2
)

// Polish (DESIGN §8, M8h): presentation only, with no effect on the rules.
const (
	DustUpdates   = 6  // Freddy updates a landing puff lasts: 3 frames, 2 each
	DustOffsetX   = -4 // the puff's left edge from Freddy's X
	HitShakeTicks = 6  // master ticks the screen shakes when he is knocked
	HitShakePx    = 2
	FlickerEvery  = 40 // master ticks between chances of a lit window flickering
	FlickerOdds   = 4  // about one lit window in this many flickers at each chance
	FlickerTicks  = 6  // a flicker's length; the lamp dips every other tick
)
