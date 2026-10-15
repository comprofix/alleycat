package rooms

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// HeartCat is a cat in the heart room: rivals in slots 0–5, one per row
// from the bottom, and Felicia in slot 6 (DESIGN §3.5).
type HeartCat struct {
	X, Y  int
	Dir   int // −1, 0 (stopped) or 1
	Bump  int // steps left backing off after bumping Freddy
	Frame int
	// Away is set while a rival is off with a gift (or, for the bottom
	// cat, before it first comes in); it returns once its delay has passed
	// since AwaySince.
	Away      bool
	AwaySince int
}

// Arrow is Cupid's arrow: one at a time, fired again as soon as it goes.
type Arrow struct {
	Active   bool
	X, Y     int
	Dir      int
	Phase    int // grows by 4 a tick to 160; its frame is Phase ÷ 32
	lastCell int
}

// Gift is a gift box: Y is the row line it sits on.
type Gift struct {
	X, Y int
	On   bool // lying on a row (not carried, not taken)
}

// HeartRoom: climb the hearts to Felicia (DESIGN §3.5).
type HeartRoom struct {
	Base
	// Broken marks the broken hearts, by row (bottom first) and column.
	Broken [len(config.HeartRows)][config.HeartCells]bool
	Cats   [config.HeartCatSlots]HeartCat
	Arrow  Arrow
	Gifts  [config.GiftMax]Gift
	// Carrying is the gift Freddy carries, or −1.
	Carrying int
	// WithGift is set when he reaches Felicia carrying a gift: it doubles
	// the kiss multiplier.
	WithGift bool

	justDropped int
	catSlot     int
}

func newHeart(difficulty, roomsCleared int, rng *rand.Rand) *HeartRoom {
	l := config.RoomLayout{StartX: config.HeartStartX, StartY: config.HeartStartY}
	r := &HeartRoom{Base: newBase(l, difficulty, false, rng), Carrying: -1, justDropped: -1}
	f := r.Freddy
	f.MinX, f.MaxX = config.HeartMinX, config.HeartMaxX
	f.EnterRoom(config.HeartStartX, config.HeartStartY, 0) // keeps the limits
	f.NoFloor = true

	for row := 1; row < len(r.Broken); row++ {
		for col := range config.HeartCols {
			r.Broken[row][col] = rng.IntN(2) == 0
		}
	}
	for i := range r.Cats {
		r.Cats[i] = HeartCat{
			X: config.CatStartX + rng.IntN(config.CatStartRoll),
			Y: config.HeartRows[i] + config.HeartCatDY,
		}
	}
	r.Cats[0].Away = true // the bottom rival comes in later
	for i := range min(max(roomsCleared, 1), config.GiftMax) {
		r.Gifts[i] = Gift{X: config.GiftX[i], Y: config.HeartRows[0], On: true}
	}
	return r
}

// Kind implements Room.
func (r *HeartRoom) Kind() Kind { return Heart }

// Update implements Room.
func (r *HeartRoom) Update(ticks, freddySteps int, in entity.Controls, rng *rand.Rand) {
	for range ticks {
		if r.Out != Playing {
			return
		}
		r.Ticks++
		r.arrowTick(rng)
		r.catsTick(rng)
		r.giftTick(in)
	}
	for range freddySteps {
		if r.Out != Playing {
			return
		}
		f := r.Freddy
		f.Update(in, r)
		if f.Y >= config.HeartFallOffY {
			r.end(Failed)
		}
	}
}

// RowAt is the row whose line is line, or −1.
func RowAt(line int) int {
	for i, y := range config.HeartRows {
		if y == line {
			return i
		}
	}
	return -1
}

// Support implements entity.World: Freddy stands on a row only on an
// intact heart.
func (r *HeartRoom) Support(f *entity.Freddy) bool {
	if f.Y < config.HeartStandDY {
		return false
	}
	line := (f.Y - config.HeartStandDY) &^ 7
	row := RowAt(line)
	if row < 0 {
		return false
	}
	col := min(max((f.X+config.HeartStandDX)/16-2, 0), config.HeartCells-1)
	if r.Broken[row][col] {
		return false
	}
	f.Y = line + config.HeartStandDY
	return true
}

// ArrowRect is the arrow's box.
func (r *HeartRoom) ArrowRect() physics.Rect {
	return physics.Rect{X: r.Arrow.X, Y: r.Arrow.Y, W: config.ArrowW, H: config.ArrowH}
}

