package scene

import (
	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/rooms"
)

// runPatter runs the background patter for this frame's master ticks
// (DESIGN §7.2). The clicks go on the music track, so effects pre-empt them.
func runPatter(ctx *Context, s audio.PatterState) {
	for i := range ctx.Ticks {
		if n := ctx.Patter.Tick(s); n != nil {
			// Each click belongs to its own tick, so the steps fall
			// exactly the right time apart (DESIGN §7).
			ctx.Audio.Now(float64(ctx.TickCount - int64(ctx.Ticks) + 1 + int64(i)))
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

// playAlleySounds plays the effects the alley made this frame (DESIGN
// §7.3). Effects go on the effect track and pre-empt the patter.
func playAlleySounds(ctx *Context, a *alley.Alley) {
	fx := ctx.FX
	for _, s := range a.Sounds {
		var n []audio.Note
		switch s {
		case alley.SoundMouseCaught:
			n = fx.MouseCaught()
		case alley.SoundSqueak:
			n = fx.Squeak()
		case alley.SoundBonk:
			n = fx.Bonk()
		case alley.SoundTrashCat:
			n = fx.TrashCat()
		case alley.SoundCanOn:
			n = fx.CanOn()
		case alley.SoundCanOff:
			n = fx.CanOff()
		case alley.SoundFenceDrop:
			n = fx.FenceDrop()
		}
		if n != nil {
			ctx.Audio.PlaySFX(n)
		}
	}
}

// followFight plays the dog-fight sound from the moment a dog catches
// Freddy until it has gone (DESIGN §7.3). fighting holds the state between
// frames.
func followFight(ctx *Context, d *entity.Dog, fighting *bool) {
	switch caught := d.Capture != 0; {
	case caught && !*fighting:
		*fighting = true
		if n := ctx.FX.DogFight(); n != nil {
			ctx.Audio.PlaySFX(n)
		}
	case !caught && *fighting:
		*fighting = false
		ctx.Audio.StopSFX()
	}
}
