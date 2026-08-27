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

	// boundaryLabelDuration is how many ticks the boundary mode stays on screen
	// after it is switched. Long enough to read at 60 ticks a second, short
	// enough that the view goes back to being nothing but space.
	boundaryLabelDuration = 120

	// boundaryLabelMargin insets that label from the top left corner.
	boundaryLabelMargin = 10
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

	// boundaryLabelTicks counts down the announcement of a boundary switch.
	// Without it the three modes are indistinguishable until something happens
	// to reach an edge.
	boundaryLabelTicks int
}

// New builds a simulator looking at the middle of an empty world of the given
// size. The world may be larger than the screen showing it.
//
// With centralStar set, the world is not quite empty: a fixed star sits at the
// middle of it, so that objects spawned afterwards have something to fall
// towards from the first tick instead of only each other.
func New(screenWidth, screenHeight, worldWidth, worldHeight int, centralStar bool) *Simulator {
	w := world.New(worldWidth, worldHeight)
	if centralStar {
		w.SpawnStar(w.Width()/2, w.Height()/2)
	}

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

	// The label is part of the interface, not the simulation, so it keeps
	// fading even while the world is paused.
	if s.boundaryLabelTicks > 0 {
		s.boundaryLabelTicks--
	}

	return nil
}

func (s *Simulator) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	s.drawWorldEdge(screen)
	s.drawSpace(screen)
	s.drawTrajectory(screen)
	s.drawBoundaryLabel(screen)
	s.drawMenusAndInfo(screen)
}

// Layout keeps the simulation at a fixed resolution regardless of window size.
func (s *Simulator) Layout(outsideWidth, outsideHeight int) (int, int) {
	return s.screenWidth, s.screenHeight
}

// drawWorldEdge outlines where the world stops and objects wrap around.
func (s *Simulator) drawWorldEdge(screen *ebiten.Image) {
	left, top := s.camera.worldToScreen(0, 0)
	right, bottom := s.camera.worldToScreen(s.world.Width(), s.world.Height())

	vector.StrokeRect(
		screen,
		float32(left),
		float32(top),
		float32(right-left),
		float32(bottom-top),
		worldBorderWidth,
		dimGray,
		true,
	)
}

// drawSpace draws every object, colored by how large it has grown. The ramp
// reads the object's own radius rather than the drawn one, so an object keeps
// its color as the view zooms.
func (s *Simulator) drawSpace(screen *ebiten.Image) {
	for _, object := range s.world.Objects {
		x, y := s.camera.worldToScreen(object.X, object.Y)
		radius := object.Radius * s.camera.zoom
		if radius < minDrawnRadius {
			radius = minDrawnRadius
		}

		if s.offScreen(x, y, radius) {
			continue
		}

		// A star is not a pile of merged objects, so it is not colored like one.
		fill := colorForRadius(object.Radius)
		if object.Fixed {
			fill = starColor
		}

		vector.DrawFilledCircle(
			screen,
			float32(x),
			float32(y),
			float32(radius),
			fill,
			true, // antialias -- small objects look like specks without it
		)
	}
}

// offScreen reports whether a circle falls entirely outside the view, so a
// world far bigger than the screen costs nothing to draw the empty parts of.
func (s *Simulator) offScreen(x, y, radius float64) bool {
	return x+radius < 0 || y+radius < 0 ||
		x-radius > float64(s.screenWidth) || y-radius > float64(s.screenHeight)
}

// drawTrajectory previews where an object being aimed will head once it is
// released: a line from the spawn point out to the cursor.
func (s *Simulator) drawTrajectory(screen *ebiten.Image) {
	if !s.dragging {
		return
	}

	startX, startY := s.camera.worldToScreen(s.dragStartX, s.dragStartY)
	cursorX, cursorY := ebiten.CursorPosition()

	vector.StrokeLine(
		screen,
		float32(startX),
		float32(startY),
		float32(cursorX),
		float32(cursorY),
		trajectoryWidth,
		dimYellow,
		true,
	)
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

// drawMenusAndInfo draws everything that belongs to the screen rather than to
// the world, so none of it moves when the view does. It only appears while
// paused: mid-simulation the view should be nothing but space.
func (s *Simulator) drawMenusAndInfo(screen *ebiten.Image) {
	if !s.world.Paused {
		return
	}

	vector.DrawFilledRect(screen, s.cornerOfScreenX, s.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)
	vector.DrawFilledRect(screen, s.cornerOfScreenX+spaceBetweenPauseBars, s.cornerOfScreenY,
		pauseBarWidth, pauseBarHeight, red, false)

	s.drawStats(screen)
}

// drawStats prints a small readout in the top left while the simulation is
// stopped and there is time to actually read it.
func (s *Simulator) drawStats(screen *ebiten.Image) {
	lines := []string{
		fmt.Sprintf("Objects:  %d", len(s.world.Objects)),
		fmt.Sprintf("Mass:     %.0f", s.world.TotalMass()),
		fmt.Sprintf("Largest:  %.1f", s.largestRadius()),
		fmt.Sprintf("Zoom:     %.2fx", s.camera.zoom),
	}

	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, statsMargin, statsMargin+i*statsLineHeight)
	}
}

// largestRadius is the radius of the biggest object out there, which is the
// quickest read on how much merging has happened. It rescans on every call, but
// the only caller is the paused readout -- so the scan happens exactly when the
// physics loop is idle, which is cheaper than keeping a cached value honest.
//
// Stars sit it out. One starts larger than any amount of merging is likely to
// reach, so counting it would pin the number to the star and say nothing about
// what the rest of the world has been doing.
func (s *Simulator) largestRadius() float64 {
	var largest float64
	for _, object := range s.world.Objects {
		if object.Fixed {
			continue
		}
		if object.Radius > largest {
			largest = object.Radius
		}
	}
	return largest
}
