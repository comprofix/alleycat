// Package session holds the state of one game in progress: skill, difficulty,
// lives and score (DESIGN §3.6).
package session

import (
	"fmt"

	"github.com/comprofix/alleycat/internal/config"
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

// LevelUp applies a kiss: +1 life, difficulty +1 (to 7), rooms reset.
func (s *Session) LevelUp() {
	s.GainLife()
	s.Difficulty = min(s.Difficulty+1, config.MaxDifficulty)
	s.RoomsCleared = 0
}
