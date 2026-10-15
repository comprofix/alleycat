// Package audio is the PC speaker emulation (DESIGN §7): a one-voice
// square-wave synth with a music track and a sound-effect track. Tunes and
// effects are note tables; there are no audio files.
package audio

import (
	"time"

	"github.com/comprofix/alleycat/internal/clock"
)

// Note is one step of a tune or effect.
type Note struct {
	Hz  float64 // 0 is a rest
	Dur time.Duration
}

// Ticks returns the length of n master ticks (DESIGN §7.1), so tables can
// be written in the original's timing unit.
func Ticks(n float64) time.Duration {
	return time.Duration(n * float64(time.Second) / clock.TickHz)
}
