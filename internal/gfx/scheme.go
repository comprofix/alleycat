package gfx

// Scheme is a scene's colour scheme: named roles, each a Sweetie 16 index
// (DESIGN §2.3). Scenes keep a distinct mood within the one palette.
type Scheme struct {
	Wall      int // the main backdrop
	WallShade int // detail on the backdrop
	Glass     int // closed windows
	Glint     int // highlights on glass
	Trim      int // window frames, lines
	Fence     int // foreground boards
	Gap       int // gaps and shade in the foreground
	Ground    int
	Metal     int // trash cans
	Shade     int // shading on props
	Outline   int
	Cheese    int // the cheese room's wedge
	Hole      int // its holes and pits
	Water     int // the underwater scene's water
}

// AlleyScheme is the alley at night (DESIGN §8): a cool magenta-plum
// building above a teal fence, nodding to CGA's magenta/cyan set (§2.3).
var AlleyScheme = Scheme{
	Wall:      Plum,
	WallShade: Charcoal,
	Glass:     Navy,
	Glint:     Sky,
	Trim:      Silver,
	Fence:     Teal,
	Gap:       Charcoal,
	Ground:    Charcoal,
	Metal:     Silver,
	Shade:     Slate,
	Outline:   Night,
}

// BirdcageScheme is the birdcage room: a red room with green furniture,
// after the original's red/green set (DESIGN §2.3, §3.4).
var BirdcageScheme = Scheme{
	Wall:      Red,
	WallShade: Plum,
	Glass:     Night,
	Glint:     Sky,
	Trim:      Yellow,
	Fence:     Green,
	Gap:       Teal,
	Ground:    Red,
	Metal:     Silver,
	Shade:     Plum,
	Outline:   Night,
}

// CheeseScheme is the cheese room: the birdcage room's red walls and green
// furniture, with a yellow wedge and orange holes (DESIGN §3.4, "Cheese room
// layout").
var CheeseScheme = Scheme{
	Wall:      Red,
	WallShade: Plum,
	Glass:     Night,
	Glint:     Sky,
	Trim:      Yellow,
	Fence:     Green,
	Gap:       Teal,
	Ground:    Red,
	Metal:     Silver,
	Shade:     Plum,
	Outline:   Night,
	Cheese:    Yellow,
	Hole:      Orange,
}

// FishbowlScheme is the fishbowl room: a plum room with teal and cyan
// furniture, after the original's magenta/cyan set (DESIGN §3.4,
// "Fishbowl room layout").
var FishbowlScheme = Scheme{
	Wall:      Plum,
	WallShade: Charcoal,
	Glass:     Night,
	Glint:     Sky,
	Trim:      Cyan,
	Fence:     Teal,
	Gap:       Cyan,
	Ground:    Plum,
	Metal:     Silver,
	Shade:     Charcoal,
	Outline:   Night,
}

// UnderwaterScheme is the underwater scene: teal water under a plum
// surface band (DESIGN §3.4, "Underwater layout").
var UnderwaterScheme = Scheme{
	Wall:      Teal,
	WallShade: Sky, // ripple marks
	Trim:      Plum,
	Shade:     Charcoal,
	Outline:   Night,
	Water:     Teal,
}

// AirColours are Freddy's and the eels' colours underwater by air warning
// stage 0–3 (DESIGN §3.4, "Underwater": the remake's recolouring).
var AirColours = [4]int{Night, Navy, Plum, Red}
