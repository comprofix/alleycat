package input

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type fakeKeys struct{ down []ebiten.Key }

func (f *fakeKeys) AppendPressedKeys(keys []ebiten.Key) []ebiten.Key {
	return append(keys, f.down...)
}

func step(in *Input, f *fakeKeys, keys ...ebiten.Key) {
	f.down = keys
	in.Update()
}

func TestEdges(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyArrowLeft)
	if !in.JustPressed(Left) || !in.Held(Left) {
		t.Fatal("Left should be just pressed and held")
	}
	step(in, f, ebiten.KeyArrowLeft)
	if in.JustPressed(Left) || !in.Held(Left) {
		t.Fatal("Left should be held, not just pressed")
	}
	step(in, f)
	if !in.JustReleased(Left) || in.Held(Left) {
		t.Fatal("Left should be just released")
	}
}

func TestAltIsAct(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyAlt, ebiten.KeyAltLeft)
	if !in.JustPressed(Act) {
		t.Fatal("Alt should trigger Act")
	}
}

func TestAltEnterIsFullscreenAndBlocksAct(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyAlt)
	step(in, f, ebiten.KeyAlt, ebiten.KeyEnter)
	if !in.JustPressed(Fullscreen) {
		t.Fatal("Alt+Enter should trigger Fullscreen")
	}
	if in.Held(Act) {
		t.Fatal("Act must not be held during Alt+Enter")
	}
	step(in, f, ebiten.KeyAlt) // Enter released, Alt still held
	if in.Held(Act) {
		t.Fatal("Act stays blocked until Alt is released")
	}
	step(in, f)
	step(in, f, ebiten.KeyAlt)
	if !in.JustPressed(Act) {
		t.Fatal("a fresh Alt press should trigger Act again")
	}
}

func TestCtrlCombos(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyS)
	if in.Held(Sound) {
		t.Fatal("S alone is not Ctrl+S")
	}
	step(in, f, ebiten.KeyControl, ebiten.KeyS)
	if !in.JustPressed(Sound) {
		t.Fatal("Ctrl+S should toggle sound")
	}
	step(in, f, ebiten.KeyControl, ebiten.KeyM)
	if !in.JustPressed(Menu) {
		t.Fatal("Ctrl+M should be Menu")
	}
}

func TestWindowKeys(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyF9)
	if !in.JustPressed(Aspect) {
		t.Fatal("F9 should be Aspect")
	}
	step(in, f, ebiten.KeyF10)
	if !in.JustPressed(Quit) {
		t.Fatal("F10 should be Quit")
	}
	step(in, f, ebiten.KeyF11)
	if !in.JustPressed(Fullscreen) {
		t.Fatal("F11 should be Fullscreen")
	}
}

func TestAnyKeyIgnoresModifiersAndWindowKeys(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	step(in, f, ebiten.KeyShift, ebiten.KeyF9)
	if in.AnyKeyJustPressed() {
		t.Fatal("modifiers and F9 should not count as any key")
	}
	step(in, f, ebiten.KeySpace)
	if !in.AnyKeyJustPressed() {
		t.Fatal("Space should count as any key")
	}
}

// Ctrl combinations are commands, and Alt+Enter is fullscreen: on the title
// screen, Ctrl+S must toggle sound, not start the game.
func TestAnyKeyIgnoresCommands(t *testing.T) {
	f := &fakeKeys{}
	in := New(f)
	for _, keys := range [][]ebiten.Key{
		{ebiten.KeyControlLeft, ebiten.KeyS},
		{ebiten.KeyControlRight, ebiten.KeyR},
		{ebiten.KeyControl, ebiten.KeyM},
		{ebiten.KeyAltLeft, ebiten.KeyEnter},
	} {
		step(in, f) // all keys up
		step(in, f, keys...)
		if in.AnyKeyJustPressed() {
			t.Errorf("%v counted as any key", keys)
		}
	}
	step(in, f)
	step(in, f, ebiten.KeyEnter)
	if !in.AnyKeyJustPressed() {
		t.Error("Enter alone should count as any key")
	}
}
