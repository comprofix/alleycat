package gfx

import "testing"

// In every room with a broom, Freddy's footprints must show against the
// floor (DESIGN §3.4): Plum prints once vanished on the Plum rooms' floor.
func TestFootprintsShow(t *testing.T) {
	for name, s := range map[string]Scheme{
		"birdcage": BirdcageScheme, "cheese": CheeseScheme, "fishbowl": FishbowlScheme,
		"sleeping dogs": DogsScheme, "bookshelf": BookshelfScheme,
	} {
		if s.Prints == s.Ground {
			t.Errorf("%s: footprints are drawn in the floor's own colour", name)
		}
	}
}
