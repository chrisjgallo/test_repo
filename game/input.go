package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) handleUserInput() {
	// check if mouse is clicked
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		g.world.Spawn(float64(x), float64(y))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.world.TogglePause()
	}
}
