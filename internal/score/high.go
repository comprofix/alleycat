// Package score saves the high score between runs (DESIGN §9).
package score

import (
	"encoding/json"
	"os"

	"github.com/comprofix/alleycat/internal/config"
)

// MaxHigh is the largest score the fence can show (7 digits).
const MaxHigh = 9_999_999

type highFile struct {
	High int `json:"high"`
}

// LoadHigh reads the saved high score at path. A missing or corrupt file
// gives 0, and a value out of range is clamped, so it never fails.
func LoadHigh(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	var f highFile
	if json.Unmarshal(data, &f) != nil {
		return 0
	}
	return min(max(f.High, 0), MaxHigh)
}

// SaveHigh writes the high score to path.
func SaveHigh(path string, high int) error {
	data, err := json.MarshalIndent(highFile{High: high}, "", "  ")
	if err != nil {
		return err
	}
	return config.WriteFile(path, append(data, '\n'))
}
