package scene

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/audio"
)

type stub struct {
	updates int
	next    Scene
	cut     bool // change to next with a cut, not a fade
}

func (s *stub) Update(ctx *Context) error {
	s.updates++
	if s.next != nil {
		if s.cut {
			ctx.Cut(s.next)
		} else {
			ctx.Switch(s.next)
		}
		s.next = nil
	}
	return nil
}

func (s *stub) Draw(*Context, *ebiten.Image) {}

func TestManagerFadesInThenRuns(t *testing.T) {
	a := &stub{}
	m := NewManager(a)
	ctx := &Context{}
	if m.Darkness() != 1 {
		t.Fatalf("should start fully dark, got %v", m.Darkness())
	}
	for range FadeTicks {
		m.Update(ctx)
	}
	if a.updates != 0 || m.Darkness() != 0 {
		t.Fatalf("after fade-in: updates=%d darkness=%v", a.updates, m.Darkness())
	}
	m.Update(ctx)
	if a.updates != 1 {
		t.Fatal("scene should update once idle")
	}
}

func TestManagerSwitchFadesOutAndIn(t *testing.T) {
	b := &stub{}
	a := &stub{next: b}
	m := NewManager(a)
	ctx := &Context{}
	for range FadeTicks {
		m.Update(ctx)
	}
	m.Update(ctx) // a requests the switch
	if m.Current() != a {
		t.Fatal("should still show a while fading out")
	}
	for range FadeTicks {
		m.Update(ctx)
	}
	if m.Current() != b || m.Darkness() != 1 {
		t.Fatalf("after fade-out: current b=%v darkness=%v", m.Current() == b, m.Darkness())
	}
	for range FadeTicks {
		m.Update(ctx)
	}
	if b.updates != 0 {
		t.Fatal("b must not update during fade-in")
	}
	m.Update(ctx)
	if b.updates != 1 || a.updates != 1 {
		t.Fatalf("updates: a=%d b=%d", a.updates, b.updates)
	}
}

// A cut changes scene at once, with no fade, and leaves the sound alone:
// the pause menu opens and closes over a game this way.
func TestManagerCut(t *testing.T) {
	b := &stub{}
	a := &stub{next: b, cut: true}
	m := NewManager(a)
	ctx := &Context{Audio: audio.NewSynth()}
	for range FadeTicks {
		m.Update(ctx)
	}
	ctx.Audio.PlayMusic([]audio.Note{{Hz: 500, Dur: 10 * time.Second}}, false)
	m.Update(ctx) // a asks for the cut
	if m.Current() != b || m.Darkness() != 0 {
		t.Fatalf("after the cut: current b=%v darkness=%v", m.Current() == b, m.Darkness())
	}
	m.Update(ctx)
	if b.updates != 1 {
		t.Fatalf("b updated %d times, want 1 straight after the cut", b.updates)
	}
	buf := make([]byte, 4*4410)
	ctx.Audio.Read(buf)
	if maxJump(buf) < 1000 {
		t.Error("the cut stopped the music")
	}
}

// Paws mode leaves the game's sounds playing, as the Game Skeleton's pause
// does (DESIGN §7).
func TestPawsLeavesSoundPlaying(t *testing.T) {
	under := &stub{}
	m := NewManager(under)
	ctx := &Context{Audio: audio.NewSynth()}
	for range FadeTicks {
		m.Update(ctx)
	}
	ctx.Audio.PlayMusic([]audio.Note{{Hz: 500, Dur: 10 * time.Second}}, false)
	pause(ctx, under)
	if _, ok := ctx.next.(*Pause); !ok {
		t.Fatalf("pause asked for %T, want paws mode", ctx.next)
	}
	buf := make([]byte, 4*4410)
	ctx.Audio.Read(buf)
	if maxJump(buf) < 1000 {
		t.Error("paws mode silenced the music")
	}
}

// A scene change silences every sound, so nothing carries over, such as a
// dog fight into the title after the last life (DESIGN §7.1). Only the
// screen-change figure plays through the fade, then all is quiet (§7.4).
func TestManagerSwitchStopsSound(t *testing.T) {
	a := &stub{next: &stub{}}
	m := NewManager(a)
	ctx := &Context{Audio: audio.NewSynth()}
	for range FadeTicks {
		m.Update(ctx)
	}
	ctx.Audio.PlayMusic([]audio.Note{{Hz: 500, Dur: 10 * time.Second}}, false)
	ctx.Audio.PlaySFX([]audio.Note{{Hz: 2000, Dur: 10 * time.Second}})
	m.Update(ctx) // a requests the switch

	figure := make([]byte, 4*4410*6) // the 0.5 s figure, and a little more
	ctx.Audio.Read(figure)
	if peak := maxJump(figure); peak < 1000 {
		t.Error("no screen-change figure during the fade")
	}
	buf := make([]byte, 4*4410) // the next 100 ms
	ctx.Audio.Read(buf)
	for i := 4; i+4 <= len(buf); i += 4 {
		prev := int16(binary.LittleEndian.Uint16(buf[i-4:]))
		cur := int16(binary.LittleEndian.Uint16(buf[i:]))
		if d := int(cur) - int(prev); d > 100 || d < -100 {
			t.Fatalf("sample %d jumps by %d: a tone is still playing after the switch", i/4, d)
		}
	}
}

// maxJump is the largest sample-to-sample step in 16-bit stereo audio.
func maxJump(buf []byte) int {
	best := 0
	for i := 4; i+4 <= len(buf); i += 4 {
		d := int(int16(binary.LittleEndian.Uint16(buf[i:]))) - int(int16(binary.LittleEndian.Uint16(buf[i-4:])))
		best = max(best, d, -d)
	}
	return best
}
