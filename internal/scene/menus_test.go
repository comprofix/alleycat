package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/alley"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/session"
	"github.com/comprofix/alleycat/internal/ui"
)

// menuContext is a context for driving the menus by key.
func menuContext() (*Context, *heldKeys) {
	ctx, keys := titleContext()
	set := config.DefaultSettings(config.Classic)
	ctx.Settings = &set
	return ctx, keys
}

// press holds keys for one tick of s and returns the scene s asked to
// change to, if any. The keys are let go of after, without running s again.
func press(ctx *Context, keys *heldKeys, s Scene, k ...ebiten.Key) Scene {
	keys.keys = k
	next := step(ctx, s)
	keys.keys = nil
	ctx.Input.Update()
	return next
}

// Play opens the skill menu on the skill last played.
func TestMainMenuPlay(t *testing.T) {
	ctx, keys := menuContext()
	ctx.LastSkill = config.Tomcat
	next := press(ctx, keys, NewMainMenu(), ebiten.KeyEnter)
	sk, ok := next.(*SkillSelect)
	if !ok {
		t.Fatalf("Play went to %T, want the skill menu", next)
	}
	if sk.menu.Cursor != int(config.Tomcat) {
		t.Errorf("skill menu opened on %d, want Tomcat", sk.menu.Cursor)
	}
}

// Each item of the main menu leads where it says, and Esc to the title.
func TestMainMenuItems(t *testing.T) {
	ctx, keys := menuContext()
	m := NewMainMenu()
	if _, ok := press(ctx, keys, m, ebiten.KeyEscape).(*Title); !ok {
		t.Error("Esc did not go back to the title")
	}
	press(ctx, keys, m, ebiten.KeyArrowDown)
	if _, ok := press(ctx, keys, m, ebiten.KeyEnter).(*Options); !ok {
		t.Error("the second item did not open the options")
	}
	press(ctx, keys, m, ebiten.KeyArrowDown)
	if _, ok := press(ctx, keys, m, ebiten.KeyEnter).(*Help); !ok {
		t.Error("the third item did not open the help")
	}
	press(ctx, keys, m, ebiten.KeyArrowDown)
	press(ctx, keys, m, ebiten.KeyEnter)
	if !ctx.Quitting() {
		t.Error("Quit did not quit")
	}
}

// The skill letters and numbers start a game at once, as on the
// original's menu (DESIGN §3.1).
func TestSkillMenuKeys(t *testing.T) {
	for k, want := range skillKeys {
		ctx, keys := menuContext()
		next := press(ctx, keys, NewSkillSelect(config.Kitten), k)
		if _, ok := next.(*Play); !ok {
			t.Fatalf("%v went to %T, want a game", k, next)
		}
		if ctx.Session.Skill != want || ctx.LastSkill != want {
			t.Errorf("%v started %v (last %v), want %v", k, ctx.Session.Skill, ctx.LastSkill, want)
		}
	}
}

// The skill menu works by arrows and Enter too, and Esc goes back.
func TestSkillMenuArrows(t *testing.T) {
	ctx, keys := menuContext()
	s := NewSkillSelect(config.Kitten)
	press(ctx, keys, s, ebiten.KeyArrowDown)
	press(ctx, keys, s, ebiten.KeyArrowDown)
	if _, ok := press(ctx, keys, s, ebiten.KeyEnter).(*Play); !ok || ctx.Session.Skill != config.Tomcat {
		t.Errorf("down twice and Enter did not start Tomcat")
	}
	if _, ok := press(ctx, keys, NewSkillSelect(config.Kitten), ebiten.KeyEscape).(*MainMenu); !ok {
		t.Error("Esc did not go back to the main menu")
	}
}

// Esc or P in a game cuts to paws mode over it, which stands the game
// still; either again cuts straight back (DESIGN §4).
func TestPawsMode(t *testing.T) {
	for _, k := range []ebiten.Key{ebiten.KeyEscape, ebiten.KeyP} {
		ctx, keys := menuContext()
		ctx.Session = session.New(config.Kitten)
		p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
		next := press(ctx, keys, p, k)
		paws, ok := next.(*Pause)
		if !ok || !ctx.cut || paws.under != p {
			t.Fatalf("%v went to %T (cut %v), want paws mode over the game", k, next, ctx.cut)
		}
		ticks := ctx.Session.LevelTicks
		if next := press(ctx, keys, paws, k); next != p || !ctx.cut {
			t.Fatalf("%v in paws mode went to %T (cut %v), want the game back", k, next, ctx.cut)
		}
		if ctx.Session.LevelTicks != ticks {
			t.Error("the game ran while in paws mode")
		}
	}
}

// Paws mode's menu: Resume goes back, Quit to menu leaves the game.
func TestPawsMenu(t *testing.T) {
	ctx, keys := menuContext()
	ctx.Session = session.New(config.Kitten)
	p := &Play{alley: alley.New(ctx.Session, ctx.RNG)}
	if next := press(ctx, keys, NewPause(p), ebiten.KeyEnter); next != p {
		t.Errorf("Resume went to %T, want the game", next)
	}
	paws := NewPause(p)
	press(ctx, keys, paws, ebiten.KeyArrowUp) // wraps round to the last item
	if _, ok := press(ctx, keys, paws, ebiten.KeyEnter).(*MainMenu); !ok || ctx.cut {
		t.Errorf("Quit to menu did not fade to the main menu")
	}
}

// The options change the settings at once and say so.
func TestOptions(t *testing.T) {
	ctx, keys := menuContext()
	o := NewOptions(func() Scene { return NewMainMenu() })
	set := ctx.Settings
	press(ctx, keys, o, ebiten.KeyArrowRight) // Scale
	if set.ScaleMode != config.ScaleCRT || !ctx.TakeSettingsChanged() {
		t.Errorf("scale %q, want CRT and the change noted", set.ScaleMode)
	}
	press(ctx, keys, o, ebiten.KeyArrowDown)
	press(ctx, keys, o, ebiten.KeyArrowDown)
	press(ctx, keys, o, ebiten.KeyEnter) // Sound
	if set.SoundOn || !ctx.TakeSettingsChanged() {
		t.Error("sound not switched off")
	}
	press(ctx, keys, o, ebiten.KeyArrowDown) // skips the volume and softening while sound is off
	if o.menu.Cursor != 5 {
		t.Errorf("cursor on %d with sound off, want Back", o.menu.Cursor)
	}
	if _, ok := press(ctx, keys, o, ebiten.KeyEnter).(*MainMenu); !ok {
		t.Error("Back did not go back")
	}
}

// Any key leaves the help.
func TestHelp(t *testing.T) {
	ctx, keys := menuContext()
	if _, ok := press(ctx, keys, NewHelp(), ebiten.KeyQ).(*MainMenu); !ok {
		t.Error("a key did not leave the help")
	}
}

// Every menu moves and chooses with the menu sounds (DESIGN §7.4).
func TestMenusHaveSounds(t *testing.T) {
	ctx, _ := menuContext()
	main, skill := NewMainMenu(), NewSkillSelect(config.Kitten)
	opts, paws := NewOptions(func() Scene { return main }), NewPause(main)
	for _, c := range []struct {
		s    Scene
		menu *ui.Menu
	}{{main, &main.menu}, {skill, &skill.menu}, {opts, &opts.menu}, {paws, &paws.menu}} {
		step(ctx, c.s)
		if c.menu.OnMove == nil || c.menu.OnConfirm == nil {
			t.Errorf("%T has no menu sounds", c.s)
		}
	}
}
