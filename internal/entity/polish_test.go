package entity

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

// A landing leaves a puff where he landed, which fades (DESIGN §8, "Polish").
func TestLandingPuff(t *testing.T) {
	f := inAlley()
	f.Update(Controls{V: -1}, floorOnly{})
	for f.Airborne() {
		f.Update(Controls{}, floorOnly{})
	}
	if f.Dust != config.DustUpdates || f.DustX != 150 || f.DustY != config.FloorY {
		t.Fatalf("puff %d at %d,%d, want %d at 150,%d", f.Dust, f.DustX, f.DustY, config.DustUpdates, config.FloorY)
	}
	run(f, Controls{}, floorOnly{}, config.DustUpdates)
	if f.Dust != 0 {
		t.Errorf("puff still showing after %d updates: %d", config.DustUpdates, f.Dust)
	}
}

// Each knock counts once, however long the thing that hit him touches him.
func TestKnocksCountOnce(t *testing.T) {
	f := inAlley()
	f.KnockDown()
	f.KnockDown()
	if f.Hits != 1 {
		t.Errorf("thrown object: %d hits, want 1", f.Hits)
	}
	g := inAlley()
	g.Bounce(10, 10)
	g.Bounce(10, 10)
	if g.Hits != 1 {
		t.Errorf("broom: %d hits, want 1", g.Hits)
	}
}
