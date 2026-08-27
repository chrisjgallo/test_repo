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

	// collisionSlop widens the collision check so a pair is caught slightly
	// before their edges actually touch. It has to be this generous because a
	// step is a whole tick long: objects near contact are travelling more than
	// ten units a step, so a narrower band would let them jump clean through
	// each other between one step and the next.
	collisionSlop = 10

	// restitution is the fraction of closing speed a bounce hands back. The
	// rest goes where it goes in a real impact -- heat, sound, deformation --
	// which is what eventually settles a bouncing pair down into one object.
	restitution = 0.9

	// mergeVelocityFraction is the share of escape velocity a bounce has to
	// return for the pair to stay two objects. See mergeSpeed.
	mergeVelocityFraction = 0.2

	// gravityScale is how much stronger this simulation's gravity is than the
	// plain inverse-square law. It lives inside forceOfGravity; escapeSpeed has
	// to use the same number to agree with it.
	gravityScale = 10

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

// collide resolves two objects that have run into each other. A solid hit sends
// them bouncing apart; a feeble one merges them into a single object, the way
// every collision used to.
func collide(object, other *SpaceObject, distance float64) {
	// Two objects in exactly the same spot have no direction to bounce along.
	if distance == 0 {
		object.absorb(other)
		return
	}

	// The collision normal points from object toward other, and approachSpeed
	// is how fast the gap along it is closing.
	normalX := (other.X - object.X) / distance
	normalY := (other.Y - object.Y) / distance
	approachSpeed := (object.VelocityX-other.VelocityX)*normalX +
		(object.VelocityY-other.VelocityY)*normalY

	// What matters is how fast the pair will be leaving once this step is over:
	// the rebound if they are still closing, and whatever they already have if
	// they are on their way out of an earlier bounce.
	separationSpeed := -approachSpeed
	if approachSpeed > 0 {
		separationSpeed = restitution * approachSpeed
	}

	if separationSpeed < mergeSpeed(object, other, distance) {
		object.absorb(other)
		return
	}

	// Already separating fast enough to get away: leave them to it. This is
	// also what stops a bounce from being applied twice, since every pair comes
	// up once from each object's point of view.
	if approachSpeed <= 0 {
		return
	}

	bounce(object, other, normalX, normalY, approachSpeed)
}

// bounce drives two objects apart along the collision normal. Momentum is
// conserved exactly; energy is not, since restitution keeps back the share of
// the impact that a real collision would shed as heat and sound.
func bounce(object, other *SpaceObject, normalX, normalY, approachSpeed float64) {
	impulse := (1 + restitution) * approachSpeed / (1/object.Mass + 1/other.Mass)

	object.VelocityX -= impulse * normalX / object.Mass
	object.VelocityY -= impulse * normalY / object.Mass
	other.VelocityX += impulse * normalX / other.Mass
	other.VelocityY += impulse * normalY / other.Mass
}

// mergeSpeed is the separation speed a collision has to leave behind for the
// pair to stay two objects. Below it they are too slow to get clear of each
// other: gravity hauls them straight back for another, weaker bounce, and then
// another, so they are merged now rather than after a run of ever-smaller hops
// to the same end.
//
// The bar is a fraction of escape velocity rather than escape velocity itself,
// because at the full value nothing would ever bounce at all. Two objects that
// fall together from rest arrive at contact travelling at just under the speed
// needed to escape from it -- that is the same energy sum read backwards -- and
// a bounce only ever takes energy away.
func mergeSpeed(object, other *SpaceObject, distance float64) float64 {
	return mergeVelocityFraction * escapeSpeed(object.Mass+other.Mass, distance)
}

// escapeSpeed is how fast two objects have to be pulling apart at the given
// separation to be free of each other for good.
func escapeSpeed(totalMass, distance float64) float64 {
	return math.Sqrt(2 * gravityScale * G * totalMass / distance)
}

// BoundaryMode is what becomes of an object that reaches the edge of the world.
type BoundaryMode int

const (
	// BoundaryWrap sends an object out one side and straight back in the other,
	// so nothing is ever lost and the world has no real edge.
	BoundaryWrap BoundaryMode = iota

	// BoundaryWall bounces an object back off the edge, the same way it would
	// bounce off another object.
	BoundaryWall

	// BoundaryVoid lets an object go. Anything that leaves is gone for good.
	BoundaryVoid

	// boundaryModeCount is how many modes there are, so cycling can wrap around.
	boundaryModeCount
)

func (m BoundaryMode) String() string {
	switch m {
	case BoundaryWrap:
		return "Wrap"
	case BoundaryWall:
		return "Wall"
	case BoundaryVoid:
		return "Void"
	}
	return "Unknown"
}

// World is the simulated space and everything in it.
type World struct {
	Objects  []SpaceObject
	Paused   bool
	Boundary BoundaryMode

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

			// Collision. Whether the pair bounces or merges, gravity between
			// the two is left off for this step: at touching distance the
			// inverse-square force is enormous, and the collision has already
			// decided what happens to them.
			if distance < object.Radius+other.Radius+collisionSlop {
				collide(object, other, distance)
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

// forceOfGravity is the usual inverse-square law turned up by gravityScale, so
// that objects fall together on a timescale worth watching.
func forceOfGravity(mass1, mass2, distance float64) float64 {
	return (G * mass1 * mass2) / (distance * (distance / gravityScale))
}
