package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Book is what one cell of the bookcase shows.
type Book int

const (
	BookcaseLeft  Book = iota // the left side panel
	BookShelf                 // an empty stretch of shelf
	BookcaseRight             // the right side panel
	BookUpright               // the book kinds: upright ...
	BookPair                  // ... two thin ones ...
	BookLeaning               // ... leaning, always followed by another book ...
	BookStack                 // ... and a short stack
)

// Spider is the bookshelf room's spider (DESIGN §3.4, "Bookshelf room in
// detail"). State is 0 at the top tracking Freddy, 1 dropping, −1 climbing.
type Spider struct {
	X, Y  int
	State int
	Frame int

	steps     int // drop and climb steps, for the changing speed
	frameWait int
	wait      int
}

// BookshelfRoom: climb the bookcase and touch the 3 ferns on top without
// the spider biting.
type BookshelfRoom struct {
	Base
	Ferns     [len(config.FernX)]bool // taken
	FernsLeft int
	Spider    Spider
	// Bite counts down the ticks of "BYTE!" before the room fails.
	Bite         int
	ByteX, ByteY int
	// Books is the bookcase's random mix of books, picked on entering.
	Books [config.BookRows][config.BookCols]Book
}

func newBookshelf(difficulty int, rng *rand.Rand) *BookshelfRoom {
	r := &BookshelfRoom{
		Base:      newBase(config.Bookshelf, difficulty, false, rng),
		FernsLeft: len(config.FernX),
		Spider:    Spider{X: config.SpiderStartX, Y: config.SpiderTopY, frameWait: config.SpiderFrameEvery},
	}
	r.fillBooks(rng)
	return r
}

// fillBooks picks the bookcase's books: side panels at the ends, full rows
// of books alternating with sparse rows that are mostly bare shelf.
func (r *BookshelfRoom) fillBooks(rng *rand.Rand) {
	for row := range r.Books {
		cells := &r.Books[row]
		for col := range cells {
			switch {
			case col == 0:
				cells[col] = BookcaseLeft
			case col == len(cells)-1:
				cells[col] = BookcaseRight
			case cells[col-1] != BookLeaning && row%2 == 0 && rng.IntN(4) != 0:
				cells[col] = BookShelf
			default:
				cells[col] = BookUpright + Book(rng.IntN(4))
			}
		}
	}
}

// Kind implements Room.
func (r *BookshelfRoom) Kind() Kind { return Bookshelf }

// Update implements Room.
func (r *BookshelfRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	for range ticks {
		if r.Out != Playing {
			return
		}
		if r.Bite > 0 {
			r.Ticks++
			if r.Bite--; r.Bite == 0 {
				r.end(Failed)
			}
			continue
		}
		r.tick(rng)
		r.fernTick()
		r.spiderTick()
	}
	for range freddySteps {
		if r.Out != Playing || r.Bite > 0 {
			return
		}
		r.step(in)
	}
}

// FernRect is fern i's box.
func FernRect(i int) physics.Rect {
	return physics.Rect{X: config.FernX[i], Y: config.FernY, W: config.FernW, H: config.FernH}
}

// fernTick takes a fern Freddy touches, one a tick; the last clears the
// room.
func (r *BookshelfRoom) fernTick() {
	for i := len(r.Ferns) - 1; i >= 0; i-- {
		if r.Ferns[i] || !FernRect(i).Overlaps(r.Freddy.HitBox()) {
			continue
		}
		r.Ferns[i] = true
		if r.FernsLeft--; r.FernsLeft == 0 {
			r.end(Cleared)
		}
		return
	}
}

// SpiderRect is the spider's box.
func (r *BookshelfRoom) SpiderRect() physics.Rect {
	return physics.Rect{X: r.Spider.X, Y: r.Spider.Y, W: config.SpiderW, H: config.SpiderH}
}

func (r *BookshelfRoom) spiderTick() {
	s := &r.Spider
	if s.wait++; s.wait < config.SpiderEvery {
		return
	}
	s.wait = 0
	r.spiderStep()
}

// spiderStep moves the spider once: frozen by the broom; along the top
// towards Freddy until it lines up with him, then down on him, and back up
// when he gets away.
func (r *BookshelfRoom) spiderStep() {
	s, f := &r.Spider, r.Freddy
	broom := r.Broom.Rect()
	if r.SpiderRect().Overlaps(broom) {
		return
	}
	if r.SpiderRect().Overlaps(f.HitBox()) {
		r.bite()
		return
	}
	ox, oy := s.X, s.Y
	if s.Y == config.SpiderTopY {
		if sx, fx := s.X&^(config.SpiderAlign-1), f.X&^(config.SpiderAlign-1); sx != fx {
			if sx < fx {
				s.X += config.SpiderSpeed[r.Difficulty]
			} else {
				s.X = max(s.X-config.SpiderSpeed[r.Difficulty], 0)
			}
			r.spiderMoved(ox, oy)
			return
		}
		s.State, s.steps = 1, 1
	}
	s.steps++
	step := (s.steps/config.SpiderDropEvery)%4 + config.SpiderDropMin
	if s.State == 1 {
		y := s.Y + step
		switch {
		case y > f.Y || abs(s.X-f.X) > config.SpiderReach:
			s.State = -1
		case y > config.SpiderMaxY:
			y, s.State = config.SpiderMaxY, -1
		}
		s.Y = y
	} else if s.Y -= step; s.Y <= config.SpiderTopY {
		s.Y, s.State = config.SpiderTopY, 0
	}
	r.spiderMoved(ox, oy)
}

// spiderMoved undoes a move into the broom, bites Freddy if the spider has
// reached him, and otherwise runs its animation.
func (r *BookshelfRoom) spiderMoved(ox, oy int) {
	s := &r.Spider
	if r.SpiderRect().Overlaps(r.Broom.Rect()) {
		s.X, s.Y = ox, oy
		return
	}
	if r.SpiderRect().Overlaps(r.Freddy.HitBox()) {
		r.bite()
		return
	}
	if s.frameWait--; s.frameWait == 0 {
		s.frameWait = config.SpiderFrameEvery
		s.Frame ^= 1
	}
}

// bite: "BYTE!" by Freddy, then the room fails.
func (r *BookshelfRoom) bite() {
	f := r.Freddy
	r.Bite = config.ByteTicks
	r.ByteX = min(max(f.X-config.ByteDX, 0), config.ByteMaxX)
	r.ByteY = max(f.Y-config.ByteDY, 0)
}
