package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
)

// Laundry is what hangs in one 8-px half of a laundry piece.
type Laundry uint8

const (
	NoLaundry  Laundry = iota
	BigLeft            // a big item starts here and covers the right half
	BigRight           // covered by the big item to its left
	MediumLeft         // likewise for a medium item
	MediumRight
	Small // Small+0..2: the three small kinds
)

// Piece is a 16-px stretch of laundry: its left and right halves.
type Piece [2]Laundry

// linePieces covers the screen width at any scroll offset.
const linePieces = 320/config.LinePiecePx + 1

// Line is one clothesline (DESIGN §3.8, "Clotheslines"). Pieces[0] starts
// at X = Offset, which stays between −16 and 0.
type Line struct {
	Row    int // 0 top, 1 middle, 2 bottom
	Dir    int // 1 right, −1 left
	Pieces [linePieces]Piece
	Offset int
}

func newLine(row, gap int, rng *rand.Rand) Line {
	l := Line{Row: row, Dir: config.LineDir[row]}
	for i := range l.Pieces {
		l.Pieces[i] = newPiece(row, gap, rng)
	}
	// Start at a piece boundary: the next step brings in a new piece.
	if l.Dir < 0 {
		l.Offset = config.LineStepPx - config.LinePiecePx
	}
	return l
}

// newPiece generates a piece of laundry (DESIGN §3.8, "New laundry").
func newPiece(row, gap int, rng *rand.Rand) Piece {
	if rng.IntN(256) < gap || rng.IntN(256) <= config.LineFillAbove[row] {
		return Piece{}
	}
	switch r := rng.IntN(256); {
	case r < config.LaundryBigChance:
		return Piece{BigLeft, BigRight}
	case r < config.LaundryBigChance+config.LaundryMediumChance:
		return Piece{MediumLeft, MediumRight}
	}
	var p Piece
	for i := range p {
		if k := rng.IntN(config.LaundrySmallKinds + 1); k < config.LaundrySmallKinds {
			p[i] = Small + Laundry(k)
		}
	}
	return p
}

// AtBoundary reports whether the line sits at a piece boundary, which is
// where a resting line always stops.
func (l *Line) AtBoundary() bool {
	if l.Dir > 0 {
		return l.Offset == 0
	}
	return l.Offset == config.LineStepPx-config.LinePiecePx
}

// step scrolls the line by one step, bringing in a new piece at a boundary.
func (l *Line) step(gap int, rng *rand.Rand) {
	n := len(l.Pieces)
	l.Offset += l.Dir * config.LineStepPx
	switch {
	case l.Offset > 0:
		copy(l.Pieces[1:], l.Pieces[:n-1])
		l.Pieces[0] = newPiece(l.Row, gap, rng)
		l.Offset -= config.LinePiecePx
	case l.Offset <= -config.LinePiecePx:
		copy(l.Pieces[:n-1], l.Pieces[1:])
		l.Pieces[n-1] = newPiece(l.Row, gap, rng)
		l.Offset += config.LinePiecePx
	}
}

// Supports reports whether laundry hangs under Freddy at X (DESIGN §3.8,
// "Alley surfaces": the test point is X + 10).
func (l *Line) Supports(x int) bool {
	c := x + config.LineProbeX - l.Offset
	if c < 0 || c >= len(l.Pieces)*config.LinePiecePx {
		return false
	}
	half := c % config.LinePiecePx / (config.LinePiecePx / 2)
	return l.Pieces[c/config.LinePiecePx][half] != NoLaundry
}
