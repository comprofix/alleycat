package scene

import (
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/input"
)

// The title cat walks in from the left, then wanders between the turn
// points, sometimes stopping to sit (DESIGN §4, "The title cat").
func TestTitleCatWanders(t *testing.T) {
	var c titleCat
	rng := rand.New(rand.NewPCG(1, 2))
	c.update(1, rng)
	if c.dir != 1 || c.x != config.TitleCatSpeed {
		t.Fatalf("start: dir %d x %d, want walking right", c.dir, c.x)
	}
	var sat, left, right bool
	for range 2000 {
		c.update(1, rng)
		if c.x < 0 || c.x > config.TitleCatTurnRightX+config.TitleCatSpeed {
			t.Fatalf("x %d off the fence", c.x)
		}
		switch c.dir {
		case 0:
			sat = true
		case -1:
			left = true
		case 1:
			right = true
		}
	}
	if !sat || !left || !right {
		t.Fatalf("sat %v left %v right %v, want all", sat, left, right)
	}
}

// heldKeys is a key source whose held keys a test sets.
type heldKeys struct{ keys []ebiten.Key }

func (h *heldKeys) AppendPressedKeys(k []ebiten.Key) []ebiten.Key { return append(k, h.keys...) }

func titleContext() (*Context, *heldKeys) {
	keys := &heldKeys{}
	return &Context{Input: input.New(keys), RNG: rand.New(rand.NewPCG(1, 2))}, keys
}

// Before the first game the title waits for a key, however long it is
// left, then goes to the skill menu (DESIGN §4, "No attract mode").
func TestTitleWaitsForKeyBeforeFirstGame(t *testing.T) {
	ctx, keys := titleContext()
	ti := NewTitle()
	for range 2 * (config.TitleHoldTicks + config.TitleRepeatExtra) {
		if next := step(ctx, ti); next != nil {
			t.Fatalf("title switched to %T with no key", next)
		}
	}
	keys.keys = []ebiten.Key{ebiten.KeySpace}
	if _, ok := step(ctx, ti).(*SkillSelect); !ok {
		t.Fatal("key did not go to the skill menu")
	}
}

// After a game the title stays, restarting its tune, and a key replays at
// the same skill.
func TestTitleAfterGame(t *testing.T) {
	ctx, keys := titleContext()
	ctx.Played, ctx.LastSkill = true, config.Tomcat
	ti := NewTitle()
	for range 3 * (config.TitleHoldTicks + config.TitleRepeatExtra) {
		if next := step(ctx, ti); next != nil {
			t.Fatalf("title switched to %T with no key", next)
		}
	}
	keys.keys = []ebiten.Key{ebiten.KeySpace}
	next := step(ctx, ti)
	if _, ok := next.(*Play); !ok {
		t.Fatalf("key went to %T, want a new game", next)
	}
	if ctx.Session == nil || ctx.Session.Skill != config.Tomcat {
		t.Errorf("new game not at the same skill")
	}
}

// The tune starts again every hold + 72 ticks after a game.
func TestTitleRepeatsTune(t *testing.T) {
	ctx, _ := titleContext()
	ctx.Played = true
	ti := NewTitle()
	restarts := 0
	for range 2*(config.TitleHoldTicks+config.TitleRepeatExtra) + 1 {
		step(ctx, ti)
		if !ti.played {
			restarts++
		}
	}
	if restarts != 2 {
		t.Errorf("tune restarted %d times, want 2", restarts)
	}
}
