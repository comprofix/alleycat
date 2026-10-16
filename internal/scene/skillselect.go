package scene

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
)

// SkillSelect is the skill menu with help text (DESIGN §3.1). There is no
// joystick prompt. K/H/T/A or 1–4 start a game; Esc returns to the title.
type SkillSelect struct{}

// NewSkillSelect returns the skill menu.
func NewSkillSelect() *SkillSelect { return &SkillSelect{} }

var skillKeys = map[ebiten.Key]config.Skill{
	ebiten.KeyK: config.Kitten, ebiten.KeyDigit1: config.Kitten, ebiten.KeyNumpad1: config.Kitten,
	ebiten.KeyH: config.HouseCat, ebiten.KeyDigit2: config.HouseCat, ebiten.KeyNumpad2: config.HouseCat,
	ebiten.KeyT: config.Tomcat, ebiten.KeyDigit3: config.Tomcat, ebiten.KeyNumpad3: config.Tomcat,
	ebiten.KeyA: config.AlleyCat, ebiten.KeyDigit4: config.AlleyCat, ebiten.KeyNumpad4: config.AlleyCat,
}

// Update implements Scene.
func (s *SkillSelect) Update(ctx *Context) error {
	if ctx.Input.JustPressed(input.Pause) {
		ctx.Switch(NewTitle())
		return nil
	}
	for k, skill := range skillKeys {
		if ctx.Input.KeyJustPressed(k) {
			startGame(ctx, skill)
			return nil
		}
	}
	return nil
}

// startGame starts a new game at skill and remembers it for the title.
func startGame(ctx *Context, skill config.Skill) {
	ctx.Played, ctx.LastSkill = true, skill
	ctx.Session = session.New(skill)
	ctx.Switch(NewPlay(ctx))
}

// Draw implements Scene.
func (s *SkillSelect) Draw(ctx *Context, dst *ebiten.Image) {
	// A framed navy panel (M8g); the text keeps the plain font for reading.
	dst.Fill(gfx.Palette[gfx.Night])
	panel := func(x, y, w, h, idx int) {
		dst.SubImage(image.Rect(x, y, x+w, y+h)).(*ebiten.Image).Fill(gfx.Palette[idx])
	}
	panel(2, 2, gfx.ScreenWidth-4, gfx.ScreenHeight-4, gfx.Slate)
	panel(3, 3, gfx.ScreenWidth-6, gfx.ScreenHeight-6, gfx.Navy)
	f := ctx.Font
	y := 10
	line := func(text string, idx int) {
		if idx == gfx.Yellow {
			f.DrawShadowed(dst, text, 10, y, idx, gfx.Night, 1)
		} else {
			f.Draw(dst, text, 10, y, idx, 1)
		}
		y += 10
	}
	line("Choose your skill level:", gfx.Yellow)
	y += 2
	for i, sk := range config.Skills {
		line("  ("+string(sk.Letter())+") or "+string(rune('1'+i))+"  "+sk.String(), gfx.White)
	}
	y += 8
	line("During play:", gfx.Yellow)
	line("  Ctrl+S  sound on/off", gfx.Silver)
	line("  Ctrl+R  restart the game", gfx.Silver)
	line("  Ctrl+M  back to this menu", gfx.Silver)
	line("  Esc     paws (pause)", gfx.Silver)
	line("  F9 aspect  F10 quit  F11 fullscreen", gfx.Silver)
	y += 8
	line("Arrow keys move the cat.", gfx.Cyan)
	line("Alt does special actions.", gfx.Cyan)
	y += 6
	line("Esc returns to the title.", gfx.Slate)
}
