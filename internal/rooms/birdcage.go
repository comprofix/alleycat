package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Cage is the birdcage (DESIGN §3.4, "The birdcage").
type Cage struct {
	X, Y    int
	Pending bool // pushed: it moves on the next tick
	PushDir int
	Falling bool
	Broken  bool
}

// Bird is the bird, once the cage has broken (DESIGN §3.4, "The bird").
type Bird struct {
	X, Y   int
	VX, VY int
	Frame  int

	waypoint int // 0–10, or none above that
	steps    int
}

// BirdcageRoom: knock the cage off the table and catch the bird.
type BirdcageRoom struct {
	Base
	Cage Cage
	Bird Bird
}

func newBirdcage(difficulty int, rng *rand.Rand) *BirdcageRoom {
	return &BirdcageRoom{
		Base: newBase(config.Birdcage, difficulty, true, rng),
		Cage: Cage{X: config.CageStartX, Y: config.CageStartY},
	}
}

// Kind implements Room.
func (r *BirdcageRoom) Kind() Kind { return Birdcage }

// Update implements Room. While the cage falls the rest of the room waits,
// unless the dog is out, in which case the cage waits for it to go.
func (r *BirdcageRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	frozen := false
	for range ticks {
		if r.Out != Playing {
			return
		}
		c := &r.Cage
		if c.Falling && !c.Broken && !r.Dog.Out() {
			r.fall(rng)
			frozen = true
			continue
		}
		r.cageTick()
		r.birdTick(rng)
		r.tick(rng)
	}
	if frozen {
		return
	}
	for range freddySteps {
		if r.Out != Playing {
			return
		}
		r.step(in)
		r.checkBird()
	}
}

func (r *BirdcageRoom) cageRect() physics.Rect {
	return physics.Rect{X: r.Cage.X, Y: r.Cage.Y, W: config.CageW, H: config.CageH}
}

// onTable reports whether Freddy is standing on the table top.
func (r *BirdcageRoom) onTable() bool {
	f := r.Freddy
	return f.Standing() && f.Y >= 0 && physics.Band(f.Y) == config.TableBand
}

func (r *BirdcageRoom) cageTick() {
	c, f, b := &r.Cage, r.Freddy, &r.Broom
	if c.Broken || c.Y >= config.CageFloorY {
		return
	}
	// The broom at the table knocks Freddy off it and holds the cage.
	if b.Y >= config.CageBroomNearY && b.X >= c.X-config.CageBroomBefore && b.X <= c.X+config.CageBroomAfter {
		if r.onTable() {
			f.Topple(config.CageKnockNoLand)
		}
		return
	}
	if c.Falling {
		if r.Dog.Out() && r.onTable() {
			f.Topple(config.CageKnockNoLand)
		}
		return
	}
	if !f.Gone && r.cageRect().Overlaps(f.HitBox()) {
		if !c.Pending {
			c.PushDir = f.Dir
			if c.PushDir == 0 {
				c.PushDir = -1
				if c.X > f.X {
					c.PushDir = 1
				}
			}
		}
		c.Pending = true
		r.separate()
		return
	}
	if !c.Pending {
		return
	}
	c.X += c.PushDir * config.CagePush
	if r.cageRect().Overlaps(f.HitBox()) {
		return
	}
	c.Pending = false
	if c.X < config.CageSafeMinX || c.X > config.CageSafeMaxX {
		c.Falling = true
	}
}

// separate pushes Freddy back until he is clear of the cage.
func (r *BirdcageRoom) separate() {
	c, f := &r.Cage, r.Freddy
	for range config.CageSeparateMax {
		f.X -= c.PushDir * config.CagePush
		f.Dir = -c.PushDir
		switch f.Arc.Dir {
		case physics.Rising:
			f.Y -= config.CageSeparateY
		case physics.Falling:
			f.Y += config.CageSeparateY
		}
		if !r.cageRect().Overlaps(f.HitBox()) {
			return
		}
	}
}

// fall drops the cage 5 px; on landing it breaks and frees the bird.
func (r *BirdcageRoom) fall(rng *rand.Rand) {
	c := &r.Cage
	if c.Y += config.CageFallStep; c.Y < config.CageFloorY {
		return
	}
	c.Broken = true
	f := r.Freddy
	r.Bird = Bird{X: c.X, Y: min(c.Y, config.BirdMaxY), VX: 1, waypoint: 0xff}
	if r.Bird.X < f.X {
		r.Bird.VX = -1
	}
}

// birdTick runs the bird: about 5 steps every 4 ticks.
func (r *BirdcageRoom) birdTick(rng *rand.Rand) {
	if !r.Cage.Broken || r.Out != Playing {
		return
	}
	r.birdStep(rng)
	if r.Bird.steps%config.BirdExtraEvery == 0 {
		r.birdStep(rng)
	}
}

func (r *BirdcageRoom) birdRect() physics.Rect {
	return physics.Rect{X: r.Bird.X, Y: r.Bird.Y, W: config.BirdW, H: config.BirdH}
}

func (r *BirdcageRoom) checkBird() {
	if r.Cage.Broken && !r.Freddy.Gone && r.birdRect().Overlaps(r.Freddy.HitBox()) {
		r.end(Cleared)
	}
}

func (r *BirdcageRoom) birdStep(rng *rand.Rand) {
	bd, f := &r.Bird, r.Freddy
	bd.steps++
	r.checkBird()
	if r.birdRect().Overlaps(r.Broom.Rect()) {
		bd.VY = -1
		return
	}

	steer := true
	if rng.IntN(256) < config.BirdCheckChance {
		ax, dx := 1, bd.X-f.X
		if dx < 0 {
			ax, dx = -1, -dx
		}
		ay, dy := 1, bd.Y-f.Y
		if dy < 0 {
			ay, dy = -1, -dy
		}
		if dx+2*dy <= config.BirdFlee[r.Difficulty] {
			bd.waypoint = 0xff
			bd.VX, bd.VY = ax, ay
			steer = false
		}
	}
	if steer {
		r.birdSteer(rng)
	}

	if bd.Y += bd.VY * config.BirdStepY; bd.Y > config.BirdMaxY {
		bd.Y, bd.VY = config.BirdMaxY, -1
	} else if bd.Y < config.BirdMinY {
		bd.Y, bd.VY = config.BirdMinY, 1
	}
	if bd.X += bd.VX * config.BirdStepX; bd.X > config.BirdMaxX {
		bd.X, bd.VX = config.BirdMaxX, -1
	} else if bd.X < config.BirdMinX {
		bd.X, bd.VX = config.BirdMinX, 1
	}
	bd.Frame++
	r.checkBird()
}

// birdSteer: fly to a waypoint, hover there a while, or wander.
func (r *BirdcageRoom) birdSteer(rng *rand.Rand) {
	bd := &r.Bird
	if bd.waypoint < len(config.BirdWaypoints) {
		if rng.IntN(256) >= config.BirdLeaveChance {
			wp := config.BirdWaypoints[bd.waypoint]
			bd.VX = sign(wp[0] - bd.X&^3)
			bd.VY = sign(wp[1] - bd.Y&^1)
			if bd.VX != 0 || bd.VY != 0 || rng.IntN(256) >= config.BirdBoredChance {
				return
			}
		}
	}
	if rng.IntN(256) < config.BirdCheckChance {
		bd.VX, bd.VY = entity.RandDir(rng), entity.RandDir(rng)
	}
	bd.waypoint = rng.IntN(config.BirdWaypointRoll)
}

func sign(v int) int {
	switch {
	case v > 0:
		return 1
	case v < 0:
		return -1
	}
	return 0
}
