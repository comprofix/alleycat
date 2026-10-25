package entity

import "github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"

// mouseFeet are where a mouse's feet are, by walk frame and standing
// (last), as its sprite's.
var mouseFeet = [3][2]float64{{14, 42}, {22, 34}, {18, 38}}

// DrawMouse draws a grey mouse side on, facing right, 64×32; frame is its
// walk frame, or -1 standing. The alley's mice and the cheese room's
// caught mice are this mouse.
func DrawMouse(p Pen, frame int) {
	feet := mouseFeet[2]
	if frame >= 0 {
		feet = mouseFeet[frame%2]
	}
	p.Stroke(resurrect64.Lilac, 2, true, 10, 21, 4, 22, 1, 25, 4, 29, 8, 30)
	for _, x := range feet {
		p.Stroke(resurrect64.Lilac, 2.5, false, x, 24, x+1, 29)
		p.Ellipse(x+2.5, 29.5, 3, 1.6, 0, resurrect64.Lilac)
	}
	p.Ellipse(30, 18, 21, 9, 0, resurrect64.Fog)
	p.Stroke(resurrect64.Mist, 3, true, 14, 13, 22, 10.5, 32, 10)
	p.Ellipse(37, 24, 10, 3, 0, resurrect64.Lilac)
	p.Ellipse(50, 16, 11, 7.5, 0.15, resurrect64.Fog)
	p.Poly(resurrect64.Fog, 54, 11, 63, 18, 54, 22)
	p.Circle(62, 18, 2.2, resurrect64.Cerise)
	p.Circle(44, 7, 5.5, resurrect64.Fog)
	p.Circle(44.5, 7.5, 3, resurrect64.Pink)
	p.Circle(52, 13.5, 2, resurrect64.Ink)
	p.Circle(51.4, 12.8, 0.7, resurrect64.White)
	p.Stroke(resurrect64.Lilac, 0.8, false, 58, 19, 64, 16)
	p.Stroke(resurrect64.Lilac, 0.8, false, 58, 20, 64, 22)
}
