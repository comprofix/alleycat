package gfx

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

func TestFitSquareUsesLargestIntegerScale(t *testing.T) {
	tests := []struct {
		outW, outH int
		want       Rect
	}{
		{1280, 800, Rect{0, 0, 1280, 800}},      // exactly 4×
		{1920, 1080, Rect{160, 40, 1600, 1000}}, // 5×, letterboxed
		{1000, 700, Rect{20, 50, 960, 600}},     // 3×
	}
	for _, tt := range tests {
		if got := Fit(config.ScaleInteger, ScreenWidth, ScreenHeight, tt.outW, tt.outH); got != tt.want {
			t.Errorf("Fit(square, %d, %d) = %+v, want %+v", tt.outW, tt.outH, got, tt.want)
		}
	}
}

func TestFitSquareSmallWindowFallsBackToFractional(t *testing.T) {
	got := Fit(config.ScaleInteger, ScreenWidth, ScreenHeight, 160, 100)
	if got.W != 160 || got.H != 100 {
		t.Errorf("Fit(square, 160, 100) = %+v, want 160×100", got)
	}
}

func TestFitCRTIsFourByThree(t *testing.T) {
	tests := []struct {
		outW, outH int
		want       Rect
	}{
		{1920, 1080, Rect{240, 0, 1440, 1080}},
		{1280, 800, Rect{106, 0, 1066.6666666666667, 800}},
		{800, 1000, Rect{0, 200, 800, 600}},
	}
	for _, tt := range tests {
		if got := Fit(config.ScaleCRT, ScreenWidth, ScreenHeight, tt.outW, tt.outH); got != tt.want {
			t.Errorf("Fit(crt, %d, %d) = %+v, want %+v", tt.outW, tt.outH, got, tt.want)
		}
	}
}

func TestPrescaleFactor(t *testing.T) {
	if k := PrescaleFactor(Rect{W: 1440, H: 1080}, ScreenWidth, ScreenHeight); k != 6 {
		t.Errorf("PrescaleFactor(1440×1080) = %d, want 6", k)
	}
	if k := PrescaleFactor(Rect{W: 100, H: 60}, ScreenWidth, ScreenHeight); k != 1 {
		t.Errorf("PrescaleFactor(100×60) = %d, want 1", k)
	}
}

// The smooth fit scales by any factor, still letterboxed.
func TestFitSmooth(t *testing.T) {
	tests := []struct {
		outW, outH int
		want       Rect
	}{
		{1280, 800, Rect{0, 0, 1280, 800}},
		{1366, 768, Rect{68, 0, 1228.8, 768}},
		{1920, 1080, Rect{96, 0, 1728, 1080}},
	}
	for _, tt := range tests {
		if got := Fit(config.ScaleSmooth, 1280, 800, tt.outW, tt.outH); got != tt.want {
			t.Errorf("Fit(smooth, %d, %d) = %+v, want %+v", tt.outW, tt.outH, got, tt.want)
		}
	}
}