// arrowHits knocks Freddy down if the arrow touches him, and reports it.
func (r *HeartRoom) arrowHits() bool {
	if !r.Arrow.Active || !r.ArrowRect().Overlaps(r.Freddy.HitBox()) {
		return false
	}
	r.Freddy.Knock(config.ArrowHitSpeed, config.ArrowHitStep, config.ArrowHitNoLand)
	r.Arrow.Active = false
	return true
}

// arrowTick runs the arrow once a tick: a new one at once if there is
// none, then 5 px across and 2 down, flipping the hearts it passes.
func (r *HeartRoom) arrowTick(rng *rand.Rand) {
	a := &r.Arrow
	if r.arrowHits() {
		return
	}
	if !a.Active {
		r.fireArrow(rng)
	}
	if a.Phase < config.ArrowGrowTicks*4 {
		a.Phase += 4
	}
	if a.Y += config.ArrowStepY; a.Y > config.ArrowMaxY {
		a.Active = false
		return
	}
	if a.X += a.Dir * config.ArrowStepX; a.X < 0 || a.X >= config.ArrowMaxX {
		a.Active = false
		return
	}
	if r.arrowHits() {
		return
	}
	r.flipHeart()
}

// fireArrow launches a new arrow from a side (16 in 26) or the top.
func (r *HeartRoom) fireArrow(rng *rand.Rand) {
	a := &r.Arrow
	v := rng.IntN(config.ArrowSpawnRoll)
	*a = Arrow{Active: true, lastCell: -1}
	if v < config.ArrowSides {
		a.X, a.Dir = config.ArrowLeftX, 1
		if v&8 != 0 {
			a.X, a.Dir = config.ArrowRightX, -1
		}
		a.Y = config.ArrowSideY[v&7] + config.ArrowSideDY
		return
	}
	v -= config.ArrowSides
	a.Dir = 1
	if v >= config.ArrowTopRight {
		a.Dir = -1
	}
	a.X, a.Y = config.ArrowTopX[v]+config.ArrowTopDX, config.ArrowTopY
}

// flipHeart flips the heart under the arrow, once per heart it passes.
func (r *HeartRoom) flipHeart() {
	a := &r.Arrow
	row := RowAt((a.Y - config.ArrowCellDY) &^ 7)
	col := a.X/16 - config.ArrowCellOffsetX
	if row < 0 || col < 0 || col >= config.ArrowCellCols {
		return
	}
	if cell := row*config.HeartCells + col; cell != a.lastCell {
		a.lastCell = cell
		r.Broken[row][col] = !r.Broken[row][col]
	}
}

// ArrowFrame is the arrow's frame as it grows in, 0–5.
func (a *Arrow) Frame() int { return min(a.Phase/32, config.ArrowFrames-1) }

// CatRect is cat i's box.
func (r *HeartRoom) CatRect(i int) physics.Rect {
	c := &r.Cats[i]
	return physics.Rect{X: c.X, Y: c.Y, W: config.CatW, H: config.CatH}
}

// catsTick moves the cats in three groups, one group a tick: 1; 2, 3 and
// 4; 5, 6 and 0.
func (r *HeartRoom) catsTick(rng *rand.Rand) {
	for {
		if r.catSlot++; r.catSlot >= config.HeartCatSlots {
			r.catSlot = 0
		}
		s := r.catSlot
		r.catStep(s, rng)
		if s == 0 || s == 1 || s == 4 || r.Out != Playing {
			return
		}
	}
}

func (r *HeartRoom) underArrow(i int) bool {
	return r.Arrow.Active && r.ArrowRect().Overlaps(r.CatRect(i))
}

// catStep runs cat i once: frozen under an arrow; a lured rival waits to
// come back; then it backs off, chases or wanders, and may take a gift.
func (r *HeartRoom) catStep(i int, rng *rand.Rand) {
	c, f := &r.Cats[i], r.Freddy
	if r.underArrow(i) {
		return
	}
	if c.Away {
		delay := r.table.RivalHideTicks
		if i == 0 {
			delay *= config.BottomCatLureTimes
		}
		if r.Ticks-c.AwaySince < delay {
			return
		}
		c.Away, c.Dir = false, 0
		c.X = config.CatReturnRightX
		if f.X > config.CatReturnSplitX {
			c.X = config.CatReturnLeftX
		}
	}
	if r.catTouch(i) {
		return
	}
	saved := *c
	switch {
	case c.Bump > 0:
		if c.Bump--; c.Bump == 0 {
			c.Dir = -c.Dir
		}
	case c.Y <= f.Y:
		r.chase(i, rng)
	}
	r.catMove(c, rng)
	if r.underArrow(i) {
		*c = saved // a step under the arrow is lost
		return
	}
	if !r.catTouch(i) {
		r.lure(i)
	}
}

