package entity

import "github.com/comprofix/alleycat/internal/config"

// SwimPose is Freddy's pose underwater (DESIGN §3.4, "Underwater").
type SwimPose int

const (
	SwimSide SwimPose = iota // 4 frames, facing Dir (or left with none)
	SwimUp                   // 2 frames
	SwimDown                 // 2 frames
	SwimTurn
	SwimSink
)

// Swim is Freddy's state underwater. His sideways direction is Freddy.Dir.
type Swim struct {
	HAcc, VAcc int // speeds are HAcc ÷ 8 and VAcc ÷ 16
	VDir       int // −1 up, 0 none, 1 down
	Pose       SwimPose
	Frame      int
	Sinking    bool

	anim int
}

// StartSwim puts Freddy at the surface of the underwater scene.
func (f *Freddy) StartSwim() {
	*f = Freddy{MinX: f.MinX, MaxX: f.MaxX, Facing: f.Facing}
	if f.Facing == 0 {
		f.Facing = 1
	}
	f.X, f.Y = config.UnderwaterStartX, config.UnderwaterStartY
	f.Swim = Swim{HAcc: config.SwimHStart, VAcc: config.SwimVStart, Pose: SwimUp}
}

// Sink starts drowning: he stops responding and sinks.
func (f *Freddy) Sink() {
	s := &f.Swim
	s.Sinking = true
	f.Dir = 0
	s.VDir, s.VAcc = 1, config.SinkSpeed
}

// SwimUpdate runs one Freddy update underwater with the difficulty's speed
// caps, and reports whether he surfaced in the upward pose, which refills
// his air.
func (f *Freddy) SwimUpdate(in Controls, maxX, maxY int) (surfaced bool) {
	s := &f.Swim
	prevH, prevV, prevPose := f.Dir, s.VDir, s.Pose

	if !s.Sinking {
		f.swimSideways(in.H, maxX)
		f.swimUpDown(in.V)
	}
	surfaced = f.swimMoveY(maxY, prevPose)

	switch {
	case s.Sinking:
		s.Pose, s.Frame = SwimSink, 0
	case f.Dir != prevH || s.VDir != prevV:
		s.Pose, s.Frame = SwimTurn, 0
	default:
		f.swimPose(in)
	}
	return surfaced
}

// swimSideways: holding a direction speeds up; a new one turns at once;
// none glides on, slowing, then stops.
func (f *Freddy) swimSideways(h, maxX int) {
	s := &f.Swim
	switch {
	case h == 0 && s.HAcc >= config.SwimHStart:
		s.HAcc--
	case h != 0 && h == f.Dir:
		if s.HAcc < config.SwimHLimit {
			s.HAcc += config.SwimHAdd
		}
	default:
		f.Dir, s.HAcc = h, config.SwimHTurn
	}
	if f.Dir != 0 {
		f.Facing = f.Dir
	}
	f.RunSpeed = min(s.HAcc>>config.SwimHShift, maxX)
	f.moveH()
}

// swimUpDown: like sideways, but with no arrow he floats up.
func (f *Freddy) swimUpDown(v int) {
	s := &f.Swim
	want := v
	if want == 0 {
		want = -1
		if s.VAcc >= config.SwimVStart {
			s.VAcc--
			return
		}
	}
	switch {
	case v != 0 && v == s.VDir:
		if s.VAcc < config.SwimVLimit {
			s.VAcc += config.SwimVAdd
		}
	case v != 0 || s.VAcc < config.SwimVStart:
		s.VDir, s.VAcc = want, config.SwimVTurn
	}
}

// swimMoveY moves him up or down within the water. Reaching the top in the
// second upward frame is surfacing.
func (f *Freddy) swimMoveY(maxY int, prevPose SwimPose) (surfaced bool) {
	s := &f.Swim
	speed := min(s.VAcc>>config.SwimVShift, maxY)
	switch s.VDir {
	case 1:
		f.Y = min(f.Y+speed, config.SwimMaxY)
	case -1:
		if f.Y -= speed; f.Y <= config.SwimTopY {
			surfaced = prevPose == SwimUp && s.Frame == 1
			f.Y = config.SwimMinY
		}
	}
	return surfaced
}

// swimPose picks the sideways or the vertical pose and its frame.
func (f *Freddy) swimPose(in Controls) {
	s := &f.Swim
	s.anim++
	c := s.anim
	if in.H == 0 && in.V == 0 {
		c >>= 1
	}
	side := f.Dir != 0 && (s.VAcc>>1) < s.HAcc
	if f.Y <= config.SwimTopY+1 && in.V != 0 {
		side = false
	}
	if f.Y >= config.SwimMaxY && s.VDir == 1 {
		side = true
	}
	if side {
		s.Pose, s.Frame = SwimSide, (c>>1)&3
		return
	}
	s.Pose, s.Frame = SwimUp, (c>>1)&1
	if s.VDir == 1 {
		s.Pose = SwimDown
	}
}
