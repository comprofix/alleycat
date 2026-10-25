package gfx

import "github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"

// Scheme is a scene's colour scheme: named roles, each a Resurrect 64 index
// (DESIGN §2.3). Scenes keep a distinct mood within the one palette.
type Scheme struct {
	Wall        int // the main backdrop
	WallShade   int // detail on the backdrop
	WallLight   int // lit edges on the backdrop
	Glass       int // closed windows
	Glint       int // highlights on glass
	Trim        int // window frames, lines
	Fence       int // foreground boards
	FenceLight  int // their lit edges
	Gap         int // gaps and shade in the foreground
	Ground      int
	GroundLight int // lit edges on the ground
	Metal       int // trash cans
	Shade       int // shading on props
	Outline     int
	Cheese      int // the cheese room's wedge
	Hole        int // its holes and pits
	Water       int // the underwater scene's water
	Prints      int // Freddy's footprints on a room's floor, darker than Ground
}

// AlleyScheme is the alley at night (DESIGN §8): a cool magenta-plum
// building above a teal fence, nodding to CGA's magenta/cyan set (§2.3).
// Each surface is shaded along its own ramp: plum brick, teal boards, grey
// cobbles.
var AlleyScheme = Scheme{
	Wall:        resurrect64.Grape,
	WallShade:   resurrect64.Plum,
	WallLight:   resurrect64.Violet,
	Glass:       resurrect64.Navy,
	Glint:       resurrect64.Blue,
	Trim:        resurrect64.Fog,
	Fence:       resurrect64.DeepTeal,
	FenceLight:  resurrect64.Teal,
	Gap:         resurrect64.Shadow,
	Ground:      resurrect64.Shadow,
	GroundLight: resurrect64.Dusk,
	Metal:       resurrect64.Fog,
	Shade:       resurrect64.Lilac,
	Outline:     resurrect64.Ink,
}

// BirdcageScheme is the birdcage room: a red room with green furniture,
// after the original's red/green set (DESIGN §2.3, §3.4).
var BirdcageScheme = Scheme{
	Wall:        resurrect64.Cerise,
	WallShade:   resurrect64.Grape,
	WallLight:   resurrect64.Cerise, // unlit until the red rooms are shaded
	Glass:       resurrect64.Ink,
	Glint:       resurrect64.Blue,
	Trim:        resurrect64.Honey,
	Fence:       resurrect64.Emerald,
	Gap:         resurrect64.DeepTeal,
	Ground:      resurrect64.Cerise,
	GroundLight: resurrect64.Cerise,
	Metal:       resurrect64.Fog,
	Shade:       resurrect64.Grape,
	Outline:     resurrect64.Ink,
	Prints:      resurrect64.Grape,
}

// CheeseScheme is the cheese room: the birdcage room's red walls and green
// furniture, with a yellow wedge and orange holes (DESIGN §3.4, "Cheese room
// layout").
var CheeseScheme = Scheme{
	Wall:        resurrect64.Cerise,
	WallShade:   resurrect64.Grape,
	WallLight:   resurrect64.Cerise, // unlit until the red rooms are shaded
	Glass:       resurrect64.Ink,
	Glint:       resurrect64.Blue,
	Trim:        resurrect64.Honey,
	Fence:       resurrect64.Emerald,
	Gap:         resurrect64.DeepTeal,
	Ground:      resurrect64.Cerise,
	GroundLight: resurrect64.Cerise,
	Metal:       resurrect64.Fog,
	Shade:       resurrect64.Grape,
	Outline:     resurrect64.Ink,
	Prints:      resurrect64.Grape,
	Cheese:      resurrect64.Honey,
	Hole:        resurrect64.Coral,
}

// FishbowlScheme is the fishbowl room: a plum room with teal and cyan
// furniture, after the original's magenta/cyan set (DESIGN §3.4,
// "Fishbowl room layout").
var FishbowlScheme = Scheme{
	Wall:        resurrect64.Grape,
	WallShade:   resurrect64.Plum,
	WallLight:   resurrect64.Violet,
	Glass:       resurrect64.Ink,
	Glint:       resurrect64.Blue,
	Trim:        resurrect64.Seafoam,
	Fence:       resurrect64.DeepTeal,
	Gap:         resurrect64.Seafoam,
	Ground:      resurrect64.Grape,
	GroundLight: resurrect64.Violet,
	Metal:       resurrect64.Fog,
	Shade:       resurrect64.Shadow,
	Outline:     resurrect64.Ink,
	Prints:      resurrect64.Ink, // Grape prints vanished on the Grape floor
}

// UnderwaterScheme is the underwater scene: teal water under a plum
// surface band (DESIGN §3.4, "Underwater layout").
var UnderwaterScheme = Scheme{
	Wall:      resurrect64.DeepTeal,
	WallShade: resurrect64.Blue, // ripple marks
	Trim:      resurrect64.Grape,
	Shade:     resurrect64.Shadow,
	Outline:   resurrect64.Ink,
	Water:     resurrect64.DeepTeal,
}

// AirColours are Freddy's and the eels' colours underwater by air warning
// stage 0–3 (DESIGN §3.4, "Underwater": the remake's recolouring).
var AirColours = [4]int{resurrect64.Ink, resurrect64.Navy, resurrect64.Grape, resurrect64.Cerise}

// DogsScheme is the sleeping dogs room: the fishbowl room's plum walls and
// teal furniture (DESIGN §3.4, "Sleeping dogs room layout").
var DogsScheme = FishbowlScheme

// BookshelfScheme is the bookshelf room: the other magenta rooms' plum walls
// and teal furniture (DESIGN §3.4, "Bookshelf room layout").
var BookshelfScheme = FishbowlScheme

// HeartScheme is the heart room: a green backdrop, after the original's
// red/green set (DESIGN §3.5, "Heart room layout"). The hearts, cats and
// cherubs carry their own colours.
var HeartScheme = Scheme{
	Wall:    resurrect64.Emerald,
	Outline: resurrect64.Ink,
}

// KissScheme is the heart room at the kiss: the cyan/magenta set, with
// teal ground, plum hearts and silver cats (DESIGN §3.5, "The kiss").
var KissScheme = Scheme{
	Wall:    resurrect64.DeepTeal,
	Shade:   resurrect64.Grape, // intact hearts
	Gap:     resurrect64.Ink,   // broken hearts
	Metal:   resurrect64.Fog,   // the rivals
	Outline: resurrect64.Ink,
}
