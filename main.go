package main

import (
	"flag"
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
	// The star is off by default, so plain `go run .` still opens on empty space
	// and the world is whatever you put in it.
	centralStar := flag.Bool("star", false,
		"begin with a fixed, massive star at the middle of the world")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Space Simulator")

	game := sim.New(screenWidth, screenHeight, worldWidth, worldHeight, *centralStar)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
