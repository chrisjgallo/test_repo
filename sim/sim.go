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

	// velocityPerPixel turns the length of a drag into a starting speed. The
	// drag is measured in world units, so aiming behaves the same at any zoom.
	velocityPerPixel = 0.02

	trajectoryWidth = 2

	// minDrawnRadius keeps objects from disappearing entirely when zoomed out.
	minDrawnRadius = 1

	worldBorderWidth = 2

	// statsMargin insets the paused readout from the top left corner, and
	// statsLineHeight is the height of a line of the built-in debug font.
	statsMargin     = 10
	statsLineHeight = 16
)

// red matches raylib's RED, so the pause indicator looks like it did before.
var red = color.RGBA{R: 230, G: 41, B: 55, A: 255}

// dimYellow draws the trajectory line: visible against black space without
// competing with the objects themselves.
var dimYellow = color.RGBA{R: 140, G: 138, B: 30, A: 255}

// dimGray outlines the edge of the world, so panning around a world bigger than
// the screen never leaves you wondering which way is back.
var dimGray = color.RGBA{R: 55, G: 55, B: 65, A: 255}

// Simulator satisfies ebiten.Game. Ebitengine calls Update on a fixed 60 Hz tick
// and Draw once per frame.
type Simulator struct {
	world  *world.World
	camera *camera

	screenWidth  int
	screenHeight int

	cornerOfScreenX float32
	cornerOfScreenY float32

	// dragging is true between pressing the mouse down and letting it go, while
	// the object waiting to be spawned is being aimed. The start is kept in
	// world coordinates so the spawn point stays put if the view moves mid-aim.
	dragging   bool
	dragStartX float64
	dragStartY float64

	// panning tracks a middle or right mouse drag that slides the view. The
	// last position is in screen coordinates, since that is what the cursor
	// moves through.
	panning  bool
	panLastX float64
	panLastY float64
}

// New builds a simulator looking at the middle of an empty world of the given
// size. The world may be larger than the screen showing it.
func New(screenWidth, screenHeight, worldWidth, worldHeight int) *Simulator {
	w := world.New(worldWidth, worldHeight)

	return &Simulator{
		world:           w,
		camera:          newCamera(screenWidth, screenHeight, w.Width()/2, w.Height()/2),
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
