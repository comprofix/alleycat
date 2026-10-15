package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Mouse is one of the cheese room's 4 mice (DESIGN §3.8, "Cheese room
// mice"). Count is its countdown: hidden at 20 and above, emerging at
// 19–18, out at 17–2 and ducking at 1–0.
type Mouse struct {
	Hole   int
	Count  int
	Caught bool
	Frame  int // flips each step while it is out
}

// Visible reports whether the mouse shows in its hole.
func (m *Mouse) Visible() bool { return !m.Caught && m.Count < config.MouseHidden }

// Transfer is Freddy going through a hole (DESIGN §3.4, "Going through a
// hole"). Frame runs 0–5 while he is hidden: 0–2 at the hole he went into,
// 3–5 at the one he comes out of.
type Transfer struct {
	Active   bool
	Frame    int
	From, To int
}

// CheeseRoom: catch the 4 mice that pop out of the cheese's 16 holes.
type CheeseRoom struct {
	Base
	Mice     [config.CheeseMice]Mouse
	Caught   int
	Transfer Transfer

	cycle     int // master ticks, for the mice's 3-tick round
	lastStart int // room tick of the last transfer's start
	started   bool
}

func newCheese(difficulty int, rng *rand.Rand) *CheeseRoom {
	r := &CheeseRoom{Base: newBase(config.Cheese, difficulty, true, rng)}
	for i := range r.Mice {
		r.Mice[i].Hole = -1
	}
	for i := len(r.Mice) - 1; i >= 0; i-- {
		r.pickHole(i, rng)
		r.Mice[i].Count = config.MouseHidden + rng.IntN(config.MouseStartRoll)
	}
	return r
}

// Kind implements Room.
func (r *CheeseRoom) Kind() Kind { return Cheese }

// Update implements Room.
func (r *CheeseRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	r.Sounds = r.Sounds[:0]
	for range ticks {
		if r.Out != Playing {
			return
		}
		r.transferTick(in)
		r.miceTick(rng)
		r.tick(rng)
	}
	for range freddySteps {
		if r.Out != Playing || r.Transfer.Active {
			return
		}
		r.step(in)
	}
}

// HoleAt is the hole Freddy is in, or −1.
func (r *CheeseRoom) HoleAt(f *entity.Freddy) int {
	if f.Airborne() || f.Gone {
		return -1
	}
	h := r.SurfaceAt(f) - config.CheeseFirstHole
	if h < 0 || h >= config.CheeseHoleCount {
		return -1
	}
	return h
}

// Partner is the hole that hole h leads to at the room's difficulty.
func (r *CheeseRoom) Partner(h int) int {
	for _, p := range config.CheesePairs[r.Difficulty%4] {
		switch h {
		case p[0]:
			return p[1]
		case p[1]:
			return p[0]
		}
	}
	return h
}

// freddyRect is Freddy's 24×14 box, used for the mice and the broom here.
func (r *CheeseRoom) freddyRect() physics.Rect { return r.Freddy.HitBox() }

// transferTick starts or runs a transfer: Alt held in a hole, clear of the
// broom, at least 12 ticks after the last one began.
func (r *CheeseRoom) transferTick(in entity.Controls) {
	f, t := r.Freddy, &r.Transfer
	if t.Active {
		t.Frame++
		if t.Frame == config.TransferSwapAfter {
			s := config.CheeseHoles[t.To]
			f.X, f.Y = s.X+config.HoleArriveDX, s.Band
		}
		if t.Frame == config.TransferFrames-1 {
			t.Active = false
			f.Gone = false
			if config.CheeseHoles[t.To].Perch {
				f.Seat = entity.JustSeated
			}
		}
		return
	}
	if !in.Act || f.Bouncing || (r.started && r.Ticks-r.lastStart < config.TransferLockout) {
		return
	}
	h := r.HoleAt(f)
	if h < 0 || r.Broom.Rect().Overlaps(r.freddyRect()) {
		return
	}
	r.started, r.lastStart = true, r.Ticks
	*t = Transfer{Active: true, From: h, To: r.Partner(h)}
	f.Seat = entity.NotSeated
	f.Dir = 0
	f.Gone = true
}

