package config

// Beeper synth (DESIGN §7).
const (
	AudioSampleRate = 44100
	AudioBufferMs   = 50 // output buffer; keeps effects close to their events

	BeeperAmplitude = 0.3  // peak level at full volume, out of 1
	BeeperSlewMs    = 2.0  // attack and release, so notes start and stop without clicks
	SoftenCutoffHz  = 3000 // "Soften beeper" one-pole low-pass
	DefaultVolume   = 0.8  // out of 1
)

// PatterParams are the background patter's settings for the alley or the
// rooms (DESIGN §7.2). All are the original's [asm].
type PatterParams struct {
	Every         int     // ticks between steps
	ClickMs       float64 // an ordinary step's click
	FillMs        float64 // a fill click, played in a rest
	BarsPerPhrase int     // a phrase may change pitch mode and start fills
	FillChance    int     // out of 256, per phrase: 0–7 fill clicks follow
}

var (
	PatterAlley = PatterParams{Every: 3, ClickMs: 0.21, FillMs: 0.64, BarsPerPhrase: 4, FillChance: 32}
	PatterRoom  = PatterParams{Every: 2, ClickMs: 0.64, FillMs: 0.21, BarsPerPhrase: 8, FillChance: 16}
)

// Other patter modes and clicks (DESIGN §7.2). Lengths and step intervals
// are the original's [asm]; the bounce's start pitch too; its rise is ours.
const (
	PatterFallEvery  = 1
	PatterAlarmEvery = 3

	DownbeatMs    = 1.7 // the first step of every bar
	AlarmClickMs  = 1.9
	FallClickMs   = 2.6
	BounceClickMs = 3.4

	// FallClickK and FallClickY0 set the falling-object click's pitch:
	// K ÷ (Y + Y0) Hz [asm].
	FallClickK  = 74574
	FallClickY0 = 32

	// The broom bounce's clicks start at 1,193,182 ÷ 1024 Hz every bounce
	// [asm], rise every tick, and wrap back to the start above BounceTopHz.
	BounceStartHz = 1165
	BounceTopHz   = 6000
	BounceRise    = 1.07 // pitch ratio from one click to the next

	PatterSweepChance = 128 // out of 256, per phrase: sweep rather than the pitch sequence [asm]
	PatterSweepRise   = 1.07
)
