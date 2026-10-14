// Command alleycat runs the Alley Cat remake.
package main

import (
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/comprofix/alleycat/internal/game"
)

func main() {
	ebiten.SetWindowTitle("Alley Cat")
	ebiten.SetWindowSize(1280, 800)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetTPS(game.TPS)

	g, err := game.New(uint64(time.Now().UnixNano()))
	if err != nil {
		log.Fatal(err)
	}
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
