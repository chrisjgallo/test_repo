package sim

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// keyPanSpeed is how fast the arrow keys slide the view, in screen pixels per
// tick. It goes through panByScreen for the same reason a drag does: the divide
// by zoom converts those pixels into world units, so the view always travels
// keyPanSpeed pixels per tick on screen, at any zoom. Panning a fixed number of
// world units instead would crawl when zoomed out and bolt when zoomed in.
const keyPanSpeed = 12

// keyZoomStep is the zoom multiplier applied per tick while a zoom key is held.
// It is far gentler than zoomStep, which is spent one notch at a time: held down
// at 60 Hz, a wheel-sized step would cross the whole zoom range in a blink.
const keyZoomStep = 1.02

func (s *Simulator) handleUserInput() {
	s.handleSpawning()
	s.handlePanning()
	s.handleZooming()

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.world.TogglePause()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		s.resetView()
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		s.world.CycleBoundary()
		s.boundaryLabelTicks = boundaryLabelDuration
	}

	s.camera.clampToWorld(s.world.Width(), s.world.Height())
}

// handleSpawning aims and launches new objects. Press where the object should
// appear and drag out the direction it should head: the longer the drag, the
// faster it launches. A plain click is a zero-length drag, so it still drops an
// object at rest.
func (s *Simulator) handleSpawning() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		s.dragging = true
		s.dragStartX, s.dragStartY = s.camera.screenToWorld(float64(x), float64(y))
	}

	if s.dragging && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		endX, endY := s.camera.screenToWorld(float64(x), float64(y))

		s.world.SpawnMoving(
			s.dragStartX,
			s.dragStartY,
			(endX-s.dragStartX)*velocityPerPixel,
			(endY-s.dragStartY)*velocityPerPixel,
		)
		s.dragging = false
	}
}

// handlePanning slides the view, either by dragging with the right or middle
// mouse button or by holding the arrow keys. Left-drag is left alone: that one
// already belongs to spawning.
func (s *Simulator) handlePanning() {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle) {
		x, y := ebiten.CursorPosition()
		s.panning = true
		s.panLastX, s.panLastY = float64(x), float64(y)
	}

	if s.panning {
		if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) &&
			!ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle) {
			s.panning = false
		} else {
			x, y := ebiten.CursorPosition()
			s.camera.panByScreen(float64(x)-s.panLastX, float64(y)-s.panLastY)
			s.panLastX, s.panLastY = float64(x), float64(y)
		}
	}

	var dx, dy float64
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		dx += keyPanSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		dx -= keyPanSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		dy += keyPanSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		dy -= keyPanSpeed
	}
	if dx != 0 || dy != 0 {
		s.camera.panByScreen(dx, dy)
	}
}

// handleZooming scales the view with the scroll wheel, anchored on the cursor,
// or with the plus and minus keys, anchored on the middle of the screen.
func (s *Simulator) handleZooming() {
	if _, wheelY := ebiten.Wheel(); wheelY != 0 {
		x, y := ebiten.CursorPosition()
		s.camera.zoomBy(math.Pow(zoomStep, wheelY), float64(x), float64(y))
	}

	centerX, centerY := float64(s.screenWidth)/2, float64(s.screenHeight)/2
	if ebiten.IsKeyPressed(ebiten.KeyEqual) {
		s.camera.zoomBy(keyZoomStep, centerX, centerY)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) {
		s.camera.zoomBy(1/keyZoomStep, centerX, centerY)
	}
}

// resetView returns to the starting view: the whole middle of the world at
// natural scale, for when panning has left you somewhere unrecognizable.
func (s *Simulator) resetView() {
	s.camera.centerX = s.world.Width() / 2
	s.camera.centerY = s.world.Height() / 2
	s.camera.zoom = 1
}
