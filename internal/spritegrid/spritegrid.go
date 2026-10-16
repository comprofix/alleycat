// Package spritegrid parses text-grid sprite sources and packs them into a
// sheet (DESIGN §8). Format:
//
//	; comment
//	sprite NAME WIDTH HEIGHT
//	<HEIGHT rows of WIDTH characters>   frame 0
//	<blank line>
//	<HEIGHT rows>                       frame 1, and so on
//	recolour NEW from SRC 0=3 E=4 ...    a copy of sprite SRC (from the same file)
//	                                     with palette indices swapped
//
// Each character is '.' for transparent or a hex digit 0–F naming a
// Sweetie 16 palette index.
package spritegrid

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"strconv"
	"strings"

	"github.com/comprofix/alleycat/internal/gfx"
)

// Transparent marks an empty pixel.
const Transparent = -1

// Sprite is one named sprite with one or more frames of equal size.
type Sprite struct {
	Name   string
	W, H   int
	Frames [][]int // each W*H palette indices, row-major
}

// Entry locates a sprite's frames in a packed sheet. Frames run left to right.
type Entry struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	W      int `json:"w"`
	H      int `json:"h"`
	Frames int `json:"frames"`
}

// Parse reads all sprites from r.
func Parse(r io.Reader) ([]Sprite, error) {
	var sprites []Sprite
	var cur *Sprite
	var rows []string
	lineNo := 0

	flush := func() error {
		if cur == nil || len(rows) == 0 {
			return nil
		}
		if len(rows) != cur.H {
			return fmt.Errorf("line %d: sprite %q frame %d has %d rows, want %d", lineNo, cur.Name, len(cur.Frames), len(rows), cur.H)
		}
		frame := make([]int, 0, cur.W*cur.H)
		for _, row := range rows {
			for _, ch := range row {
				if ch == '.' {
					frame = append(frame, Transparent)
					continue
				}
				v, err := strconv.ParseInt(string(ch), 16, 8)
				if err != nil {
					return fmt.Errorf("sprite %q: bad pixel %q", cur.Name, ch)
				}
				frame = append(frame, int(v))
			}
		}
		cur.Frames = append(cur.Frames, frame)
		rows = rows[:0]
		return nil
	}

	var recolours []recolour

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		lineNo++
		line := strings.TrimRight(sc.Text(), " \t\r")
		switch {
		case strings.HasPrefix(line, ";"):
			continue
		case strings.HasPrefix(line, "recolour "):
			if err := flush(); err != nil {
				return nil, err
			}
			rc, err := parseRecolour(line)
			if err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNo, err)
			}
			recolours = append(recolours, rc)
			cur = nil
		case strings.HasPrefix(line, "sprite "):
			if err := flush(); err != nil {
				return nil, err
			}
			var s Sprite
			if _, err := fmt.Sscanf(line, "sprite %s %d %d", &s.Name, &s.W, &s.H); err != nil {
				return nil, fmt.Errorf("line %d: %v", lineNo, err)
			}
			if s.W <= 0 || s.H <= 0 {
				return nil, fmt.Errorf("line %d: sprite %q has invalid size", lineNo, s.Name)
			}
			sprites = append(sprites, s)
			cur = &sprites[len(sprites)-1]
		case line == "":
			if err := flush(); err != nil {
				return nil, err
			}
		default:
			if cur == nil {
				return nil, fmt.Errorf("line %d: pixels before any sprite header", lineNo)
			}
			if len(line) != cur.W {
				return nil, fmt.Errorf("line %d: sprite %q row is %d wide, want %d", lineNo, cur.Name, len(line), cur.W)
			}
			rows = append(rows, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if err := flush(); err != nil {
		return nil, err
	}
	for _, s := range sprites {
		if len(s.Frames) == 0 {
			return nil, fmt.Errorf("sprite %q has no frames", s.Name)
		}
	}
	for _, rc := range recolours {
		s, err := rc.apply(sprites)
		if err != nil {
			return nil, err
		}
		sprites = append(sprites, s)
	}
	return sprites, nil
}

// recolour is a sprite made from another by swapping palette indices.
type recolour struct {
	name, from string
	swap       map[int]int
}

// parseRecolour reads "recolour NEW from SRC a=b ...", a and b hex digits.
func parseRecolour(line string) (recolour, error) {
	f := strings.Fields(line)
	if len(f) < 5 || f[2] != "from" {
		return recolour{}, fmt.Errorf("want \"recolour NEW from SRC a=b ...\", got %q", line)
	}
	rc := recolour{name: f[1], from: f[3], swap: map[int]int{}}
	for _, p := range f[4:] {
		a, b, ok := strings.Cut(p, "=")
		av, err1 := strconv.ParseInt(a, 16, 8)
		bv, err2 := strconv.ParseInt(b, 16, 8)
		if !ok || err1 != nil || err2 != nil || len(a) != 1 || len(b) != 1 {
			return recolour{}, fmt.Errorf("bad colour swap %q", p)
		}
		rc.swap[int(av)] = int(bv)
	}
	return rc, nil
}

// apply builds the recoloured sprite from its source.
func (rc recolour) apply(sprites []Sprite) (Sprite, error) {
	for _, s := range sprites {
		if s.Name != rc.from {
			continue
		}
		out := Sprite{Name: rc.name, W: s.W, H: s.H}
		for _, fr := range s.Frames {
			nf := make([]int, len(fr))
			for i, v := range fr {
				if to, ok := rc.swap[v]; ok {
					v = to
				}
				nf[i] = v
			}
			out.Frames = append(out.Frames, nf)
		}
		return out, nil
	}
	return Sprite{}, fmt.Errorf("recolour %q: no sprite %q in this file", rc.name, rc.from)
}

// Pack lays sprites out top to bottom, frames left to right, and returns
// the sheet image and the location of each sprite.
func Pack(sprites []Sprite) (*image.RGBA, map[string]Entry, error) {
	w, h := 1, 1
	for _, s := range sprites {
		w = max(w, s.W*len(s.Frames))
		h += s.H
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	index := make(map[string]Entry, len(sprites))
	y := 0
	for _, s := range sprites {
		if _, dup := index[s.Name]; dup {
			return nil, nil, fmt.Errorf("duplicate sprite name %q", s.Name)
		}
		for f, frame := range s.Frames {
			for i, p := range frame {
				if p == Transparent {
					continue
				}
				img.SetRGBA(f*s.W+i%s.W, y+i/s.W, gfx.Palette[p])
			}
		}
		index[s.Name] = Entry{X: 0, Y: y, W: s.W, H: s.H, Frames: len(s.Frames)}
		y += s.H
	}
	return img, index, nil
}