// chase: a cat level with or above Freddy may turn towards him; Felicia
// turns away until he is near the top, then always comes to him.
func (r *HeartRoom) chase(i int, rng *rand.Rand) {
	c, f := &r.Cats[i], r.Freddy
	comes := i == config.Felicia && f.Y < config.FeliciaComeY
	if !comes && rng.IntN(256) > r.table.RivalChaseChance {
		return
	}
	dir := 0
	if cx, fx := c.X&^(config.CatAlign-1), f.X&^(config.CatAlign-1); cx < fx {
		dir = 1
	} else if cx > fx {
		dir = -1
	}
	c.Dir = dir
	if i == config.Felicia && !comes {
		c.Dir = -1
		if dir == -1 {
			c.Dir = 1
		}
	}
}

// catMove steps a cat along its row, turning at the ends; a stopped cat
// may set off, a walking one may stop.
func (r *HeartRoom) catMove(c *HeartCat, rng *rand.Rand) {
	if c.Dir == 0 {
		if rng.IntN(256) < config.CatStartChance {
			c.Dir = entity.RandDir(rng)
		}
		return
	}
	step := config.RivalCatStepPx
	if c.Bump > 0 {
		step = config.CatBumpStep
	}
	if c.X += c.Dir * step; c.X > config.CatMaxX {
		c.X, c.Dir, c.Bump = config.CatMaxX, -1, 0
	} else if c.X < config.CatMinX {
		c.X, c.Dir, c.Bump = config.CatMinX, 1, 0
	}
	c.Frame = (c.Frame + 1) % config.CatFrames
	if c.Bump == 0 && rng.IntN(256) < config.CatStopChance {
		c.Dir = 0
	}
}

// catTouch: touching Felicia clears the room; touching a rival knocks
// Freddy down and sends it backing off.
func (r *HeartRoom) catTouch(i int) bool {
	c, f := &r.Cats[i], r.Freddy
	if f.Gone || !r.CatRect(i).Overlaps(f.HitBox()) {
		return false
	}
	if i == config.Felicia {
		r.WithGift = r.Carrying >= 0
		f.Gone = true
		r.end(Cleared)
		return true
	}
	f.Knock(config.CatKnockSpeed, config.CatKnockStep, config.CatKnockNoLand)
	c.Bump = config.CatBumpSteps
	c.Dir = -1
	if c.X > f.X {
		c.Dir = 1
	}
	return true
}

// lure: a cat that walks into a gift takes it; a rival goes off with it.
func (r *HeartRoom) lure(i int) {
	cat := r.CatRect(i)
	for g := len(r.Gifts) - 1; g >= 0; g-- {
		if !r.Gifts[g].On || !cat.Overlaps(r.GiftRect(g)) {
			continue
		}
		r.Gifts[g].On = false
		if i != config.Felicia {
			r.Cats[i].Away, r.Cats[i].AwaySince = true, r.Ticks
		}
		return
	}
}

// GiftRect is gift g's box.
func (r *HeartRoom) GiftRect(g int) physics.Rect {
	return physics.Rect{X: r.Gifts[g].X, Y: r.Gifts[g].Y, W: config.GiftW, H: config.GiftH}
}

// giftTick picks up a gift Freddy touches, or with Alt puts the one he
// carries down on the nearest row.
func (r *HeartRoom) giftTick(in entity.Controls) {
	if r.Carrying >= 0 {
		if in.Act {
			r.putDown()
		}
		return
	}
	for g := len(r.Gifts) - 1; g >= 0; g-- {
		if !r.Gifts[g].On || !r.GiftRect(g).Overlaps(r.Freddy.HitBox()) {
			continue
		}
		if g != r.justDropped {
			r.Gifts[g].On = false
			r.Carrying = g
		}
		return
	}
	r.justDropped = -1
}

// putDown puts the carried gift on the row line nearest Freddy, unless it
// would overlap another gift.
func (r *HeartRoom) putDown() {
	f := r.Freddy
	row, best := 0, 1<<30
	for i := len(config.HeartRows) - 1; i >= 0; i-- {
		if d := abs(f.Y - config.HeartRows[i]); d <= best {
			row, best = i, d
		}
	}
	g := r.Carrying
	place := Gift{X: min(f.X, config.GiftMaxX) &^ 3, Y: config.HeartRows[row], On: true}
	box := physics.Rect{X: place.X, Y: place.Y, W: config.GiftW, H: config.GiftH}
	for o := range r.Gifts {
		if o != g && r.Gifts[o].On && box.Overlaps(r.GiftRect(o)) {
			return
		}
	}
	r.Gifts[g] = place
	r.Carrying, r.justDropped = -1, g
}
