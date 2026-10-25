package gfx

import (
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/bitfont"
	"github.com/comprofix/alleycat/internal/config"
)

// Font draws text in any palette colour, as the Game Skeleton's: the menus
// use it. Every character is CellWidth pixels wide and each line
// LineHeight pixels tall, so text lays out in character cells.
type Font interface {
	// CellWidth is how far each character moves the next one on, in pixels.
	CellWidth() int
	// LineHeight is the height of one line of text, in pixels.
	LineHeight() int
	// Width is the width of s in pixels.
	Width(s string) int
	// Draw renders s with its top-left corner at (x, y) in palette colour c.
	Draw(dst *ebiten.Image, s string, x, y, c int)
	// DrawShadowed renders s with a drop shadow in palette colour shadow.
	DrawShadowed(dst *ebiten.Image, s string, x, y, c, shadow int)
	// CentreX is the x that centres s on the screen.
	CentreX(s string) int
}

// Fonts are the menus' two fonts.
type Fonts struct {
	Small Font // 8×8
	Large Font // 16×16, made from Small with Scale2x
}

// LoadFont reads the 8×8 bitmap font at path in fsys (DESIGN §8).
func LoadFont(fsys fs.FS, path string) (*bitfont.Font, error) {
	f, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bf, err := bitfont.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return bf, nil
}

// NewFonts makes the menus' fonts from the 8×8 font.
func NewFonts(small *bitfont.Font) Fonts {
	return Fonts{Small: NewFont(small), Large: NewFont(bitfont.Scale2x(small))}
}

// atlas draws each of bf's glyphs into a row of cells, ink white so it can
// be tinted when drawn, and returns it with each rune's cell.
func atlas(bf *bitfont.Font) (*ebiten.Image, map[rune]int) {
	runes := bf.Runes()
	n := bf.Size
	img := image.NewRGBA(image.Rect(0, 0, n*len(runes), n))
	index := make(map[rune]int, len(runes))
	for i, r := range runes {
		bf.Draw(img.SubImage(image.Rect(i*n, 0, (i+1)*n, n)).(*image.RGBA), string(r), i*n, 0, color.RGBA{0xff, 0xff, 0xff, 0xff})
		index[r] = i
	}
	return ebiten.NewImageFromImage(img), index
}

// BitmapFont is a Font drawn from a bitmap font's glyphs.
type BitmapFont struct {
	bf    *bitfont.Font
	atlas *ebiten.Image
	index map[rune]int
}

// NewFont builds a glyph atlas for bf.
func NewFont(bf *bitfont.Font) *BitmapFont {
	a, index := atlas(bf)
	return &BitmapFont{bf: bf, atlas: a, index: index}
}

// CellWidth is the font's cell size: 8 or 16, as its cells are square.
func (f *BitmapFont) CellWidth() int { return f.bf.Size }

// LineHeight is the font's cell size: 8 or 16, as its cells are square.
func (f *BitmapFont) LineHeight() int { return f.bf.Size }

// Width is the width of s in pixels.
func (f *BitmapFont) Width(s string) int { return f.bf.Width(s) }

// Draw renders s at (x, y) in palette colour c. Characters missing from the
// font are drawn as '?'.
func (f *BitmapFont) Draw(dst *ebiten.Image, s string, x, y, c int) {
	n := f.bf.Size
	op := &ebiten.DrawImageOptions{}
	op.ColorScale.ScaleWithColor(RGBA(c))
	for _, r := range s {
		i, ok := f.index[r]
		if !ok {
			i, ok = f.index['?']
		}
		if ok && r != ' ' {
			src := f.atlas.SubImage(image.Rect(i*n, 0, (i+1)*n, n)).(*ebiten.Image)
			op.GeoM.Reset()
			op.GeoM.Translate(float64(x), float64(y))
			dst.DrawImage(src, op)
		}
		x += n
	}
}

// DrawShadowed renders s with a drop shadow in palette colour shadow, offset
// down and right by one pixel per 8 pixels of font size.
func (f *BitmapFont) DrawShadowed(dst *ebiten.Image, s string, x, y, c, shadow int) {
	d := max(1, f.bf.Size/8)
	f.Draw(dst, s, x+d, y+d, shadow)
	f.Draw(dst, s, x, y, c)
}

// CentreX is the x that centres s on the screen.
func (f *BitmapFont) CentreX(s string) int {
	return (config.Project.ScreenWidth - f.Width(s)) / 2
}

// GlyphSize is the pixel font's cell size in pixels.
const GlyphSize = 8

// PixelFont draws the 8×8 font at whole-number scales, with the lettering
// of the pixel look: shadows, outlines and bands of colour (DESIGN §8).
type PixelFont struct {
	atlas *ebiten.Image
	index map[rune]int
}

// NewPixelFont builds the glyph atlas for the 8×8 font bf.
func NewPixelFont(bf *bitfont.Font) *PixelFont {
	a, index := atlas(bf)
	return &PixelFont{atlas: a, index: index}
}

// Width returns the width of s in pixels at scale 1.
func Width(s string) int {
	return utf8.RuneCountInString(s) * GlyphSize
}

// Draw renders s at (x, y) in palette colour idx, scaled by an integer factor.
// Characters missing from the font are drawn as '?'.
func (f *PixelFont) Draw(dst *ebiten.Image, s string, x, y, idx, scale int) {
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
func (f *PixelFont) DrawCentred(dst *ebiten.Image, s string, y, idx, scale int) {
	f.Draw(dst, s, (ScreenWidth-Width(s)*max(scale, 1))/2, y, idx, scale)
}

// DrawShadowed renders s with a drop shadow, one pixel per scale step down
// and right, in palette colour shadow.
func (f *PixelFont) DrawShadowed(dst *ebiten.Image, s string, x, y, idx, shadow, scale int) {
	d := max(scale, 1)
	f.Draw(dst, s, x+d, y+d, shadow, scale)
	f.Draw(dst, s, x, y, idx, scale)
}

// DrawOutlined renders s with a one-pixel outline round every letter in
// palette colour outline.
func (f *PixelFont) DrawOutlined(dst *ebiten.Image, s string, x, y, idx, outline, scale int) {
	for _, d := range [8][2]int{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		f.Draw(dst, s, x+d[0], y+d[1], outline, scale)
	}
	f.Draw(dst, s, x, y, idx, scale)
}

// DrawBanded renders s in three bands of colour, top to bottom, as painted
// lettering catches the light: the top quarter in light, the middle in
// idx and the lower three eighths in dark. Each band is the text drawn
// again through a clip of its rows.
func (f *PixelFont) DrawBanded(dst *ebiten.Image, s string, x, y, light, idx, dark, scale int) {
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
