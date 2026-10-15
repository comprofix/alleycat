// Package entity holds the actors shared by the alley and the rooms.
package entity

import (
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/physics"
)

// Controls is the arrow-key state for one Freddy update: H is −1 (left),
// 0 or 1 (right); V is −1 (Up), 0 or 1 (Down).
type Controls struct {
	H, V int
}

// World is what Freddy stands on.
type World interface {
	// Support looks for a surface under f at its current position. If it
	// finds one it snaps f onto it and returns true.
	Support(f *Freddy) bool
}

// Depth is Freddy's place relative to the alley fence (DESIGN §3.8, "Fence
// depth").
type Depth int

const (
	InFront Depth = iota
	OnFence
	Behind // jumped up off the fence: the fence top no longer holds him
)

// Seat is Freddy's state on a clothesline.
type Seat int

const (
	NotSeated Seat = iota
	JustSeated
	Seated
)

// Freddy is the player's cat. X and Y are the top-left of his sprite; Y can
// go above the screen, up to config.FreddyTopY. All movement follows
// DESIGN §3.8 and runs once per Freddy update (2 per master tick).
type Freddy struct {
	X, Y int

	Dir      int // sideways motion: −1, 0 or 1
	prevDir  int
	Facing   int // last non-zero direction, for drawing
	RunSpeed int
	Arc      physics.Arc
	noLand   int // updates before he may land again

	Seat  Seat
	Depth Depth

	// MinX and MaxX clamp X (8–290 in the alley).
	MinX, MaxX int

	// WalkFrame is the walk-cycle frame, 0–5.
	WalkFrame int

	enterWait   int // updates before he appears
	enterSlices int // updates left coming into view at the screen edge

	// Knocked is set when a thrown object knocks him down: he falls behind
	// the fence and cannot land until he walks back in (DESIGN §3.8,
	// "Alley systems").
	Knocked bool

	// HoldEntry pauses his wait before walking in, while the alley dog is
	// out (DESIGN §3.8, "Re-entering the alley").
	HoldEntry bool
}

// NewFreddy returns Freddy about to walk into the alley from the left.
func NewFreddy() *Freddy {
	f := &Freddy{MinX: config.FreddyMinX, MaxX: config.FreddyMaxX}
	f.Enter(true)
	return f
}

// Enter starts Freddy walking in at floor level from the left or right
// edge (DESIGN §3.8, "Re-entering the alley").
func (f *Freddy) Enter(fromLeft bool) {
	*f = Freddy{MinX: f.MinX, MaxX: f.MaxX}
	f.X, f.Dir = config.FreddyEnterLeftX, 1
	if !fromLeft {
		f.X, f.Dir = config.FreddyEnterRightX, -1
	}
	f.prevDir, f.Facing = f.Dir, f.Dir
	f.Y = config.FloorY
	f.RunSpeed = config.FreddyWalkStartSpeed
	f.enterWait = config.FreddyEnterWait
	f.enterSlices = config.FreddyEnterSlices
}

// Hidden reports whether Freddy is waiting off-screen before walking in.
func (f *Freddy) Hidden() bool { return f.enterWait > 0 }

// EnterSlice reports whether Freddy is walking in at the screen edge, and
// if so how many pixels of him show there.
func (f *Freddy) EnterSlice() (px int, entering bool) {
	if f.enterWait > 0 || f.enterSlices == 0 {
		return 0, false
	}
	return (config.FreddyEnterSlices - f.enterSlices) * 8, true
}

// Entering reports whether he is waiting to walk in or coming into view.
func (f *Freddy) Entering() bool { return f.enterWait > 0 || f.enterSlices > 0 }

// Airborne reports whether he is rising or falling.
func (f *Freddy) Airborne() bool { return f.Arc.Dir != physics.Still }

// Height is the height of his current sprite.
func (f *Freddy) Height() int {
	if f.Seat != NotSeated {
		return config.FreddySittingH
	}
	return config.FreddyH
}

// HitBox is his box for windows, thrown objects, mice and the trash cat.
func (f *Freddy) HitBox() physics.Rect {
	return physics.Rect{X: f.X, Y: f.Y, W: config.FreddyHitW, H: config.FreddyHitH}
}

// Update runs one Freddy update.
func (f *Freddy) Update(in Controls, w World) {
	if f.enterWait > 0 {
		if !f.HoldEntry {
			f.enterWait--
		}
		return
	}
	if f.enterSlices > 0 {
		f.enterSlices--
		f.WalkFrame = (f.WalkFrame + 1) % 6
		if f.enterSlices > 0 {
			return
		}
		f.RunSpeed = config.FreddyEnterStep
		f.moveH()
	}

	switch {
	case f.Seat != NotSeated:
		f.updateSeated(in)
	case f.Airborne():
		f.updateAirborne(w)
	default:
		f.updateStanding(in, w)
	}
}

// updateSeated: on a clothesline he can only jump or drop (DESIGN §3.8,
// "Clotheslines").
func (f *Freddy) updateSeated(in Controls) {
	if f.Seat == JustSeated {
		f.Seat = Seated
		f.RunSpeed = config.FreddyLineRunSpeed
	}
	if in.V != 0 {
		f.takeInput(in)
	}
}

