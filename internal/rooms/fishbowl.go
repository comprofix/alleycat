package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Phase is where Freddy is in the fishbowl room.
type Phase int

const (
	InRoom     Phase = iota
	Underwater       // in the bowl: the underwater scene
	Zapped           // an eel got him: "ZAP!" before the room fails
)

// Swimmer is a fish (slots 0–11) or an eel (12–23) underwater (DESIGN §3.4,
// "Underwater").
type Swimmer struct {
	X, Y       int
	HDir, VDir int
	Off        bool // eaten, or an eel not yet awake
}

// Bubble is a bubble left behind by a drowning Freddy.
type Bubble struct{ X, Y int }

const swimmers = 2 * config.FishCount

// FishbowlRoom: dive into the bowl and eat the 12 fish without touching an
// eel or running out of air.
type FishbowlRoom struct {
	Base
	Phase     Phase
	BowlFrame int

	Swimmers  [swimmers]Swimmer
	FishLeft  int
	FishFrame int // 0 or 1, flipping every 2 ticks
	EelFrame  int // 0–3, advancing every 2 ticks
	Surface   [config.SurfaceColumns]int
	Drowning  bool
	Bubbles   []Bubble
	Zap       int // ticks of "ZAP!" left
	ZapX      int
	ZapY      int

	bowlWait    int
	slot        int // the next swimmer slot to move
	lastSurface int // room tick he last surfaced
	bubbleWait  int
	bubbles     int // drowning bubbles so far, for the air warning's falling pitch
	bubbleDist  int
	surfaceCol  int
	surfaceWait int
}

func newFishbowl(difficulty int, rng *rand.Rand) *FishbowlRoom {
	return &FishbowlRoom{Base: newBase(config.Fishbowl, difficulty, false, rng)}
}

// Kind implements Room.
func (r *FishbowlRoom) Kind() Kind { return Fishbowl }

// Update implements Room.
func (r *FishbowlRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	r.Sounds = r.Sounds[:0]
	for range ticks {
		if r.Out != Playing {
			return
		}
		switch r.Phase {
		case InRoom:
			r.bowlTick()
			r.tick(rng)
		case Underwater:
			r.Ticks++
			r.swimmersTick(rng)
			r.surfaceTick()
			// Underwater Freddy updates once per tick, not twice (DESIGN
			// §3.4, "Underwater").
			r.swimStep(in)
		case Zapped:
			r.Ticks++
			if r.Zap--; r.Zap <= 0 {
				r.end(Failed)
			}
		}
	}
	for range freddySteps {
		if r.Out != Playing {
			return
		}
		switch r.Phase {
		case InRoom:
			r.step(in)
			if !r.Freddy.Gone && r.BowlRect().Overlaps(r.Freddy.HitBox()) {
				r.dive(rng)
			}
		}
	}
}

// BowlRect is the box that dives Freddy into the bowl.
func (r *FishbowlRoom) BowlRect() physics.Rect {
	return physics.Rect{X: config.BowlX, Y: config.BowlY, W: config.BowlW, H: config.BowlH}
}

func (r *FishbowlRoom) bowlTick() {
	if r.bowlWait++; r.bowlWait >= config.BowlEvery {
		r.bowlWait = 0
		r.BowlFrame = (r.BowlFrame + 1) % config.BowlFrames
	}
}

// dive starts the underwater scene: Freddy at the surface with a full
// breath, and the fish and eels in their lanes.
func (r *FishbowlRoom) dive(rng *rand.Rand) {
	r.Phase = Underwater
	r.Freddy.StartSwim()
	r.lastSurface = r.Ticks
	r.bubbleWait, r.bubbleDist = 1, config.BubbleStart
	r.FishLeft = config.FishCount
	for i := swimmers - 1; i >= 0; i-- {
		s := &r.Swimmers[i]
		*s = Swimmer{Y: laneTop(i), VDir: 1, HDir: entity.RandDir(rng), X: rng.IntN(config.SwimmerStartX)}
	}
	for n := config.FishCount - r.table.ActiveEels; n > 0; {
		s := &r.Swimmers[config.FishCount+rng.IntN(config.FishCount)]
		if !s.Off {
			s.Off = true
			n--
		}
	}
}

func laneTop(i int) int {
	if i < config.FishCount {
		return config.FishLanes[i]
	}
	return config.EelLanes[i-config.FishCount]
}

// IsEel reports whether slot i is an eel.
func IsEel(i int) bool { return i >= config.FishCount }

// SwimmerRect is slot i's box.
func (r *FishbowlRoom) SwimmerRect(i int) physics.Rect {
	s := &r.Swimmers[i]
	if IsEel(i) {
		return physics.Rect{X: s.X, Y: s.Y, W: config.EelW, H: config.EelH}
	}
	return physics.Rect{X: s.X, Y: s.Y, W: config.FishW, H: config.FishH}
}

// swimmersTick moves the swimmers in two halves: slots 1–12 on one tick,
// then 13–23 and 0 on the next, so each moves every 2 ticks.
func (r *FishbowlRoom) swimmersTick(rng *rand.Rand) {
	for {
		r.slot++
		if r.slot >= swimmers {
			r.slot = 0
			r.FishFrame ^= 1
			r.EelFrame = (r.EelFrame + 1) % 4
		}
		r.swimmerStep(r.slot, rng)
		if r.slot == 0 || r.slot == config.FishCount {
			return
		}
	}
}

