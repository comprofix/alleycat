package rooms

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
)

// Bookshelf room layout, measured from the reference for layout only (DESIGN
// §3.4, "Bookshelf room layout"). The art is our own.
const (
	bookOpeningX             = 80
	bookLampLX, bookLampRX   = 16, 136
	bookChairLX, bookChairRX = 48, 96
)

// buildBookshelf draws the bookshelf room's backdrop: the walls, the window
// with its curtains, the lamps, the chairs and the bookcase with this
// visit's books.
func (r *Renderer) buildBookshelf(room *BookshelfRoom) {
	b := r.back
	r.buildWalls(wallY0, wallY1)
	r.buildWindow(bookOpeningX, openingY, "curtain_teal")
	lamp := r.sheet.Sprite("lamp_teal").Frame(0)
	entity.DrawAt(b, lamp, bookLampLX, lampY, false)
	entity.DrawAt(b, lamp, bookLampRX, lampY, false)
	chair := r.sheet.Sprite("chair_teal").Frame(0)
	entity.DrawAt(b, chair, bookChairLX, furnitureY, false)
	entity.DrawAt(b, chair, bookChairRX, furnitureY, true)
	book := r.sheet.Sprite("book")
	for row, cells := range room.Books {
		for col, k := range cells {
			x := config.BookcaseX + col*config.BookCell
			y := config.BookcaseY + row*config.BookCell
			entity.DrawAt(b, book.Frame(int(k)), x, y, false)
		}
	}
}

// drawBookshelf draws the ferns still standing, the spider and "BYTE!".
func (r *Renderer) drawBookshelf(dst *ebiten.Image, room *BookshelfRoom) {
	fern := r.sheet.Sprite("fern").Frame(0)
	for i, taken := range room.Ferns {
		if !taken {
			entity.DrawAt(dst, fern, config.FernX[i], config.FernY, false)
		}
	}
	s := &room.Spider
	entity.DrawAt(dst, r.sheet.Sprite("spider").Frame(s.Frame), s.X, s.Y, false)
}

// drawBite draws "BYTE!" over Freddy while the bite plays.
func (r *Renderer) drawBite(dst *ebiten.Image, room *BookshelfRoom) {
	if room.Bite > 0 {
		entity.DrawAt(dst, r.actors.Sprite("byte").Frame(0), room.ByteX, room.ByteY, false)
	}
}
