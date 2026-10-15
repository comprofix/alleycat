// Package entity holds the actors shared by the alley and the rooms.
package entity

import (
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/physics"
)

// Controls is the input for one Freddy update: H is −1 (left), 0 or 1
// (right); V is −1 (Up), 0 or 1 (Down); Act is Action (Alt) held.
type Controls struct {
	H, V int
	Act  bool
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

	// Gone is set while a room dog fights him: he is off the screen and does
	// not move (DESIGN §3.4).
	Gone bool

	// Bouncing is set by the broom's bounce until he lands; he cannot be
	// bounced again meanwhile (DESIGN §3.4, "The broom in detail").
	Bouncing bool

	// Sank reports that he sank out of sight behind the fence and was reset.
	// It stays set until the alley clears it, so it can pause and play the
	// crackle (DESIGN §7.3, §12).
	Sank bool

	// Leapt reports that he started a jump or a drop. It stays set until the
	// alley clears it, for the trash-can step-off click (DESIGN §7.3).
	Leapt bool

	// Swim is his state underwater (DESIGN §3.4, "Underwater").
	Swim Swim

	// NoFloor is set in the heart room: there is no floor to land on, and
	// the bottom row is held up by its hearts like any other (DESIGN §3.5).
	NoFloor bool
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
	if f.Gone {
		return
	}
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
		if f.Y >= config.FloorY && !f.NoFloor {
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
			f.Sank = true
			return
		}
		f.Arc.Speed = config.FreddyClampFallSpeed
	}
}

func (f *Freddy) land() {
	f.Arc.Dir = physics.Still
	f.Knocked = false
	f.Bouncing = false
	f.RunSpeed = config.FreddyWalkStartSpeed
	f.noLand = 0
}

func (f *Freddy) updateStanding(in Controls, w World) {
	if (f.Y < config.FloorY || f.NoFloor) && !w.Support(f) {
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
	case in.V > 0 && (f.Y < config.FloorY || f.NoFloor):
		f.startDrop()
	default:
		f.walk()
	}
}

func (f *Freddy) startJump() {
	f.Leapt = true
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
	f.Leapt = true
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
	f.noLand = config.RoomEnterNoLand
}

// Bounce is the broom's knock: it throws him towards the room's window at
// (wx, wy), harder the nearer he is to it (DESIGN §3.4, "The broom in
// detail").
func (f *Freddy) Bounce(wx, wy int) {
	if f.Bouncing {
		return
	}
	f.Bouncing = true
	f.Seat = NotSeated
	dir := physics.Falling
	if f.Y >= wy {
		dir = physics.Rising
	}
	var d int
	if f.X > wx {
		f.Dir, d = -1, f.X-wx
	} else {
		f.Dir, d = 1, wx-f.X-1
	}
	s := max(255-min(d, 255), config.BounceMinStrength)
	step := s - s/4
	f.RunSpeed = step >> 5
	f.Arc = physics.Arc{Dir: dir, Speed: config.BounceSpeed, Step: uint8(step), Acc: 1}
	f.noLand = config.BounceNoLand
}

// EnterRoom places him at a room's start point, falling, unable to land for
// a moment (DESIGN §3.4, "Room framework in detail").
func (f *Freddy) EnterRoom(x, y, noLand int) {
	*f = Freddy{MinX: f.MinX, MaxX: f.MaxX, X: x, Y: y, Facing: 1}
	f.RunSpeed = config.FreddyWalkStartSpeed
	f.Arc.Start(physics.Falling, 1, config.RoomEnterStep)
	f.noLand = noLand
}

// Topple knocks him off whatever he is standing on: he falls and cannot
// land for noLand updates. His sideways motion is unchanged.
func (f *Freddy) Topple(noLand int) {
	f.Seat = NotSeated
	f.noLand = noLand
	f.Arc.Dir = physics.Falling
	f.Arc.Speed = max(f.Arc.Speed, 1)
}

// Knock drives him downward at speed with gravity step, unable to land for
// noLand updates; his sideways motion is unchanged (DESIGN §3.5: an arrow
// or a rival cat in the heart room).
func (f *Freddy) Knock(speed int, step uint8, noLand int) {
	f.Seat = NotSeated
	f.Arc.Dir, f.Arc.Speed, f.Arc.Step = physics.Falling, speed, step
	f.noLand = noLand
}

// Standing reports whether he is standing on a walkable surface or the
// floor (not in the air, not seated).
func (f *Freddy) Standing() bool { return !f.Airborne() && f.Seat == NotSeated }
