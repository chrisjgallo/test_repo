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
	// before their edges actually touch, rather than leaving contact to depend
	// on two floating point positions lining up. It used to have to be generous
	// enough to catch a fast pair mid-flight as well; substepping took that job
	// over, and the value stayed as it was because the merge and bounce
	// behavior is tuned around this much room.
	collisionSlop = 10

	// substepSafeFraction is the share of the narrowest collision band an
	// object may cross in a single substep. A pair has to cover twice their
	// band between one substep and the next to get through it unseen, so
	// holding each of them to half a band leaves a factor of two in hand.
	substepSafeFraction = 0.5

	// maxSubsteps is where the slicing stops. A substep costs a full pass over
	// every pair, so without a ceiling a fast enough world could price itself
	// out of its own frame. It is set high enough that gravity alone does not
	// reach it: the worlds that do are ones where a few survivors are flinging
	// each other about at hundreds of units a step, and by then there are few
	// enough of them left that the passes are cheap. Past this point collisions
	// go back to being missable, which is the deliberate trade -- a frame that
	// arrives beats a frame that is right.
	maxSubsteps = 64

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

	// starMass is the mass of a fixed star: a little over sixty spawned objects'
	// worth, so the star rather than the crowd decides where everything goes.
	//
	// It is tuned against how fast a drag can throw something (velocityPerPixel,
	// in the sim package). Launched sideways from half a screen out, an object
	// needs about a fifth-screen drag to hold a circular orbit and about a
	// quarter-screen one to escape the star altogether, which puts both inside a
	// comfortable mouse movement and leaves the interesting range -- ellipses,
	// slow inward spirals -- in between.
	starMass = 5000

	// starRadius is how large a fixed star is, to draw and to collide with.
	// Merging would have to swallow nine hundred objects to reach it, which is
	// well past the far end of the color ramp -- so a star is drawn in a color of
	// its own rather than as the biggest blob on screen.
	starRadius = 30
)

// SpaceObject is a single body in the simulation. A mass of zero marks an object
// that is done for -- absorbed by another, or drifted out of the world -- and is
// waiting to be swept up.
type SpaceObject struct {
	X, Y      float64
	Radius    float64
	VelocityX float64
	VelocityY float64
	Mass      float64

	// Fixed marks an object that is anchored in place: it pulls on everything
	// around it, and nothing -- gravity, a collision, or the edge of the world --
	// moves it. Stars are the only thing that sets it.
	Fixed bool
}

func (o SpaceObject) surfaceArea() float64 {
	return math.Pi * o.Radius * o.Radius
}

// absorb merges other into o, conserving momentum and area, and leaves other
// with zero mass so the next step removes it.
//
// A fixed o keeps the position and velocity it had: it is anchored, so the
// momentum handed to it has nowhere to go. Mass and area still accumulate, which
// is how a star grows -- and pulls harder -- on what it eats.
func (o *SpaceObject) absorb(other *SpaceObject) {
	totalMass := o.Mass + other.Mass

	if !o.Fixed {
		o.VelocityX = (o.VelocityX*o.Mass + other.VelocityX*other.Mass) / totalMass
		o.VelocityY = (o.VelocityY*o.Mass + other.VelocityY*other.Mass) / totalMass
		o.X = (o.X*o.Mass + other.X*other.Mass) / totalMass
		o.Y = (o.Y*o.Mass + other.Y*other.Mass) / totalMass
	}

	o.Radius = math.Sqrt((o.surfaceArea() + other.surfaceArea()) / math.Pi)
	o.Mass = totalMass

	other.Mass = 0
}

// collide resolves two objects that have run into each other. A solid hit sends
// them bouncing apart; a feeble one merges them into a single object, the way
// every collision used to.
func collide(object, other *SpaceObject, distance float64) {
	// Nothing here survives a zero mass: it divides through the bounce impulse
	// and the momentum averages of a merge alike, and the NaN it leaves behind
	// spreads to every object either one touches from then on. Callers skip
	// objects already marked for removal, so this only holds the line if one
	// ever stops doing so.
	if object.Mass <= 0 || other.Mass <= 0 {
		return
	}

	// A fixed object gives no ground: it swallows whatever reaches it and stays
	// where it is. Nothing below applies to it. There is no bounce to divide up,
	// since one side of the pair cannot take any of it -- the impulse would come
	// back off something immovable entirely into whatever hit it -- and so no
	// bounce for the merge test to weigh either. All that is left to decide is
	// which one absorbs, and being fixed decides it.
	if object.Fixed || other.Fixed {
		if other.Fixed {
			object, other = other, object
		}
		object.absorb(other)
		return
	}

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

	width  float64
	height float64
}

