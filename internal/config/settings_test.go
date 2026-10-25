package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	want := Settings{Look: LookClassic, ScaleMode: ScaleCRT, Fullscreen: true, WindowW: 960, WindowH: 600, SoundOn: false, Muted: true, Volume: 0.5, Soften: false}
	if err := SaveSettings(path, want); err != nil {
		t.Fatal(err)
	}
	if got := LoadSettings(path); got != want {
		t.Errorf("loaded %+v, want %+v", got, want)
	}
}

// A missing or corrupt file gives the defaults (DESIGN §9).
func TestSettingsFallBack(t *testing.T) {
	dir := t.TempDir()
	def := DefaultSettings()
	if got := LoadSettings(filepath.Join(dir, "none.json")); got != def {
		t.Errorf("missing file: %+v", got)
	}
	for _, bad := range []string{"{not json", `{"volume": "loud"}`, ""} {
		path := filepath.Join(dir, "bad.json")
		os.WriteFile(path, []byte(bad), 0o644)
		if got := LoadSettings(path); got != def {
			t.Errorf("%q: %+v, want the defaults", bad, got)
		}
	}
}

// The defaults: the new look, whole-number scaling for the classic look,
// a 1280×800 window, sound on and softened.
func TestDefaultSettings(t *testing.T) {
	want := Settings{Look: LookNew, ScaleMode: ScaleInteger, WindowW: 1280, WindowH: 800, SoundOn: true, Volume: DefaultVolume, Soften: true}
	if got := DefaultSettings(); got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// Missing fields keep their defaults; values out of range are clamped.
func TestSettingsClamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	os.WriteFile(path, []byte(`{"look":"odd","scale_mode":"wide","window_width":5,"window_height":99999,"volume":3}`), 0o644)
	got := LoadSettings(path)
	want := DefaultSettings()
	want.WindowW, want.WindowH, want.Volume = 640, MaxWindowH, 1 // half the new look's screen
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
	os.WriteFile(path, []byte(`{"look":"classic","window_width":5}`), 0o644)
	got = LoadSettings(path)
	want = DefaultSettings()
	want.Look, want.WindowW = LookClassic, 320
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

// A file saved by v1.2.1 or before has "aspect" in place of "scale_mode",
// and keeps its choice for the classic look; with no look saved, it opens
// in the new look, as for every player.
func TestSettingsFromOlderVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	for aspect, want := range map[string]ScaleMode{"crt": ScaleCRT, "square": ScaleInteger} {
		os.WriteFile(path, []byte(`{"aspect":"`+aspect+`","window_width":960,"window_height":600}`), 0o644)
		got := LoadSettings(path)
		if got.ScaleMode != want || got.WindowW != 960 || got.Look != LookNew {
			t.Errorf("aspect %q: %+v, want scale mode %q", aspect, got, want)
		}
	}
	// A newer file's scale mode wins.
	os.WriteFile(path, []byte(`{"aspect":"crt","scale_mode":"integer"}`), 0o644)
	if got := LoadSettings(path); got.ScaleMode != ScaleInteger {
		t.Errorf("scale_mode beside aspect: %q, want integer", got.ScaleMode)
	}
}

func TestNextMode(t *testing.T) {
	if got := Classic.NextMode(ScaleInteger); got != ScaleCRT {
		t.Errorf("integer → %q, want crt", got)
	}
	if got := Classic.NextMode(ScaleCRT); got != ScaleInteger {
		t.Errorf("crt → %q, want integer", got)
	}
	if got := Classic.NextMode("bogus"); got != ScaleInteger {
		t.Errorf("unknown mode → %q, want the first mode", got)
	}
}

func TestInGoBuild(t *testing.T) {
	for dir, want := range map[string]bool{
		`C:\Users\me\AppData\Local\Temp\go-build123\b001\exe`: true,
		"/tmp/go-build99/b001/exe":                            true,
		`C:\Games\AlleyCat`:                                   false,
		"/home/me/alleycat/bin":                               false,
	} {
		if got := inGoBuild(dir); got != want {
			t.Errorf("%s: %v, want %v", dir, got, want)
		}
	}
}

// The new look shows the screen smoothed whatever the classic look's
// choice; the classic look shows its choice.
func TestMode(t *testing.T) {
	s := Settings{Look: LookNew, ScaleMode: ScaleCRT}
	if got := s.Mode(); got != ScaleSmooth {
		t.Errorf("new look: %q, want smooth", got)
	}
	s.Look = LookClassic
	if got := s.Mode(); got != ScaleCRT {
		t.Errorf("classic look: %q, want crt", got)
	}
}

// Each look has its preset: the classic look 320×200, the new look four
// times that, so the game's coordinates are the same in both.
func TestLookPresets(t *testing.T) {
	c, n := LookClassic.Preset(), LookNew.Preset()
	if c.ScreenWidth != 320 || c.ScreenHeight != 200 || c.Art != ArtPixel {
		t.Errorf("classic: %d×%d %s", c.ScreenWidth, c.ScreenHeight, c.Art)
	}
	if n.ScreenWidth != 4*c.ScreenWidth || n.ScreenHeight != 4*c.ScreenHeight || n.Art != ArtSmooth {
		t.Errorf("new: %d×%d %s, want four times classic, smooth", n.ScreenWidth, n.ScreenHeight, n.Art)
	}
	if w, h := n.DefaultWindow(); w != 1280 || h != 800 {
		t.Errorf("new look's window %d×%d, want 1280×800", w, h)
	}
	if Look("odd").Preset().Name != Smooth.Name {
		t.Error("an unknown look is not the new look")
	}
}
