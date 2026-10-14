package config

import "testing"

func TestSkillStartDifficulty(t *testing.T) {
	want := map[Skill]int{Kitten: 0, HouseCat: 1, Tomcat: 2, AlleyCat: 3}
	for s, d := range want {
		if got := s.StartDifficulty(); got != d {
			t.Errorf("%v.StartDifficulty() = %d, want %d", s, got, d)
		}
	}
}

func TestSkillLetters(t *testing.T) {
	got := ""
	for _, s := range Skills {
		got += string(s.Letter())
	}
	if got != "KHTA" {
		t.Errorf("letters = %q, want KHTA", got)
	}
}

func TestMultiplierIsTwoDPlusOne(t *testing.T) {
	for d, row := range Difficulties {
		if row.Multiplier != 2*d+1 {
			t.Errorf("difficulty %d multiplier = %d, want %d", d, row.Multiplier, 2*d+1)
		}
	}
}

func TestCanCounts(t *testing.T) {
	want := [8]int{6, 5, 4, 3, 4, 3, 4, 3} // DESIGN §3.8
	for d, row := range Difficulties {
		if len(row.Cans) != want[d] {
			t.Errorf("difficulty %d has %d cans, want %d", d, len(row.Cans), want[d])
		}
	}
}

func TestDrowningBeforeFailure(t *testing.T) {
	for d, row := range Difficulties {
		if row.DrowningTicks >= row.AirFailTicks {
			t.Errorf("difficulty %d: drowning %d should come before failure %d", d, row.DrowningTicks, row.AirFailTicks)
		}
	}
}

func TestForLevelClamps(t *testing.T) {
	if ForLevel(-3).Multiplier != 1 || ForLevel(99).Multiplier != 15 {
		t.Error("ForLevel should clamp to 0–7")
	}
}
