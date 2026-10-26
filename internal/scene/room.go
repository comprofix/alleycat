package scene

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/debug"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/rooms"
)

// RoomScene runs a room (DESIGN §3.4) and sends Freddy on according to how
// it ends: the tally, "OOPS!", or a lost life.
type RoomScene struct {
	room   rooms.Room
	view   *rooms.Renderer
	smooth *rooms.SmoothRenderer
	x, y   int // where Freddy went in, in the alley

	fighting bool // the dog-fight sound is playing
	shake    shaker
}

func newRoomScene(ctx *Context, room rooms.Room, x, y int) *RoomScene {
	return &RoomScene{room: room, view: rooms.NewRenderer(ctx.Rooms, ctx.Alley), smooth: rooms.NewSmoothRenderer(), x: x, y: y}
}

// Update implements Scene.
func (r *RoomScene) Update(ctx *Context) error {
	if handleGameKeys(ctx, r) {
		return nil
	}
	s := ctx.Session
	s.LevelTicks += ctx.Ticks
	r.room.Update(ctx.Ticks, ctx.FreddySteps, controls(ctx.Input), ctx.RNG)
	runPatter(ctx, roomPatter(r.room))
	b := r.room.Shared()
	r.shake.update(b.Freddy.Hits, ctx.Ticks)
	playRoomSounds(ctx, r.room)
	followFight(ctx, &b.Dog, &r.fighting)
	if hr, ok := r.room.(*rooms.HeartRoom); ok {
		r.heartEnds(ctx, hr)
		return nil
	}
	switch b.Out {
	case rooms.Cleared:
		bonus := rooms.Bonus(r.room.Kind(), b.Ticks)
		s.AddScore(bonus)
		s.RoomsCleared++
		s.FeliciaWaiting = true
		ctx.Switch(newTally(bonus, rooms.TimeBonus(r.room.Kind(), b.Ticks), r.x, r.y))
	case rooms.LeftEarly:
		ctx.Switch(newResult(resultOops, returnTo(r.x, r.y)))
	case rooms.Failed:
		s.LoseLife()
		if b.ByDog {
			ctx.Switch(newHereKitty(afterLostLife))
		} else {
			ctx.Switch(newResult(resultLost, afterLostLife))
		}
	}
	return nil
}

// heartEnds sends Freddy on from the heart room: the kiss, or the
// rejection scene and back to the alley with no life lost (DESIGN §3.5).
func (r *RoomScene) heartEnds(ctx *Context, hr *rooms.HeartRoom) {
	switch hr.Out {
	case rooms.Cleared:
		ctx.Switch(newKiss(ctx, hr, r.view, r.smooth, returnTo(r.x, r.y)))
	case rooms.Failed:
		ctx.Switch(newReject(returnTo(r.x, r.y)))
	}
}

func (*RoomScene) native() {}

// Draw implements Scene: the room in the look dst is the size of, or its
// pixel art enlarged while its new look is still to come.
func (r *RoomScene) Draw(ctx *Context, dst *ebiten.Image) {
	if dst.Bounds().Dx() != gfx.ScreenWidth && r.smooth.Drawn(r.room) {
		r.shake.draw(dst, func(d *ebiten.Image) { r.smooth.Draw(d, r.room) })
	} else {
		drawPixels(ctx, dst, func(d *ebiten.Image) {
			r.shake.draw(d, func(d *ebiten.Image) { r.view.Draw(d, r.room) })
		})
	}
	if debug.Enabled {
		drawPixels(ctx, dst, func(d *ebiten.Image) {
			r.view.DrawDebug(d, r.room)
			ctx.PixelFont.Draw(d, fmt.Sprintf("TPS %.1f FPS %.1f", ebiten.ActualTPS(), ebiten.ActualFPS()), 4, 190, resurrect64.White, 1)
		})
	}
}

// returnTo is the alley with Freddy back where he went in.
func returnTo(x, y int) func(*Context) Scene {
	return func(ctx *Context) Scene { return newReturnPlay(ctx, x, y) }
}

// afterLostLife: game over at 0 lives (straight to the title), otherwise
// Freddy walks back into the alley from the side (DESIGN §3.4, "How a room
// ends").
func afterLostLife(ctx *Context) Scene {
	if ctx.Session.Lives == 0 {
		ctx.HighScore = max(ctx.HighScore, ctx.Session.Score)
		return NewTitle()
	}
	return NewPlay(ctx)
}

// resultWord is the word shown on leaving a room (DESIGN §3.7).
type resultWord int

const (
	resultOops resultWord = iota // left without losing a life
	resultLost                   // a life was lost: one of four words
)

// lostWords are picked at random when a life is lost (owner decision).
var lostWords = []string{"OUCH!", "ALAS!", "PITY!", "ARGH!"}

// resultInks are the colours the word may appear in (DESIGN §3.7).
var resultInks = []int{resurrect64.White, resurrect64.Seafoam, resurrect64.Cerise}

// resultBands are each ink's lit and shaded steps, for the word's bands.
var resultBands = map[int][2]int{
	resurrect64.White:   {resurrect64.White, resurrect64.Mist},
	resurrect64.Seafoam: {resurrect64.White, resurrect64.Aqua},
	resurrect64.Cerise:  {resurrect64.Flamingo, resurrect64.Mulberry},
}

// Result shows the result word for 28 ticks, then moves on (DESIGN §3.8,
// "Returning from a room").
type Result struct {
	word  resultWord
	text  string
	ink   int
	next  func(*Context) Scene
	ticks int

	font *gfx.VectorFont // the new look's lettering; a render cache
	big  banner
}

func newResult(word resultWord, next func(*Context) Scene) *Result {
	return &Result{word: word, next: next, ink: -1}
}

// Update implements Scene.
func (r *Result) Update(ctx *Context) error {
	if r.ink < 0 {
		r.ink = resultInks[ctx.RNG.IntN(len(resultInks))]
		r.text = "OOPS!"
		if r.word == resultLost {
			r.text = lostWords[ctx.RNG.IntN(len(lostWords))]
		}
		// The result tune fills the screen's 28 ticks (DESIGN §7.4).
		tune := audio.RoomLeft()
		if r.word == resultLost {
			tune = audio.Failure()
		}
		ctx.Audio.PlayMusic(tune, false)
	}
	r.ticks += ctx.Ticks
	if r.ticks >= config.ResultTicks {
		ctx.Switch(r.next(ctx))
	}
	return nil
}

// Draw implements Scene.
// The result word is sprayed like the fence graffiti (DESIGN §8, M8g): big,
// with a dark shadow and paint running down from some letters.
const (
	resultScale = 3
	resultY     = 84
)

func (r *Result) Draw(ctx *Context, dst *ebiten.Image) {
	if smoothScreen(dst) {
		r.drawSmooth(dst)
		return
	}
	dst.Fill(gfx.Palette[resurrect64.Ink])
	if r.ink < 0 {
		return
	}
	x := gfx.CentreX(r.text, resultScale)
	ctx.PixelFont.DrawShadowed(dst, r.text, x, resultY, r.ink, resurrect64.Shadow, resultScale)
	band := resultBands[r.ink]
	ctx.PixelFont.DrawBanded(dst, r.text, x, resultY, band[0], r.ink, band[1], resultScale)
	glyph := gfx.GlyphSize * resultScale
	for i := range len(r.text) {
		if i%2 == 0 {
			continue
		}
		dx := x + i*glyph + glyph/2
		dst.SubImage(image.Rect(dx, resultY+glyph-3, dx+2, resultY+glyph+3+i%3*3)).(*ebiten.Image).Fill(gfx.Palette[band[1]])
	}
}
