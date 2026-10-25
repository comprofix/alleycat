// Package gfx holds the palette, screen geometry and presentation helpers.
package gfx

import (
	"github.com/comprofix/alleycat/internal/gfx/palettes"
	"github.com/comprofix/alleycat/internal/gfx/palettes/sweetie16"
)

// Logical screen size. All game logic and art use this grid (DESIGN §2.1).
const (
	ScreenWidth  = 320
	ScreenHeight = 200
)

// Palette indices for Sweetie 16 (DESIGN §2.2).
const (
	Night    = sweetie16.Night
	Plum     = sweetie16.Plum
	Red      = sweetie16.Red
	Orange   = sweetie16.Orange
	Yellow   = sweetie16.Yellow
	Lime     = sweetie16.Lime
	Green    = sweetie16.Green
	Teal     = sweetie16.Teal
	Navy     = sweetie16.Navy
	Blue     = sweetie16.Blue
	Sky      = sweetie16.Sky
	Cyan     = sweetie16.Cyan
	White    = sweetie16.White
	Silver   = sweetie16.Silver
	Slate    = sweetie16.Slate
	Charcoal = sweetie16.Charcoal
)

// Named is the game's palette with its colour names, which sprite files use
// in their colours line (DESIGN §8). It is the only source of colour in the
// game: art and code reference its entries, and its hex codes are written
// in its own package and nowhere else.
var Named palettes.Palette = sweetie16.Palette

// Palette is the colours of Named, by index.
var Palette = Named.Colors
