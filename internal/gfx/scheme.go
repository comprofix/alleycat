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