// miceTick steps the mice in turn: 1, then 2, then 3 and 0, so each steps
// once every 3 ticks.
func (r *CheeseRoom) miceTick(rng *rand.Rand) {
	switch r.cycle % config.CheeseCycleTicks {
	case 0:
		r.mouseStep(1, rng)
	case 1:
		r.mouseStep(2, rng)
	default:
		r.mouseStep(3, rng)
		r.mouseStep(0, rng)
	}
	r.cycle++
}

// MouseRect is mouse i's box, shifted down while it emerges or ducks.
func (r *CheeseRoom) MouseRect(i int) physics.Rect {
	m := &r.Mice[i]
	x, y := mouseAt(m.Hole)
	if m.Count >= config.MouseEmerge || m.Count < config.MouseOutMin {
		if m.Hole < config.CheeseTopHoles {
			y += config.CheeseTopShift
		} else {
			y += config.CheeseMouseShift
		}
	}
	return physics.Rect{X: x, Y: y, W: config.CheeseMouseW, H: config.CheeseMouseH}
}

// mouseAt is where a mouse sits in hole h.
func mouseAt(h int) (x, y int) {
	s := config.CheeseHoles[h]
	if h < config.CheeseTopHoles {
		return s.X + config.CheeseMouseDX, s.Band + config.CheeseTopMouseDY
	}
	return s.X + config.CheeseMouseDX, s.Band + config.CheeseMouseDY
}

// near is |dx| + |dy| from Freddy to (x, y).
func (r *CheeseRoom) near(x, y int) int {
	f := r.Freddy
	return abs(x-f.X) + abs(y-f.Y)
}

// touching reports whether Freddy touches mouse i. Mid-transfer he is only
// hidden, so he still counts where he is.
func (r *CheeseRoom) touching(i int) bool {
	return (!r.Freddy.Gone || r.Transfer.Active) && r.MouseRect(i).Overlaps(r.freddyRect())
}

// mouseStep runs one mouse's step: catching, the broom's freeze, the
// countdown, a new hole at 0, and the shy retreat.
func (r *CheeseRoom) mouseStep(i int, rng *rand.Rand) {
	m := &r.Mice[i]
	if m.Caught {
		return
	}
	if r.touching(i) {
		r.tryCatch(i)
		return
	}
	if r.MouseRect(i).Overlaps(r.Broom.Rect()) {
		return
	}
	if m.Count == 0 {
		r.pickHole(i, rng)
		m.Count = config.MouseHidden + rng.IntN(config.MouseRestartRoll)
	}
	m.Count--
	if r.touching(i) {
		r.tryCatch(i)
		return
	}
	x, y := mouseAt(m.Hole)
	if m.Count >= config.MouseOutMin && r.near(x, y) < config.CheeseRetreat[r.Difficulty] {
		switch {
		case m.Count < config.MouseEmerge:
			m.Count = 1
		case m.Count < config.MouseHidden:
			m.Count = 0
		}
	}
	m.Frame ^= 1
}

// tryCatch catches mouse i if it is showing; the 4th clears the room.
func (r *CheeseRoom) tryCatch(i int) {
	m := &r.Mice[i]
	if !m.Visible() {
		return
	}
	m.Caught = true
	r.Caught++
	r.sound(SoundMouseCaught)
	r.Freddy.Seat = entity.NotSeated
	if r.Caught == config.CheeseMice {
		r.end(Cleared)
	}
}

// pickHole moves mouse i to a random hole no other mouse is using,
// re-rolling up to 32 times if it is within 50 px of Freddy.
func (r *CheeseRoom) pickHole(i int, rng *rand.Rand) {
	tries := config.HoleRerolls
	for {
		h := rng.IntN(config.CheeseHoleCount)
		if r.holeTaken(i, h) {
			continue
		}
		r.Mice[i].Hole = h
		if tries == 0 {
			return
		}
		if x, y := mouseAt(h); r.near(x, y) >= config.HoleAwayDist {
			return
		}
		tries--
	}
}

func (r *CheeseRoom) holeTaken(i, h int) bool {
	for j := range r.Mice {
		if j != i && r.Mice[j].Hole == h {
			return true
		}
	}
	return false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
