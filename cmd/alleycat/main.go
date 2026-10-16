// Command alleycat runs the Alley Cat remake.
package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/config"
	"github.com/comprofix/alleycat/internal/game"
)

func main() {
	// The settings live beside the game (DESIGN §9).
	dir := config.SaveDir()
	saved := game.Saved{
		Settings: config.LoadSettings(filepath.Join(dir, config.SettingsFile)),
		Dir:      dir,
	}

	ebiten.SetWindowTitle("Alley Cat")
	ebiten.SetWindowSize(saved.Settings.WindowW, saved.Settings.WindowH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetFullscreen(saved.Settings.Fullscreen)
	ebiten.SetWindowClosingHandled(true) // to save the window size on close
	ebiten.SetTPS(game.TPS)

	g, err := game.New(uint64(time.Now().UnixNano()), saved)
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