// New returns an empty world of the given size.
func New(width, height int) *World {
	return &World{
		width:  float64(width),
		height: float64(height),
	}
}

// Width is how far the world stretches horizontally before it wraps.
func (w *World) Width() float64 { return w.width }

// Height is how far the world stretches vertically before it wraps.
func (w *World) Height() float64 { return w.height }

// Spawn adds a new object at rest at the given position.
func (w *World) Spawn(x, y float64) {
	w.SpawnMoving(x, y, 0, 0)
}

// SpawnMoving adds a new object at the given position, already travelling at
// the given velocity.
func (w *World) SpawnMoving(x, y, velocityX, velocityY float64) {
	w.Objects = append(w.Objects, SpaceObject{
		X:         x,
		Y:         y,
		Radius:    defaultRadius,
		VelocityX: velocityX,
		VelocityY: velocityY,
		Mass:      defaultMass,
	})
}

// SpawnStar adds a fixed star at the given position: far heavier than anything
// spawned by hand, anchored where it is put, and fatal to whatever runs into it.
// Nothing stops a world having several, though one in the middle is the point of
// them -- everything else then has something to fall around.
func (w *World) SpawnStar(x, y float64) {
	w.Objects = append(w.Objects, SpaceObject{
		X:      x,
		Y:      y,
		Radius: starRadius,
		Mass:   starMass,
		Fixed:  true,
	})
}

// CycleBoundary moves on to the next way of treating the edge of the world.
func (w *World) CycleBoundary() {
	w.Boundary = (w.Boundary + 1) % boundaryModeCount
}

// TotalMass is the mass of everything in the world added together. Merging
// conserves it, so it only ever changes when something new is spawned.
func (w *World) TotalMass() float64 {
	var total float64
	for _, object := range w.Objects {
		total += object.Mass
	}
	return total
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
	substeps := w.substepCount()
	for substep := 0; substep < substeps; substep++ {
		w.advance(1 / float64(substeps))
	}
}

// substepCount is how many slices a step has to be taken in for the collision
// check to see everything that happens during it.
//
// Taken in one go, a step is long enough for a heavy pair to pass clean through
// each other: clear on one side at the end of one step, clear on the far side at
// the end of the next, never once landing inside the band meant to catch them.
// It takes no unusual speed to manage, either -- gravity at close range alone
// can move an object further in a single step than the band is wide.
//
// So the worst case is worked out in advance: the fastest object there is, plus
// the hardest shove the heaviest one can give it over a step, measured against
// the narrowest band there is to catch anything.
func (w *World) substepCount() int {
	smallestRadius, fastestSpeed := math.Inf(1), 0.0
	heaviestMass, heaviestRadius := 0.0, 0.0
	live := 0

	for i := range w.Objects {
		object := &w.Objects[i]
		if object.Mass == 0 {
			continue
		}

		live++
		smallestRadius = math.Min(smallestRadius, object.Radius)
		fastestSpeed = math.Max(fastestSpeed, math.Hypot(object.VelocityX, object.VelocityY))
		if object.Mass > heaviestMass {
			heaviestMass, heaviestRadius = object.Mass, object.Radius
		}
	}

	// Nothing to collide with, so nothing to outrun.
	if live < 2 {
		return 1
	}

	// The narrowest band any pair can have, and so the least room a collision
	// has to be noticed in: two of the smallest object there is.
	narrowestBand := 2*smallestRadius + collisionSlop

	// The hardest pull anything can feel is the heaviest object's, from as close
	// in as gravity is ever applied to it -- the edge of its own band with the
	// smallest object going, since inside that the collision takes over.
	pullRange := smallestRadius + heaviestRadius + collisionSlop
	strongestPull := G * gravityScale * heaviestMass / (pullRange * pullRange)

	reach := fastestSpeed + strongestPull
	return min(max(int(math.Ceil(reach/(substepSafeFraction*narrowestBand))), 1), maxSubsteps)
}

