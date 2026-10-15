package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Outcome is how a room is going.
type Outcome int

const (
	Playing   Outcome = iota
	Cleared           // the objective is done: the tally follows
	LeftEarly         // Freddy went out of the window: "OOPS!"
	Failed            // a life is lost
)

// Room is a room behind an alley window (DESIGN §3.4). Its logic runs
// headless.
type Room interface {
	Kind() Kind
	// Update advances the room by this frame's master ticks and Freddy
	// steps.
	Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand)
	// Shared is the shared part: Freddy, the broom, the dog and the outcome.
	Shared() *Base
}

// Base is what every room shares (DESIGN §3.4, "Room framework in detail"):
// Freddy, the broom, the room dog, the surfaces and the window.
type Base struct {
	Layout     config.RoomLayout
	Freddy     *entity.Freddy
	Broom      entity.Broom
	Dog        entity.Dog
	Difficulty int
	// Ticks counts master ticks since Freddy came in, for the time bonus.
	Ticks int
	// Out is how the room is going.
	Out Outcome
	// ByDog is set when a dog caused the failure ("HERE KITTY").
	ByDog bool

	table   config.Difficulty
	hasDog  bool
	stopped bool
}

func newBase(l config.RoomLayout, difficulty int, hasDog bool, rng *rand.Rand) Base {
	f := entity.NewFreddy()
	f.EnterRoom(l.StartX, l.StartY, config.RoomEnterNoLand)
	return Base{
		Layout:     l,
		Freddy:     f,
		Broom:      entity.NewBroom(l.BroomEvery, rng),
		Difficulty: difficulty,
		table:      config.ForLevel(difficulty),
		hasDog:     hasDog,
	}
}

// Shared implements Room.
func (b *Base) Shared() *Base { return b }

// HasDog reports whether the room dog runs in this room.
func (b *Base) HasDog() bool { return b.hasDog }

// ExitRect is the box Freddy falls into to leave by the window.
func (b *Base) ExitRect() physics.Rect {
	return physics.Rect{
		X: b.Layout.WindowX - config.ExitDX, Y: b.Layout.WindowY - config.ExitDY,
		W: config.ExitW, H: config.ExitH,
	}
}

// Support implements entity.World: the window, then the room's surfaces.
func (b *Base) Support(f *entity.Freddy) bool {
	if f.Arc.Dir == physics.Falling && b.ExitRect().Overlaps(f.HitBox()) {
		b.end(LeftEarly)
		return false
	}
	if f.Y < 0 {
		return false
	}
	band := physics.Band(f.Y)
	for _, s := range b.Layout.Surfaces {
		if band != s.Band || physics.Band(f.X) < s.X || max(f.X-s.Reach, 0)&^3 > s.X {
			continue
		}
		f.Y = s.Band
		if s.Perch {
			f.Seat = entity.JustSeated
			f.X &^= 3
		}
		return true
	}
	return false
}

func (b *Base) end(o Outcome) {
	if b.Out == Playing {
		b.Out = o
	}
}

// tick runs the broom and the dog for one master tick.
func (b *Base) tick(rng *rand.Rand) {
	b.Ticks++
	l := b.Layout
	b.Broom.Tick(b.Freddy, &b.Dog, b.Ticks, l.WindowX, l.WindowY, rng)
	if !b.hasDog {
		return
	}
	switch b.Dog.Tick(b.Freddy, b.table, rng) {
	case entity.DogCaught:
		b.Freddy.Gone = true
	case entity.DogFightOver:
		b.ByDog = true
		b.end(Failed)
	}
}

// step runs one Freddy update.
func (b *Base) step(in entity.Controls) {
	f := b.Freddy
	f.Update(in, b)
	b.Broom.Footprint(f)
	if b.hasDog && b.Dog.CheckCatch(f) {
		f.Gone = true
	}
}

// update is the plain room loop: ticks for the broom and dog, steps for
// Freddy, stopping as soon as the room is over.
func (b *Base) update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	for range ticks {
		if b.Out != Playing {
			return
		}
		b.tick(rng)
	}
	for range freddySteps {
		if b.Out != Playing {
			return
		}
		b.step(in)
	}
}

// baseBonus is each room's base bonus (DESIGN §3.6).
var baseBonus = map[Kind]int{
	Fishbowl: 3000, Bookshelf: 2000, Cheese: 1500, Birdcage: 3500, Dogs: 2500,
}

// Bonus is the room bonus for clearing a room after t ticks: the base plus
// the time bonus (DESIGN §3.6).
func Bonus(k Kind, t int) int {
	if k == Dogs {
		return baseBonus[k] + max(0, config.TallyDogsTimeWindow-t)
	}
	return baseBonus[k] + 2*max(0, config.TallyTimeWindow-t)
}

// New returns the room of kind k, or nil if it is not built yet.
func New(k Kind, difficulty int, rng *rand.Rand) Room {
	switch k {
	case Birdcage:
		return newBirdcage(difficulty, rng)
	}
	return nil
}

// TimeBonus is the time part of the room bonus.
func TimeBonus(k Kind, t int) int { return Bonus(k, t) - baseBonus[k] }
