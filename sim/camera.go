package sim

const (
	// minZoom and maxZoom bound how far the view can pull back or push in.
	// Below minZoom objects are specks; above maxZoom there is nothing left
	// on screen to look at.
	minZoom = 0.15
	maxZoom = 8

	// zoomStep is the zoom multiplier applied per notch of the scroll wheel.
	zoomStep = 1.1
)

// camera is the window onto the world: which world point sits at the middle of
// the screen, and how many screen pixels a world unit is worth.
type camera struct {
	centerX float64
	centerY float64
	zoom    float64

	screenWidth  float64
	screenHeight float64
}

func newCamera(screenWidth, screenHeight int, centerX, centerY float64) *camera {
	return &camera{
		centerX:      centerX,
		centerY:      centerY,
		zoom:         1,
		screenWidth:  float64(screenWidth),
		screenHeight: float64(screenHeight),
	}
}

// worldToScreen converts a point in the world to where it lands on screen.
func (c *camera) worldToScreen(x, y float64) (float64, float64) {
	return (x-c.centerX)*c.zoom + c.screenWidth/2,
		(y-c.centerY)*c.zoom + c.screenHeight/2
}

// screenToWorld converts a point on screen back to the world point under it.
func (c *camera) screenToWorld(x, y float64) (float64, float64) {
	return (x-c.screenWidth/2)/c.zoom + c.centerX,
		(y-c.screenHeight/2)/c.zoom + c.centerY
}

// panByScreen slides the view by a distance measured in screen pixels, so a
// drag keeps up with the cursor no matter how far zoomed in or out it is.
func (c *camera) panByScreen(dx, dy float64) {
	c.centerX -= dx / c.zoom
	c.centerY -= dy / c.zoom
}

// zoomBy scales the view by factor, keeping whatever world point is under
// (anchorX, anchorY) pinned to that same spot on screen. Anchoring to the
// cursor means the thing being looked at stays put instead of sliding away.
func (c *camera) zoomBy(factor, anchorX, anchorY float64) {
	beforeX, beforeY := c.screenToWorld(anchorX, anchorY)

	c.zoom = clamp(c.zoom*factor, minZoom, maxZoom)

	afterX, afterY := c.screenToWorld(anchorX, anchorY)
	c.centerX += beforeX - afterX
	c.centerY += beforeY - afterY
}

// clampToWorld keeps the view centered somewhere inside the world, so panning
// can never strand the camera in empty space with nothing to navigate back by.
func (c *camera) clampToWorld(worldWidth, worldHeight float64) {
	c.centerX = clamp(c.centerX, 0, worldWidth)
	c.centerY = clamp(c.centerY, 0, worldHeight)
}

func clamp(value, low, high float64) float64 {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
