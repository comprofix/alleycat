package config

import (
	"github.com/comprofix/alleycat/internal/gfx/palettes"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// Title is the game's name, shown in the window title and on the menus.
const Title = "Alley Cat"

// TPS is the game loop rate. The master clock runs off it (DESIGN §3.8).
const TPS = 60

// ScaleMode is how the internal screen is shown in the window (DESIGN
// §2.1). The values are what settings.json stores.
type ScaleMode string

const (
	// ScaleInteger is the largest whole-number scale that fits, letterboxed.
	ScaleInteger ScaleMode = "integer"
	// ScaleCRT stretches the screen to a 4:3 shape, as on a CGA monitor
	// (pixels 1.2× taller than wide).
	ScaleCRT ScaleMode = "crt"
	// ScaleSmooth is the largest size that fits, by any factor, smoothed
	// and letterboxed: for the smooth art style.
	ScaleSmooth ScaleMode = "smooth"
)

// ArtStyle is how a game's art is made, which decides its fonts and
// scaling.
type ArtStyle string

const (
	// ArtPixel is pixel art: sprites and bitmap fonts at a low resolution,
	// scaled up by whole numbers.
	ArtPixel ArtStyle = "pixel"
	// ArtSmooth is smooth art: shapes drawn in code and an outline font at
	// a high resolution, scaled by any factor.
	ArtSmooth ArtStyle = "smooth"
)

// Preset is a tested set of display choices, as the Game Skeleton's.
type Preset struct {
	Name                      string
	ScreenWidth, ScreenHeight int
	ScaleModes                []ScaleMode // the modes F9 cycles through; the first is the default
	Palette                   palettes.Palette
	Art                       ArtStyle
}

// Classic is the 320×200 pixel look (DESIGN §2.1): Resurrect 64, scaled by
// whole numbers or to CRT 4:3.
var Classic = Preset{
	Name:        "Classic",
	ScreenWidth: 320, ScreenHeight: 200,
	ScaleModes: []ScaleMode{ScaleInteger, ScaleCRT},
	Palette:    resurrect64.Palette,
	Art:        ArtPixel,
}

// Smooth is the new look: the screen four times the 320×200 grid, so the
// game's coordinates are unchanged, with Resurrect 64 and smooth art,
// scaled by any factor.
var Smooth = Preset{
	Name:        "Smooth",
	ScreenWidth: 1280, ScreenHeight: 800,
	ScaleModes: []ScaleMode{ScaleSmooth},
	Palette:    resurrect64.Palette,
	Art:        ArtSmooth,
}

// Look is which of the game's two looks is shown. The values are what
// settings.json stores.
type Look string

const (
	// LookNew is the smooth look, what players see first.
	LookNew Look = "new"
	// LookClassic is the 320×200 pixel look of v1.2.1, after the 1984
	// game.
	LookClassic Look = "classic"
)

// Looks lists the looks in the options menu's order.
var Looks = []Look{LookNew, LookClassic}

// Preset is the look's preset.
func (l Look) Preset() Preset {
	if l == LookClassic {
		return Classic
	}
	return Smooth
}

// Name is the look as the options menu shows it.
func (l Look) Name() string {
	if l == LookClassic {
		return "Classic"
	}
	return "New"
}

// Project is the preset of the look being shown: its resolution, scale
// modes and palette. The game sets it from the settings at start and when
// the look changes, and nothing else does.
var Project = Classic

// HasMode reports whether the preset offers mode m.
func (p Preset) HasMode(m ScaleMode) bool {
	for _, x := range p.ScaleModes {
		if x == m {
			return true
		}
	}
	return false
}

// NextMode is the mode after m in the preset's cycle, wrapping around.
func (p Preset) NextMode(m ScaleMode) ScaleMode {
	for i, x := range p.ScaleModes {
		if x == m {
			return p.ScaleModes[(i+1)%len(p.ScaleModes)]
		}
	}
	return p.ScaleModes[0]
}

// The window's allowed size, in window pixels (DESIGN §2.1).
const (
	MaxWindowW, MaxWindowH = 7680, 4800
	defaultWindowW         = 1280 // the default window is about this wide
)

// DefaultWindow is the largest whole-number multiple of the screen that is
// no wider than 1280 pixels: 1280×800 for Classic.
func (p Preset) DefaultWindow() (w, h int) {
	k := max(1, defaultWindowW/p.ScreenWidth)
	return p.ScreenWidth * k, p.ScreenHeight * k
}

// MinWindow is the smallest window: one window pixel per screen pixel for
// pixel art, and half the screen for smooth art, which scales by any
// factor.
func (p Preset) MinWindow() (w, h int) {
	if p.Art == ArtSmooth {
		return p.ScreenWidth / 2, p.ScreenHeight / 2
	}
	return p.ScreenWidth, p.ScreenHeight
}
