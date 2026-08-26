// Package world holds the simulation state -- the space objects and the physics
// that moves them around. It knows nothing about how any of it gets drawn.
package world

import "math"

// G is the gravity constant, changed by orders of magnitude for the sake of the
// simulation.
const G = .667430

const (
	// roomForError is how far past an edge an object drifts before it wraps
	// around to the other side: go far enough and you'll come back.
	roomForError = 8

	// collisionSlop widens the collision check so two objects merge slightly
	// before their edges actually touch.
	collisionSlop = 10

	// defaultMass is the mass every newly spawned object starts with.
	defaultMass = 80

	// defaultRadius is the radius every newly spawned object starts with.
	defaultRadius = 1.0
)

// SpaceObject is a single body in the simulation. A mass of zero marks an object
// that has been absorbed by another and is waiting to be swept up.
type SpaceObject struct {
	X, Y      float64
	Radius    float64
	VelocityX float64
	VelocityY float64
	Mass      float64
}

func (o SpaceObject) surfaceArea() float64 {
	return math.Pi * o.Radius * o.Radius
}

// absorb merges other into o, conserving momentum and area, and leaves other
// with zero mass so the next step removes it.
func (o *SpaceObject) absorb(other *SpaceObject) {
	totalMass := o.Mass + other.Mass

	o.VelocityX = (o.VelocityX*o.Mass + other.VelocityX*other.Mass) / totalMass
	o.VelocityY = (o.VelocityY*o.Mass + other.VelocityY*other.Mass) / totalMass
	o.X = (o.X*o.Mass + other.X*other.Mass) / totalMass
	o.Y = (o.Y*o.Mass + other.Y*other.Mass) / totalMass
	o.Radius = math.Sqrt((o.surfaceArea() + other.surfaceArea()) / math.Pi)
	o.Mass = totalMass

	other.Mass = 0
}

// World is the simulated space and everything in it.
type World struct {
	Objects []SpaceObject
	Paused  bool

	screenWidth  float64
	screenHeight float64
}

// New returns an empty world sized to the screen it will be drawn on.
func New(screenWidth, screenHeight int) *World {
	return &World{
		screenWidth:  float64(screenWidth),
		screenHeight: float64(screenHeight),
	}
}

// Spawn adds a new object at the given position.
func (w *World) Spawn(x, y float64) {
	w.Objects = append(w.Objects, SpaceObject{
		X:      x,
		Y:      y,
		Radius: defaultRadius,
		Mass:   defaultMass,
	})
}

// TogglePause freezes or resumes the simulation. Objects can still be added
// while paused; they just don't move until time starts again.
func (w *World) TogglePause() {
	w.Paused = !w.Paused
}

// UpdateSpace advances the simulation by one step.
func (w *World) UpdateSpace() {
	w.removeDestroyed()

	// Don't update anything after this point if the world is paused.
	if w.Paused {
		return
	}

	w.handleObjectVelocityAndGravity()
}

func (w *World) removeDestroyed() {
	remaining := w.Objects[:0]
	for _, object := range w.Objects {
		if object.Mass != 0 {
			remaining = append(remaining, object)
		}
	}
	w.Objects = remaining
}

func (w *World) handleObjectVelocityAndGravity() {
	for i := range w.Objects {
		w.wrapAroundEdges(&w.Objects[i])
	}

	for i := range w.Objects {
		object := &w.Objects[i]
		if object.Mass == 0 {
			continue // absorbed earlier in this same step
		}

		for j := range w.Objects {
			if i == j {
				continue
			}

			other := &w.Objects[j]
			if other.Mass == 0 {
				continue
			}

			distance := math.Hypot(other.X-object.X, other.Y-object.Y)

			// Collision. Merging leaves the pair at one position, so there is
			// no gravity left to apply between them this step.
			if distance < object.Radius+other.Radius+collisionSlop {
				object.absorb(other)
				continue
			}

			force := forceOfGravity(object.Mass, other.Mass, distance)
			xAcceleration := ((other.X - object.X) / distance) * force / object.Mass
			yAcceleration := ((other.Y - object.Y) / distance) * force / object.Mass

			object.VelocityX += xAcceleration
			object.VelocityY += yAcceleration
		}

		object.X += object.VelocityX
		object.Y += object.VelocityY
	}
}

func (w *World) wrapAroundEdges(object *SpaceObject) {
	if object.X < -roomForError {
		object.X = w.screenWidth - roomForError
	}
	if object.X > w.screenWidth+roomForError {
		object.X = roomForError
	}
	if object.Y < -roomForError {
		object.Y = w.screenHeight - roomForError
	}
	if object.Y > w.screenHeight+roomForError {
		object.Y = roomForError
	}
}

// forceOfGravity scales the usual inverse-square law to make gravity have
// slightly more effect far away and slightly less close up than normal.
func forceOfGravity(mass1, mass2, distance float64) float64 {
	return (G * mass1 * mass2) / (distance * (distance / 10))
}
