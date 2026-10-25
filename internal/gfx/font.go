package gfx

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
)

// GlyphSize is the font cell size in pixels.
const GlyphSize = 8

// Glyph is one 8×8 character; bit 7 of each row is the leftmost pixel.
type Glyph [GlyphSize]uint8

// ParseFont reads the text-grid font format: "@" plus the character on its
// own line ("@space" for space), then 8 rows of 8 '#'/'.' characters.
// Lines starting with ';' are comments.
func ParseFont(r io.Reader) (map[rune]Glyph, error) {
	glyphs := map[rune]Glyph{}
	sc := bufio.NewScanner(r)
	lineNo := 0
	var cur rune
	var g Glyph
	row := -1
	for sc.Scan() {
		lineNo++
		line := strings.TrimSuffix(sc.Text(), "\r")
		if strings.HasPrefix(line, ";") {
			continue
		}
		if row < 0 {
			switch {
			case line == "@space":
				cur = ' '
			case strings.HasPrefix(line, "@") && utf8.RuneCountInString(line) == 2:
				cur, _ = utf8.DecodeRuneInString(line[1:])
			default:
				return nil, fmt.Errorf("line %d: want glyph header like \"@A\" or \"@space\", got %q", lineNo, line)
			}
			if _, dup := glyphs[cur]; dup {
				return nil, fmt.Errorf("line %d: duplicate glyph %q", lineNo, cur)
			}
			g, row = Glyph{}, 0
			continue
		}
		if len(line) != GlyphSize {
			return nil, fmt.Errorf("line %d: glyph %q row is %d wide, want 8", lineNo, cur, len(line))
		}
		for x, ch := range line {
			switch ch {
			case '#':
				g[row] |= 0x80 >> x
			case '.':
			default:
				return nil, fmt.Errorf("line %d: glyph %q has bad pixel %q", lineNo, cur, ch)
			}
		}
		row++
		if row == GlyphSize {
			glyphs[cur] = g
			row = -1
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if row >= 0 {
		return nil, fmt.Errorf("glyph %q is incomplete", cur)
	}
	return glyphs, nil
}

// Font draws text from an 8×8 glyph atlas in any palette colour.
type Font struct {
	atlas *ebiten.Image
	index map[rune]int
}

// NewFont builds the glyph atlas. Ink is stored white and tinted at draw time.
func NewFont(glyphs map[rune]Glyph) *Font {
	f := &Font{index: make(map[rune]int, len(glyphs))}
	img := image.NewRGBA(image.Rect(0, 0, GlyphSize*len(glyphs), GlyphSize))
	i := 0
	for r := rune(0); r < utf8.MaxRune && len(f.index) < len(glyphs); r++ {
		g, ok := glyphs[r]
		if !ok {
			continue
		}
		for y, bits := range g {
			for x := range GlyphSize {
				if bits&(0x80>>x) != 0 {
					img.SetRGBA(i*GlyphSize+x, y, color.RGBA{0xff, 0xff, 0xff, 0xff})
				}
			}
		}
		f.index[r] = i
		i++
	}
	f.atlas = ebiten.NewImageFromImage(img)
	return f
}

// Width returns the width of s in pixels.
func Width(s string) int {
	return utf8.RuneCountInString(s) * GlyphSize
}

// Draw renders s at (x, y) in palette colour idx, scaled by an integer factor.
// Characters missing from the font are drawn as '?'.
func (f *Font) Draw(dst *ebiten.Image, s string, x, y, idx, scale int) {
	scale = max(scale, 1)
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleWithColor(Palette[idx])
	for _, r := range s {
		i, ok := f.index[r]
		if !ok {
			i = f.index['?']
		}
		if r != ' ' {
			src := f.atlas.SubImage(image.Rect(i*GlyphSize, 0, (i+1)*GlyphSize, GlyphSize)).(*ebiten.Image)
			op.GeoM.Reset()
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(src, op)
		}
		x += GlyphSize * scale
	}
}

// DrawCentred renders s horizontally centred on the logical screen.
func (f *Font) DrawCentred(dst *ebiten.Image, s string, y, idx, scale int) {
	f.Draw(dst, s, (ScreenWidth-Width(s)*max(scale, 1))/2, y, idx, scale)
}

// DrawShadowed renders s with a drop shadow, one pixel per scale step down
// and right, in palette colour shadow.
func (f *Font) DrawShadowed(dst *ebiten.Image, s string, x, y, idx, shadow, scale int) {
	d := max(scale, 1)
	f.Draw(dst, s, x+d, y+d, shadow, scale)
	f.Draw(dst, s, x, y, idx, scale)
}

// DrawOutlined renders s with a one-pixel outline round every letter in
// palette colour outline.
func (f *Font) DrawOutlined(dst *ebiten.Image, s string, x, y, idx, outline, scale int) {
	for _, d := range [8][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		f.Draw(dst, s, x+d[0], y+d[1], outline, scale)
	}
	f.Draw(dst, s, x, y, idx, scale)
}

// DrawBanded renders s in three bands of colour, top to bottom, as painted
// lettering catches the light: the top quarter in light, the middle in
// idx and the lower three eighths in dark. Each band is the text drawn
// again through a clip of its rows.
func (f *Font) DrawBanded(dst *ebiten.Image, s string, x, y, light, idx, dark, scale int) {
	h := GlyphSize * max(scale, 1)
	cuts := [4]int{0, h / 4, h * 5 / 8, h}
	b := dst.Bounds()
	for i, ink := range [3]int{light, idx, dark} {
		clip := dst.SubImage(image.Rect(b.Min.X, y+cuts[i], b.Max.X, y+cuts[i+1])).(*ebiten.Image)
		f.Draw(clip, s, x, y, ink, scale)
	}
}

// CentreX is the X that centres s at scale on the logical screen.
func CentreX(s string, scale int) int { return (ScreenWidth - Width(s)*max(scale, 1)) / 2 }
