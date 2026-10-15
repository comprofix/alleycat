package scene

import (
	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/rooms"
)

// runPatter runs the background patter for this frame's master ticks
// (DESIGN §7.2). The clicks go on the music track, so effects pre-empt them.
func runPatter(ctx *Context, s audio.PatterState) {
	for range ctx.Ticks {
		if n := ctx.Patter.Tick(s); n != nil {
			ctx.Audio.PlayMusic(n, false)
		}
	}
}

// alleyPatter is the patter's view of the alley.
func alleyPatter(a *alley.Alley) audio.PatterState {
	s := audio.PatterState{
		Mode:     audio.Alley,
		Bouncing: a.Freddy.Bouncing,
		DogOut:   a.Dog.Out(),
		Falling:  a.Thrown.Active,
		FallY:    a.Thrown.Y,
	}
	if a.Freeze > 0 {
		s.Mode = audio.Quiet
	}
	return s
}

// roomPatter is the patter's view of a room.
func roomPatter(r rooms.Room) audio.PatterState {
	b := r.Shared()
	s := audio.PatterState{
		Mode:     audio.Room,
		Bouncing: b.Freddy.Bouncing,
		DogOut:   b.HasDog() && b.Dog.Out(),
	}
	if rooms.Hushed(r) {
		s.Mode = audio.Quiet
	}
	return s
}
