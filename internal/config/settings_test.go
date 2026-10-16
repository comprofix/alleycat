package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	want := Settings{Aspect: AspectCRT, Fullscreen: true, WindowW: 960, WindowH: 600, SoundOn: false, Muted: true, Volume: 0.5, Soften: false}
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
	if got := LoadSettings(filepath.Join(dir, "none.json")); got != DefaultSettings() {
		t.Errorf("missing file: %+v", got)
	}
	for _, bad := range []string{"{not json", `{"volume": "loud"}`, ""} {
		path := filepath.Join(dir, "bad.json")
		os.WriteFile(path, []byte(bad), 0o644)
		if got := LoadSettings(path); got != DefaultSettings() {
			t.Errorf("%q: %+v, want the defaults", bad, got)
		}
	}
}

// Missing fields keep their defaults; values out of range are clamped.
func TestSettingsClamp(t *testing.T) {
	path := filepath.Join(t.TempDir(), SettingsFile)
	os.WriteFile(path, []byte(`{"aspect":"wide","window_width":5,"window_height":99999,"volume":3}`), 0o644)
	got := LoadSettings(path)
	want := DefaultSettings()
	want.WindowW, want.WindowH, want.Volume = MinWindowW, MaxWindowH, 1
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
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
