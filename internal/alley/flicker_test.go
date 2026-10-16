package alley

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

// The lit windows flicker now and then, briefly (DESIGN §8, "Polish").
func TestWindowsFlickerNowAndThen(t *testing.T) {
	const span = 100 * config.FlickerEvery
	for i := range litWindows {
		dim := 0
		for tick := range span {
			if flickering(tick, i) {
				dim++
				if tick%config.FlickerEvery >= config.FlickerTicks {
					t.Fatalf("window %d dim at tick %d, outside a flicker", i, tick)
				}
			}
		}
		// One chance in FlickerOdds, 3 dim ticks a flicker: allow a wide band.
		if dim < 20 || dim > 150 {
			t.Errorf("window %d dim on %d of %d ticks", i, dim, span)
		}
	}
	// Each window keeps its own pattern: they rarely dip all together.
	all := 0
	for tick := range span {
		n := 0
		for i := range litWindows {
			if flickering(tick, i) {
				n++
			}
		}
		if n == len(litWindows) {
			all++
		}
	}
	if all > 6 {
		t.Errorf("all lit windows dim together on %d ticks", all)
	}
}
