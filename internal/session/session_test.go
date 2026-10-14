package session

import (
	"testing"

	"github.com/comprofix/alleycat/internal/config"
)

func TestNewStartsAtSkillDifficulty(t *testing.T) {
	s := New(config.Tomcat)
	if s.Difficulty != 2 || s.Lives != 3 || s.Score != 0 {
		t.Errorf("New(Tomcat) = %+v", s)
	}
}

func TestLivesClamp(t *testing.T) {
	s := New(config.Kitten)
	for range 20 {
		s.GainLife()
	}
	if s.Lives != 9 {
		t.Errorf("lives = %d, want max 9", s.Lives)
	}
	for range 8 {
		if s.LoseLife() {
			t.Fatal("game over too early")
		}
	}
	if !s.LoseLife() || s.Lives != 0 {
		t.Errorf("last life should end the game; lives = %d", s.Lives)
	}
	s.LoseLife()
	if s.Lives != 0 {
		t.Errorf("lives went negative: %d", s.Lives)
	}
}

func TestLevelUpCapsAtSeven(t *testing.T) {
	s := New(config.AlleyCat)
	s.RoomsCleared = 4
	for range 10 {
		s.LevelUp()
	}
	if s.Difficulty != 7 || s.RoomsCleared != 0 {
		t.Errorf("after level-ups: difficulty %d rooms %d", s.Difficulty, s.RoomsCleared)
	}
}

func TestFormatScore(t *testing.T) {
	cases := map[int]string{0: "000-0000", 4408: "000-4408", 1059901: "105-9901", MaxScore: "999-9999"}
	for n, want := range cases {
		if got := FormatScore(n); got != want {
			t.Errorf("FormatScore(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestScoreCaps(t *testing.T) {
	s := New(config.Kitten)
	s.AddScore(MaxScore + 500)
	if s.Score != MaxScore {
		t.Errorf("score = %d, want %d", s.Score, MaxScore)
	}
}
