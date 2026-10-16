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
)

// The window's default and allowed sizes, in window pixels (DESIGN §2.1).
const (
	DefaultWindowW, DefaultWindowH = 1280, 800
	MinWindowW, MinWindowH         = 320, 200
	MaxWindowW, MaxWindowH         = 7680, 4800
)

// Aspect modes as saved.
const (
	AspectSquare = "square"
	AspectCRT    = "crt"
)

// Settings are the player's choices saved between runs (DESIGN §9). They
// are presentation and sound only and never touch the rules.
type Settings struct {
	Aspect     string  `json:"aspect"` // AspectSquare or AspectCRT
	Fullscreen bool    `json:"fullscreen"`
	WindowW    int     `json:"window_width"`
	WindowH    int     `json:"window_height"`
	SoundOn    bool    `json:"sound_on"`
	Muted      bool    `json:"muted"`
	Volume     float64 `json:"volume"` // 0–1
	Soften     bool    `json:"soften_beeper"`
}

// DefaultSettings are the settings with no file.
func DefaultSettings() Settings {
	return Settings{
		Aspect:  AspectSquare,
		WindowW: DefaultWindowW, WindowH: DefaultWindowH,
		SoundOn: true, Volume: DefaultVolume, Soften: true,
	}
}

// LoadSettings reads the settings at path. A missing or corrupt file gives
// the defaults, a missing field keeps its default, and values out of range
// are clamped, so it never fails.
func LoadSettings(path string) Settings {
	s := DefaultSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		return s
	}
	got := s
	if json.Unmarshal(data, &got) != nil {
		return s
	}
	return got.clamped()
}

func (s Settings) clamped() Settings {
	if s.Aspect != AspectCRT {
		s.Aspect = AspectSquare
	}
	s.WindowW = min(max(s.WindowW, MinWindowW), MaxWindowW)
	s.WindowH = min(max(s.WindowH, MinWindowH), MaxWindowH)
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
	for _, part := range strings.Split(filepath.ToSlash(dir), "/") {
		if strings.HasPrefix(part, "go-build") {
			return true
		}
	}
	return false
}
