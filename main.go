package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"spacesim/sim"
)

const (
	screenWidth  = 1700
	screenHeight = 1000

	// The world is three times the screen in each direction, so there is room
	// to pan around and to let objects travel a while before they wrap.
	worldWidth  = screenWidth * 3
	worldHeight = screenHeight * 3
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Simulator")

	if err := ebiten.RunGame(sim.New(screenWidth, screenHeight, worldWidth, worldHeight)); err != nil {
		log.Fatal(err)
	}
}
