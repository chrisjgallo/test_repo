// Package game wires the simulation to Ebitengine: input in, pixels out.
package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"spacesim/world"
)

const (
	pauseBarWidth         = 20
	pauseBarHeight        = 100
	spaceBetweenPauseBars = 30
)

// red matches raylib's RED, so the pause indicator looks like it did before.
var red = color.RGBA{R: 230, G: 41, B: 55, A: 255}

// Game satisfies ebiten.Game. Ebitengine calls Update on a fixed 60 Hz tick and
// Draw once per frame.
type Game struct {
	world *world.World

	screenWidth  int
	screenHeight int

	cornerOfScreenX float32
	cornerOfScreenY float32
}

// New builds a game with an empty world sized to the screen.
func New(screenWidth, screenHeight int) *Game {
	return &Game{
		world:           world.New(screenWidth, screenHeight),
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		cornerOfScreenX: float32(screenWidth) * 0.9,
		cornerOfScreenY: float32(screenHeight) / 10,
	}
}

func (g *Game) Update() error {
	g.handleUserInput()
	g.world.UpdateSpace()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	g.drawSpace(screen)
	g.drawMenusAndInfo(screen)
}

// Layout keeps the simulation at a fixed resolution regardless of window size.
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.screenWidth, g.screenHeight
}

func (g *Game) drawSpace(screen *ebiten.Image) {
	for _, object := range g.world.Objects {
		vector.DrawFilledCircle(
			screen,
			float32(object.X),
			float32(object.Y),
			float32(object.Radius),
			color.White,
			true, // antialias -- small objects look like specks without it
		)
	}
}

func (g *Game) drawMenusAndInfo(screen *ebiten.Image) {
	if !g.world.Paused {
		return
	}

	vector.DrawFilledRect(screen, g.cornerOfScreenX, g.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)
	vector.DrawFilledRect(screen, g.cornerOfScreenX+spaceBetweenPauseBars, g.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)
}
