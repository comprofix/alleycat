// Package gfx holds the palette, screen geometry and presentation helpers.
package gfx

import (
	"image/color"

	"github.com/comprofix/alleycat/internal/gfx/palettes"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Logical screen size. All game logic and art use this grid (DESIGN §2.1).
const (
	ScreenWidth  = 320
	ScreenHeight = 200
)

// Named is the game's palette, Resurrect 64 (DESIGN §2.2), with its colour
// names. Code names colours with the constants of its package, e.g.
// resurrect64.Ink, and sprite files by name in their colours line (DESIGN
// §8). It is the only source of colour in the game: its hex codes are
// written in its own package and nowhere else.
var Named palettes.Palette = resurrect64.Palette

// Palette is the colours of Named, by index.
var Palette = Named.Colors

// RGBA returns colour i of the palette.
func RGBA(i int) color.RGBA { return Palette[i] }
