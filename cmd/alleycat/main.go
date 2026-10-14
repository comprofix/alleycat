// Command alleycat runs the Alley Cat remake.
package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/game"
)

func main() {
	ebiten.SetWindowTitle("Alley Cat")
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(60)

	if err := ebiten.RunGame(game.New()); err != nil {
		log.Fatal(err)
	}
}
