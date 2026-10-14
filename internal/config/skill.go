// Package config holds skill levels, the per-difficulty tables and the
// tuning constants. Values come from DESIGN.md; each cites its section.
package config

// Skill is the level chosen on the menu (DESIGN §3.1).
type Skill int

const (
	Kitten Skill = iota
	HouseCat
	Tomcat
	AlleyCat
)

// Skills lists the skills in menu order.
var Skills = [...]Skill{Kitten, HouseCat, Tomcat, AlleyCat}

func (s Skill) String() string {
	switch s {
	case Kitten:
		return "Kitten"
	case HouseCat:
		return "House Cat"
	case Tomcat:
		return "Tomcat"
	case AlleyCat:
		return "Alley Cat"
	}
	return "?"
}

// Letter is the menu key and the letter painted on the fence.
func (s Skill) Letter() rune {
	return [...]rune{'K', 'H', 'T', 'A'}[s]
}

// StartDifficulty is the difficulty level a new game starts at.
func (s Skill) StartDifficulty() int {
	return int(s)
}
