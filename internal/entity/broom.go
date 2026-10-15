package entity

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/physics"
)

// Broom is the broom in the rooms (DESIGN §3.4, "The broom in detail"). It
// sweeps Freddy's footprints off the floor, and otherwise wanders or homes
// in on him; touching it bounces him towards the room's window.
type Broom struct {
	X, Y       int
	HDir, VDir int
	Frame      int // 0–11: its swing, 7 poses there and back
	Marks      [config.BroomColumns]int

	column   int // the floor column it is looking at
	lastFoot int // the column Freddy last left a mark in
	acc      uint8
	counter  uint8
	every    int // ticks between steps
	wait     int
}

// NewBroom places the broom at its start, stepping every `every` ticks.
func NewBroom(every int, rng *rand.Rand) Broom {
	return Broom{
		X: config.BroomStartX, Y: config.BroomStartY,
		column:   config.BroomColumns - 2,
		lastFoot: -1,
		acc:      uint8(rng.IntN(256)),
		counter:  0x6c,
		every:    every,
	}
}

// Rect is the broom's box.
func (b *Broom) Rect() physics.Rect {
	return physics.Rect{X: b.X, Y: b.Y, W: config.BroomW, H: config.BroomH}
}

// Footprint leaves a mark where Freddy walks along the floor: one per column
// he enters, up to four.
func (b *Broom) Footprint(f *Freddy) {
	if f.Y < config.FloorY || f.Dir == 0 || f.Airborne() || f.Seat != NotSeated || f.Gone {
		return
	}
	col := (f.X + config.FootprintColumnOff) / 8
	if col < 0 || col >= config.BroomColumns || col == b.lastFoot {
		return
	}
	b.lastFoot = col
	if b.Marks[col] < config.FootprintMax {
		b.Marks[col]++
	}
}

func dogRect(d *Dog) physics.Rect {
	return physics.Rect{X: d.X, Y: config.DogY, W: config.DogW, H: config.DogH}
}

// touch checks the broom against Freddy, bouncing him towards the window at
// (wx, wy), and against the dog. It reports either.
func (b *Broom) touch(f *Freddy, d *Dog, wx, wy int) bool {
	if d.Capture != 0 {
		return false
	}
	r := b.Rect()
	if !f.Gone && r.Overlaps(f.HitBox()) {
		f.Bounce(wx, wy)
		return true
	}
	return d.Out() && r.Overlaps(dogRect(d))
}

// Tick runs the broom for one master tick. roomTicks is the time Freddy has
// spent in the room; (wx, wy) is the room's window.
func (b *Broom) Tick(f *Freddy, d *Dog, roomTicks, wx, wy int, rng *rand.Rand) {
	if b.wait > 1 {
		b.wait--
		return
	}
	b.wait = b.every
	if b.touch(f, d, wx, wy) {
		return
	}
	b.counter++
	b.acc ^= b.counter & uint8(rng.IntN(256))

	toX := 1
	dx := f.X - b.X
	if b.X >= f.X {
		toX, dx = -1, b.X-f.X
	}
	toY := 1
	dy := f.Y - (b.Y + config.BroomTargetDY)
	if b.Y+config.BroomTargetDY >= f.Y {
		toY, dy = -1, b.Y+config.BroomTargetDY-f.Y
	}
	dist := uint8(dx>>2) + uint8(dy>>1)

	if b.column < 0 || b.column >= config.BroomColumns-1 {
		b.column = config.BroomColumns - 2
	}
	if b.Marks[b.column] == 0 {
		b.column--
		t := uint8(roomTicks / config.BroomTimeDivisor)
		v := uint8(0)
		if dist > t {
			v = dist - t
		}
		switch {
		case v >= b.acc:
			b.VDir = 1
			switch r := rng.IntN(256); {
			case r == 0:
				b.HDir = 0
			case r <= 7:
				b.HDir = RandDir(rng)
			}
		case b.acc&0x2f == 0:
			b.HDir, b.VDir = RandDir(rng), RandDir(rng)
		case b.acc&7 == 0:
			b.HDir, b.VDir = toX, toY
		}
	} else {
		b.VDir = 1
		tx := b.column * 8
		switch {
		case b.X < tx:
			b.HDir = 1
		case b.X > tx:
			b.HDir = -1
		default:
			b.HDir = 0
			if b.Y == config.BroomFloorY {
				b.VDir = 0
				if b.Frame%(config.BroomFrames/2) == config.BroomSweepFrame {
					b.Marks[b.column]--
				}
			}
		}
	}

	px, py := b.X, b.Y
	b.X = min(max(b.X+b.HDir*config.BroomStepX, 0), config.BroomMaxX) &^ 7
	b.Y = min(max(b.Y+b.VDir*config.BroomStepY, 0), config.BroomFloorY)
	if b.touch(f, d, wx, wy) {
		b.X, b.Y = px, py
		b.HDir, b.VDir = 0, 0
		return
	}
	b.Frame = (b.Frame + 1) % config.BroomFrames
}
