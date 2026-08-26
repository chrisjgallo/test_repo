package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"spacesim/sim"
)

const (
	screenWidth  = 1700
	screenHeight = 1000
)

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Simulator")

	if err := ebiten.RunGame(sim.New(screenWidth, screenHeight)); err != nil {
		log.Fatal(err)
	}
}
