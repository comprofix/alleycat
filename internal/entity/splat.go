package entity

import (
	"math"

	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
)

// DrawSplat draws word in font f on a ragged splat filling a w×h box on
// the layout, its letters climbing to the right: the new look's impact
// words, BONK!, SQUEEK! and BYTE!.
func DrawSplat(p Pen, f gfx.Font, word string, w, h int) {
	splat(p, float64(w*K), float64(h*K))
	step := (float64(w*K) - 24) / float64(len(word))
	for i, ch := range word {
		lx := 12 + float64(i)*step
		ly := float64(h*K)/2 - 18 - float64(i)*float64(h*K)*0.3/float64(len(word))
		p.Text(f, string(ch), lx+2, ly+2, resurrect64.Plum)
		p.Text(f, string(ch), lx, ly, resurrect64.Grape)
	}
}

// splat draws a ragged splat filling a w×h box: a blob with a spiky
// edge, lighter in the middle, with drops flung round it.
func splat(p Pen, w, h float64) {
	cx, cy := w/2, h/2
	const n = 36
	edge := func(scale float64) []float64 {
		pts := make([]float64, 0, 2*n)
		for i := range n {
			a := 2 * math.Pi * float64(i) / n
			d := scale * (0.86 + 0.14*float64((i*7)%5)/4)
			if i%2 == 1 {
				d *= 0.84
			}
			pts = append(pts, cx+(w/2-4)*d*math.Cos(a), cy+(h/2-3)*d*math.Sin(a))
		}
		return pts
	}
	p.Poly(resurrect64.Fog, edge(1)...)
	p.Poly(resurrect64.Mist, edge(0.72)...)
	for i := range 8 {
		a := float64(i)*math.Pi/4 + 0.3
		p.Circle(cx+(w/2-1)*math.Cos(a), cy+(h/2)*math.Sin(a), 2.5, resurrect64.Fog)
	}
}
