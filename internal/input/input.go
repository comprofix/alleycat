// Package input maps the keyboard to game actions with edge detection
// (DESIGN §5). Keyboard only; gamepad support is skipped for now.
package input

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Action is a logical input.
type Action int

const (
	Left Action = iota
	Right
	Up
	Down
	Act        // Alt: cheese hole, drink, drop present
	Pause      // Esc: "paws mode"
	Sound      // Ctrl+S
	Restart    // Ctrl+R
	Menu       // Ctrl+M
	Aspect     // F9
	Quit       // F10
	Fullscreen // Alt+Enter or F11
	numActions
)

// KeySource reports which keys are currently down.
type KeySource interface {
	AppendPressedKeys(keys []ebiten.Key) []ebiten.Key
}

// Ebiten reads the real keyboard.
type Ebiten struct{}

// AppendPressedKeys implements KeySource.
func (Ebiten) AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	return inpututil.AppendPressedKeys(keys)
}

// Input holds this frame's and last frame's state.
type Input struct {
	src        KeySource
	buf        []ebiten.Key
	keys, prev [ebiten.KeyMax + 1]bool
	cur, last  [numActions]bool
	// actBlocked is set when Enter is pressed while Alt is held (Alt+Enter is
	// fullscreen), so that Alt press does not count as Act until released.
	actBlocked bool
}

// New returns an Input reading from src.
func New(src KeySource) *Input {
	return &Input{src: src}
}

// Update samples the keyboard. Call once per game-loop frame.
func (in *Input) Update() {
	in.prev = in.keys
	in.last = in.cur
	in.keys = [ebiten.KeyMax + 1]bool{}
	in.buf = in.src.AppendPressedKeys(in.buf[:0])
	for _, k := range in.buf {
		if k >= 0 && k <= ebiten.KeyMax {
			in.keys[k] = true
		}
	}

	alt := in.keys[ebiten.KeyAlt] || in.keys[ebiten.KeyAltLeft] || in.keys[ebiten.KeyAltRight]
	ctrl := in.keys[ebiten.KeyControl] || in.keys[ebiten.KeyControlLeft] || in.keys[ebiten.KeyControlRight]
	enterNow := in.keyJustPressed(ebiten.KeyEnter) || in.keyJustPressed(ebiten.KeyNumpadEnter)

	if !alt {
		in.actBlocked = false
	} else if enterNow {
		in.actBlocked = true
	}

	in.cur = [numActions]bool{
		Left:       in.keys[ebiten.KeyArrowLeft],
		Right:      in.keys[ebiten.KeyArrowRight],
		Up:         in.keys[ebiten.KeyArrowUp],
		Down:       in.keys[ebiten.KeyArrowDown],
		Act:        alt && !in.actBlocked && !in.keys[ebiten.KeyEnter],
		Pause:      in.keys[ebiten.KeyEscape],
		Sound:      ctrl && in.keys[ebiten.KeyS],
		Restart:    ctrl && in.keys[ebiten.KeyR],
		Menu:       ctrl && in.keys[ebiten.KeyM],
		Aspect:     in.keys[ebiten.KeyF9],
		Quit:       in.keys[ebiten.KeyF10],
		Fullscreen: in.keys[ebiten.KeyF11] || (alt && in.keys[ebiten.KeyEnter]),
	}
}

// Held reports whether the action is active this frame.
func (in *Input) Held(a Action) bool { return in.cur[a] }

// JustPressed reports whether the action became active this frame.
func (in *Input) JustPressed(a Action) bool { return in.cur[a] && !in.last[a] }

// JustReleased reports whether the action stopped this frame.
func (in *Input) JustReleased(a Action) bool { return !in.cur[a] && in.last[a] }

// KeyJustPressed reports whether a specific key went down this frame.
func (in *Input) KeyJustPressed(k ebiten.Key) bool { return in.keyJustPressed(k) }

func (in *Input) keyJustPressed(k ebiten.Key) bool { return in.keys[k] && !in.prev[k] }

// AnyKeyJustPressed reports whether any key other than a modifier or a
// window key (F9–F11) went down this frame. Keys pressed with Ctrl held are
// commands (Ctrl+S, Ctrl+R, Ctrl+M), and Enter with Alt held is fullscreen,
// so neither counts.
func (in *Input) AnyKeyJustPressed() bool {
	if in.keys[ebiten.KeyControl] || in.keys[ebiten.KeyControlLeft] || in.keys[ebiten.KeyControlRight] {
		return false
	}
	alt := in.keys[ebiten.KeyAlt] || in.keys[ebiten.KeyAltLeft] || in.keys[ebiten.KeyAltRight]
	for k := ebiten.Key(0); k <= ebiten.KeyMax; k++ {
		if !in.keyJustPressed(k) {
			continue
		}
		switch k {
		case ebiten.KeyAlt, ebiten.KeyAltLeft, ebiten.KeyAltRight,
			ebiten.KeyControl, ebiten.KeyControlLeft, ebiten.KeyControlRight,
			ebiten.KeyShift, ebiten.KeyShiftLeft, ebiten.KeyShiftRight,
			ebiten.KeyMeta, ebiten.KeyMetaLeft, ebiten.KeyMetaRight,
			ebiten.KeyF9, ebiten.KeyF10, ebiten.KeyF11:
			continue
		}
		if alt && (k == ebiten.KeyEnter || k == ebiten.KeyNumpadEnter) {
			continue
		}
		return true
	}
	return false
}
