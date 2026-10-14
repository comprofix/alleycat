package gfx

import "math"

// Aspect selects how the 320×200 screen is presented (DESIGN §2.1).
type Aspect int

const (
	// AspectSquare shows square pixels at the largest integer scale that fits.
	AspectSquare Aspect = iota
	// AspectCRT stretches the image to 4:3, as on a CGA monitor (pixels 1.2× taller than wide).
	AspectCRT
)

func (a Aspect) String() string {
	if a == AspectCRT {
		return "CRT 4:3"
	}
	return "square pixels"
}

// Next cycles to the other aspect mode.
func (a Aspect) Next() Aspect {
	if a == AspectSquare {
		return AspectCRT
	}
	return AspectSquare
}

// Rect is a destination rectangle on the output surface, in output pixels.
type Rect struct {
	X, Y, W, H float64
}

// Fit returns where the logical screen is drawn on an output surface of
// outW×outH pixels, centred with letterboxing.
func Fit(a Aspect, outW, outH int) Rect {
	ow, oh := float64(outW), float64(outH)
	var w, h float64
	switch a {
	case AspectCRT:
		// Largest 4:3 rectangle that fits.
		w = math.Min(ow, oh*4/3)
		h = w * 3 / 4
	default:
		s := math.Floor(math.Min(ow/ScreenWidth, oh/ScreenHeight))
		if s < 1 {
			// Window smaller than 320×200: fall back to a fractional fit.
			s = math.Min(ow/ScreenWidth, oh/ScreenHeight)
		}
		w, h = ScreenWidth*s, ScreenHeight*s
	}
	return Rect{X: math.Floor((ow - w) / 2), Y: math.Floor((oh - h) / 2), W: w, H: h}
}

// PrescaleFactor is the integer nearest-neighbour upscale applied before a
// smooth (linear) resample to a non-integer CRT size. Upscaling first keeps
// pixel edges sharp while evening out the uneven rows.
func PrescaleFactor(r Rect) int {
	k := math.Ceil(math.Max(r.W/ScreenWidth, r.H/ScreenHeight))
	return int(math.Max(k, 1))
}