// advance moves the world on by one slice of a step. Velocities and positions
// are scaled by the slice; the collisions found along the way are not, since an
// impact changes a velocity outright however long the step it lands in.
func (w *World) advance(slice float64) {
	for i := range w.Objects {
		w.applyBoundary(&w.Objects[i])
	}

	for i := range w.Objects {
		object := &w.Objects[i]
		if object.Mass == 0 {
			continue // absorbed earlier in this same slice
		}

		// A fixed object pulls on everything and is moved by nothing, so there is
		// no acceleration to add up and no position to advance from its point of
		// view. Its collisions are still caught, from the other half of the pair:
		// every pair comes up once from each side.
		if object.Fixed {
			continue
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
			// the two is left off for this slice: at touching distance the
			// inverse-square force is enormous, and the collision has already
			// decided what happens to them.
			if distance < object.Radius+other.Radius+collisionSlop {
				collide(object, other, distance)

				// Ordinarily the object being worked on is the one left standing
				// after a collision, whichever way it goes: a merge is always the
				// outer object absorbing the inner one. A fixed object is the
				// exception -- it does the absorbing whichever side of the pair it
				// turns up on -- so this is the one collision that can leave the
				// object under consideration at zero mass, with nothing further to
				// work out and a zero mass to divide the next pair's force by.
				if object.Mass == 0 {
					break
				}
				continue
			}

			force := forceOfGravity(object.Mass, other.Mass, distance)
			xAcceleration := ((other.X - object.X) / distance) * force / object.Mass
			yAcceleration := ((other.Y - object.Y) / distance) * force / object.Mass

			object.VelocityX += xAcceleration * slice
			object.VelocityY += yAcceleration * slice
		}

		object.X += object.VelocityX * slice
		object.Y += object.VelocityY * slice
	}
}

// applyBoundary does whatever the current mode says to an object that has
// reached the edge of the world.
func (w *World) applyBoundary(object *SpaceObject) {
	// A fixed object is exempt from all three modes: every one of them either
	// moves an object or takes it away, and a fixed object stays where it was put.
	if object.Fixed {
		return
	}

	switch w.Boundary {
	case BoundaryWall:
		w.bounceOffEdges(object)
	case BoundaryVoid:
		w.removeIfOutside(object)
	default:
		w.wrapAroundEdges(object)
	}
}

// bounceOffEdges turns the edge of the world into a solid wall. An object keeps
// the same share of its speed off a wall that it would keep off another object,
// so the edge is not a free source of energy.
func (w *World) bounceOffEdges(object *SpaceObject) {
	// Each edge nudges the object back inside first, so nothing can end up
	// buried in a wall, and only reverses the velocity that is driving it in.
	// An object already on its way out is left alone, or it would stick.
	if object.X-object.Radius < 0 {
		object.X = object.Radius
		if object.VelocityX < 0 {
			object.VelocityX = -object.VelocityX * restitution
		}
	}
	if object.X+object.Radius > w.width {
		object.X = w.width - object.Radius
		if object.VelocityX > 0 {
			object.VelocityX = -object.VelocityX * restitution
		}
	}
	if object.Y-object.Radius < 0 {
		object.Y = object.Radius
		if object.VelocityY < 0 {
			object.VelocityY = -object.VelocityY * restitution
		}
	}
	if object.Y+object.Radius > w.height {
		object.Y = w.height - object.Radius
		if object.VelocityY > 0 {
			object.VelocityY = -object.VelocityY * restitution
		}
	}
}

// removeIfOutside marks an object that has left the world entirely, leaving the
// usual sweep to clear it out on the next step.
func (w *World) removeIfOutside(object *SpaceObject) {
	if object.X+object.Radius < 0 || object.X-object.Radius > w.width ||
		object.Y+object.Radius < 0 || object.Y-object.Radius > w.height {
		object.Mass = 0
	}
}

func (w *World) wrapAroundEdges(object *SpaceObject) {
	if object.X < -roomForError {
		object.X = w.width - roomForError
	}
	if object.X > w.width+roomForError {
		object.X = roomForError
	}
	if object.Y < -roomForError {
		object.Y = w.height - roomForError
	}
	if object.Y > w.height+roomForError {
		object.Y = roomForError
	}
}

// forceOfGravity is the usual inverse-square law turned up by gravityScale, so
// that objects fall together on a timescale worth watching.
func forceOfGravity(mass1, mass2, distance float64) float64 {
	return (G * mass1 * mass2) / (distance * (distance / gravityScale))
}
