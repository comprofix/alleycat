package scene

import (
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
			ctx.Session = session.New(skill)
			ctx.Switch(NewPlay())
			return nil
		}
	}
	return nil
}

// Draw implements Scene.
func (s *SkillSelect) Draw(ctx *Context, dst *ebiten.Image) {
	dst.Fill(gfx.Palette[gfx.Night])
	f := ctx.Font
	y := 8
	line := func(text string, idx int) {
		f.Draw(dst, text, 8, y, idx, 1)
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
