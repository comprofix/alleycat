package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Saved files (DESIGN §9).
const (
	SettingsFile = "settings.json"
	ScoresFile   = "scores.json"
)

// Settings are the player's choices saved between runs (DESIGN §9). They
// are presentation and sound only and never touch the rules.
type Settings struct {
	ScaleMode  ScaleMode `json:"scale_mode"`
	Fullscreen bool      `json:"fullscreen"`
	WindowW    int       `json:"window_width"`
	WindowH    int       `json:"window_height"`
	SoundOn    bool      `json:"sound_on"`
	Muted      bool      `json:"muted"`
	Volume     float64   `json:"volume"` // 0–1
	Soften     bool      `json:"soften_beeper"`
}

// DefaultSettings are the settings for preset p with no file.
func DefaultSettings(p Preset) Settings {
	w, h := p.DefaultWindow()
	return Settings{
		ScaleMode: p.ScaleModes[0],
		WindowW:   w, WindowH: h,
		SoundOn: true, Volume: DefaultVolume, Soften: true,
	}
}

// LoadSettings reads the settings at path for preset p. A missing or
// corrupt file gives the defaults, a missing field keeps its default, and
// values out of range are clamped, so it never fails. A file from v1.2.1
// or before keeps its aspect, saved then as "aspect": "crt".
func LoadSettings(path string, p Preset) Settings {
	s := DefaultSettings(p)
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	// The outer scale_mode hides the inner one, to tell whether it is there.
	got := struct {
		Settings
		Mode   *ScaleMode `json:"scale_mode"`
		Aspect string     `json:"aspect"`
	}{Settings: s}
	if json.Unmarshal(data, &got) != nil {
		return s
	}
	switch {
	case got.Mode != nil:
		got.ScaleMode = *got.Mode
	case got.Aspect == string(ScaleCRT):
		got.ScaleMode = ScaleCRT
	}
	return got.Settings.Clamped(p)
}

// Clamped returns s with every value made valid for preset p.
func (s Settings) Clamped(p Preset) Settings {
	if !p.HasMode(s.ScaleMode) {
		s.ScaleMode = p.ScaleModes[0]
	}
	minW, minH := p.MinWindow()
	s.WindowW = min(max(s.WindowW, minW), MaxWindowW)
	s.WindowH = min(max(s.WindowH, minH), MaxWindowH)
	if !(s.Volume >= 0) { // also catches NaN
		s.Volume = 0
	}
	s.Volume = min(s.Volume, 1)
	return s
}

// SaveSettings writes the settings to path.
func SaveSettings(path string, s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return WriteFile(path, append(data, '\n'))
}

// WriteFile writes data to a temporary file beside path and renames it into
// place, so a crash mid-write cannot leave a half-written file.
func WriteFile(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return err
	}
	_, werr := tmp.Write(data)
	cerr := tmp.Close()
	if werr == nil {
		werr = cerr
	}
	if werr == nil {
		werr = os.Rename(tmp.Name(), path)
	}
	if werr != nil {
		os.Remove(tmp.Name())
	}
	return werr
}

// SaveDir is the folder the saved files live in (DESIGN §9): the folder
// holding the game's executable. Under go run the executable is built in a
// throwaway go-build folder, so the working folder is used instead.
func SaveDir() string {
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		dir := filepath.Dir(exe)
		if !inGoBuild(dir) {
			return dir
		}
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return "."
}

// inGoBuild reports whether dir is inside a go-build temporary folder.
func inGoBuild(dir string) bool {
	for _, part := range strings.FieldsFunc(dir, func(r rune) bool { return r == '/' || r == '\\' }) {
		if strings.HasPrefix(part, "go-build") {
			return true
		}
	}
	return false
}
