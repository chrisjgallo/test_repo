// Package sim drives the world with Ebitengine: input in, pixels out.
package sim

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"spacesim/world"
)

const (
	pauseBarWidth         = 20
	pauseBarHeight        = 100
	spaceBetweenPauseBars = 30

	// boundaryLabelDuration is how many ticks the boundary mode stays on screen
	// after it is switched. Long enough to read at 60 ticks a second, short
	// enough that the view goes back to being nothing but space.
	boundaryLabelDuration = 120

	// boundaryLabelMargin insets that label from the top left corner.
	boundaryLabelMargin = 10
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

	// boundaryLabelTicks counts down the announcement of a boundary switch.
	// Without it the three modes are indistinguishable until something happens
	// to reach an edge.
	boundaryLabelTicks int
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

	// The label is part of the interface, not the simulation, so it keeps
	// fading even while the world is paused.
	if s.boundaryLabelTicks > 0 {
		s.boundaryLabelTicks--
	}

	return nil
}

func (s *Simulator) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	s.drawSpace(screen)
	s.drawBoundaryLabel(screen)
	s.drawMenusAndInfo(screen)
}

// Layout keeps the simulation at a fixed resolution regardless of window size.
func (s *Simulator) Layout(outsideWidth, outsideHeight int) (int, int) {
	return s.screenWidth, s.screenHeight
}

// drawSpace draws every object, colored by how large it has grown.
func (s *Simulator) drawSpace(screen *ebiten.Image) {
	for _, object := range s.world.Objects {
		vector.DrawFilledCircle(
			screen,
			float32(object.X),
			float32(object.Y),
			float32(object.Radius),
			colorForRadius(object.Radius),
			true, // antialias -- small objects look like specks without it
		)
	}
}

// drawBoundaryLabel names the boundary mode for a moment after it is switched,
// so it is clear which of the three is in force.
func (s *Simulator) drawBoundaryLabel(screen *ebiten.Image) {
	if s.boundaryLabelTicks <= 0 {
		return
	}

	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Boundary: %s", s.world.Boundary),
		boundaryLabelMargin, boundaryLabelMargin)
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
