package scene

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/gfx"
	"github.com/comprofix/alleycat/internal/input"
	"github.com/comprofix/alleycat/internal/session"
	"github.com/comprofix/alleycat/internal/ui"
)

// The menus are the Game Skeleton's, as Night Crossing has them (DESIGN
// §4): a heading in the large font over a panel, worked by key or mouse.

// menuPad is the margin inside menu panels: one line of the small font.
func menuPad(f gfx.Fonts) int { return f.Small.LineHeight() }

// drawMenuScreen draws a full-screen menu: a heading in the large font and
// the menu in a panel below it, with an optional line of help under that.
func drawMenuScreen(ctx *Context, dst *ebiten.Image, heading string, m *ui.Menu, help string) {
	sc, f := ctx.Scheme, ctx.Fonts
	dst.Fill(gfx.RGBA(sc.Backdrop))
	ui.TextCentred(dst, f.Large, heading, f.Large.LineHeight(), sc.Title, sc.Shadow)
	drawMenuPanel(ctx, dst, m, menuTop(f))
	drawHelpLine(ctx, dst, help)
}

// drawHelpLine draws a line of help along the bottom of a menu screen.
func drawHelpLine(ctx *Context, dst *ebiten.Image, help string) {
	if help == "" {
		return
	}
	sc, f := ctx.Scheme, ctx.Fonts
	ui.TextCentred(dst, f.Small, help, config.Project.ScreenHeight-2*f.Small.LineHeight(), sc.Dim, sc.Shadow)
}

// menuTop is the top of a menu screen's menu: the panel's border sits a
// line gap below the heading.
func menuTop(f gfx.Fonts) int { return 2*f.Large.LineHeight() + ui.LineGap + menuPad(f) }

// updateMenu runs m for one tick, placed as drawMenuPanel draws it with its
// top at y, so the mouse works. Headless tests have no fonts and no mouse.
func updateMenu(ctx *Context, m *ui.Menu, y func(gfx.Fonts) int) {
	if f := ctx.Fonts; f.Small != nil && f.Large != nil {
		m.PlaceCentred(f.Small, y(f))
	}
	m.Update(ctx.Input)
}

// menuSounds gives m the menus' sounds, as Night Crossing's (DESIGN §7.4).
func menuSounds(ctx *Context, m *ui.Menu) {
	m.OnMove = func() { ctx.Audio.PlayMenu(audio.MenuMove()) }
	m.OnConfirm = func() { ctx.Audio.PlayMenu(audio.MenuConfirm()) }
}

// drawMenuPanel draws m in a panel centred across the screen, its top at y.
func drawMenuPanel(ctx *Context, dst *ebiten.Image, m *ui.Menu, y int) {
	sc, f := ctx.Scheme, ctx.Fonts
	w, h := m.Size(f.Small)
	ui.Panel(dst, ui.PanelAround(w, h, y, menuPad(f)), sc.Panel, sc.Border)
	m.DrawCentred(dst, f.Small, y, sc)
}

// MainMenu offers Play, Options, Help and Quit.
type MainMenu struct {
	menu ui.Menu
}

// NewMainMenu returns the main menu.
func NewMainMenu() *MainMenu { return &MainMenu{} }

// Update implements Scene.
func (s *MainMenu) Update(ctx *Context) error {
	if s.menu.Items == nil {
		s.menu.Items = []ui.Item{
			{Label: "Play", Confirm: func() { ctx.Switch(NewSkillSelect(ctx.LastSkill)) }},
			{Label: "Options", Confirm: func() { ctx.Switch(NewOptions(func() Scene { return NewMainMenu() })) }},
			{Label: "Help", Confirm: func() { ctx.Switch(NewHelp()) }},
			{Label: "Quit", Confirm: ctx.Quit},
		}
		menuSounds(ctx, &s.menu)
	}
	if ctx.Input.JustPressed(input.Back) {
		ctx.Switch(NewTitle())
		return nil
	}
	updateMenu(ctx, &s.menu, menuTop)
	return nil
}

func (*MainMenu) native()    {}
func (*SkillSelect) native() {}
func (*Options) native()     {}
func (*Help) native()        {}
func (*Pause) native()       {}

// Draw implements Scene.
func (s *MainMenu) Draw(ctx *Context, dst *ebiten.Image) {
	drawMenuScreen(ctx, dst, "ALLEY CAT", &s.menu, "Enter: choose   Esc: back")
}

// SkillSelect chooses the skill level and starts a game (DESIGN §3.1).
// K, H, T and A or 1–4 start one at once, as on the original's menu.
type SkillSelect struct {
	menu ui.Menu
}

// NewSkillSelect returns the skill menu with the cursor on skill.
func NewSkillSelect(skill config.Skill) *SkillSelect {
	return &SkillSelect{menu: ui.Menu{Cursor: int(skill)}}
}

