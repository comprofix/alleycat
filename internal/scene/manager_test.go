package scene

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type stub struct {
	updates int
	next    Scene
}

func (s *stub) Update(ctx *Context) error {
	s.updates++
	if s.next != nil {
		ctx.Switch(s.next)
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
	for range FadeFrames {
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
	for range FadeFrames {
		m.Update(ctx)
	}
	m.Update(ctx) // a requests the switch
	if m.Current() != a {
		t.Fatal("should still show a while fading out")
	}
	for range FadeFrames {
		m.Update(ctx)
	}
	if m.Current() != b || m.Darkness() != 1 {
		t.Fatalf("after fade-out: current b=%v darkness=%v", m.Current() == b, m.Darkness())
	}
	for range FadeFrames {
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
