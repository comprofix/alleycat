package gfx

import "github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"

// Scheme gives the menus their colours by role, as the Game Skeleton's.
// Each role is a palette index.
type Scheme struct {
	Backdrop  int // screen background
	Panel     int // menu and dialog panels
	Border    int // panel borders
	Text      int // normal text
	Dim       int // disabled or secondary text
	Highlight int // the selected menu item
	Title     int // titles
	Shadow    int // text shadows and fades
}

// MenuScheme is the menus' colours, after the old skill screen's: a navy
// panel on black, honey headings (DESIGN §2.3).
var MenuScheme = Scheme{
	Backdrop:  resurrect64.Ink,
	Panel:     resurrect64.Navy,
	Border:    resurrect64.Lilac,
	Text:      resurrect64.White,
	Dim:       resurrect64.Lilac,
	Highlight: resurrect64.Honey,
	Title:     resurrect64.Honey,
	Shadow:    resurrect64.Ink,
}