var skillKeys = map[ebiten.Key]config.Skill{
	ebiten.KeyK: config.Kitten, ebiten.KeyDigit1: config.Kitten, ebiten.KeyNumpad1: config.Kitten,
	ebiten.KeyH: config.HouseCat, ebiten.KeyDigit2: config.HouseCat, ebiten.KeyNumpad2: config.HouseCat,
	ebiten.KeyT: config.Tomcat, ebiten.KeyDigit3: config.Tomcat, ebiten.KeyNumpad3: config.Tomcat,
	ebiten.KeyA: config.AlleyCat, ebiten.KeyDigit4: config.AlleyCat, ebiten.KeyNumpad4: config.AlleyCat,
}

// Update implements Scene.
func (s *SkillSelect) Update(ctx *Context) error {
	if s.menu.Items == nil {
		for _, sk := range config.Skills {
			s.menu.Items = append(s.menu.Items, ui.Item{Label: sk.String(), Confirm: func() { startGame(ctx, sk) }})
		}
		menuSounds(ctx, &s.menu)
	}
	if ctx.Input.JustPressed(input.Back) {
		ctx.Switch(NewMainMenu())
		return nil
	}
	for k, skill := range skillKeys {
		if ctx.Input.KeyJustPressed(k) {
			ctx.Audio.PlayMenu(audio.MenuConfirm())
			startGame(ctx, skill)
			return nil
		}
	}
	updateMenu(ctx, &s.menu, menuTop)
	return nil
}

// startGame starts a new game at skill and remembers it for the menu.
func startGame(ctx *Context, skill config.Skill) {
	ctx.LastSkill = skill
	ctx.Session = session.New(skill)
	ctx.Switch(NewPlay(ctx))
}

// Draw implements Scene.
func (s *SkillSelect) Draw(ctx *Context, dst *ebiten.Image) {
	drawMenuScreen(ctx, dst, "SKILL", &s.menu, "K, H, T, A or 1-4 start at once")
}

// Options changes the saved settings (DESIGN §9). Changes apply at once.
// Back returns to the scene back makes.
type Options struct {
	back func() Scene
	menu ui.Menu
}

// NewOptions returns the options menu; leaving it switches to back().
func NewOptions(back func() Scene) *Options { return &Options{back: back} }

// VolumeSteps is how many steps the volume has.
const VolumeSteps = 10

// Update implements Scene.
func (s *Options) Update(ctx *Context) error {
	if s.menu.Items == nil {
		s.menu.Items = s.items(ctx)
		menuSounds(ctx, &s.menu)
	}
	if ctx.Input.JustPressed(input.Back) {
		ctx.Switch(s.back())
		return nil
	}
	updateMenu(ctx, &s.menu, menuTop)
	return nil
}

func (s *Options) items(ctx *Context) []ui.Item {
	set := ctx.Settings
	toggle := func(v *bool) func(int) {
		return func(int) { *v = !*v; ctx.SettingsChanged() }
	}
	soundOff := func() bool { return !set.SoundOn }
	return []ui.Item{
		{Label: "Look", Value: func() string { return set.Look.Name() },
			Adjust: func(d int) {
				set.Look = next(config.Looks, set.Look, d)
				ctx.SettingsChanged()
			}},
		// The new look has one scale mode, so Scale is dimmed under it.
		{Label: "Scale", Value: func() string { return ScaleModeName(set.Mode()) },
			DisabledIf: func() bool { return len(set.Look.Preset().ScaleModes) < 2 },
			Adjust: func(d int) {
				set.ScaleMode = next(set.Look.Preset().ScaleModes, set.ScaleMode, d)
				ctx.SettingsChanged()
			}},
		{Label: "Fullscreen", Value: ui.OnOff(&set.Fullscreen), Adjust: toggle(&set.Fullscreen)},
		{Label: "Sound", Value: ui.OnOff(&set.SoundOn), Adjust: toggle(&set.SoundOn)},
		{
			// The mouse sets the volume to the nearest step.
			Label:      "Volume",
			Slider:     func() float64 { return set.Volume },
			DisabledIf: soundOff,
			SetSlider: func(x float64) {
				if n := float64(steps(x)) / VolumeSteps; n != set.Volume {
					set.Volume = n
					ctx.SettingsChanged()
				}
			},
			Adjust: func(d int) {
				set.Volume = float64(min(max(steps(set.Volume)+d, 0), VolumeSteps)) / VolumeSteps
				ctx.SettingsChanged()
			},
		},
		{Label: "Soften beeper", Value: ui.OnOff(&set.Soften), DisabledIf: soundOff, Adjust: toggle(&set.Soften)},
		{Label: "Back", Confirm: func() { ctx.Switch(s.back()) }},
	}
}

// next is the item d places after v in list, wrapping round; the first if
// v is not in it.
func next[T comparable](list []T, v T, d int) T {
	i := 0
	for j, x := range list {
		if x == v {
			i = j + d
		}
	}
	return list[(i%len(list)+len(list))%len(list)]
}

// steps is a 0–1 volume as a whole number of steps.
func steps(v float64) int { return int(math.Round(v * VolumeSteps)) }

