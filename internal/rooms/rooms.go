// Package rooms holds the rooms behind the alley windows (DESIGN §3.4) and
// the rule that picks one when Freddy enters a window (§3.8).
package rooms

import "math/rand/v2"

// Kind is a room.
type Kind int

const (
	Fishbowl Kind = iota
	Bookshelf
	Cheese
	Birdcage
	Dogs
	Heart // Felicia's room; reached only through her window
)

var names = [...]string{"FISHBOWL", "BOOKSHELF", "CHEESE", "BIRDCAGE", "SLEEPING DOGS", "FELICIA"}

func (k Kind) String() string { return names[k] }

// pools are the weighted pools by difficulty mod 4; mod 4 = 3 is uniform.
var pools = [3][4]Kind{
	{Cheese, Cheese, Fishbowl, Fishbowl},
	{Bookshelf, Bookshelf, Birdcage, Birdcage},
	{Dogs, Dogs, Cheese, Fishbowl},
}

var all = [5]Kind{Fishbowl, Bookshelf, Cheese, Birdcage, Dogs}

// Picker picks rooms, remembering the last two so that no room comes up
// three times running.
type Picker struct {
	last   [2]Kind
	visits int
}

// Pick chooses the room behind an ordinary window (DESIGN §3.8, "Room
// selection"): about 3 times in 4 from the difficulty's weighted pool,
// otherwise from all five.
func (p *Picker) Pick(difficulty int, rng *rand.Rand) Kind {
	for {
		var k Kind
		if pool := difficulty % 4; rng.IntN(4) != 0 && pool < len(pools) {
			k = pools[pool][rng.IntN(4)]
		} else {
			k = all[rng.IntN(len(all))]
		}
		if p.visits >= 2 && k == p.last[0] && k == p.last[1] {
			continue
		}
		p.last[1], p.last[0] = p.last[0], k
		p.visits++
		return k
	}
}
