// Package sim drives the world with Ebitengine: input in, pixels out.
package sim

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

// Simulator satisfies ebiten.Game. Ebitengine calls Update on a fixed 60 Hz tick
// and Draw once per frame.
type Simulator struct {
	world *world.World

	screenWidth  int
	screenHeight int

	cornerOfScreenX float32
	cornerOfScreenY float32
}

// New builds a simulator with an empty world sized to the screen.
func New(screenWidth, screenHeight int) *Simulator {
	return &Simulator{
		world:           world.New(screenWidth, screenHeight),
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		cornerOfScreenX: float32(screenWidth) * 0.9,
		cornerOfScreenY: float32(screenHeight) / 10,
	}
}

func (s *Simulator) Update() error {
	s.handleUserInput()
	s.world.UpdateSpace()
	return nil
}

func (s *Simulator) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	s.drawSpace(screen)
	s.drawMenusAndInfo(screen)
}

// Layout keeps the simulation at a fixed resolution regardless of window size.
func (s *Simulator) Layout(outsideWidth, outsideHeight int) (int, int) {
	return s.screenWidth, s.screenHeight
}

func (s *Simulator) drawSpace(screen *ebiten.Image) {
	for _, object := range s.world.Objects {
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

func (s *Simulator) drawMenusAndInfo(screen *ebiten.Image) {
	if !s.world.Paused {
		return
	}

	vector.DrawFilledRect(screen, s.cornerOfScreenX, s.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)
	vector.DrawFilledRect(screen, s.cornerOfScreenX+spaceBetweenPauseBars, s.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)
}
