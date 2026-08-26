package sim

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (s *Simulator) handleUserInput() {
	// check if mouse is clicked
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		s.world.Spawn(float64(x), float64(y))
	}

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.world.TogglePause()
	}
}
