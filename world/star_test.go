package world

import (
	"math"
	"testing"
)

// fallBudget is how long TestStarStaysPutWhilePullingHard waits for an object
// dropped six hundred units out to reach the star. The fall takes about ninety
// steps at the current starMass, and the budget is deliberately several times
// that: what the test is for is that the fall happens at all and that the star
// does not budge while it does, so a bound tight enough to notice a retuned star
// would only ever be a thing to come back and fix. Time to fall goes as the
// inverse square root of the star's mass, so this covers dividing it by ten.
const fallBudget = 300

// TestStarStaysPutWhilePullingHard drops an object next to a star and checks the
// star is unmoved by it -- through the fall, the impact, and the merge. The star
// is object zero throughout: the sweep that clears absorbed objects keeps the
// order of the survivors.
func TestStarStaysPutWhilePullingHard(t *testing.T) {
	w := New(2000, 2000)
	w.SpawnStar(1000, 1000)
	w.Spawn(1600, 1000)

	eaten := false
	for step := 0; step < fallBudget; step++ {
		w.UpdateSpace()

		star := w.Objects[0]
		if !star.Fixed {
			t.Fatalf("step %d: lost track of the star: %+v", step, w.Objects)
		}
		assertClose(t, "star X", star.X, 1000)
		assertClose(t, "star Y", star.Y, 1000)
		assertClose(t, "star VelocityX", star.VelocityX, 0)
		assertClose(t, "star VelocityY", star.VelocityY, 0)

		if len(w.Objects) == 1 {
			eaten = true
			break
		}
	}

	// Falling from six hundred units out, the object should have reached the star.
	// If it has not, the pull is not strong enough to be worth calling a star.
	if !eaten {
		t.Errorf("object never reached the star in %d steps: %+v", fallBudget, w.Objects[1])
	}
}

// TestStarGravityBeatsAnOrdinaryPair checks the star is actually the thing in
// charge: an object sitting next to another object is pulled towards the star
// rather than towards its neighbour.
func TestStarGravityBeatsAnOrdinaryPair(t *testing.T) {
	w := New(4000, 4000)
	w.SpawnStar(2000, 2000)
	w.Spawn(2600, 2000)
	w.Spawn(2700, 2000)

	w.UpdateSpace()

	// The outer object has its neighbour pulling it inwards too, so the telling
	// one is the inner object: its neighbour pulls it out, the star pulls it in.
	if got := w.Objects[1].VelocityX; got >= 0 {
		t.Errorf("inner object should be pulled towards the star, got VelocityX %v", got)
	}
}

func TestStarSwallowsWhatHitsIt(t *testing.T) {
	w := New(2000, 2000)
	w.SpawnStar(1000, 1000)
	// Just inside the collision band, and travelling fast enough that a bounce
	// rather than a merge would be the outcome between two ordinary objects.
	w.SpawnMoving(1000+starRadius+defaultRadius, 1000, -40, 0)

	w.UpdateSpace() // absorbed, leaving the object at zero mass
	w.UpdateSpace() // swept up

	if len(w.Objects) != 1 {
		t.Fatalf("want 1 object left after the star ate one, got %d", len(w.Objects))
	}

	star := w.Objects[0]
	if !star.Fixed {
		t.Fatal("the star should be the survivor, not the object that hit it")
	}
	assertClose(t, "star X", star.X, 1000)
	assertClose(t, "star Y", star.Y, 1000)
	assertClose(t, "star VelocityX", star.VelocityX, 0)
	assertClose(t, "star mass", star.Mass, starMass+defaultMass)

	// Area is conserved by a merge whether the eater is anchored or not.
	wantRadius := math.Sqrt(starRadius*starRadius + defaultRadius*defaultRadius)
	assertClose(t, "star radius", star.Radius, wantRadius)
}

// mealBudget is how long TestStarKeepsGrowingAsItEats runs a ring of objects
// falling into a star. Most of them are eaten in the first couple of hundred
// steps; the rest of the budget is for the stragglers, which a bounce on the way
// in can throw into an orbit that takes a while to come down. Some may still be
// up there at the end, which the test allows for rather than waiting out.
const mealBudget = 2000

// TestStarKeepsGrowingAsItEats covers the state the single-meal tests never
// reach. A star is not what it was at launch after a while: mass climbs by every
// object it takes, and the radius climbs with it, since absorb conserves area for
// an anchored body as much as for any other.
//
// Both are worth pinning together. The radius is what a star is drawn at and what
// it captures with, and it is also what holds gravity at arm's length in
// substepCount -- so the gap between mass climbing linearly and radius climbing
// as a square root is exactly the cost curve BenchmarkStep measures.
func TestStarKeepsGrowingAsItEats(t *testing.T) {
	const meals = 60

	w := New(4000, 4000)
	w.SpawnStar(2000, 2000)
	for i := 0; i < meals; i++ {
		angle := 2 * math.Pi * float64(i) / meals
		w.Spawn(2000+300*math.Cos(angle), 2000+300*math.Sin(angle))
	}

	for step := 0; step < mealBudget && len(w.Objects) > 1; step++ {
		w.UpdateSpace()
	}

	star := w.Objects[0]
	if !star.Fixed {
		t.Fatalf("the star should be object zero, got %+v", star)
	}

	// What is left loose is whatever survived in orbit, plus whatever it merged
	// with up there. Counting the star's meals from its mass rather than assuming
	// all sixty arrived is what keeps this test about growth instead of about how
	// long a straggler takes to come down.
	eaten := (star.Mass - starMass) / defaultMass
	if eaten < meals/2 {
		t.Fatalf("want the star to have eaten most of the ring, got %v of %d: %+v",
			eaten, meals, w.Objects)
	}

	// Mass is the sum of the meals, and area is too -- so the radius, which is the
	// area read back out, is the square root of the count. That square root against
	// a mass climbing straight is the whole of the star's cost curve.
	assertClose(t, "star mass", star.Mass, starMass+eaten*defaultMass)
	assertClose(t, "star radius",
		star.Radius, math.Sqrt(starRadius*starRadius+eaten*defaultRadius*defaultRadius))

	// And the growth is real growth: this is not the star it was launched as.
	if star.Radius <= starRadius {
		t.Errorf("want a radius past the %v it was born with, got %v", float64(starRadius), star.Radius)
	}

	// Still exactly where it was put, after a ring of impacts rather than one.
	assertClose(t, "star X", star.X, 2000)
	assertClose(t, "star Y", star.Y, 2000)
	assertClose(t, "star VelocityX", star.VelocityX, 0)
	assertClose(t, "star VelocityY", star.VelocityY, 0)
}

