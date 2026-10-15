package scene

import (
	"testing"
	"time"

	"github.com/comprofix/alleycat/internal/audio"
)

// loudness plays a long tone and returns its largest sample-to-sample step
// after the first 50 ms, so the last tone's fade is over. Below quiet is
// silence: the speaker model only settles slowly then.
func loudness(ctx *Context) int {
	ctx.Audio.StopMusic()
	ctx.Audio.PlayMusic([]audio.Note{{Hz: 500, Dur: time.Second}}, false)
	ctx.Audio.Read(make([]byte, 4*2205))
	buf := make([]byte, 4*11025)
	ctx.Audio.Read(buf)
	return maxJump(buf)
}

const quiet = 50

// The sound settings reach the synth (DESIGN §9): off or muted is silent,
// and the volume scales the level.
func TestApplySound(t *testing.T) {
	ctx := &Context{Audio: audio.NewSynth(), Sound: DefaultSound()}
	ctx.ApplySound()
	full := loudness(ctx)
	if full <= quiet {
		t.Fatal("no sound with the default settings")
	}

	ctx.Sound.On = false
	ctx.ApplySound()
	if loudness(ctx) > quiet {
		t.Error("sound off (Ctrl+S) still plays")
	}
	ctx.Sound.On, ctx.Sound.Muted = true, true
	ctx.ApplySound()
	if loudness(ctx) > quiet {
		t.Error("muted still plays")
	}

	ctx.Sound.Muted = false
	ctx.Sound.Volume /= 2
	ctx.ApplySound()
	if half := loudness(ctx); half < full*4/10 || half > full*6/10 {
		t.Errorf("half volume gives %d against %d at full", half, full)
	}
}

func TestDefaultSound(t *testing.T) {
	s := DefaultSound()
	if !s.On || s.Muted || s.Volume <= 0 || !s.Soften {
		t.Errorf("DefaultSound = %+v: want on, unmuted, audible, softened", s)
	}
}
