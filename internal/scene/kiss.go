package scene

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/gfx/palettes/resurrect64"
	"github.com/comprofix/alleycat/internal/rooms"
)

// kissStage is where the kiss sequence is.
type kissStage int

const (
	kissOpen   kissStage = iota // the big heart alone
	kissText                    // the bonus and the flashing multiplier
	kissDrop                    // the heart drops, leaving a column of hearts
	kissBursts                  // three bursts of hearts
	kissDone
)

// stamp is a picture left on the screen: the kiss draws without erasing,
// so the heart's drop and the bursts leave trails.
type stamp struct {
	sprite string
	x, y   int
	cats   bool // the big heart with Freddy and Felicia in it
}

// Kiss is the kiss sequence (DESIGN §3.5, "The kiss"): the heart room
// freezes in the kiss colours, the big heart holding both cats shows with
// the bonus, drops to the middle, and three bursts of hearts fill the
// screen. Then Freddy levels up and, after the dancing cats at difficulty
// 2 and up, goes back to the alley.
type Kiss struct {
	room *rooms.HeartRoom
	view *rooms.Renderer

	stage  kissStage
	wait   int
	ticks  int
	x, y   int // the big heart
	slide  int
	stamps []stamp

	burst      int
	bx, by     [8]int
	firstFrame bool

	bonus, multiplier int
	next              func(*Context) Scene
	smooched          bool
}

func newKiss(ctx *Context, room *rooms.HeartRoom, view *rooms.Renderer, next func(*Context) Scene) *Kiss {
	f := room.Freddy
	x := max(min(f.X, config.KissMaxX)-config.KissStartDX, 0) &^ 15
	k := &Kiss{
		room: room, view: view, next: next,
		x: x, y: config.KissHeartY,
		slide: min(abs(x-config.KissCentreX)/8, config.KissSlideMax) + config.KissSlideAdd,
		wait:  config.KissOpenTicks,
	}
	k.dropStep()
	k.bonus, k.multiplier = ctx.Session.KissBonus(room.WithGift)
	return k
}

// dropStep moves the big heart a step down and towards the middle, and
// leaves it there.
func (k *Kiss) dropStep() {
	switch cx := k.x &^ 15; {
	case cx == config.KissCentreX:
		k.x = config.KissCentreX
	case cx > config.KissCentreX:
		k.x -= k.slide
	default:
		k.x += k.slide
	}
	k.y += config.KissHeartStep
	k.stamps = append(k.stamps, stamp{sprite: "heart_big", x: k.x + config.KissHeartDX, y: k.y, cats: true})
}

// Update implements Scene.
func (k *Kiss) Update(ctx *Context) error {
	// The smooch plays as the big heart first appears (DESIGN §7.3).
	if !k.smooched {
		k.smooched = true
		if n := ctx.FX.Smooch(); n != nil {
			ctx.Audio.PlaySFX(n)
		}
		// The kiss tune loops through the sequence. The smooch is an
		// effect, so the tune waits under it and starts as it ends.
		ctx.Audio.PlayMusic(audio.Kiss(), true)
	}
	for range ctx.Ticks {
		k.ticks++
		if k.wait--; k.wait > 0 {
			continue
		}
		switch k.stage {
		case kissOpen:
			for range k.multiplier {
				ctx.Session.AddScore(k.bonus)
			}
			k.stage, k.wait = kissText, config.KissTextTicks
		case kissText:
			k.stage, k.wait = kissDrop, config.KissStepTicks
		case kissDrop:
			if k.y >= config.KissHeartStopY {
				k.startBurst(0)
				continue
			}
			k.dropStep()
			k.wait = config.KissStepTicks
		case kissBursts:
			k.burstFrame()
		case kissDone:
			ctx.Session.LevelUp()
			if d := newDance(ctx.Session.Difficulty, k.next); d != nil {
				ctx.Switch(d)
			} else {
				ctx.Switch(k.next(ctx))
			}
			return nil
		}
	}
	return nil
}

func (k *Kiss) startBurst(n int) {
	b := config.KissBursts[n]
	k.stage, k.burst, k.firstFrame = kissBursts, n, true
	for i := range k.bx {
		k.bx[i], k.by[i] = b.StartX, b.StartY
	}
	k.burstFrame()
}

// burstFrame leaves the burst's hearts where they are, then moves them on;
// once one reaches a limit the next burst starts.
func (k *Kiss) burstFrame() {
	b := config.KissBursts[k.burst]
	for i := range k.bx {
		if !k.firstFrame || i == len(k.bx)-1 { // they start all in one place
			k.stamps = append(k.stamps, stamp{sprite: b.Sprite, x: k.bx[i], y: k.by[i]})
		}
	}
	k.firstFrame = false
	done := false
	for i, d := range config.KissDirs {
		k.bx[i], done = burstMove(k.bx[i], d[0]*b.StepX, b.MaxX, done)
		k.by[i], done = burstMove(k.by[i], d[1]*b.StepY, b.MaxY, done)
	}
	k.wait = b.Every
	if !done {
		return
	}
	if k.burst+1 < len(config.KissBursts) {
		k.startBurstAfter()
		return
	}
	k.stage = kissDone
}

// startBurstAfter begins the next burst once this frame's wait is over.
func (k *Kiss) startBurstAfter() {
	n := k.burst + 1
	b := config.KissBursts[n]
	k.burst, k.firstFrame = n, true
	for i := range k.bx {
		k.bx[i], k.by[i] = b.StartX, b.StartY
	}
}

// burstMove moves v by step within 0–limit, noting when it hits an end.
func burstMove(v, step, limit int, done bool) (int, bool) {
	v += step
	switch {
	case v > limit:
		return limit, true
	case v < 0:
		return 0, true
	}
	return v, done
}

// Draw implements Scene.
func (k *Kiss) Draw(ctx *Context, dst *ebiten.Image) {
	k.view.DrawHeartKiss(dst, k.room)
	sheet := ctx.Rooms
	for _, s := range k.stamps {
		entity.DrawAt(dst, sheet.Sprite(s.sprite).Frame(0), s.x, s.y, false)
		if s.cats {
			entity.DrawAt(dst, sheet.Sprite("kiss_cats").Frame(0), s.x+config.KissCatsDX, s.y+config.KissCatsDY, false)
		}
	}
	if k.stage != kissText {
		return
	}
	f := ctx.PixelFont
	score := fmt.Sprint(k.bonus)
	f.DrawShadowed(dst, score, gfx.CentreX(score, 1), config.KissScoreY, resurrect64.White, resurrect64.Ink, 1)
	text := fmt.Sprintf("BONUS MULTIPLIER: %02d", k.multiplier)
	bar := resurrect64.Ink
	if (k.ticks/config.KissFlashTicks)%2 == 1 {
		bar = resurrect64.Cobalt
	}
	w := len(text)*8 + 8
	dst.SubImage(image.Rect((gfx.ScreenWidth-w)/2, config.KissBarY-2, (gfx.ScreenWidth+w)/2, config.KissBarY+10)).(*ebiten.Image).Fill(gfx.Palette[bar])
	f.DrawShadowed(dst, text, gfx.CentreX(text, 1), config.KissBarY, resurrect64.White, resurrect64.Ink, 1)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