func (f *Freddy) updateAirborne(w World) {
	if f.moveH() {
		// Hit the side clamp: stop and fall from here.
		f.Dir = 0
		f.Arc.Dir, f.Arc.Speed = physics.Falling, config.FreddyClampFallSpeed
		f.noLand = 0
	} else {
		f.Arc.Gravity(config.FreddyMaxFallSpeed)
	}

	canLand := !f.Knocked
	if canLand && f.noLand > 0 {
		f.noLand--
		canLand = f.noLand == 0
	}
	if canLand && f.Arc.Dir == physics.Falling && w.Support(f) {
		f.land()
		return
	}

	if f.Arc.Dir == physics.Rising {
		f.Y -= f.Arc.Speed
		if f.Y < config.FreddyTopY {
			f.Y = config.FreddyTopY
			f.Arc.Dir, f.Arc.Speed = physics.Falling, 1
		}
	} else {
		f.Y += f.Arc.Speed
		if f.Y >= config.FloorY {
			f.Y = config.FloorY
			f.Depth = InFront
			f.land()
			return
		}
	}

	// Behind the fence he sinks out of sight and walks back in.
	if f.Depth == Behind && f.Y+config.FreddyH >= config.FenceHideY {
		if f.Y >= config.FenceHideY {
			f.Enter(f.X < config.FreddyEnterSplitX)
			return
		}
		f.Arc.Speed = config.FreddyClampFallSpeed
	}
}

func (f *Freddy) land() {
	f.Arc.Dir = physics.Still
	f.Knocked = false
	f.RunSpeed = config.FreddyWalkStartSpeed
	f.noLand = 0
}

func (f *Freddy) updateStanding(in Controls, w World) {
	if f.Y < config.FloorY && !w.Support(f) {
		// Walked off a can or the fence.
		f.Dir = 0
		f.startDrop()
		return
	}
	f.takeInput(in)
}

// takeInput handles the arrows while standing or seated.
func (f *Freddy) takeInput(in Controls) {
	f.prevDir = f.Dir
	f.Dir = in.H
	if f.Dir != 0 {
		f.Facing = f.Dir
	}
	switch {
	case in.V < 0:
		f.startJump()
	case in.V > 0 && f.Y < config.FloorY:
		f.startDrop()
	default:
		f.walk()
	}
}

func (f *Freddy) startJump() {
	step := uint8((15 - f.RunSpeed) * 16)
	if f.RunSpeed > config.FreddyWalkStartSpeed {
		f.RunSpeed -= config.FreddyJumpSpeedCost
	}
	f.noLand = 0
	if f.Depth == OnFence {
		f.Depth = Behind
	}
	f.Arc.Start(physics.Rising, config.FreddyJumpSpeed, step)
	f.Seat = NotSeated
}

func (f *Freddy) startDrop() {
	f.noLand = config.FreddyDropNoLand
	if f.Depth == OnFence {
		f.Depth = InFront
	}
	f.Arc.Start(physics.Falling, 1, config.FreddyDropStep)
	f.Seat = NotSeated
}

// FallOffLine drops him off the end of a clothesline he was riding.
func (f *Freddy) FallOffLine() {
	f.Seat = NotSeated
	f.RunSpeed = config.FreddyLineFallRunSpeed
	f.noLand = config.FreddyLineFallNoLand
	f.Arc.Dir, f.Arc.Speed, f.Arc.Step = physics.Falling, 1, config.FreddyLineFallStep
}

func (f *Freddy) walk() {
	f.moveH()
	if f.Dir == 0 {
		f.RunSpeed = config.FreddyWalkStartSpeed
		f.Arc.Acc = 8
		return
	}
	if f.Dir != f.prevDir {
		f.RunSpeed = config.FreddyWalkStartSpeed
	}
	if f.RunSpeed < config.FreddyWalkMaxSpeed {
		f.Arc.Acc--
		if f.Arc.Acc%config.FreddyWalkAccelEvery == 0 {
			f.RunSpeed++
		}
	}
	f.WalkFrame = (f.WalkFrame + 1) % 6
}

// moveH moves him sideways at his run speed and reports whether he hit the
// clamp.
func (f *Freddy) moveH() bool {
	switch f.Dir {
	case 1:
		f.X += f.RunSpeed
		if f.X > f.MaxX {
			f.X = f.MaxX
			return true
		}
	case -1:
		f.X -= f.RunSpeed
		if f.X < f.MinX {
			f.X = f.MinX
			return true
		}
	}
	return false
}

// KnockDown is a thrown object's hit: he drops straight down behind the
// fence at a steady speed, unable to land, and walks back in (DESIGN §3.8,
// "Alley systems").
func (f *Freddy) KnockDown() {
	f.Seat = NotSeated
	f.Dir = 0
	f.Depth = Behind
	f.Knocked = true
	f.Arc = physics.Arc{Dir: physics.Falling, Speed: config.KnockFallSpeed, Step: 1, Acc: 0xff}
}

// KnockOffLine is a mouse knocking him off a clothesline: he falls and
// cannot land for a while.
func (f *Freddy) KnockOffLine() {
	f.Seat = NotSeated
	f.Dir = 0
	f.noLand = config.MouseKnockNoLand
	f.Arc.Dir = physics.Falling
	f.Arc.Speed = max(f.Arc.Speed, 1)
}

// ReturnFromRoom puts him back where he went in, falling behind the fence
// (DESIGN §3.8, "Returning from a room").
func (f *Freddy) ReturnFromRoom(x, y int) {
	*f = Freddy{MinX: f.MinX, MaxX: f.MaxX, X: x, Y: y, Facing: 1}
	f.Depth = Behind
	f.RunSpeed = config.FreddyWalkStartSpeed
	f.Arc.Start(physics.Falling, 1, config.ReturnFallStep)
}