func (r *FishbowlRoom) swimmerStep(i int, rng *rand.Rand) {
	s := &r.Swimmers[i]
	if s.Off {
		return
	}
	if rng.IntN(256) < config.SwimmerTurn {
		s.HDir, s.VDir = entity.RandDir(rng), entity.RandDir(rng)
	}
	step := config.FishStep
	if IsEel(i) {
		step = config.EelStep
	}
	if s.X += s.HDir * step; s.X < 0 {
		s.X, s.HDir = 0, 1
	} else if s.X > config.SwimmerMaxX {
		s.X, s.HDir = config.SwimmerMaxX, -1
	}
	top := laneTop(i)
	if s.Y += s.VDir; s.Y < top {
		s.Y, s.VDir = top, 1
	} else if s.Y > top+config.LaneDepth {
		s.Y, s.VDir = top+config.LaneDepth, -1
	}
}

// surfaceTick shifts one wave column of the surface every 8 ticks, leaving
// it alone right by Freddy when he is at the surface.
func (r *FishbowlRoom) surfaceTick() {
	if r.surfaceWait++; r.surfaceWait < config.SurfaceEvery {
		return
	}
	r.surfaceWait = 0
	r.surfaceCol = (r.surfaceCol + 1) % config.SurfaceColumns
	f := r.Freddy
	if f.Y <= config.SurfaceBandY && abs(f.X+4-r.surfaceCol*8) < config.SurfaceSkipPx {
		return
	}
	r.Surface[r.surfaceCol] = (r.Surface[r.surfaceCol] + 1) % 4
}

// AirTicks is how long Freddy has been under since he last surfaced.
func (r *FishbowlRoom) AirTicks() int { return r.Ticks - r.lastSurface }

// AirStage is the air warning, 0–3; 0 again while drowning.
func (r *FishbowlRoom) AirStage() int {
	if r.Drowning {
		return 0
	}
	stage := 0
	for _, t := range config.AirStages[r.Difficulty] {
		if r.AirTicks() >= t {
			stage++
		}
	}
	return stage
}

// swimStep runs one Freddy update underwater: air, swimming, bubbles and
// contacts.
func (r *FishbowlRoom) swimStep(in entity.Controls) {
	f, d := r.Freddy, r.table
	if air := r.AirTicks(); air >= d.DrowningTicks {
		if air >= d.AirFailTicks {
			r.end(Failed)
			return
		}
		if !r.Drowning {
			r.Drowning = true
			f.Sink()
		}
	}
	if f.SwimUpdate(in, d.SwimMaxX, d.SwimMaxY) {
		r.lastSurface = r.Ticks
	}
	if r.Drowning {
		r.bubble()
	}
	r.contacts()
}

// bubble leaves a bubble above him every 6 updates, higher each time while
// he lies on the bottom.
func (r *FishbowlRoom) bubble() {
	if r.bubbleWait--; r.bubbleWait > 0 {
		return
	}
	r.bubbleWait = config.BubbleEvery
	f := r.Freddy
	if f.Y >= config.SwimMaxY && r.bubbleDist < config.BubbleMax {
		r.bubbleDist += config.BubbleRise
	}
	r.soundArg(SoundAirWarning, r.bubbles)
	r.bubbles++
	r.Bubbles = append(r.Bubbles, Bubble{X: f.X, Y: max(f.Y-r.bubbleDist, 0) &^ 7})
}

// contacts eats the fish Freddy touches, each waking a sleeping eel, and
// starts "ZAP!" on touching an eel.
func (r *FishbowlRoom) contacts() {
	f := r.Freddy
	eaten := 0
	for i := range r.Swimmers {
		s := &r.Swimmers[i]
		if s.Off || !r.SwimmerRect(i).Overlaps(f.HitBox()) {
			continue
		}
		if IsEel(i) && r.Out == Playing && !r.Drowning {
			r.zap()
			return
		}
		eaten++
		s.Off = true
		if !IsEel(i) {
			r.sound(SoundFishEaten)
			if r.FishLeft--; r.FishLeft == 0 && !r.Drowning {
				r.end(Cleared)
			}
		}
	}
	for ; eaten > 0; eaten-- {
		r.wakeEel()
	}
}

// wakeEel brings in a sleeping eel from the side away from Freddy.
func (r *FishbowlRoom) wakeEel() {
	for i := swimmers - 1; i >= config.FishCount; i-- {
		s := &r.Swimmers[i]
		if !s.Off {
			continue
		}
		s.Off = false
		s.X, s.HDir = 0, 1
		if r.Freddy.X <= config.EelWakeSplitX {
			s.X, s.HDir = config.SwimmerMaxX, -1
		}
		return
	}
}

func (r *FishbowlRoom) zap() {
	r.sound(SoundZap)
	f := r.Freddy
	r.Phase = Zapped
	r.Zap = config.ZapTicks
	r.ZapX = min(max(f.X-config.ZapDX, 0), config.ZapMaxX)
	r.ZapY = min(f.Y, config.ZapMaxY)
}
