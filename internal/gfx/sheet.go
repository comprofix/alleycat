package gfx

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// sheetEntry mirrors spritegrid.Entry in the generated JSON.
type sheetEntry struct {
	X, Y, W, H, Frames int
}

// Sprite is a named sprite's frames within a sheet.
type Sprite struct {
	frames []*ebiten.Image
	W, H   int
}

// Frames returns the number of frames.
func (s *Sprite) Frames() int { return len(s.frames) }

// Frame returns frame i, wrapping around.
func (s *Sprite) Frame(i int) *ebiten.Image {
	n := len(s.frames)
	return s.frames[((i%n)+n)%n]
}

// Sheet holds the sprites from one generated sheet.
type Sheet struct {
	sprites map[string]*Sprite
}

// LoadSheet reads base.png and base.json (written by tools/spritegen).
func LoadSheet(fsys fs.FS, base string) (*Sheet, error) {
	pf, err := fsys.Open(base + ".png")
	if err != nil {
		return nil, err
	}
	defer pf.Close()
	img, err := png.Decode(pf)
	if err != nil {
		return nil, fmt.Errorf("%s.png: %w", base, err)
	}
	data, err := fs.ReadFile(fsys, base+".json")
	if err != nil {
		return nil, err
	}
	var index map[string]sheetEntry
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("%s.json: %w", base, err)
	}

	atlas := ebiten.NewImageFromImage(img)
	sh := &Sheet{sprites: make(map[string]*Sprite, len(index))}
	for name, e := range index {
		sp := &Sprite{W: e.W, H: e.H}
		for f := range e.Frames {
			r := image.Rect(e.X+f*e.W, e.Y, e.X+(f+1)*e.W, e.Y+e.H)
			sp.frames = append(sp.frames, atlas.SubImage(r).(*ebiten.Image))
		}
		sh.sprites[name] = sp
	}
	return sh, nil
}

// Sprite returns the named sprite; it panics if missing, since sprite names
// are fixed at build time.
func (sh *Sheet) Sprite(name string) *Sprite {
	sp, ok := sh.sprites[name]
	if !ok {
		panic("gfx: no sprite named " + name)
	}
	return sp
}
