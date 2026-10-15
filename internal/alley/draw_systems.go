package alley

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/session"
)

// HUD positions on the fence (DESIGN §3.8, "HUD positions").
const (
	hudHighX, hudHighY   = 40, 120
	hudLivesX, hudLivesY = 256, 116
	hudScoreX, hudScoreY = 240, 128
	hudSkillX, hudSkillY = 8, 160
)

// graffiti is the decorative writing across the fence (DESIGN §3.3, kept
// by the owner as a nod to the original).
var graffiti = []struct {
	text string
	x, y int
}{
	{"LOVE", 116, 134},
	{"THEM", 160, 116},
	{"MOUSIES", 200, 146},
}

var thrownSprites = [...]string{"thrown_boot", "thrown_phone", "thrown_bottle", "thrown_pot"}

// rows returns the top n rows of img (from row from).
func rows(img *ebiten.Image, from, n int) *ebiten.Image {
	b := img.Bounds()
	return img.SubImage(image.Rect(b.Min.X, b.Min.Y+from, b.Max.X, b.Min.Y+from+n)).(*ebiten.Image)
}

func (r *Renderer) drawGraffiti(dst *ebiten.Image) {
	for _, g := range graffiti {
		r.font.Draw(dst, g.text, g.x, g.y, r.scheme.Outline, 1)
	}
}

func (r *Renderer) drawHUD(dst *ebiten.Image, a *Alley, high int) {
	s, ink := a.Session, r.scheme.Outline
	r.font.Draw(dst, "HI!", hudHighX-gfx.Width("HI! "), hudHighY, ink, 1)
	r.font.Draw(dst, session.FormatScore(high), hudHighX, hudHighY, ink, 1)
	r.font.Draw(dst, fmt.Sprintf("%d CAT", s.Lives), hudLivesX, hudLivesY, ink, 1)
	r.font.Draw(dst, session.FormatScore(s.Score), hudScoreX, hudScoreY, ink, 1)
	r.font.Draw(dst, string(s.Skill.Letter()), hudSkillX, hudSkillY, ink, 1)
}

// drawWindow shows the open window: darkness, or Felicia, rising from the
// bottom of the glass as it opens.
func (r *Renderer) drawWindow(dst *ebiten.Image, w *Window) {
	n := w.Openness(w.Index)
	if n == 0 {
		return
	}
	o := WindowRect(w.Index)
	glassTop := o.Y + 1
	y := glassTop + config.WindowPhaseHold - 1 - n
	fill(dst, o.X-1, y, o.W+2, n, r.scheme.Outline)
	if w.Felicia {
		img := r.sheet.Sprite("felicia").Frame(0)
		h := img.Bounds().Dy()
		from := max(h-n, 0)
		entity.DrawAt(dst, rows(img, from, h-from), o.X, y, false)
	}
}

func (r *Renderer) drawMice(dst *ebiten.Image, a *Alley) {
	for i, m := range a.Mice {
		y := config.MouseY[i]
		if m.Caught {
			r.font.Draw(dst, fmt.Sprint(config.MousePoints[i]), m.X, y, gfx.White, 1)
			continue
		}
		img := r.sheet.Sprite("mouse_stand").Frame(0)
		if m.Dir != 0 {
			img = r.sheet.Sprite("mouse_walk").Frame(m.Frame)
		}
		entity.DrawAt(dst, img, m.X, y, m.Dir < 0)
	}
}

func (r *Renderer) drawThrown(dst *ebiten.Image, t *Thrown) {
	if !t.Active {
		return
	}
	img := r.sheet.Sprite(thrownSprites[t.Kind]).Frame(0)
	entity.DrawAt(dst, rows(img, 0, t.Visible()), t.X, t.Y, t.Dir < 0)
}

func (r *Renderer) drawTrashCat(dst *ebiten.Image, t *TrashCat) {
	if t.Phase == 0 || t.H <= 0 {
		return
	}
	entity.DrawAt(dst, rows(r.sheet.Sprite("trash_cat").Frame(0), 0, t.H), t.X, t.Top, false)
}

// drawImpact shows the impact word while the game is frozen.
func (r *Renderer) drawImpact(dst *ebiten.Image, im Impact) {
	var word string
	var bg, ink int
	switch im.Kind {
	case Bonk:
		// A ragged splat with slanting lettering, as in the original
		// (checked against the HD video at 10:19.4).
		entity.DrawAt(dst, r.sheet.Sprite("bonk").Frame(0), im.X, im.Y, false)
		return
	case Squeak:
		word, bg, ink = "SQUEEK!", gfx.White, gfx.Plum
	default:
		return
	}
	w := gfx.Width(word)
	x := min(im.X, gfx.ScreenWidth-w-4)
	fill(dst, x, im.Y, w+4, 12, bg)
	r.font.Draw(dst, word, x+2, im.Y+2, ink, 1)
}