// ScaleModeName is a scale mode as the options menu shows it.
func ScaleModeName(m config.ScaleMode) string {
	switch m {
	case config.ScaleCRT:
		return "CRT 4:3"
	case config.ScaleSmooth:
		return "Smooth"
	default:
		return "Whole pixels"
	}
}

// Draw implements Scene.
func (s *Options) Draw(ctx *Context, dst *ebiten.Image) {
	drawMenuScreen(ctx, dst, "OPTIONS", &s.menu, "Left/Right: change   Esc: back")
}

// Help lists the keys (DESIGN §5). Any key or a click goes back.
type Help struct{}

// NewHelp returns the help screen.
func NewHelp() *Help { return &Help{} }

// helpLines are the keys and what they do.
var helpLines = [][2]string{
	{"Arrow keys", "move the cat"},
	{"Alt", "special actions"},
	{"Esc or P", "paws (pause)"},
	{"Ctrl+S", "sound on/off"},
	{"Ctrl+R", "restart the game"},
	{"Ctrl+M", "back to the menu"},
	{"F9", "scale"},
	{"F10", "quit"},
	{"F11", "fullscreen"},
}

// helpKeyCells is the width of the keys' column, in characters.
const helpKeyCells = 12

// Update implements Scene.
func (h *Help) Update(ctx *Context) error {
	if ctx.Input.AnyJustPressed() {
		ctx.Audio.PlayMenu(audio.MenuConfirm())
		ctx.Switch(NewMainMenu())
	}
	return nil
}

// Draw implements Scene.
func (h *Help) Draw(ctx *Context, dst *ebiten.Image) {
	sc, f := ctx.Scheme, ctx.Fonts
	dst.Fill(gfx.RGBA(sc.Backdrop))
	ui.TextCentred(dst, f.Large, "HELP", f.Large.LineHeight(), sc.Title, sc.Shadow)
	cell, line := f.Small.CellWidth(), f.Small.LineHeight()+ui.LineGap
	w := 0
	for _, l := range helpLines {
		w = max(w, (helpKeyCells+len(l[1]))*cell)
	}
	y := menuTop(f)
	ui.Panel(dst, ui.PanelAround(w, len(helpLines)*line-ui.LineGap, y, menuPad(f)), sc.Panel, sc.Border)
	x := (config.Project.ScreenWidth - w) / 2
	for i, l := range helpLines {
		f.Small.DrawShadowed(dst, l[0], x, y+i*line, sc.Highlight, sc.Shadow)
		f.Small.DrawShadowed(dst, l[1], x+helpKeyCells*cell, y+i*line, sc.Text, sc.Shadow)
	}
	drawHelpLine(ctx, dst, "Press any key")
}

// Pause is paws mode (DESIGN §4): the game stands still, dimmed, under a
// menu. Esc or P again resumes it.
type Pause struct {
	under Scene
	menu  ui.Menu
}

// NewPause pauses under.
func NewPause(under Scene) *Pause { return &Pause{under: under} }

// pause stops the game under and opens paws mode over it.
func pause(ctx *Context, under Scene) {
	ctx.Audio.SetPaused(true)
	ctx.Cut(NewPause(under))
}

// Update implements Scene.
func (p *Pause) Update(ctx *Context) error {
	resume := func() {
		ctx.Audio.SetPaused(false)
		ctx.Cut(p.under)
	}
	if p.menu.Items == nil {
		p.menu.Items = []ui.Item{
			{Label: "Resume", Confirm: resume},
			{Label: "Options", Confirm: func() { ctx.Switch(NewOptions(func() Scene { return p })) }},
			{Label: "Quit to menu", Confirm: func() { ctx.Switch(NewMainMenu()) }},
		}
		menuSounds(ctx, &p.menu)
	}
	if ctx.Input.JustPressed(input.Back) || ctx.Input.JustPressed(input.Pause) {
		resume()
		return nil
	}
	updateMenu(ctx, &p.menu, pauseMenuTop)
	return nil
}

// PauseDim is how far the game is dimmed under paws mode, 0–1.
const PauseDim = 0.6

// pauseTop is the top of the pause screen's heading, and pauseMenuTop the
// top of its menu, under the heading: the two sit around the middle.
func pauseTop(f gfx.Fonts) int     { return config.Project.ScreenHeight/2 - 3*f.Large.LineHeight()/2 }
func pauseMenuTop(f gfx.Fonts) int { return pauseTop(f) + 2*f.Large.LineHeight() }

// Draw implements Scene.
func (p *Pause) Draw(ctx *Context, dst *ebiten.Image) {
	sc, f := ctx.Scheme, ctx.Fonts
	drawScene(ctx, p.under, dst)
	Dim(dst, sc.Shadow, PauseDim)
	ui.TextCentred(dst, f.Large, "PAWS", pauseTop(f), sc.Title, sc.Shadow)
	drawMenuPanel(ctx, dst, &p.menu, pauseMenuTop(f))
}
