package entity

import (
	"math"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// DrawSplat draws word in font f on a ragged splat filling a w×h box on
// the layout, its letters climbing a little to the right: the new look's
// impact words, BONK!, SQUEEK! and BYTE!. The letters are spread inside
// the splat's solid middle, so none sticks out past its spikes; f should
// be 36 pixels.
func DrawSplat(p Pen, f gfx.Font, word string, w, h int) {
	bw, bh := float64(w*K), float64(h*K)
	splat(p, bw, bh)
	// The letters are spaced over the box less an inset at either end and
	// centred on it, the middle of their capitals on its middle: in Go Mono
	// that is about halfway down the line, below the room left for accents.
	const inset, rise, capMiddle = 22, 8, 0.49
	cell := float64(f.CellWidth())
	step := (bw - 2*inset - cell) / float64(len(word)-1)
	left := (bw - float64(len(word)-1)*step - cell) / 2
	for i, ch := range word {
		t := float64(i)/float64(len(word)-1) - 0.5
		lx := left + float64(i)*step
		ly := bh/2 - capMiddle*float64(f.LineHeight()) - t*rise - 1 // less half the shadow's drop
		p.Text(f, string(ch), lx+2, ly+2, resurrect64.Plum)
		p.Text(f, string(ch), lx, ly, resurrect64.Grape)
	}
}

// splat draws a ragged splat over a w×h box, its spikes reaching a little
// past the box: a blob with shallow dents between the spikes, so its
// middle is solid, lighter inside, with drops flung round it.
func splat(p Pen, w, h float64) {
	cx, cy := w/2, h/2
	rx, ry := w/2+8, h/2+8
	const n = 36
	edge := func(scale float64) []float64 {
		pts := make([]float64, 0, 2*n)
		for i := range n {
			a := 2 * math.Pi * float64(i) / n
			d := scale * (0.94 + 0.06*float64((i*7)%5)/4)
			if i%2 == 1 {
				d = scale * 0.86
			}
			pts = append(pts, cx+rx*d*math.Cos(a), cy+ry*d*math.Sin(a))
		}
		return pts
	}
	p.Poly(resurrect64.Fog, edge(1)...)
	p.Poly(resurrect64.Mist, edge(0.78)...)
	for i := range 8 {
		a := float64(i)*math.Pi/4 + 0.3
		p.Circle(cx+rx*1.04*math.Cos(a), cy+ry*1.04*math.Sin(a), 2.5, resurrect64.Fog)
	}
}
