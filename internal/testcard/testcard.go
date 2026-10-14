// Package testcard draws the M0 palette test card: all 16 Sweetie 16
// swatches plus pixel-level patterns for checking scaling and aspect.
package testcard

import (
	"image"

	"github.com/comprofix/alleycat/internal/gfx"
)

// Build returns the 320×200 test card. It uses palette colours only.
func Build() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, gfx.ScreenWidth, gfx.ScreenHeight))
	fill(img, image.Rect(0, 0, gfx.ScreenWidth, gfx.ScreenHeight), gfx.Night)

	// 1-pixel border marks the exact edge of the logical screen.
	outline(img, image.Rect(0, 0, gfx.ScreenWidth, gfx.ScreenHeight), gfx.White)

	// Two rows of eight 32×32 swatches, in palette order.
	for i := range gfx.Palette {
		x := 16 + (i%8)*36
		y := 16 + (i/8)*40
		fill(img, image.Rect(x, y, x+32, y+32), i)
		outline(img, image.Rect(x-1, y-1, x+33, y+33), gfx.Slate)
	}

	// 1-pixel checkerboard: any blur or uneven scaling shows up here.
	checker(img, image.Rect(16, 112, 96, 184), gfx.White, gfx.Night, 1)
	// 2-pixel checkerboard for comparison.
	checker(img, image.Rect(104, 112, 184, 184), gfx.White, gfx.Night, 2)

	// Alternating 1-pixel horizontal and vertical lines.
	for y := 112; y < 184; y++ {
		for x := 192; x < 232; x++ {
			if y%2 == 0 {
				img.Set(x, y, gfx.Palette[gfx.Yellow])
			}
		}
		for x := 240; x < 280; x++ {
			if x%2 == 0 {
				img.Set(x, y, gfx.Palette[gfx.Cyan])
			}
		}
	}

	// A 24×24 square: square in square-pixel mode, 1.2× taller in CRT mode.
	outline(img, image.Rect(288-4, 112, 288+28, 112+32), gfx.Orange)
	fill(img, image.Rect(288, 116, 288+24, 116+24), gfx.Red)

	return img
}

func fill(img *image.RGBA, r image.Rectangle, idx int) {
	c := gfx.Palette[idx]
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			img.SetRGBA(x, y, c)
		}
	}
}

func outline(img *image.RGBA, r image.Rectangle, idx int) {
	c := gfx.Palette[idx]
	for x := r.Min.X; x < r.Max.X; x++ {
		img.SetRGBA(x, r.Min.Y, c)
		img.SetRGBA(x, r.Max.Y-1, c)
	}
	for y := r.Min.Y; y < r.Max.Y; y++ {
		img.SetRGBA(r.Min.X, y, c)
		img.SetRGBA(r.Max.X-1, y, c)
	}
}

func checker(img *image.RGBA, r image.Rectangle, a, b, size int) {
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			idx := b
			if ((x-r.Min.X)/size+(y-r.Min.Y)/size)%2 == 0 {
				idx = a
			}
			img.SetRGBA(x, y, gfx.Palette[idx])
		}
	}
}
