package entity

import (
	"os"
	"testing"

	"github.com/comprofix/alleycat/internal/spritegrid"
)

// Every recoloured set of Freddy's sprites exists with his sizes and frame
// counts, so the heart room cats and the gift Freddy can stand in for him
// (DESIGN §8).
func TestFreddySets(t *testing.T) {
	f, err := os.Open("../../assets/sprites/alley.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	sprites, err := spritegrid.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]spritegrid.Sprite{}
	for _, s := range sprites {
		byName[s.Name] = s
	}
	sets := map[string][]string{
		"gift":    {"walk", "stand", "jump", "sit"},
		"rival":   {"walk", "jump"},
		"felicia": {"walk", "jump"},
		"silver":  {"walk", "jump"},
	}
	for set, poses := range sets {
		for _, pose := range poses {
			want, got := byName["freddy_"+pose], byName[set+"_"+pose]
			if got.Name == "" {
				t.Errorf("no %s_%s", set, pose)
				continue
			}
			if got.W != want.W || got.H != want.H || len(got.Frames) != len(want.Frames) {
				t.Errorf("%s_%s is %dx%d with %d frames, want Freddy's %dx%d with %d",
					set, pose, got.W, got.H, len(got.Frames), want.W, want.H, len(want.Frames))
			}
		}
	}
}
