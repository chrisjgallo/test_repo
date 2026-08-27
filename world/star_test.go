package world

import (
	"math"
	"testing"
)

// TestStarStaysPutWhilePullingHard drops an object next to a star and checks the
// star is unmoved by it -- through the fall, the impact, and the merge. The star
// is object zero throughout: the sweep that clears absorbed objects keeps the
// order of the survivors.
func TestStarStaysPutWhilePullingHard(t *testing.T) {
	w := New(2000, 2000)
	w.SpawnStar(1000, 1000)
	w.Spawn(1600, 1000)

	eaten := false
	for step := 0; step < 100; step++ {
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

	// Falling from six hundred units out, the object should have reached the star
	// well inside a hundred steps. If it has not, the pull is not strong enough
	// to be worth calling a star.
	if !eaten {
		t.Errorf("object never reached the star in 100 steps: %+v", w.Objects[1])
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

		// A circular orbit stays at its radius. The tolerance is loose enough for
		// the drift a fixed-step integrator accumulates over a few laps, and tight
		// enough that falling in or spiralling out would fail it.
		distance := distanceBetween(w.Objects[0], w.Objects[1])
		if distance < radius*0.5 || distance > radius*1.5 {
			t.Fatalf("step %d: orbit radius drifted to %v, want near %v",
				step, distance, float64(radius))
		}
	}
}
