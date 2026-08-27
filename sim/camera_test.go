package sim

import (
	"math"
	"testing"
)

func TestScreenAndWorldRoundTrip(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)
	c.zoom = 2.5

	worldX, worldY := c.screenToWorld(400, 700)
	screenX, screenY := c.worldToScreen(worldX, worldY)

	if math.Abs(screenX-400) > 1e-9 || math.Abs(screenY-700) > 1e-9 {
		t.Errorf("want (400, 700) back, got (%v, %v)", screenX, screenY)
	}
}

func TestCameraCenterSitsAtScreenCenter(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)

	x, y := c.worldToScreen(2550, 1500)

	if x != 850 || y != 500 {
		t.Errorf("want the center of the world at (850, 500), got (%v, %v)", x, y)
	}
}

func TestZoomKeepsAnchorPinned(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)

	// Whatever sits under the cursor before zooming should still be there after.
	const anchorX, anchorY = 300, 200
	beforeX, beforeY := c.screenToWorld(anchorX, anchorY)

	c.zoomBy(2, anchorX, anchorY)

	afterX, afterY := c.screenToWorld(anchorX, anchorY)
	if math.Abs(afterX-beforeX) > 1e-9 || math.Abs(afterY-beforeY) > 1e-9 {
		t.Errorf("anchor drifted from (%v, %v) to (%v, %v)", beforeX, beforeY, afterX, afterY)
	}
	if c.zoom != 2 {
		t.Errorf("want zoom 2, got %v", c.zoom)
	}
}

func TestZoomStaysWithinLimits(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)

	c.zoomBy(1000, 850, 500)
	if c.zoom != maxZoom {
		t.Errorf("want zoom capped at %v, got %v", maxZoom, c.zoom)
	}

	c.zoomBy(0.00001, 850, 500)
	if c.zoom != minZoom {
		t.Errorf("want zoom floored at %v, got %v", minZoom, c.zoom)
	}
}

func TestPanTracksTheCursorAtAnyZoom(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)
	c.zoom = 4

	// Dragging 100 screen pixels should move the view 100 screen pixels' worth
	// of world, which at 4x zoom is only 25 world units.
	c.panByScreen(100, 0)

	if want := 2550 - 25.0; c.centerX != want {
		t.Errorf("want center x %v, got %v", want, c.centerX)
	}
}

func TestCameraStaysInsideTheWorld(t *testing.T) {
	c := newCamera(1700, 1000, 2550, 1500)

	c.panByScreen(100000, 100000)
	c.clampToWorld(5100, 3000)

	if c.centerX != 0 || c.centerY != 0 {
		t.Errorf("want the camera clamped to (0, 0), got (%v, %v)", c.centerX, c.centerY)
	}

	c.panByScreen(-100000, -100000)
	c.clampToWorld(5100, 3000)

	if c.centerX != 5100 || c.centerY != 3000 {
		t.Errorf("want the camera clamped to (5100, 3000), got (%v, %v)", c.centerX, c.centerY)
	}
}
