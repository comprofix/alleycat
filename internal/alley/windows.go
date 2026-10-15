package alley

import (
	"math/rand/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/entity"
	"github.com/comprofix/alleycat/internal/physics"
)

// Window is the one window event that can be under way (DESIGN §3.8,
// "Alley systems"). Phase counts down from 29: 29–16 opening, 15 fully open,
// 14–1 closing, 0 idle.
type Window struct {
	Index   int // 0–11; column = Index mod 4, row = Index / 4
	Phase   int
	Felicia bool // Felicia is in this window
	forced  bool // opened because Freddy camped on the fence
	hold    int
	held    int // tick the hold began
	throws  int
	// HeldByMouse is set while a mouse in front of the window holds it.
	HeldByMouse bool
}

// WindowRect is the opening of window i.
func WindowRect(i int) physics.Rect {
	return physics.Rect{
		X: config.WindowX0 + i%config.WindowCols*config.WindowPitchX,
		Y: config.WindowY0 + i/config.WindowCols*config.WindowPitchY,
		W: config.WindowW, H: config.WindowH,
	}
}

// Openness is how far window i is open, from 0 (closed) to 14 (fully open).
func (w *Window) Openness(i int) int {
	switch {
	case w.Phase == 0 || i != w.Index:
		return 0
	case w.Phase > config.WindowPhaseHold:
		return config.WindowPhaseStart + 1 - w.Phase
	case w.Phase == config.WindowPhaseHold:
		return config.WindowPhaseHold - 1
	}
	return w.Phase
}

func (a *Alley) updateWindow(rng *rand.Rand) {
	w := &a.Window
	w.HeldByMouse = false
	if w.Phase != 0 && a.windowTouch() {
		return
	}
	if a.Thrown.Active {
		return
	}
	if w.Phase == 0 {
		if !a.openWindow(rng) {
			return
		}
		if a.windowTouch() {
			return
		}
	}
	if a.mouseHoldsWindow() {
		w.HeldByMouse = true
		return
	}
	if w.Phase == config.WindowPhaseHoldEnd {
		w.held = a.Tick
	}
	if w.Phase == config.WindowPhaseHold && a.Tick-w.held < w.hold {
		if w.throws > 0 && !w.Felicia {
			w.throws--
			a.launch(rng)
		}
		return
	}
	w.Phase--
}

// openWindow picks a window to open, if any.
func (a *Alley) openWindow(rng *rand.Rand) bool {
	f := a.Freddy
	if f.Y > config.WindowMaxFreddyY {
		return false
	}
	felicia := a.Session.FeliciaWaiting
	forced := f.Depth == entity.OnFence && !felicia && a.Tick-a.fenceSince >= config.FenceCampTicks
	for {
		var i int
		if forced {
			i = rng.IntN(config.WindowCols) // the top row
		} else {
			i = rng.IntN(config.WindowRoll)
		}
		if i >= config.WindowCols*config.WindowRows {
			return false
		}
		if WindowRect(i).Overlaps(f.HitBox()) {
			continue
		}
		a.Window = Window{
			Index: i, Phase: config.WindowPhaseStart, Felicia: felicia,
			forced: forced, hold: a.table.WindowHoldTicks, throws: 1,
		}
		return true
	}
}

// windowTouch handles Freddy touching the open window: he goes in if he is
// falling into it while it is open enough. Touching it at all holds it.
func (a *Alley) windowTouch() bool {
	f, w := a.Freddy, &a.Window
	if !WindowRect(w.Index).Overlaps(f.HitBox()) {
		return false
	}
	if f.Arc.Dir == physics.Falling && !f.Knocked && f.Y < config.WindowEnterMaxY &&
		w.Phase >= config.WindowEnterMinPhase && w.Phase <= config.WindowEnterMaxPhase {
		if a.Session.FeliciaWaiting {
			a.Exit = ToHeart
		} else {
			a.Exit = ToRoom
		}
	}
	return true
}

// mouseHoldsWindow: the middle-line mouse in front of a top-row window, or
// the bottom-line mouse in front of a middle-row window, holds it still.
func (a *Alley) mouseHoldsWindow() bool {
	row := a.Window.Index / config.WindowCols
	if row >= 2 {
		return false
	}
	mx := a.Mice[row+1].X
	wx := WindowRect(a.Window.Index).X
	return mx+config.WindowMouseBefore >= wx && max(mx-config.WindowMouseAfter, 0) <= wx
}
