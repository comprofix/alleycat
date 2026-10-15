// Package session holds the state of one game in progress: skill, difficulty,
// lives and score (DESIGN §3.6).
package session

import (
	"fmt"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/rooms"
)

// FormatScore renders a score as the fence shows it: 7 digits, "000-0000".
func FormatScore(score int) string {
	return fmt.Sprintf("%03d-%04d", score/10000, score%10000)
}

// MaxScore is the largest score the 7-digit display can show.
const MaxScore = 9_999_999

// Session is one game in progress.
type Session struct {
	Skill        config.Skill
	Difficulty   int
	Lives        int
	Score        int
	RoomsCleared int // since the last kiss; sets the heart-room gift count
	// LevelTicks counts master ticks since the level began (the game start
	// or the last kiss), for the kiss bonus (DESIGN §3.6).
	LevelTicks int

	// FeliciaWaiting is set when a room is cleared: the next window to open
	// is Felicia's (DESIGN §3.3). Entering it clears the flag.
	FeliciaWaiting bool

	// Rooms picks the room behind each ordinary window (DESIGN §3.8).
	Rooms rooms.Picker
}

// New starts a game at the skill's starting difficulty with 3 lives.
func New(skill config.Skill) *Session {
	return &Session{
		Skill:      skill,
		Difficulty: skill.StartDifficulty(),
		Lives:      config.StartLives,
	}
}

// Table returns the difficulty table for the current level.
func (s *Session) Table() config.Difficulty {
	return config.ForLevel(s.Difficulty)
}

// AddScore adds points, capped at the display maximum.
func (s *Session) AddScore(points int) {
	s.Score = min(s.Score+points, MaxScore)
}

// GainLife adds a life, up to the maximum of 9.
func (s *Session) GainLife() {
	s.Lives = min(s.Lives+1, config.MaxLives)
}

// LoseLife removes a life and reports whether the game is over.
func (s *Session) LoseLife() (gameOver bool) {
	s.Lives = max(s.Lives-1, 0)
	return s.Lives == 0
}

// LevelUp applies a kiss: +1 life, difficulty +1 (to 7), rooms and the
// level timer reset.
func (s *Session) LevelUp() {
	s.GainLife()
	s.Difficulty = min(s.Difficulty+1, config.MaxDifficulty)
	s.RoomsCleared = 0
	s.LevelTicks = 0
}

// KissBonus is the kiss bonus and its multiplier at the current difficulty
// (DESIGN §3.6): 4000 + max(0, 10800 − T) ÷ 2, added multiplier times, the
// multiplier being 2d + 1, doubled with a gift.
func (s *Session) KissBonus(withGift bool) (bonus, multiplier int) {
	bonus = config.KissBase + max(0, config.KissWindow-s.LevelTicks)/2
	multiplier = 2*s.Difficulty + 1
	if withGift {
		multiplier *= 2
	}
	return bonus, multiplier
}
