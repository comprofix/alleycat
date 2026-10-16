package scene

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
)

// shaker gives the screen a small shake each time Freddy takes a new knock
// (DESIGN §8, "Polish"). It is presentation only.
type shaker struct {
	hits int // Freddy's knock count last seen
	left int // master ticks of shake left

	buf *ebiten.Image // a render target for the shaken frame
}

// update follows Freddy's knock count. His count starts again when a room
// resets him, so only a rise starts a shake.
func (s *shaker) update(hits, ticks int) {
	s.left = max(s.left-ticks, 0)
	if hits > s.hits {
		s.left = config.HitShakeTicks
	}
	s.hits = hits
}

// offset is the sideways shift this frame: back and forth, easing off.
func (s *shaker) offset() int {
	if s.left == 0 {
		return 0
	}
	d := config.HitShakePx
	if s.left <= config.HitShakeTicks/3 {
		d = 1
	}
	if s.left%2 == 1 {
		d = -d
	}
	return d
}

// draw calls scene to draw the frame, shifted while shaking.
func (s *shaker) draw(dst *ebiten.Image, scene func(*ebiten.Image)) {
	dx := s.offset()
	if dx == 0 {
		scene(dst)
		return
	}
	if s.buf == nil {
		s.buf = ebiten.NewImage(gfx.ScreenWidth, gfx.ScreenHeight)
	}
	s.buf.Clear()
	scene(s.buf)
	dst.Fill(gfx.Palette[gfx.Night])
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(dx), 0)
	dst.DrawImage(s.buf, op)
}