// TestStarsIgnoreEachOther pins what several stars in one world do, which is
// nothing: no mutual gravity and no collision, however close together they are
// put. advance skips fixed objects in the outer loop, so a fixed pair is never
// examined from either side.
//
// The two here overlap outright -- well inside the band that would have either of
// them swallow an ordinary object -- and still have to come out untouched.
func TestStarsIgnoreEachOther(t *testing.T) {
	w := New(4000, 4000)
	w.SpawnStar(2000, 2000)
	w.SpawnStar(2000+starRadius, 2000)
	w.Spawn(1000, 2000) // so the step has something to do

	for step := 0; step < 100; step++ {
		w.UpdateSpace()

		if len(w.Objects) < 2 || !w.Objects[0].Fixed || !w.Objects[1].Fixed {
			t.Fatalf("step %d: one star has eaten the other: %+v", step, w.Objects)
		}
		assertClose(t, "first star X", w.Objects[0].X, 2000)
		assertClose(t, "first star mass", w.Objects[0].Mass, starMass)
		assertClose(t, "second star X", w.Objects[1].X, 2000+starRadius)
		assertClose(t, "second star mass", w.Objects[1].Mass, starMass)
	}
}

// TestStarEatingLeavesNoNaN guards the one thing a star changes about the shape
// of a collision: it absorbs from either side of the pair, so it can leave the
// object the physics loop is partway through working on at zero mass. Whatever
// pairs that object still had left to consider would then be dividing by it.
//
// The third object is what makes the case: it sits clear of both the star and the
// object being eaten, so it is still to come when the eating happens.
func TestStarEatingLeavesNoNaN(t *testing.T) {
	w := New(2000, 2000)
	w.SpawnStar(1000, 1000)
	w.SpawnMoving(1000+starRadius+defaultRadius, 1000, -40, 0) // in contact
	w.Spawn(1500, 1000)                                        // well clear

	w.UpdateSpace()

	for i, object := range w.Objects {
		if math.IsNaN(object.X) || math.IsNaN(object.Y) ||
			math.IsNaN(object.VelocityX) || math.IsNaN(object.VelocityY) ||
			math.IsNaN(object.Mass) || math.IsNaN(object.Radius) {
			t.Errorf("object %d came out of the star's meal with a NaN: %+v", i, object)
		}
	}
}

// TestStarIgnoresTheBoundary covers all three edge modes at once. A star is put
// where each of them would do something to an ordinary object -- wrap it around,
// bounce it back, or delete it -- and has to come out unchanged.
func TestStarIgnoresTheBoundary(t *testing.T) {
	for _, boundary := range []BoundaryMode{BoundaryWrap, BoundaryWall, BoundaryVoid} {
		t.Run(boundary.String(), func(t *testing.T) {
			w := New(1000, 1000)
			w.Boundary = boundary
			// Outside the world entirely: past the wrap margin, past the wall, and
			// far enough out for the void to take it.
			w.SpawnStar(1000+roomForError+50, 500)
			// Something for the star to pull on, so the step is not a no-op.
			w.Spawn(500, 500)

			w.UpdateSpace()

			if len(w.Objects) == 0 || !w.Objects[0].Fixed {
				t.Fatalf("the star should still be there, got %+v", w.Objects)
			}
			assertClose(t, "star X", w.Objects[0].X, 1000+roomForError+50)
			assertClose(t, "star Y", w.Objects[0].Y, 500)
		})
	}
}

// TestStarHoldsAnObjectInOrbit is the point of a fixed central mass: launched
// sideways at the right speed, an object should circle it rather than fall in or
// fly away. The speed is the textbook one for a circular orbit, which the
// simulation's gravity obeys with its scale factor folded in.
func TestStarHoldsAnObjectInOrbit(t *testing.T) {
	const radius = 600

	w := New(4000, 4000)
	w.Boundary = BoundaryVoid // so a failed orbit is unmistakable: the object goes
	w.SpawnStar(2000, 2000)
	w.SpawnMoving(2000+radius, 2000, 0, math.Sqrt(gravityScale*G*starMass/radius))

	for step := 0; step < 2000; step++ {
		w.UpdateSpace()

		if len(w.Objects) != 2 {
			t.Fatalf("step %d: object left orbit -- eaten or gone", step)
		}

		// A circular orbit stays at its radius. Two percent is more than fifteen
		// times the drift actually observed here -- an eighth of a percent, and not
		// growing, since stepping the velocity before the position makes this
		// integrator symplectic and so stable in energy. A looser band than this
		// would sit there quietly while an orbit decayed by a third, which is the
		// whole class of regression the test is for.
		distance := distanceBetween(w.Objects[0], w.Objects[1])
		if distance < radius*0.98 || distance > radius*1.02 {
			t.Fatalf("step %d: orbit radius drifted to %v, want near %v",
				step, distance, float64(radius))
		}
	}
}
